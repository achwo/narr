package testutils

type FakeAudioFileProvider struct {
	Files []string
	Err   error
}

func (f *FakeAudioFileProvider) AudioFiles(fullPath string) ([]string, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	return f.Files, nil
}

type FileData struct {
	Title    string
	Duration float64
	Metadata string
}

type FakeMetadataProvider struct {
	Data     map[string]FileData
	ErrTitle error
	ErrMeta  error
}

func (f *FakeMetadataProvider) ReadTitleAndDuration(file string) (string, float64, error) {
	if f.ErrTitle != nil {
		return "", 0.0, f.ErrTitle
	}

	data := f.Data[file]

	return data.Title, data.Duration, nil
}

func (f *FakeMetadataProvider) ReadMetadata(file string) (string, error) {
	if f.ErrMeta != nil {
		return "", f.ErrMeta
	}
	return f.Data[file].Metadata, nil
}

type FakeAudioProcessor struct{}

// AddChapters implements m4b.AudioProcessor.
func (c *FakeAudioProcessor) AddChapters(m4bFile string, chapters string) error {
	panic("unimplemented")
}

// AddCover implements m4b.AudioProcessor.
func (c *FakeAudioProcessor) AddCover(m4bFile string, coverFile string) error {
	panic("unimplemented")
}

// AddMetadata implements m4b.AudioProcessor.
func (c *FakeAudioProcessor) AddMetadata(m4bFile string, metadata string) error {
	panic("unimplemented")
}

// Concat implements m4b.AudioProcessor.
func (c *FakeAudioProcessor) Concat(files []string, output string) error {
	return nil
}

func (c *FakeAudioProcessor) ToM4A(files []string) ([]string, error) {
	return nil, nil
}
