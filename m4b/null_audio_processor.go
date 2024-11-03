package m4b

type NullAudioProcessor struct {
}

func (p *NullAudioProcessor) ToM4A(files []string, outputPath string) ([]string, error) {
	return nil, nil
}

func (p *NullAudioProcessor) Concat(files []string, filelistPath string, outputPath string) (string, error) {
	return "", nil
}

func (p *NullAudioProcessor) AddChapters(m4bFile string, chapters string) error {
	return nil
}

func (p *NullAudioProcessor) ExtractCover(m4aFile string) (string, error) {
	return "", nil
}

func (p *NullAudioProcessor) AddCover(m4bFile string, coverFile string) error {
	return nil
}

func (p *NullAudioProcessor) AddMetadata(m4bFile string, metadata string, bookTitle string) error {
	return nil
}
