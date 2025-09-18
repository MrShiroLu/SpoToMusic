package cmd

import (
	"fmt"
	"spotomusic/internal/spotify"
	"strings"

	"github.com/spf13/cobra"
)

// debugCmd represents the debug command
var debugCmd = &cobra.Command{
	Use:   "debug [playlist-id]",
	Short: "Debug Spotify playlist parsing",
	Long: `This command helps debug Spotify playlist parsing by showing
detailed information about what data is being extracted.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		playlistID := args[0]
		playlistName, _ := cmd.Flags().GetString("name")
		
		client, err := spotify.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create Spotify client: %v", err)
		}

		fmt.Printf("Debugging playlist: %s\n", playlistID)
		fmt.Println(strings.Repeat("=", 50))

		// Get playlist info
		playlist, err := client.GetPlaylistInfo(playlistID, playlistName)
		if err != nil {
			return fmt.Errorf("failed to get playlist info: %v", err)
		}

		fmt.Printf("Playlist Name: %s\n", playlist.Name)
		fmt.Printf("Track Count: %d\n", playlist.TrackCount)
		fmt.Printf("Description: %s\n", playlist.Description)
		fmt.Printf("Public: %t\n", playlist.Public)
		fmt.Printf("Owner: %s\n", playlist.Owner)

		// Get tracks
		tracks, err := client.GetPlaylistTracks(playlistID)
		if err != nil {
			return fmt.Errorf("failed to get tracks: %v", err)
		}

		fmt.Printf("\nFound %d tracks:\n", len(tracks))
		for i, track := range tracks {
			fmt.Printf("%d. %s - %s\n", i+1, track.Artist, track.Name)
		}

		// Save HTML for analysis
		if err := client.SaveHTMLForAnalysis(playlistID); err != nil {
			fmt.Printf("Warning: Could not save HTML for analysis: %v\n", err)
		} else {
			fmt.Printf("\nHTML saved to debug.html for analysis\n")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
	debugCmd.Flags().String("name", "", "Name of the Spotify playlist (for debugging)")
}
