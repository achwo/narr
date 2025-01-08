package m4b_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/achwo/narr/m4b"
	"github.com/stretchr/testify/require"
)

func TestShowChapters(t *testing.T) {
	config := m4b.ProjectConfig{ChapterRules: []m4b.ChapterRule{}}
	deps := setupDeps()
	project, err := m4b.NewProjectWithDeps(config, *deps)
	require.NoError(t, err)

	chapters, err := project.Chapters()
	require.NoError(t, err)

	require.Equal(t, "CHAPTER0=00:00:00.000\nCHAPTER0NAME=Chapter 1\n\nCHAPTER1=01:23:20.000\nCHAPTER1NAME=Chapter 2", chapters)
}

func TestMetadata(t *testing.T) {
	metadataRule := m4b.MetadataRule{
		Type:   "regex",
		Tag:    "title",
		Regex:  "^Chapter (\\d+)-\\d+: (.+)$",
		Format: "%s - %s",
	}
	config := m4b.ProjectConfig{ChapterRules: []m4b.ChapterRule{}}
	config.MetadataRules = append(config.MetadataRules, metadataRule)

	project, err := m4b.NewProjectWithDeps(config, *setupDeps())
	require.NoError(t, err)

	metadata, err := project.Metadata()
	require.NoError(t, err)

	require.Equal(
		t,
		`;FFMETADATA1
title=01 - Star dust
artist=Hans Wurst/ read by George Washington
album=The Book?
date=2002-09-16`,
		metadata,
	)
}

func TestFilename(t *testing.T) {
	config := m4b.ProjectConfig{ChapterRules: []m4b.ChapterRule{}}
	deps := setupDeps()
	project, err := m4b.NewProjectWithDeps(config, *deps)
	require.NoError(t, err)

	filename, err := project.Filename()
	require.NoError(t, err)

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	require.Equal(t, filepath.Join(home, "narr", "Hans Wurst_ read by George Washington/The Book_/The Book_.m4b"), filename)
}

func TestTracks(t *testing.T) {
	config := m4b.ProjectConfig{ChapterRules: []m4b.ChapterRule{}}
	data := make(map[string]m4b.FileData)

	data["file1.m4a"] = m4b.FileData{
		Title:    "Chapter 1",
		Duration: 5000,
		Metadata: `;FFMETADATA1
title=Chapter 01-02: Star dust
artist=Hans Wurst read by George Washington
album=The Book
track=3/16
disc=1/10
date=2002-09-16`,
	}

	data["file2.m4a"] = m4b.FileData{
		Title:    "Chapter 1",
		Duration: 5000,
		Metadata: `;FFMETADATA1
title=Chapter 01-02: Star dust
artist=Hans Wurst read by George Washington
album=The Book
track=2/16
disc=1/10
date=2002-09-16`,
	}

	data["file3.m4a"] = m4b.FileData{
		Title:    "Chapter 2",
		Duration: 5000,
		Metadata: `;FFMETADATA1
title=Chapter 02-01: Star dust
artist=Hans Wurst read by George Washington
album=The Book
track=1/16
disc=2/10
date=2002-09-16`,
	}

	fakeAudioProcessor := &m4b.NullAudioProcessor{Data: data}

	deps := m4b.ProjectDependencies{
		AudioFileProvider: &FakeAudioFileProvider{
			Files: []string{"file1.m4a", "file2.m4a", "file3.m4a"},
		},
		AudioProcessor: fakeAudioProcessor,
		TrackFactory:   &m4b.FFmpegTrackFactory{AudioProcessor: fakeAudioProcessor},
	}

	project, err := m4b.NewProjectWithDeps(config, deps)
	require.NoError(t, err)

	files, err := project.Tracks()
	require.NoError(t, err)

	require.Equal(t, "file2.m4a", files[0].File)
	require.Equal(t, "file1.m4a", files[1].File)
	require.Equal(t, "file3.m4a", files[2].File)

}

func setupDeps() *m4b.ProjectDependencies {
	data := make(map[string]m4b.FileData)

	data["file1.m4a"] = m4b.FileData{
		Title:    "Chapter 1",
		Duration: 5000,
		Metadata: `;FFMETADATA1
title=Chapter 01-02: Star dust
artist=Hans Wurst/ read by George Washington
album=The Book?
track=1/16
disc=1/10
date=2002-09-16`,
	}
	data["file2.m4a"] = m4b.FileData{
		Title:    "Chapter 1",
		Duration: 5000,
		Metadata: `;FFMETADATA1
title=Chapter 01-02: Star dust
artist=Hans Wurst/ read by George Washington
album=The Book?
track=2/16
disc=2/10
date=2002-09-16`,
	}

	data["file2.m4a"] = m4b.FileData{
		Title:    "Chapter 2",
		Duration: 5000,
		Metadata: `;FFMETADATA1
title=Chapter 02-02: Wurst
artist=Hans Wurst/ read by George Washington
album=The Book?
track=3/16
disc=2/10
date=2002-09-16`,
	}

	fakeAudioProcessor := &m4b.NullAudioProcessor{Data: data}
	trackFactory := &m4b.FFmpegTrackFactory{AudioProcessor: fakeAudioProcessor}
	fakeAudioProvider := &FakeAudioFileProvider{
		Files: []string{"file1.m4a", "file2.m4a"},
	}

	return &m4b.ProjectDependencies{
		AudioFileProvider: fakeAudioProvider,
		AudioProcessor:    fakeAudioProcessor,
		TrackFactory:      trackFactory,
	}
}

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
