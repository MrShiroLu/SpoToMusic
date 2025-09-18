package spotify

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Client struct {
	httpClient *http.Client
}

type Playlist struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TrackCount  int    `json:"track_count"`
	Public      bool   `json:"public"`
	Owner       string `json:"owner"`
}

type Track struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	Duration int    `json:"duration_ms"`
	URI      string `json:"uri"`
}

// NewClient creates a new Spotify client for public playlists (no API key required)
func NewClient() (*Client, error) {
	// No API key needed for public playlists
	httpClient := &http.Client{}
	
	return &Client{
		httpClient: httpClient,
	}, nil
}

// GetUserPlaylists retrieves playlists from provided links
func (c *Client) GetUserPlaylists() ([]Playlist, error) {
	// Get playlist links from environment variable
	playlistLinks := os.Getenv("SPOTIFY_PLAYLIST_LINKS")
	if playlistLinks == "" {
		return nil, fmt.Errorf("SPOTIFY_PLAYLIST_LINKS environment variable required (comma-separated playlist URLs)")
	}

	// Split links by comma
	links := strings.Split(playlistLinks, ",")
	var result []Playlist

	for _, link := range links {
		link = strings.TrimSpace(link)
		if link == "" {
			continue
		}

		// Extract playlist ID from URL
		playlistID, err := c.extractPlaylistID(link)
		if err != nil {
			fmt.Printf("Warning: Invalid playlist URL %s: %v\n", link, err)
			continue
		}

		// Get playlist info
		playlist, err := c.GetPlaylistInfo(playlistID, playlistID) // Pass playlistID as name for now
		if err != nil {
			fmt.Printf("Warning: Failed to get playlist %s: %v\n", link, err)
			continue
		}

		result = append(result, playlist)
	}

	return result, nil
}

// extractPlaylistID extracts playlist ID from Spotify URL
func (c *Client) extractPlaylistID(url string) (string, error) {
	// Handle different Spotify URL formats:
	// https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M
	// https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M?si=...
	// spotify:playlist:37i9dQZF1DXcBWIGoYBM5M
	
	if strings.Contains(url, "open.spotify.com/playlist/") {
		// Extract from web URL
		parts := strings.Split(url, "/playlist/")
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid Spotify playlist URL")
		}
		playlistID := strings.Split(parts[1], "?")[0] // Remove query parameters
		return playlistID, nil
	} else if strings.Contains(url, "spotify:playlist:") {
		// Extract from URI format
		parts := strings.Split(url, "spotify:playlist:")
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid Spotify playlist URI")
		}
		return parts[1], nil
	}
	
	return "", fmt.Errorf("unsupported URL format")
}

// GetPlaylistInfo gets playlist information using Spotify embed API
func (c *Client) GetPlaylistInfo(playlistID string, playlistName string) (Playlist, error) {
	// Use Spotify embed API which provides better data
	url := fmt.Sprintf("https://open.spotify.com/embed/playlist/%s", playlistID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Playlist{}, fmt.Errorf("request oluşturulamadı: %v", err)
	}

	// Set headers to mimic a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Playlist{}, fmt.Errorf("playlist isteği başarısız: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return Playlist{}, fmt.Errorf("playlist isteği başarısız: HTTP %d", resp.StatusCode)
	}

	// Read the HTML content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Playlist{}, fmt.Errorf("HTML okunamadı: %v", err)
	}

	// Parse playlist info from HTML
	playlist, err := c.parsePlaylistFromEmbedHTML(string(body), playlistID, playlistName)
	if err != nil {
		return Playlist{}, fmt.Errorf("playlist parse edilemedi: %v", err)
	}

	// Get tracks to determine the actual track count
	tracks, err := c.GetPlaylistTracks(playlistID)
	if err == nil {
		playlist.TrackCount = len(tracks)
	}

	return playlist, nil
}

