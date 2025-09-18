package transfer

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"spotomusic/internal/spotify"
	"spotomusic/internal/youtube"
)

type Service struct {
	spotifyClient *spotify.Client
	youtubeClient *youtube.Client
}

type TransferResult struct {
	PlaylistName    string
	TotalTracks     int
	MatchedTracks   int
	FailedTracks    int
	YouTubePlaylist *youtube.YouTubePlaylist
	Errors          []string
}

// NewService creates a new transfer service
func NewService() *Service {
	return &Service{}
}

// TransferPlaylist transfers a single playlist from Spotify to YouTube Music
func (s *Service) TransferPlaylist(playlistID string, playlistName string, dryRun bool) error {
	// Initialize clients
	if err := s.initializeClients(); err != nil {
		return fmt.Errorf("clients initialize edilemedi: %v", err)
	}

	// If playlistName is not provided, try to get it from Spotify
	if playlistName == "" {
		spotifyPlaylistInfo, err := s.spotifyClient.GetPlaylistInfo(playlistID, "Unknown Playlist")
		if err != nil {
			return fmt.Errorf("failed to get playlist info: %v", err)
		}
		// Update playlist name from GetPlaylistInfo result
		playlistName = spotifyPlaylistInfo.Name
	}

	// Get tracks
	tracks, err := s.spotifyClient.GetPlaylistTracks(playlistID)
	if err != nil {
		return fmt.Errorf("playlist tracks alınamadı: %v", err)
	}

	// Update track count after getting tracks
	spotifyPlaylist := spotify.Playlist{
		ID: playlistID,
		Name: playlistName,
		TrackCount: len(tracks),
	}

	fmt.Printf("Transferring playlist: %s (%d tracks)\n", spotifyPlaylist.Name, spotifyPlaylist.TrackCount)

	// Check if playlist already exists on YouTube
	exists, existingPlaylist, err := s.youtubeClient.PlaylistExists(spotifyPlaylist.Name)
	if err != nil {
		return fmt.Errorf("playlist existence check failed: %v", err)
	}

	var youtubePlaylist *youtube.YouTubePlaylist
	if exists {
		fmt.Printf("Playlist '%s' already exists on YouTube Music. Using existing playlist.\n", spotifyPlaylist.Name)
		youtubePlaylist = existingPlaylist
	} else {
		// Create new playlist
		if dryRun {
			fmt.Printf("[DRY RUN] Would create playlist: %s\n", spotifyPlaylist.Name)
			youtubePlaylist = &youtube.YouTubePlaylist{
				Title:       spotifyPlaylist.Name,
				Description: fmt.Sprintf("Transferred from Spotify playlist: %s", playlistID),
			}
		} else {
			youtubePlaylist, err = s.youtubeClient.CreatePlaylist(spotifyPlaylist.Name, fmt.Sprintf("Transferred from Spotify playlist: %s", playlistID))
			if err != nil {
				return fmt.Errorf("YouTube playlist oluşturulamadı: %v", err)
			}
			fmt.Printf("Created YouTube playlist: %s\n", youtubePlaylist.Title)
		}
	}

	// Transfer tracks
	result := s.transferTracks(tracks, youtubePlaylist, dryRun)
	s.printTransferResult(result)

	return nil
}

