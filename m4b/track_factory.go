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

	const numWorkers = 5
	filesCh := make(chan string, numWorkers)
	tracksCh := make(chan Track, numWorkers)
	errorsCh := make(chan error, numWorkers)

	for i := 0; i < numWorkers; i++ {
		go t.worker(filesCh, tracksCh, errorsCh, metadataRules)
	}

	go func() {
		for _, file := range files {
			filesCh <- file
		}
		close(filesCh)
	}()

	for i := 0; i < len(files); i++ {
		select {
		case track := <-tracksCh:
			tracks = append(tracks, track)
		case err := <-errorsCh:
			return nil, fmt.Errorf("could not load track: %w", err)
		}
	}

	return tracks, nil
}

func (t *FFmpegTrackFactory) worker(in <-chan string, out chan<- Track, errors chan<- error, rules []MetadataRule) {
	for file := range in {
		track, err := t.LoadTrack(file, rules)
		if err != nil {
			errors <- err
			continue
		}
		out <- track
	}
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