// parsePlaylistFromEmbedHTML extracts playlist information from embed HTML content
func (c *Client) parsePlaylistFromEmbedHTML(htmlContent, playlistID, playlistName string) (Playlist, error) {
	// In this simplified approach, we are not extracting playlist name or track count from HTML directly.
	// These will be provided via CLI arguments or derived from track scraping results.
	// However, we still need to return a basic Playlist object.

	// Try to extract playlist name from the og:title meta tag as a fallback for the playlist name
	// Use the provided playlistName if available, otherwise try to extract from HTML
	if playlistName != "Unknown Playlist" {
		return Playlist{
			ID:          playlistID,
			Name:        playlistName,
			Description: "",
			TrackCount:  0,
			Public:      true,
			Owner:       "Unknown",
		}, nil
	}

	playlistName = "Unknown Playlist"
	ogTitleRegex := regexp.MustCompile(`<meta property="og:title" content="([^"]+)"\/>`)
	ogTitleMatches := ogTitleRegex.FindStringSubmatch(htmlContent)
	
	if len(ogTitleMatches) > 1 {
		playlistName = strings.TrimSuffix(strings.TrimSpace(ogTitleMatches[1]), " Spotify")
	}

	return Playlist{
		ID:          playlistID,
		Name:        playlistName,
		Description: "",
		TrackCount:  0,
		Public:      true,
		Owner:       "Unknown",
	}, nil
}

// parsePlaylistFromHTML extracts playlist information from HTML content
func (c *Client) parsePlaylistFromHTML(html, playlistID string) (Playlist, error) {
	// Look for JSON data in the HTML
	// Spotify embeds playlist data in a script tag
	jsonRegex := regexp.MustCompile(`"playlist":\s*({[^}]+})`)
	matches := jsonRegex.FindStringSubmatch(html)
	
	if len(matches) < 2 {
		// Fallback: try to extract basic info from HTML
		return c.parsePlaylistFromHTMLFallback(html, playlistID)
	}

	// Parse the JSON data
	var playlistData struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Public      bool   `json:"public"`
		Owner       struct {
			DisplayName string `json:"display_name"`
		} `json:"owner"`
		Tracks struct {
			Total int `json:"total"`
		} `json:"tracks"`
	}

	if err := json.Unmarshal([]byte(matches[1]), &playlistData); err != nil {
		// Fallback if JSON parsing fails
		return c.parsePlaylistFromHTMLFallback(html, playlistID)
	}

	return Playlist{
		ID:          playlistID,
		Name:        playlistData.Name,
		Description: playlistData.Description,
		TrackCount:  playlistData.Tracks.Total,
		Public:      playlistData.Public,
		Owner:       playlistData.Owner.DisplayName,
	}, nil
}

// parsePlaylistFromHTMLFallback extracts basic info when JSON parsing fails
func (c *Client) parsePlaylistFromHTMLFallback(html, playlistID string) (Playlist, error) {
	// Extract playlist name from title tag
	titleRegex := regexp.MustCompile(`<title>([^<]+)</title>`)
	titleMatches := titleRegex.FindStringSubmatch(html)
	
	name := "Unknown Playlist"
	if len(titleMatches) > 1 {
		name = strings.TrimSpace(titleMatches[1])
		// Remove " | Spotify" suffix if present
		name = strings.Replace(name, " | Spotify", "", 1)
	}

	// Try multiple methods to count tracks
	trackCount := 0
	
	// Method 1: Look for tracklist rows
	trackRowRegex := regexp.MustCompile(`data-testid="tracklist-row"`)
	trackMatches := trackRowRegex.FindAllString(html, -1)
	trackCount = len(trackMatches)
	
	// Method 2: If no tracks found, look for any track-related elements
	if trackCount == 0 {
		trackElementRegex := regexp.MustCompile(`class="[^"]*track[^"]*"`)
		trackMatches = trackElementRegex.FindAllString(html, -1)
		trackCount = len(trackMatches)
	}
	
	// Method 3: Look for track numbers
	if trackCount == 0 {
		trackNumberRegex := regexp.MustCompile(`<span[^>]*class="[^"]*track-number[^"]*"[^>]*>(\d+)</span>`)
		numberMatches := trackNumberRegex.FindAllStringSubmatch(html, -1)
		trackCount = len(numberMatches)
	}
	
	// Method 4: Look for duration patterns (each track usually has a duration)
	if trackCount == 0 {
		durationRegex := regexp.MustCompile(`\d+:\d+`)
		durationMatches := durationRegex.FindAllString(html, -1)
		trackCount = len(durationMatches)
	}

	return Playlist{
		ID:          playlistID,
		Name:        name,
		Description: "",
		TrackCount:  trackCount,
		Public:      true, // Assume public if we can access it
		Owner:       "Unknown",
	}, nil
}

