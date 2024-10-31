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
