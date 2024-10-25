package m4b

import (
	"strconv"
	"strings"
)

type Track struct {
	File     string
	Metadata map[string]string
	TagOrder []string
}

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
