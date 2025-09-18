package youtube

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/youtube/v3"
)

type Client struct {
	service *youtube.Service
	httpClient *http.Client
}

type YouTubePlaylist struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	VideoCount  int    `json:"video_count"`
}

type YouTubeVideo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	ChannelName string `json:"channel_name"`
	Duration    string `json:"duration"`
	URL         string `json:"url"`
}

// NewClient creates a new YouTube client with OAuth2 authentication
func NewClient() (*Client, error) {
	ctx := context.Background()

	// Load credentials from environment or file
	credentialsJSON := os.Getenv("YOUTUBE_CREDENTIALS_JSON")
	if credentialsJSON == "" {
		// Try to load from file
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("home directory bulunamadı: %v", err)
		}
		
		credsFile := filepath.Join(homeDir, ".spotomusic_youtube_credentials.json")
		data, err := os.ReadFile(credsFile)
		if err != nil {
			return nil, fmt.Errorf("YouTube credentials dosyası bulunamadı. Lütfen %s dosyasını oluşturun", credsFile)
		}
		credentialsJSON = string(data)
	}

	// Parse credentials
	config, err := google.ConfigFromJSON([]byte(credentialsJSON), youtube.YoutubeScope)
	if err != nil {
		return nil, fmt.Errorf("credentials parse edilemedi: %v", err)
	}

	// Check for saved token
	token, err := loadYouTubeToken()
	if err != nil {
		// No saved token, need to authenticate
		token, err = authenticateYouTube(config)
		if err != nil {
			return nil, fmt.Errorf("YouTube authentication failed: %v", err)
		}
		// Save token for future use
		if err := saveYouTubeToken(token); err != nil {
			fmt.Printf("Warning: YouTube token kaydedilemedi: %v\n", err)
		}
	}

	// Create HTTP client with token
	httpClient := config.Client(ctx, token)
	
	// Create YouTube service
	service, err := youtube.New(httpClient)
	if err != nil {
		return nil, fmt.Errorf("YouTube service oluşturulamadı: %v", err)
	}

	return &Client{
		service:    service,
		httpClient: httpClient,
	}, nil
}

// CreatePlaylist creates a new playlist on YouTube
func (c *Client) CreatePlaylist(title, description string) (*YouTubePlaylist, error) {
	playlist := &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			Title:       title,
			Description: description,
		},
		Status: &youtube.PlaylistStatus{
			PrivacyStatus: "private", // Private by default
		},
	}

	call := c.service.Playlists.Insert([]string{"snippet", "status"}, playlist)
	result, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("playlist oluşturulamadı: %v", err)
	}

	return &YouTubePlaylist{
		ID:          result.Id,
		Title:       result.Snippet.Title,
		Description: result.Snippet.Description,
		VideoCount:  0,
	}, nil
}

// SearchVideo searches for a video on YouTube
func (c *Client) SearchVideo(query string) ([]YouTubeVideo, error) {
	call := c.service.Search.List([]string{"snippet"}).
		Q(query).
		Type("video").
		MaxResults(5) // Limit to 5 results for better matching

	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("video search failed: %v", err)
	}

	var videos []YouTubeVideo
	for _, item := range response.Items {
		videos = append(videos, YouTubeVideo{
			ID:          item.Id.VideoId,
			Title:       item.Snippet.Title,
			ChannelName: item.Snippet.ChannelTitle,
			Duration:    "", // Duration not available in search results
			URL:         fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.Id.VideoId),
		})
	}

	return videos, nil
}

// AddVideoToPlaylist adds a video to a playlist
func (c *Client) AddVideoToPlaylist(playlistID, videoID string) error {
	playlistItem := &youtube.PlaylistItem{
		Snippet: &youtube.PlaylistItemSnippet{
			PlaylistId: playlistID,
			ResourceId: &youtube.ResourceId{
				Kind:    "youtube#video",
				VideoId: videoID,
			},
		},
	}

	call := c.service.PlaylistItems.Insert([]string{"snippet"}, playlistItem)
	_, err := call.Do()
	if err != nil {
		// Check if it's a duplicate error
		if googleapi.IsNotModified(err) || strings.Contains(err.Error(), "already exists") {
			return nil // Ignore duplicate errors
		}
		return fmt.Errorf("video playlist'e eklenemedi: %v", err)
	}

	return nil
}

// GetUserPlaylists retrieves all playlists for the authenticated user
func (c *Client) GetUserPlaylists() ([]YouTubePlaylist, error) {
	call := c.service.Playlists.List([]string{"snippet", "contentDetails"}).
		Mine(true).
		MaxResults(50)

	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("playlists alınamadı: %v", err)
	}

	var playlists []YouTubePlaylist
	for _, playlist := range response.Items {
		playlists = append(playlists, YouTubePlaylist{
			ID:          playlist.Id,
			Title:       playlist.Snippet.Title,
			Description: playlist.Snippet.Description,
			VideoCount:  int(playlist.ContentDetails.ItemCount),
		})
	}

	return playlists, nil
}

// PlaylistExists checks if a playlist with the given title exists
func (c *Client) PlaylistExists(title string) (bool, *YouTubePlaylist, error) {
	playlists, err := c.GetUserPlaylists()
	if err != nil {
		return false, nil, err
	}

	for _, playlist := range playlists {
		if strings.EqualFold(playlist.Title, title) {
			return true, &playlist, nil
		}
	}

	return false, nil, nil
}
