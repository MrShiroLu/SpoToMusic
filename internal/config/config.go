package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"spotomusic/internal/logger"
)

type Config struct {
	Spotify   SpotifyConfig   `mapstructure:"spotify"`
	YouTube   YouTubeConfig   `mapstructure:"youtube"`
	Transfer  TransferConfig  `mapstructure:"transfer"`
	Logging   LoggingConfig   `mapstructure:"logging"`
}

type SpotifyConfig struct {
}

type YouTubeConfig struct {
	CredentialsFile string `mapstructure:"credentials_file"`
}

type TransferConfig struct {
	MaxRetries     int  `mapstructure:"max_retries"`
	RetryDelay     int  `mapstructure:"retry_delay_ms"`
	SkipExisting   bool `mapstructure:"skip_existing"`
	DryRun         bool `mapstructure:"dry_run"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Verbose bool  `mapstructure:"verbose"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	viper.SetConfigName("spotomusic")
	viper.SetConfigType("yaml")
	
	// Add config paths
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home directory bulunamadı: %v", err)
	}
	
	viper.AddConfigPath(".")
	viper.AddConfigPath(homeDir)
	viper.AddConfigPath(filepath.Join(homeDir, ".spotomusic"))
	
	// Set default values
	setDefaults()
	
	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("config file okunamadı: %v", err)
		}
		// Config file not found, use defaults
		logger.Info("Config file bulunamadı, default değerler kullanılıyor")
	}
	
	// Bind environment variables
	viper.AutomaticEnv()
	
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("config unmarshal edilemedi: %v", err)
	}
	
	// Override with environment variables
	overrideWithEnv(&config)
	
	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Spotify defaults - no defaults needed for public playlists
	
	// YouTube defaults
	homeDir, _ := os.UserHomeDir()
	viper.SetDefault("youtube.credentials_file", filepath.Join(homeDir, ".spotomusic_youtube_credentials.json"))
	
	// Transfer defaults
	viper.SetDefault("transfer.max_retries", 3)
	viper.SetDefault("transfer.retry_delay_ms", 1000)
	viper.SetDefault("transfer.skip_existing", true)
	viper.SetDefault("transfer.dry_run", false)
	
	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.verbose", false)
}

// overrideWithEnv overrides config with environment variables
func overrideWithEnv(config *Config) {
	// Spotify
	// if username := os.Getenv("SPOTIFY_USERNAME"); username != "" {
	// 	config.Spotify.Username = username
	// }
	
	// YouTube
	if credentialsFile := os.Getenv("YOUTUBE_CREDENTIALS_FILE"); credentialsFile != "" {
		config.YouTube.CredentialsFile = credentialsFile
	}
	if credentialsJSON := os.Getenv("YOUTUBE_CREDENTIALS_JSON"); credentialsJSON != "" {
		// Write to temp file
		homeDir, _ := os.UserHomeDir()
		tempFile := filepath.Join(homeDir, ".spotomusic_youtube_credentials.json")
		os.WriteFile(tempFile, []byte(credentialsJSON), 0600)
		config.YouTube.CredentialsFile = tempFile
	}
	
	// Transfer
	if dryRun := os.Getenv("SPOTOMUSIC_DRY_RUN"); dryRun == "true" {
		config.Transfer.DryRun = true
	}
	if skipExisting := os.Getenv("SPOTOMUSIC_SKIP_EXISTING"); skipExisting == "false" {
		config.Transfer.SkipExisting = false
	}
	
	// Logging
	if verbose := os.Getenv("SPOTOMUSIC_VERBOSE"); verbose == "true" {
		config.Logging.Verbose = true
		config.Logging.Level = "debug"
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate Spotify config
	// if c.Spotify.Username == "" {
	// 	return fmt.Errorf("SPOTIFY_USERNAME gerekli (public playlist sahibinin kullanıcı adı)")
	// }
	
	// Validate YouTube config
	if c.YouTube.CredentialsFile == "" {
		return fmt.Errorf("YouTube credentials file gerekli")
	}
	if _, err := os.Stat(c.YouTube.CredentialsFile); os.IsNotExist(err) {
		return fmt.Errorf("YouTube credentials file bulunamadı: %s", c.YouTube.CredentialsFile)
	}
	
	return nil
}

// Save saves the configuration to file
func (c *Config) Save() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home directory bulunamadı: %v", err)
	}
	
	configDir := filepath.Join(homeDir, ".spotomusic")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("config directory oluşturulamadı: %v", err)
	}
	
	configFile := filepath.Join(configDir, "spotomusic.yaml")
	
	// Set values in viper
	viper.Set("spotify", c.Spotify)
	viper.Set("youtube", c.YouTube)
	viper.Set("transfer", c.Transfer)
	viper.Set("logging", c.Logging)
	
	return viper.WriteConfigAs(configFile)
}
