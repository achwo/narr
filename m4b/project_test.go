package m4b_test

import (
	"testing"

	"github.com/achwo/narr/m4b"
	"github.com/achwo/narr/testutils"
	"github.com/stretchr/testify/require"
)

func TestShowChapters(t *testing.T) {
	project, err := setupProject()
	if err != nil {
		t.Fatal(err)
	}

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
	project, err := setupProject()
	project.Config.MetadataRules = append(project.Config.MetadataRules, metadataRule)
	require.NoError(t, err)

	metadata, err := project.Metadata()
	require.NoError(t, err)

	require.Equal(
		t,
		`;FFMETADATA1
title=01 - Star dust
artist=Hans Wurst read by George Washington
album=The Book
date=2002-09-16`,
		metadata,
	)
}

func TestFilename(t *testing.T) {
	project, err := setupProject()
	if err != nil {
		t.Fatal(err)
	}

	filename, err := project.Filename()
	require.NoError(t, err)

	require.Equal(t, "Hans Wurst read by George Washington/The Book.m4b", filename)
}

func TestTracks(t *testing.T) {
	// should return all files within the project folder sorted by cd and track numbers
	project, err := setupProject()

	fakeAudioProvider := &testutils.FakeAudioFileProvider{
		Files: []string{"file1.m4a", "file2.m4a", "file3.m4a"},
	}

	data := make(map[string]testutils.FileData)

	data["file1.m4a"] = testutils.FileData{
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

	data["file2.m4a"] = testutils.FileData{
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

	data["file3.m4a"] = testutils.FileData{
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

	fakeMetadataProvider := &testutils.FakeMetadataProvider{Data: data}

	project.MetadataProvider = fakeMetadataProvider
	project.AudioFileProvider = fakeAudioProvider

	if err != nil {
		t.Fatal(err)
	}

	files, err := project.Tracks()
	require.NoError(t, err)

	require.Equal(t, "file2.m4a", files[0].File)
	require.Equal(t, "file1.m4a", files[1].File)
	require.Equal(t, "file3.m4a", files[2].File)

}

func setupProject() (*m4b.Project, error) {
	fakeAudioProvider := &testutils.FakeAudioFileProvider{
		Files: []string{"file1.m4a", "file2.m4a"},
	}

	data := make(map[string]testutils.FileData)

	data["file1.m4a"] = testutils.FileData{
		Title:    "Chapter 1",
		Duration: 5000,
		Metadata: `;FFMETADATA1
title=Chapter 01-02: Star dust
artist=Hans Wurst read by George Washington
album=The Book
track=1/16
disc=1/10
date=2002-09-16`,
	}
	data["file2.m4a"] = testutils.FileData{
		Title:    "Chapter 1",
		Duration: 5000,
		Metadata: `;FFMETADATA1
title=Chapter 01-02: Star dust
artist=Hans Wurst read by George Washington
album=The Book
track=2/16
disc=2/10
date=2002-09-16`,
	}

	data["file2.m4a"] = testutils.FileData{
		Title:    "Chapter 2",
		Duration: 5000,
		Metadata: `;FFMETADATA1
title=Chapter 02-02: Wurst
artist=Hans Wurst read by George Washington
album=The Book
track=3/16
disc=2/10
date=2002-09-16`,
	}

	fakeMetadataProvider := &testutils.FakeMetadataProvider{Data: data}
	fakeAudioConverter := &m4b.NullAudioProcessor{}

	config := m4b.ProjectConfig{ChapterRules: []m4b.ChapterRule{}}

	return m4b.NewProject(config, fakeAudioProvider, fakeMetadataProvider, fakeAudioConverter)
}
