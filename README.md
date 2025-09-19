# SpoToMusic

A Go CLI application that transfers your Spotify playlists to YouTube.

## Features

- Lists all your Spotify playlists
- Automatically transfers playlists to YouTube
- Fast and reliable transfer
- Detailed progress reports
- Easy setup and usage

## Installation

### Requirements

- Go 1.21 or higher
- Spotify playlist links
- Google Cloud Console 

### 1. Clone the repository

```bash
git clone https://github.com/yourusername/spotomusic.git
cd spotomusic
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Spotify setup

1. Get your Spotify playlist links
2. No API key needed for public playlists
3. Only playlist links are required

### 4. YouTube API setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project
3. Enable YouTube Data API v3
4. Create OAuth 2.0 credentials
5. Download the JSON file

### 5. Set environment variables

```bash
export SPOTIFY_PLAYLIST_LINKS="https://open.spotify.com/playlist/YOUR_PLAYLIST_ID_1,https://open.spotify.com/playlist/YOUR_PLAYLIST_ID_2"
export YOUTUBE_CREDENTIALS_FILE="/path/to/your/credentials.json"
```
## Usage

### Basic commands

```bash
# List playlists
./spotomusic list

# Transfer a specific playlist
./spotomusic transfer 37i9dQZF1DXcBWIGoYBM5M

# Transfer all playlists
./spotomusic transfer --all

# Interactive mode
./spotomusic transfer --interactive

# Dry run (simulation only)
./spotomusic transfer --all --dry-run
```

### Command options

```bash
# Verbose output
./spotomusic --verbose list

# Specify config file
./spotomusic --config /path/to/config.yaml list

# Dry run
./spotomusic --dry-run transfer --all
```

## Configuration

The application searches for configuration files in the following order:

1. File specified with `--config` flag
2. `$HOME/.spotomusic.yaml`
3. `$HOME/.spotomusic/spotomusic.yaml`
4. `spotomusic.yaml` in current directory

### Example configuration

```yaml
spotify:
  # username: "spotify_username"  # Public playlist owner's username - No longer needed with playlist links

youtube:
  credentials_file: "/path/to/credentials.json"

transfer:
  max_retries: 3
  retry_delay_ms: 1000
  skip_existing: true
  dry_run: false

logging:
  level: "info"
  verbose: false
```

## Development

### Run tests

```bash
go test ./...
```

### Build

```bash
go build -o spotomusic .
```

### Cross-compile

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o spotomusic-linux .

# Windows
GOOS=windows GOARCH=amd64 go build -o spotomusic.exe .

# macOS
GOOS=darwin GOARCH=amd64 go build -o spotomusic-macos .
```

## API Limit

- **YouTube Data API**: You have a daily quota of 10,000 credits. Each operation (like searching for a video or adding a video to a playlist) consumes credits. This limit can be quickly reached with large playlists. Refer to the [official YouTube Data API Quota Usage](https://developers.google.com/youtube/v3/guides/quota) for detailed information.

## Troubleshooting

### Common errors

1. **"YouTube credentials file not found"**
   - Check the path to the JSON file downloaded from Google Cloud Console

2. **"Playlist not found"**
   - Make sure the provided Spotify playlist links are correct and accessible
   - Verify the playlist ID is correct

3. **"HTTP 404" error (Spotify)**
   - Make sure the Spotify playlist link is correct
   - Check that the playlist is public

4. **"quotaExceeded" error (YouTube)**
   - You have exceeded your daily YouTube Data API quota. Please try again after 24 hours or request a quota increase from Google Cloud Console.

### Log files

Log files are stored at `$HOME/.spotomusic/logs/spotomusic.log`


## Acknowledgments

- [YouTube Data API](https://developers.google.com/youtube/v3)
- [Cobra CLI](https://github.com/spf13/cobra)
- [Viper Config](https://github.com/spf13/viper)
