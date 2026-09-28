package m4b

import (
	"os"
	"path/filepath"
)

// FileData represents metadata about an audio file for testing
type FileData struct {
	Title    string  // Title of the audio file
	Duration float64 // Duration in seconds
	Metadata string  // Additional metadata
	HasCover bool    // Whether the file has an embedded cover image
}

// NullAudioProcessor implements a no-op audio processor that returns empty/nil values.
// This can be useful for testing or as a placeholder implementation.
type NullAudioProcessor struct {
	Data     map[string]FileData
	Written  map[string]string
	ErrTitle error
	ErrMeta  error
	ErrWrite error
	ErrCover error

	ConcatAndEncodeCalls int
	CopyToM4BCalls       int
}

// ConcatAndEncode writes an empty M4B file into outputPath and returns its path.
// It simulates decoding, joining and encoding multiple audio files at once.
func (p *NullAudioProcessor) ConcatAndEncode(files []string, outputPath string) (string, error) {
	p.ConcatAndEncodeCalls++
	return emptyM4B(outputPath)
}

// CopyToM4B writes an empty M4B file into outputPath and returns its path.
// It simulates copying the audio of a single file without encoding it.
func (p *NullAudioProcessor) CopyToM4B(file string, outputPath string) (string, error) {
	p.CopyToM4BCalls++
	return emptyM4B(outputPath)
}

func emptyM4B(outputPath string) (string, error) {
	file := filepath.Join(outputPath, "null.m4b")
	return file, os.WriteFile(file, nil, 0600)
}

// AddChapters is a no-op implementation that returns nil.
// It simulates adding chapter markers to an M4B file.
func (p *NullAudioProcessor) AddChapters(m4bFile string, chapters string) error {
	return nil
}

// ExtractCover is a no-op implementation that returns nil values.
func (p *NullAudioProcessor) ExtractCover(m4aFile string, workDir string) (string, error) {
	return "", nil
}

// HasCoverStream returns the preconfigured cover flag for a file
func (p *NullAudioProcessor) HasCoverStream(file string) (bool, error) {
	if p.ErrCover != nil {
		return false, p.ErrCover
	}

	return p.Data[file].HasCover, nil
}

// AddCover is a no-op implementation that returns nil values.
func (p *NullAudioProcessor) AddCover(m4bFile string, coverFile string) error {
	return nil
}

// AddMetadata is a no-op implementation that returns nil values.
func (p *NullAudioProcessor) AddMetadata(m4bFile string, metadata string, bookTitle string) error {
	return nil
}

// ReadTitleAndDuration returns the preconfigured title and duration for a file
func (p *NullAudioProcessor) ReadTitleAndDuration(file string) (string, float64, error) {
	if p.ErrTitle != nil {
		return "", 0.0, p.ErrTitle
	}

	data := p.Data[file]

	return data.Title, data.Duration, nil
}

// ReadDecodedDurations returns the preconfigured durations for the files
func (p *NullAudioProcessor) ReadDecodedDurations(files []string) ([]float64, error) {
	durations := make([]float64, 0, len(files))
	for _, file := range files {
		durations = append(durations, p.Data[file].Duration)
	}
	return durations, nil
}

// WriteMetadata records the metadata instead of writing it to the file.
func (p *NullAudioProcessor) WriteMetadata(file string, metadata string, verbose bool) error {
	if p.ErrWrite != nil {
		return p.ErrWrite
	}

	if p.Written == nil {
		p.Written = make(map[string]string)
	}

	p.Written[file] = metadata
	return nil
}

// ReadMetadata returns the preconfigured metadata for a file
func (p *NullAudioProcessor) ReadMetadata(file string) (string, error) {
	if p.ErrMeta != nil {
		return "", p.ErrMeta
	}
	return p.Data[file].Metadata, nil
}
