package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

const redirectURI = "http://localhost:8080/callback"

var (
	clientID     = "SLACK_CLIENT_ID"
	clientSecret = "SLACK_CLIENT_SECRET"
	auth         = spotify.NewAuthenticator(redirectURI, spotify.ScopePlaylistModifyPublic, spotify.ScopePlaylistModifyPrivate, spotify.ScopePlaylistReadPrivate)
	ch           = make(chan *oauth2.Token)
	state        = "abc123"
)

func main() {
	auth.SetAuthInfo(clientID, clientSecret)

	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request received")
		url := auth.AuthURL(state)
		http.Redirect(w, r, url, http.StatusFound)
	})

	go func() {
		token := <-ch
		log.Printf("Access Token: %s\nRefresh Token: %s\n", token.AccessToken, token.RefreshToken)
	}()

	log.Println("Starting web server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusInternalServerError)
		return
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
		return
	}
	ch <- tok
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "Login Completed! You can close this window.")
}
