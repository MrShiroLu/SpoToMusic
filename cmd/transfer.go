package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"spotomusic/internal/transfer"
)

// transferCmd represents the transfer command
var transferCmd = &cobra.Command{
	Use:   "transfer [playlist-id]",
	Short: "Transfers the specified playlist to YouTube",
	Long: `This command transfers the specified Spotify playlist to YouTube.

Examples:
  spotomusic transfer 37i9dQZF1DXcBWIGoYBM5M --name "My Awesome Playlist"
  spotomusic transfer --all
  spotomusic transfer --interactive`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		interactive, _ := cmd.Flags().GetBool("interactive")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		playlistName, _ := cmd.Flags().GetString("name")

		transferService := transfer.NewService()

		if all {
			return transferService.TransferAllPlaylists(dryRun)
		}

		if interactive {
			return transferService.TransferInteractive(dryRun)
		}

		if len(args) == 0 {
			return fmt.Errorf("playlist ID required or use --all/--interactive flag")
		}

		playlistID := args[0]
		return transferService.TransferPlaylist(playlistID, playlistName, dryRun)
	},
}

func init() {
	rootCmd.AddCommand(transferCmd)

	transferCmd.Flags().Bool("all", false, "Transfer all playlists")
	transferCmd.Flags().Bool("interactive", false, "Interactive mode - select playlists")
	transferCmd.Flags().String("name", "", "Name of the Spotify playlist (required for single playlist transfer)")
	transferCmd.Flags().String("youtube-playlist-name", "", "Name of the playlist to create on YouTube")
	transferCmd.Flags().Bool("skip-existing", true, "Skip existing playlists")
}
