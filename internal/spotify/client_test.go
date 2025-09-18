package spotify

import (
	"strings"
	"testing"
)

func TestBuildSearchQuery(t *testing.T) {
	tests := []struct {
		name     string
		artist   string
		title    string
		expected string
	}{
		{
			name:     "Simple track",
			artist:   "Ed Sheeran",
			title:    "Shape of You",
			expected: "Ed Sheeran Shape of You",
		},
		{
			name:     "Track with ft.",
			artist:   "Ed Sheeran ft. Justin Bieber",
			title:    "I Don't Care",
			expected: "Ed Sheeran  Justin Bieber I Don't Care",
		},
		{
			name:     "Track with feat.",
			artist:   "Ariana Grande feat. Nicki Minaj",
			title:    "Side to Side",
			expected: "Ariana Grande  Nicki Minaj Side to Side",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			track := Track{
				Artist: tt.artist,
				Name:   tt.title,
			}
			
			// This would need to be a method on a service struct
			// For now, we'll test the logic directly
			query := buildSearchQuery(track)
			if query != tt.expected {
				t.Errorf("buildSearchQuery() = %v, want %v", query, tt.expected)
			}
		})
	}
}

// buildSearchQuery is a helper function for testing
func buildSearchQuery(track Track) string {
	query := track.Artist + " " + track.Name
	
	// Remove common words that might interfere with search
	query = strings.ReplaceAll(query, "ft.", "")
	query = strings.ReplaceAll(query, "feat.", "")
	query = strings.ReplaceAll(query, "featuring", "")
	
	// Remove extra spaces
	query = strings.Join(strings.Fields(query), " ")
	
	return query
}
