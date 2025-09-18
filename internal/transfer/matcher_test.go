package transfer

import (
	"testing"
	"spotomusic/internal/spotify"
	"spotomusic/internal/youtube"
)

func TestFindBestMatch(t *testing.T) {
	service := &Service{}
	
	tests := []struct {
		name        string
		track       spotify.Track
		videos      []youtube.YouTubeVideo
		expectMatch bool
	}{
		{
			name: "Perfect match",
			track: spotify.Track{
				Name:   "Shape of You",
				Artist: "Ed Sheeran",
			},
			videos: []youtube.YouTubeVideo{
				{
					ID:    "1",
					Title: "Ed Sheeran - Shape of You (Official Music Video)",
				},
				{
					ID:    "2", 
					Title: "Some Other Song",
				},
			},
			expectMatch: true,
		},
		{
			name: "Partial match",
			track: spotify.Track{
				Name:   "Shape of You",
				Artist: "Ed Sheeran",
			},
			videos: []youtube.YouTubeVideo{
				{
					ID:    "1",
					Title: "Shape of You - Ed Sheeran Cover",
				},
			},
			expectMatch: true,
		},
		{
			name: "No match",
			track: spotify.Track{
				Name:   "Shape of You",
				Artist: "Ed Sheeran",
			},
			videos: []youtube.YouTubeVideo{
				{
					ID:    "1",
					Title: "Completely Different Song",
				},
			},
			expectMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := service.findBestMatch(tt.track, tt.videos)
			
			if tt.expectMatch && match == nil {
				t.Errorf("Expected match but got nil")
			}
			if !tt.expectMatch && match != nil {
				t.Errorf("Expected no match but got %v", match)
			}
		})
	}
}
