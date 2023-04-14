# Last.fm Top Tracks to Spotify Playlist

This Google Cloud Function syncs your top tracks from Last.fm to a Spotify playlist. The function fetches your top tracks from Last.fm and updates the specified Spotify playlist with these tracks.

## Prerequisites

- A Google Cloud Platform account
- A Last.fm account with API key and secret
- A Spotify account with API key, secret, and refresh token

## Setup

1. Clone this repository to your local machine.

2. Install the [Google Cloud SDK](https://cloud.google.com/sdk/docs/install) and authenticate with your Google Cloud account.

3. Create a new Google Cloud project or use an existing one.

```bash
gcloud projects create PROJECT_ID
gcloud config set project PROJECT_ID
```

4. Enable the Cloud Functions and Cloud Build APIs.

```bash
gcloud services enable cloudfunctions.googleapis.com
gcloud services enable cloudbuild.googleapis.com
```

5. Set up the required environment variables in the `env.yaml` file:

```yaml
LASTFM_API_KEY: "your_lastfm_api_key"
LASTFM_API_SECRET: "your_lastfm_api_secret"
LASTFM_USERNAME: "your_lastfm_username"
SPOTIFY_CLIENT_ID: "your_spotify_client_id"
SPOTIFY_CLIENT_SECRET: "your_spotify_client_secret"
SPOTIFY_REFRESH_TOKEN: "your_spotify_refresh_token"
SPOTIFY_PLAYLIST_ID: "your_spotify_playlist_id"
```

You can obtain the Last.fm API key and secret from the [Last.fm API account settings](https://www.last.fm/api/account/create).

For the Spotify API key and secret, you need to create a Spotify Developer account and an app in the Spotify Developer Dashboard. Follow the Authorization Code Flow to obtain the refresh token.

To fetch the refresh token, you can use the provided spotify_token_fetcher.go script in this repository. First, update the clientID, clientSecret, and redirectURI variables in the script with your own values. Then, run the script using the following command:

```bash
go run spotify_token_fetcher/spotify_token_fetcher.go
```

This will output a JSON string containing the Spotify access token and refresh token. Copy the refresh_token value from the JSON string and set it as the value of the SPOTIFY_REFRESH_TOKEN environment variable in your Cloud Function environment.

Note that the Spotify access token is only valid for a short period of time and will expire. To avoid having to fetch a new token every time the Cloud Function runs, you should use the refresh token to obtain a new access token as needed.
6. Deploy the Google Cloud Function.

```bash
gcloud functions deploy UpdatePlaylistHandler \
--runtime go120 \
--trigger-http \
--allow-unauthenticated \
--env-vars-file env.yaml \
--timeout 540s
```

7. Set up a Cloud Scheduler job to trigger the function once a day.

```bash
gcloud scheduler jobs create http daily-update \
--schedule "0 0 * * *" \
--uri https://REGION-PROJECT_ID.cloudfunctions.net/UpdatePlaylistHandler \
--http-method GET
```

Replace `REGION` and `PROJECT_ID` with the appropriate values for your project.

8. Your Google Cloud Function is now set up and will update the specified Spotify playlist with your Last.fm top tracks every day.