// GetPlaylistTracks retrieves all tracks from a specific playlist using embed API
func (c *Client) GetPlaylistTracks(playlistID string) ([]Track, error) {
	// Use Spotify embed API which provides better track data
	url := fmt.Sprintf("https://open.spotify.com/embed/playlist/%s", playlistID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("request oluşturulamadı: %v", err)
	}

	// Set headers to mimic a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("track isteği başarısız: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("track isteği başarısız: HTTP %d", resp.StatusCode)
	}

	// Read the HTML content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HTML okunamadı: %v", err)
	}

	// Parse tracks from embed HTML
	tracks, err := c.parseTracksFromHTML(string(body))
	if err != nil {
		return nil, fmt.Errorf("track parse edilemedi: %v", err)
	}

	return tracks, nil
}

// parseTracksFromEmbedHTML extracts track information from embed HTML content
func (c *Client) parseTracksFromEmbedHTML(htmlContent string) ([]Track, error) {
	// In this simplified approach, we rely on parseTracksFromHTMLElements
	return c.parseTracksFromHTMLElements(htmlContent)
}

// parseTracksFromHTML extracts track information from HTML content
func (c *Client) parseTracksFromHTML(htmlContent string) ([]Track, error) {
	// Look for track data in the HTML
	// Spotify embeds track data in various ways, we'll try multiple approaches
	
	// Method 1: Look for JSON data in script tags (new approach)
	tracks := c.extractTracksFromScriptTags(htmlContent)
	if len(tracks) > 0 {
		return tracks, nil
	}

	// Method 2: Parse from HTML elements (fallback)
	return c.parseTracksFromHTMLElements(htmlContent)
}

// parseTracksFromJSON parses tracks from JSON data
func (c *Client) parseTracksFromJSON(jsonData string) ([]Track, error) {
	// This is a simplified parser - in reality, Spotify's JSON structure is complex
	// For now, we'll use the HTML fallback method
	return nil, fmt.Errorf("JSON parsing not implemented yet")
}

// parseTracksFromHTMLElements parses tracks from HTML elements
func (c *Client) parseTracksFromHTMLElements(html string) ([]Track, error) {
	var tracks []Track

	// Method 1: Look for track data in script tags (Spotify embeds JSON data)
	tracks = c.extractTracksFromScriptTags(html)
	if len(tracks) > 0 {
		return tracks, nil
	}

	// Method 2: Look for track elements with data-testid
	tracks = c.extractTracksFromDataTestId(html)
	if len(tracks) > 0 {
		return tracks, nil
	}

	// Method 3: Look for track information in meta tags or structured data
	tracks = c.extractTracksFromMetaData(html)
	if len(tracks) > 0 {
		return tracks, nil
	}

	// Method 4: Fallback - look for any text patterns that might be tracks
	tracks = c.extractTracksFromTextPatterns(html)
	return tracks, nil
}

