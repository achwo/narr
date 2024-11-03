package m4b

import (
	"strconv"
	"strings"
)

// Track represents an audio track with its file path and associated metadata.
// It contains the file path, a map of metadata tags, and the order in which
// the tags should be preserved.
type Track struct {
	File     string            // Path to the audio file
	Metadata map[string]string // Map of metadata tags and their values
	TagOrder []string          // Ordered list of metadata tag names
}

// DiscNumber returns the disc number from the track's metadata.
// It parses the "disc" tag which may be in the format "n" or "n/total".
// Returns the disc number and true if successfully parsed, or 0 and false
// if the tag is missing or invalid.
func (t *Track) DiscNumber() (int, bool) {
	disc, exists := t.Metadata["disc"]
	if !exists || disc == "" {
		return 0, false
	}

	parts := strings.Split(disc, "/")

	discNumber, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, false
	}
	return discNumber, true
}

// TrackNumber returns the track number from the track's metadata.
// It parses the "track" tag which may be in the format "n" or "n/total".
// Returns the track number and true if successfully parsed, or 0 and false
// if the tag is missing or invalid.
func (t *Track) TrackNumber() (int, bool) {
	track, exists := t.Metadata["track"]
	if !exists || track == "" {
		return 0, false
	}

	parts := strings.Split(track, "/")

	trackNumber, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, false
	}
	return trackNumber, true
}