// TransferAllPlaylists transfers all playlists from Spotify to YouTube Music
func (s *Service) TransferAllPlaylists(dryRun bool) error {
	// Initialize clients
	if err := s.initializeClients(); err != nil {
		return fmt.Errorf("clients initialize edilemedi: %v", err)
	}

	// Get all playlists
	playlists, err := s.spotifyClient.GetUserPlaylists()
	if err != nil {
		return fmt.Errorf("playlists alınamadı: %v", err)
	}

	fmt.Printf("Found %d playlists to transfer\n", len(playlists))

	var totalResults []TransferResult
	for i, playlist := range playlists {
		fmt.Printf("\n[%d/%d] Processing: %s\n", i+1, len(playlists), playlist.Name)
		
		// Get tracks
		tracks, err := s.spotifyClient.GetPlaylistTracks(playlist.ID)
		if err != nil {
			fmt.Printf("Error getting tracks for %s: %v\n", playlist.Name, err)
			continue
		}

		// Update playlist track count
		playlist.TrackCount = len(tracks)

		// Check if playlist exists
		exists, existingPlaylist, err := s.youtubeClient.PlaylistExists(playlist.Name)
		if err != nil {
			fmt.Printf("Error checking playlist existence: %v\n", err)
			continue
		}

		var youtubePlaylist *youtube.YouTubePlaylist
		if exists {
			youtubePlaylist = existingPlaylist
		} else {
			if dryRun {
				youtubePlaylist = &youtube.YouTubePlaylist{
					Title:       playlist.Name,
					Description: playlist.Description,
				}
			} else {
				youtubePlaylist, err = s.youtubeClient.CreatePlaylist(playlist.Name, playlist.Description)
				if err != nil {
					fmt.Printf("Error creating YouTube playlist: %v\n", err)
					continue
				}
			}
		}

		// Transfer tracks
		result := s.transferTracks(tracks, youtubePlaylist, dryRun)
		totalResults = append(totalResults, result)
	}

	// Print summary
	s.printSummary(totalResults)

	return nil
}

// TransferInteractive provides interactive playlist selection
func (s *Service) TransferInteractive(dryRun bool) error {
	// Initialize clients
	if err := s.initializeClients(); err != nil {
		return fmt.Errorf("clients initialize edilemedi: %v", err)
	}

	// Get all playlists
	playlists, err := s.spotifyClient.GetUserPlaylists()
	if err != nil {
		return fmt.Errorf("playlists alınamadı: %v", err)
	}

	// Create selection prompt
	items := make([]string, len(playlists))
	for i, playlist := range playlists {
		items[i] = fmt.Sprintf("%s (%d tracks)", playlist.Name, playlist.TrackCount)
	}

	prompt := promptui.Select{
		Label: "Transfer edilecek playlist'i seçin",
		Items: items,
		Size:  10,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("playlist selection failed: %v", err)
	}

	selectedPlaylist := playlists[index]
	return s.TransferPlaylist(selectedPlaylist.ID, selectedPlaylist.Name, dryRun)
}

// initializeClients initializes Spotify and YouTube clients
func (s *Service) initializeClients() error {
	var err error

	if s.spotifyClient == nil {
		s.spotifyClient, err = spotify.NewClient()
		if err != nil {
			return fmt.Errorf("Spotify client: %v", err)
		}
	}

	if s.youtubeClient == nil {
		s.youtubeClient, err = youtube.NewClient()
		if err != nil {
			return fmt.Errorf("YouTube client: %v", err)
		}
	}

	return nil
}