// extractTracksFromScriptTags looks for track data in script tags
func (c *Client) extractTracksFromScriptTags(htmlContent string) []Track {
	var tracks []Track

	// Look for various JSON patterns that might contain track data in script tags
	// This is a common pattern for SPAs to embed initial data
	scriptRegex := regexp.MustCompile(`window.__spotify_initial_state = ([^;]+);`)
	matches := scriptRegex.FindStringSubmatch(htmlContent)

	if len(matches) > 1 {
		jsonData := matches[1]
		
		var initialState struct {
			Entities struct {
				Tracks struct {
					Data map[string]struct {
						Name     string `json:"name"`
						Artists  []struct {
							Name string `json:"name"`
						} `json:"artists"`
						Album    struct {
							Name string `json:"name"`
						} `json:"album"`
						Duration int    `json:"duration_ms"`
						URI      string `json:"uri"`
					} `json:"data"`
				} `json:"tracks"`
			}
		}

		if err := json.Unmarshal([]byte(jsonData), &initialState); err == nil {
			for _, trackData := range initialState.Entities.Tracks.Data {
				var artistNames []string
				for _, artist := range trackData.Artists {
					artistNames = append(artistNames, artist.Name)
				}
				artist := ""
				if len(artistNames) > 0 {
					artist = artistNames[0]
					if len(artistNames) > 1 {
						artist += " ft. " + artistNames[1]
					}
				}

				tracks = append(tracks, Track{
					Name:     trackData.Name,
					Artist:   artist,
					Album:    trackData.Album.Name,
					Duration: trackData.Duration,
					URI:      trackData.URI,
				})
			}
		}
	}

	return tracks
}

// extractTracksFromDataTestId looks for tracks using data-testid attributes
func (c *Client) extractTracksFromDataTestId(htmlContent string) []Track {
	var tracks []Track

	// Look for track rows
	trackRowRegex := regexp.MustCompile(`<li[^>]*data-testid="tracklist-row-\d+"[^>]*>(.*?)</li>`)
	rowMatches := trackRowRegex.FindAllStringSubmatch(htmlContent, -1)

	for _, row := range rowMatches {
		if len(row) > 1 {
			rowHTML := row[1]
			
			// Extract track name
			nameRegex := regexp.MustCompile(`<h3[^>]*data-encore-id="text"[^>]*>([^<]+)</h3>`)
			nameMatches := nameRegex.FindStringSubmatch(rowHTML)
			
			// Extract artist name
			artistRegex := regexp.MustCompile(`<h4[^>]*data-encore-id="text"[^>]*>([^<]+)</h4>`)
			artistMatches := artistRegex.FindStringSubmatch(rowHTML)

			// Extract duration
			durationRegex := regexp.MustCompile(`<div[^>]*data-testid="duration-cell"[^>]*>(\d+:\d+)</div>`)
			durationMatches := durationRegex.FindStringSubmatch(rowHTML)
			
			if len(nameMatches) > 1 && len(artistMatches) > 1 && len(durationMatches) > 1 {
				// Convert duration string (MM:SS) to milliseconds
				durationStr := strings.TrimSpace(durationMatches[1])
				parts := strings.Split(durationStr, ":")
				minutes, _ := strconv.Atoi(parts[0])
				seconds, _ := strconv.Atoi(parts[1])
				durationMs := (minutes*60 + seconds) * 1000

				tracks = append(tracks, Track{
					Name:   html.UnescapeString(strings.TrimSpace(nameMatches[1])),
					Artist: html.UnescapeString(strings.TrimSpace(artistMatches[1])),
					Duration: durationMs,
				})
			}
		}
	}

	return tracks
}

