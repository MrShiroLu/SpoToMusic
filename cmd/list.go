package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"spotomusic/internal/spotify"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists Spotify playlists",
	Long: `This command displays all playlists from your Spotify account 
and prepares them for transfer.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := spotify.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create Spotify client: %v", err)
		}

		playlists, err := client.GetUserPlaylists()
		if err != nil {
			return fmt.Errorf("failed to get playlists: %v", err)
		}

		fmt.Printf("Found %d playlists:\n\n", len(playlists))
		for i, playlist := range playlists {
			fmt.Printf("%d. %s (%d tracks)\n", i+1, playlist.Name, playlist.TrackCount)
			if playlist.Description != "" {
				fmt.Printf("   Description: %s\n", playlist.Description)
			}
			fmt.Printf("   ID: %s\n\n", playlist.ID)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
