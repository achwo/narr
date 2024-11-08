package m4b

import (
	"strconv"
	"strings"
)

// Track represents an audio track with its file path and associated metadata.
// It contains the file path, a map of metadata tags, and the order in which
// the tags should be preserved.
type Track struct {
	File          string // Path to the audio file
	MetadataRules []MetadataRule
	metadataCache map[string]string // Map of metadata tags and their values
	tagOrder      []string          // Ordered list of metadata tag names
	rawMetadata   string
	title         string
	duration      float64
}

// DiscNumber returns the disc number from the track's metadata.
// It parses the "disc" tag which may be in the format "n" or "n/total".
// Returns the disc number and true if successfully parsed, or 0 and false
// if the tag is missing or invalid.
func (t *Track) DiscNumber() (int, bool) {
	disc, exists := t.MetadataTag("disc")
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
	track, exists := t.MetadataTag("track")
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

func (t *Track) Metadata() (map[string]string, []string, error) {
	if t.metadataCache == nil {
		metadata := t.rawMetadata
		tags, tagOrder := t.getMetadataTags(metadata)

		for _, rule := range t.MetadataRules {
			err := rule.Apply(tags)
			if err != nil {
				return nil, nil, err
			}
		}

		t.metadataCache = tags
		t.tagOrder = tagOrder
	}

	return t.metadataCache, t.tagOrder, nil
}

func (t *Track) MetadataTag(tag string) (string, bool) {
	metadata, _, err := t.Metadata()

	if err != nil {
		return "", false
	}

	value, exists := metadata[tag]
	return value, exists
}

func (t *Track) TitleAndDuration() (string, float64, error) {
	return t.title, t.duration, nil
}

func (t *Track) getMetadataTags(metadata string) (map[string]string, []string) {
	var tags = make(map[string]string)

	lines := strings.Split(metadata, "\n")[1:]
	tagOrder := make([]string, 0, len(lines))
	for _, line := range lines {
		split := strings.SplitN(line, "=", 2)
		if len(split) < 1 {
			continue
		}

		tagOrder = append(tagOrder, split[0])

		if len(split) == 1 {
			tags[split[0]] = ""
			continue
		}

		tags[split[0]] = split[1]
	}

	return tags, tagOrder
}
