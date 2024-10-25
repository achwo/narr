package m4b

type NullAudioProcessor struct {
}

// AddChapters implements AudioProcessor.
func (f *NullAudioProcessor) AddChapters(m4bFile string, chapters string) error {
	return nil
}

// AddCover implements AudioProcessor.
func (f *NullAudioProcessor) AddCover(m4bFile string, coverFile string) error {
	return nil
}

// AddMetadata implements AudioProcessor.
func (f *NullAudioProcessor) AddMetadata(m4bFile string, metadata string) error {
	return nil
}

func (f *NullAudioProcessor) Concat(files []string, output string) error {
	return nil
}

func (c *NullAudioProcessor) ToM4A(files []string) ([]string, error) {
	return nil, nil
}
