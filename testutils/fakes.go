package testutils

// FakeAudioFileProvider implements a test double for providing audio files
type FakeAudioFileProvider struct {
	Files []string // Files to return from AudioFiles
	Err   error    // Error to return, if any
}

// AudioFiles returns the preconfigured Files slice and Err value
func (f *FakeAudioFileProvider) AudioFiles(fullPath string) ([]string, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	return f.Files, nil
}

// FileData represents metadata about an audio file for testing
type FileData struct {
	Title    string  // Title of the audio file
	Duration float64 // Duration in seconds
	Metadata string  // Additional metadata
}

// FakeMetadataProvider implements a test double for providing audio metadata
type FakeMetadataProvider struct {
	Data     map[string]FileData
	ErrTitle error
	ErrMeta  error
}

// ReadTitleAndDuration returns the preconfigured title and duration for a file
func (f *FakeMetadataProvider) ReadTitleAndDuration(file string) (string, float64, error) {
	if f.ErrTitle != nil {
		return "", 0.0, f.ErrTitle
	}

	data := f.Data[file]

	return data.Title, data.Duration, nil
}

// ReadMetadata returns the preconfigured metadata for a file
func (f *FakeMetadataProvider) ReadMetadata(file string) (string, error) {
	if f.ErrMeta != nil {
		return "", f.ErrMeta
	}
	return f.Data[file].Metadata, nil
}
