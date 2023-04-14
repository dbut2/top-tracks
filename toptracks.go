package top_tracks

import (
	"context"
	"fmt"
	"github.com/shkh/lastfm-go/lastfm"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"net/http"
	"os"
)

func UpdatePlaylistHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Set up Last.fm client
	apiKey := os.Getenv("LASTFM_API_KEY")
	apiSecret := os.Getenv("LASTFM_API_SECRET")
	lastFmUsername := os.Getenv("LASTFM_USERNAME")
	lastFmApi := lastfm.New(apiKey, apiSecret)

	// Fetch the user's top n tracks from Last.fm
	n := 200
	lastFmTopTracks, err := fetchLastFmTopTracks(lastFmApi, lastFmUsername, n)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching top tracks from Last.fm: %v", err), http.StatusInternalServerError)
		return
	}

	// Set up Spotify client
	token, err := getSpotifyToken(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting Spotify access token: %v", err), http.StatusInternalServerError)
		return
	}

	spotifyClient := spotify.Authenticator{}.NewClient(token)
	spotifyClient.AutoRetry = true // Enable automatic retrying for rate limit errors

	// Update Spotify playlist
	playlistID := spotify.ID(os.Getenv("SPOTIFY_PLAYLIST_ID"))
	err = updateSpotifyPlaylistWithLastFmTracks(spotifyClient, playlistID, lastFmTopTracks)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error updating Spotify playlist: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Playlist updated successfully")
}

func getSpotifyToken(ctx context.Context) (*oauth2.Token, error) {
	// Retrieve the client ID and secret from the environment variables
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	// Retrieve the refresh token from the environment variable
	refreshToken := os.Getenv("SPOTIFY_REFRESH_TOKEN")

	// Set up the OAuth2 config
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: "https://accounts.spotify.com/api/token",
		},
	}

	// Use the OAuth2 config and refresh token to obtain a new access token
	token, err := config.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
	}).Token()
	if err != nil {
		return nil, fmt.Errorf("error getting Spotify access token: %v", err)
	}

	// Return the new access token
	return token, nil
}

func lastFmTracksToSpotifyTracks(spotifyClient spotify.Client, lastFmTopTracks *lastfm.UserGetTopTracks) ([]spotify.FullTrack, error) {
	var spotifyTracks []spotify.FullTrack

	for _, lastFmTrack := range lastFmTopTracks.Tracks {
		spotifyTrack, err := searchTrackOnSpotify(spotifyClient, lastFmTrack.Name, lastFmTrack.Artist.Name)
		if err != nil {
			fmt.Printf("Error searching for track '%s' by '%s' on Spotify: %v\n", lastFmTrack.Name, lastFmTrack.Artist.Name, err)
			continue
		}
		if spotifyTrack != nil {
			spotifyTracks = append(spotifyTracks, *spotifyTrack)
		}
	}

	return spotifyTracks, nil
}

func searchTrackOnSpotify(client spotify.Client, trackName, artistName string) (*spotify.FullTrack, error) {
	query := fmt.Sprintf("track:%s artist:%s", trackName, artistName)
	results, err := client.Search(query, spotify.SearchTypeTrack)
	if err != nil {
		return nil, err
	}

	if len(results.Tracks.Tracks) > 0 {
		return &results.Tracks.Tracks[0], nil
	}

	fmt.Printf("Track '%s' by '%s' not found on Spotify\n", trackName, artistName)
	return nil, nil
}

func fetchLastFmTopTracks(api *lastfm.Api, username string, n int) (*lastfm.UserGetTopTracks, error) {
	params := lastfm.P{
		"user":  username,
		"limit": n,
	}
	topTracks, err := api.User.GetTopTracks(params)
	if err != nil {
		return nil, err
	}
	return &topTracks, nil
}

func updateSpotifyPlaylistWithLastFmTracks(spotifyClient spotify.Client, playlistID spotify.ID, lastFmTopTracks *lastfm.UserGetTopTracks) error {
	// Convert Last.fm tracks to Spotify tracks
	spotifyTracks, err := lastFmTracksToSpotifyTracks(spotifyClient, lastFmTopTracks)
	if err != nil {
		return err
	}

	// Collect Spotify track IDs
	trackIDs := make([]spotify.ID, len(spotifyTracks))
	for i, track := range spotifyTracks {
		trackIDs[i] = track.ID
	}

	// Clear the playlist
	err = spotifyClient.ReplacePlaylistTracks(playlistID)
	if err != nil {
		return fmt.Errorf("failed to clear playlist: %v", err)
	}

	// Add tracks in chunks of 100
	chunkSize := 100
	for i := 0; i < len(trackIDs); i += chunkSize {
		end := i + chunkSize
		if end > len(trackIDs) {
			end = len(trackIDs)
		}

		_, err = spotifyClient.AddTracksToPlaylist(playlistID, trackIDs[i:end]...)
		if err != nil {
			return fmt.Errorf("failed to add tracks to playlist: %v", err)
		}
	}

	return nil
}
