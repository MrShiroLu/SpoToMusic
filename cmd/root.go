package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"spotomusic/internal/logger"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "spotomusic",
	Short: "A CLI application that transfers Spotify playlists to YouTube",
	Long: `SpoToMusic is a command-line tool that allows you to easily 
transfer your Spotify playlists to YouTube.

Features:
- Lists all playlists from your Spotify account
- Transfers selected playlists to YouTube
- High success rate with smart song matching algorithm
- Detailed progress reports`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.spotomusic.yaml)")
	rootCmd.PersistentFlags().Bool("verbose", false, "verbose output")
	rootCmd.PersistentFlags().Bool("dry-run", false, "simulation only, don't perform actual transfer")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("dry-run", rootCmd.PersistentFlags().Lookup("dry-run"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Set up logging
	logger.SetVerbose(viper.GetBool("verbose"))
	
	logger.Info("Starting SpoToMusic...")
}