// extractTracksFromMetaData looks for track information in meta tags
func (c *Client) extractTracksFromMetaData(htmlContent string) []Track {
	var tracks []Track

	// Look for structured data (JSON-LD)
	jsonLdRegex := regexp.MustCompile(`<script[^>]*type="application/ld\+json"[^>]*>(.*?)</script>`)
	matches := jsonLdRegex.FindAllStringSubmatch(htmlContent, -1)

	for _, match := range matches {
		if len(match) > 1 {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(match[1]), &data); err == nil {
				// Look for track information in the structured data
				if graph, ok := data["@graph"].([]interface{}); ok {
					for _, item := range graph {
						if itemMap, ok := item.(map[string]interface{}); ok {
							if itemType, ok := itemMap["@type"].(string); ok && itemType == "MusicRecording" {
								track := Track{}
								if name, ok := itemMap["name"].(string); ok {
									track.Name = html.UnescapeString(name)
								}
								if byArtist, ok := itemMap["byArtist"].(map[string]interface{}); ok {
									if artistName, ok := byArtist["name"].(string); ok {
										track.Artist = html.UnescapeString(artistName)
									}
								}
								// Duration (if available)
								// Album (if available)
								tracks = append(tracks, track)
							}
						}
					}
				} else if items, ok := data["itemListElement"].([]interface{}); ok {
					for _, item := range items {
						if itemMap, ok := item.(map[string]interface{}); ok {
							if name, ok := itemMap["name"].(string); ok {
								tracks = append(tracks, Track{
									Name: html.UnescapeString(name),
								})
							}
						}
					}
				}
			}
		}
	}

	return tracks
}

// extractTracksFromTextPatterns looks for track patterns in text
func (c *Client) extractTracksFromTextPatterns(html string) []Track {
	var tracks []Track

	// Look for common track patterns
	lines := strings.Split(html, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Look for lines that might contain track information
		if strings.Contains(line, " - ") && len(line) > 10 && len(line) < 200 {
			// This might be a track line like "Artist - Song"
			parts := strings.Split(line, " - ")
			if len(parts) == 2 {
				artist := strings.TrimSpace(parts[0])
				title := strings.TrimSpace(parts[1])
				
				// Basic validation
				if len(artist) > 0 && len(title) > 0 && 
				   !strings.Contains(artist, "<") && !strings.Contains(title, "<") {
					tracks = append(tracks, Track{
						Name:   title,
						Artist: artist,
					})
				}
			}
		}
	}

	return tracks
}

// SearchTrack searches for a track on Spotify using direct HTTP requests
func (c *Client) SearchTrack(query string) ([]Track, error) {
	// Search tracks using direct HTTP request
	url := fmt.Sprintf("https://api.spotify.com/v1/search?q=%s&type=track&limit=5", url.QueryEscape(query))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("request oluşturulamadı: %v", err)
	}

	// Set headers for public API access
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search isteği başarısız: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("search isteği başarısız: HTTP %d", resp.StatusCode)
	}

	var searchResponse struct {
		Tracks struct {
			Items []struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				Duration int    `json:"duration_ms"`
				URI      string `json:"uri"`
				Artists  []struct {
					Name string `json:"name"`
				} `json:"artists"`
				Album struct {
					Name string `json:"name"`
				} `json:"album"`
			} `json:"items"`
		} `json:"tracks"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("search yanıtı parse edilemedi: %v", err)
	}

	var tracks []Track
	for _, track := range searchResponse.Tracks.Items {
		var artistNames []string
		for _, artist := range track.Artists {
			artistNames = append(artistNames, artist.Name)
		}
		artist := ""
		if len(artistNames) > 0 {
			artist = artistNames[0]
			if len(artistNames) > 1 {
				artist += " ft. " + artistNames[1]
			}
		}

		tracks = append(tracks, Track{
			ID:       track.ID,
			Name:     track.Name,
			Artist:   artist,
			Album:    track.Album.Name,
			Duration: track.Duration,
			URI:      track.URI,
		})
	}

	return tracks, nil
}

// SaveHTMLForAnalysis saves the HTML content to a file for debugging
func (c *Client) SaveHTMLForAnalysis(playlistID string) error {
	url := fmt.Sprintf("https://open.spotify.com/playlist/%s", playlistID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("request oluşturulamadı: %v", err)
	}

	// Set headers to mimic a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTML isteği başarısız: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTML isteği başarısız: HTTP %d", resp.StatusCode)
	}

	// Read the HTML content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("HTML okunamadı: %v", err)
	}

	// Save to file
	return os.WriteFile("debug.html", body, 0644)
}
