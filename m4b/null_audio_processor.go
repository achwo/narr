package m4b

// NullAudioProcessor implements a no-op audio processor that returns empty/nil values.
// This can be useful for testing or as a placeholder implementation.
type NullAudioProcessor struct {
}

// ToM4A is a no-op implementation that returns nil values.
// It simulates converting audio files to M4A format.
func (p *NullAudioProcessor) ToM4A(files []string, outputPath string) ([]string, error) {
	return nil, nil
}

// Concat is a no-op implementation that returns empty values.
// It simulates concatenating multiple audio files into a single file.
func (p *NullAudioProcessor) Concat(files []string, filelistPath string, outputPath string) (string, error) {
	return "", nil
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

// AddCover is a no-op implementation that returns nil values.
func (p *NullAudioProcessor) AddCover(m4bFile string, coverFile string) error {
	return nil
}

// AddMetadata is a no-op implementation that returns nil values.
func (p *NullAudioProcessor) AddMetadata(m4bFile string, metadata string, bookTitle string) error {
	return nil
}
