package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

// authenticateYouTube performs OAuth2 authentication flow for YouTube
func authenticateYouTube(config *oauth2.Config) (*oauth2.Token, error) {
	// Set redirect URI to localhost
	config.RedirectURL = "http://localhost:8081"
	
	// Start HTTP server for callback
	http.HandleFunc("/", completeYouTubeAuth(config))

	go func() {
		err := http.ListenAndServe(":8081", nil)
		if err != nil {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Lütfen aşağıdaki URL'yi tarayıcınızda açın:\n%s\n\n", authURL)

	// Wait for callback
	token := <-youtubeAuthCh

	return token, nil
}

var youtubeAuthCh = make(chan *oauth2.Token)

func completeYouTubeAuth(config *oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.FormValue("code")
		if code == "" {
			http.Error(w, "Authorization code not found", http.StatusBadRequest)
			return
		}

		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
			fmt.Printf("Token exchange error: %v\n", err)
			return
		}

		fmt.Fprintf(w, "YouTube authentication completed! You can close this window.")
		youtubeAuthCh <- token
	}
}

// loadYouTubeToken loads saved OAuth2 token from file
func loadYouTubeToken() (*oauth2.Token, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	tokenFile := filepath.Join(homeDir, ".spotomusic_youtube_token.json")
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// saveYouTubeToken saves OAuth2 token to file
func saveYouTubeToken(token *oauth2.Token) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	tokenFile := filepath.Join(homeDir, ".spotomusic_youtube_token.json")
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	return os.WriteFile(tokenFile, data, 0600)
}