// transferTracks transfers tracks from Spotify to YouTube Music
func (s *Service) transferTracks(tracks []spotify.Track, youtubePlaylist *youtube.YouTubePlaylist, dryRun bool) TransferResult {
	result := TransferResult{
		PlaylistName:    youtubePlaylist.Title,
		TotalTracks:     len(tracks),
		YouTubePlaylist: youtubePlaylist,
	}

	fmt.Printf("Transferring %d tracks...\n", len(tracks))

	for i, track := range tracks {
		fmt.Printf("[%d/%d] %s - %s", i+1, len(tracks), track.Artist, track.Name)
		
		// Search for track on YouTube
		query := s.buildSearchQuery(track)
		youtubeVideos, err := s.youtubeClient.SearchVideo(query)
		if err != nil {
			fmt.Printf(" [ERROR: %v]\n", err)
			result.FailedTracks++
			result.Errors = append(result.Errors, fmt.Sprintf("%s - %s: %v", track.Artist, track.Name, err))
			continue
		}

		if len(youtubeVideos) == 0 {
			fmt.Printf(" [NOT FOUND]\n")
			result.FailedTracks++
			result.Errors = append(result.Errors, fmt.Sprintf("%s - %s: No matching video found", track.Artist, track.Name))
			continue
		}

		// Find best match
		bestMatch := s.findBestMatch(track, youtubeVideos)
		if bestMatch == nil {
			fmt.Printf(" [NO GOOD MATCH]\n")
			result.FailedTracks++
			result.Errors = append(result.Errors, fmt.Sprintf("%s - %s: No good match found", track.Artist, track.Name))
			continue
		}

		// Add to playlist
		if !dryRun {
			err = s.youtubeClient.AddVideoToPlaylist(youtubePlaylist.ID, bestMatch.ID)
			if err != nil {
				fmt.Printf(" [ADD ERROR: %v]\n", err)
				result.FailedTracks++
				result.Errors = append(result.Errors, fmt.Sprintf("%s - %s: %v", track.Artist, track.Name, err))
				continue
			}
		}

		fmt.Printf(" [MATCHED: %s]\n", bestMatch.Title)
		result.MatchedTracks++
		
		// Add delay to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	return result
}

// buildSearchQuery builds a search query for YouTube
func (s *Service) buildSearchQuery(track spotify.Track) string {
	// Clean up the query
	query := fmt.Sprintf("%s %s", track.Artist, track.Name)
	
	// Remove common words that might interfere with search
	query = strings.ReplaceAll(query, "ft.", "")
	query = strings.ReplaceAll(query, "feat.", "")
	query = strings.ReplaceAll(query, "featuring", "")
	
	// Remove extra spaces
	query = strings.Join(strings.Fields(query), " ")
	
	return query
}

// findBestMatch finds the best matching YouTube video
func (s *Service) findBestMatch(track spotify.Track, videos []youtube.YouTubeVideo) *youtube.YouTubeVideo {
	if len(videos) == 0 {
		return nil
	}

	// Simple matching algorithm
	trackTitle := strings.ToLower(track.Name)
	trackArtist := strings.ToLower(track.Artist)

	for _, video := range videos {
		videoTitle := strings.ToLower(video.Title)
		
		// Check if both title and artist are in the video title
		if strings.Contains(videoTitle, trackTitle) && strings.Contains(videoTitle, trackArtist) {
			return &video
		}
		
		// Check if just the title matches (for remixes, covers, etc.)
		if strings.Contains(videoTitle, trackTitle) {
			return &video
		}
	}

	// Return the first result if no good match found
	return &videos[0]
}

// printTransferResult prints the result of a transfer
func (s *Service) printTransferResult(result TransferResult) {
	fmt.Printf("\n" + strings.Repeat("=", 50) + "\n")
	fmt.Printf("Transfer Result: %s\n", result.PlaylistName)
	fmt.Printf("Total Tracks: %d\n", result.TotalTracks)
	
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	
	fmt.Printf("Matched: %s\n", green(result.MatchedTracks))
	fmt.Printf("Failed: %s\n", red(result.FailedTracks))
	
	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors:\n")
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}
	
	fmt.Printf(strings.Repeat("=", 50) + "\n\n")
}

// printSummary prints a summary of all transfers
func (s *Service) printSummary(results []TransferResult) {
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("TRANSFER SUMMARY\n")
	fmt.Printf(strings.Repeat("=", 60) + "\n")
	
	totalPlaylists := len(results)
	totalTracks := 0
	totalMatched := 0
	totalFailed := 0
	
	for _, result := range results {
		totalTracks += result.TotalTracks
		totalMatched += result.MatchedTracks
		totalFailed += result.FailedTracks
		
		green := color.New(color.FgGreen).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()
		
		fmt.Printf("%-30s | %s/%s | %s failed\n", 
			result.PlaylistName, 
			green(result.MatchedTracks), 
			fmt.Sprintf("%d", result.TotalTracks),
			red(result.FailedTracks))
	}
	
	fmt.Printf(strings.Repeat("-", 60) + "\n")
	fmt.Printf("TOTAL: %d playlists, %d tracks, %d matched, %d failed\n", 
		totalPlaylists, totalTracks, totalMatched, totalFailed)
	fmt.Printf(strings.Repeat("=", 60) + "\n")
}
