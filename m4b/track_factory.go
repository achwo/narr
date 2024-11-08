package m4b

import "fmt"

type FFmpegTrackFactory struct {
	AudioProcessor AudioProcessor
}

func (t *FFmpegTrackFactory) LoadTracks(
	files []string,
	metadataRules []MetadataRule,
) ([]Track, error) {
	tracks := make([]Track, 0, len(files))

	for _, file := range files {
		track, err := t.LoadTrack(file, metadataRules)
		if err != nil {
			return nil, fmt.Errorf("could not load track %s: %w", file, err)
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

func (t *FFmpegTrackFactory) LoadTrack(
	file string,
	metadataRules []MetadataRule,
) (Track, error) {
	metadata, err := t.AudioProcessor.ReadMetadata(file)
	if err != nil {
		return Track{}, err
	}
	title, duration, err := t.AudioProcessor.ReadTitleAndDuration(file)
	if err != nil {
		return Track{}, err
	}

	return Track{
		File:          file,
		rawMetadata:   metadata,
		title:         title,
		duration:      duration,
		MetadataRules: metadataRules,
	}, nil
}
