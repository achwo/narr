package m4b

type NullAudioProcessor struct {
}

func (c *NullAudioProcessor) ToM4A(files []string, outputPath string) ([]string, error) {
	return nil, nil
}

func (f *NullAudioProcessor) Concat(files []string, output string) error {
	return nil
}

func (f *NullAudioProcessor) AddChapters(m4bFile string, chapters string) error {
	return nil
}

func (f *NullAudioProcessor) AddCover(m4bFile string, coverFile string) error {
	return nil
}

func (f *NullAudioProcessor) AddMetadata(m4bFile string, metadata string) error {
	return nil
}
