package m4b_test

import (
	"fmt"
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

func TestMetadata_FreeformTagInSourceFile_KeepsItsSpelling(t *testing.T) {
	config := m4b.ProjectConfig{
		ChapterRules:  []m4b.ChapterRule{},
		MetadataRules: []m4b.MetadataRule{{Type: "set", Tag: "AUDIBLE_ASIN", Value: "B01ABCDEFG"}},
	}
	deps := depsForMetadata(`;FFMETADATA1
title=Star dust
artist=Hans Wurst
album=The Book
SERIES=Die Reihe`)

	project, err := m4b.NewProjectWithDeps(config, *deps)
	require.NoError(t, err)

	metadata, err := project.Metadata()
	require.NoError(t, err)

	require.Equal(
		t,
		`;FFMETADATA1
title=Star dust
artist=Hans Wurst
album=The Book
SERIES=Die Reihe
AUDIBLE_ASIN=B01ABCDEFG`,
		metadata,
	)
}

func TestMetadata_RuleMatchesTagCaseInsensitively_UpdatesTheValue(t *testing.T) {
	config := m4b.ProjectConfig{
		ChapterRules:  []m4b.ChapterRule{},
		MetadataRules: []m4b.MetadataRule{{Type: "set", Tag: "series", Value: "Andere Reihe"}},
	}
	deps := depsForMetadata(`;FFMETADATA1
album=The Book
SERIES=Die Reihe`)

	project, err := m4b.NewProjectWithDeps(config, *deps)
	require.NoError(t, err)

	metadata, err := project.Metadata()
	require.NoError(t, err)

	require.Equal(t, ";FFMETADATA1\nalbum=The Book\nSERIES=Andere Reihe", metadata)
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

func TestFilename_AlbumArtistAndArtist_UsesAlbumArtist(t *testing.T) {
	filename := filenameForMetadata(t, `;FFMETADATA1
title=Chapter 01-02: Star dust
album_artist=Testautor
artist=Test Sprecher
album=The Book
track=1/16
disc=1/10
date=2002-09-16`)

	require.Equal(t, filepath.Join(homeDir(t), "narr", "Testautor", "The Book", "The Book.m4b"), filename)
}

func TestFilename_OnlyArtist_UsesArtist(t *testing.T) {
	filename := filenameForMetadata(t, `;FFMETADATA1
title=Chapter 01-02: Star dust
artist=Test Sprecher
album=The Book
track=1/16
disc=1/10
date=2002-09-16`)

	require.Equal(t, filepath.Join(homeDir(t), "narr", "Test Sprecher", "The Book", "The Book.m4b"), filename)
}

func TestFilename_EmptyAlbumArtist_UsesArtist(t *testing.T) {
	filename := filenameForMetadata(t, `;FFMETADATA1
title=Chapter 01-02: Star dust
album_artist=
artist=Test Sprecher
album=The Book
track=1/16
disc=1/10
date=2002-09-16`)

	require.Equal(t, filepath.Join(homeDir(t), "narr", "Test Sprecher", "The Book", "The Book.m4b"), filename)
}

func TestFilename_MixedCaseAlbumArtistTag_UsesAlbumArtist(t *testing.T) {
	filename := filenameForMetadata(t, `;FFMETADATA1
title=Chapter 01-02: Star dust
ALBUM_ARTIST=Testautor
artist=Test Sprecher
album=The Book
track=1/16
disc=1/10
date=2002-09-16`)

	require.Equal(t, filepath.Join(homeDir(t), "narr", "Testautor", "The Book", "The Book.m4b"), filename)
}

func TestFilename_NoArtistTags_ReturnsError(t *testing.T) {
	project, err := m4b.NewProjectWithDeps(
		m4b.ProjectConfig{ChapterRules: []m4b.ChapterRule{}},
		*depsForMetadata(`;FFMETADATA1
title=Chapter 01-02: Star dust
album=The Book
track=1/16
disc=1/10
date=2002-09-16`),
	)
	require.NoError(t, err)

	_, err = project.Filename()
	require.Error(t, err)
}

func homeDir(t *testing.T) string {
	t.Helper()
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	return home
}

func filenameForMetadata(t *testing.T, metadata string) string {
	t.Helper()

	config := m4b.ProjectConfig{ChapterRules: []m4b.ChapterRule{}}
	project, err := m4b.NewProjectWithDeps(config, *depsForMetadata(metadata))
	require.NoError(t, err)

	filename, err := project.Filename()
	require.NoError(t, err)

	return filename
}

func depsForMetadata(metadata string) *m4b.ProjectDependencies {
	data := map[string]m4b.FileData{
		"file1.m4a": {Title: "Chapter 1", Duration: 5000, Metadata: metadata},
	}

	fakeAudioProcessor := &m4b.NullAudioProcessor{Data: data}

	return &m4b.ProjectDependencies{
		AudioFileProvider: &FakeAudioFileProvider{Files: []string{"file1.m4a"}},
		AudioProcessor:    fakeAudioProcessor,
		TrackFactory:      &m4b.FFmpegTrackFactory{AudioProcessor: fakeAudioProcessor},
	}
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

func TestCheckCover_NoCoverPathAndNoEmbeddedCover_ReturnsError(t *testing.T) {
	deps, _ := depsWithCover(false)
	config := m4b.ProjectConfig{ProjectPath: "/books/the-book"}
	project, err := m4b.NewProjectWithDeps(config, *deps)
	require.NoError(t, err)

	err = project.CheckCover()

	require.EqualError(
		t,
		err,
		"no cover for project /books/the-book: the first audio file file1.m4a has no embedded cover image "+
			"and coverPath is empty in narr.yaml. Set coverPath to an image file, or embed a cover into the audio files",
	)
}

func TestCheckCover_NoCoverPathAndEmbeddedCover_ReturnsNoError(t *testing.T) {
	deps, _ := depsWithCover(true)
	project, err := m4b.NewProjectWithDeps(m4b.ProjectConfig{}, *deps)
	require.NoError(t, err)

	require.NoError(t, project.CheckCover())
}

func TestCheckCover_CoverPathDoesNotExist_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	deps, _ := depsWithCover(true)
	config := m4b.ProjectConfig{ProjectPath: dir, CoverPath: "cover.jpg"}
	project, err := m4b.NewProjectWithDeps(config, *deps)
	require.NoError(t, err)

	err = project.CheckCover()

	require.EqualError(
		t,
		err,
		"no cover for project "+dir+": coverPath in narr.yaml points to "+filepath.Join(dir, "cover.jpg")+
			", which does not exist. Set coverPath to an existing image file, or leave it empty to use the cover "+
			"embedded in the audio files",
	)
}

func TestCheckCover_CoverPathIsDirectory_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	deps, _ := depsWithCover(true)
	config := m4b.ProjectConfig{ProjectPath: dir, CoverPath: "."}
	project, err := m4b.NewProjectWithDeps(config, *deps)
	require.NoError(t, err)

	err = project.CheckCover()

	require.ErrorContains(t, err, "which is a directory")
}

func TestCheckCover_CoverPathExists_ReturnsNoError(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "cover.jpg"), []byte("jpeg"), 0600))

	deps, _ := depsWithCover(false)
	config := m4b.ProjectConfig{ProjectPath: dir, CoverPath: "cover.jpg"}
	project, err := m4b.NewProjectWithDeps(config, *deps)
	require.NoError(t, err)

	require.NoError(t, project.CheckCover())
}

func TestCover_CoverPathExists_ReturnsConfiguredPath(t *testing.T) {
	dir := t.TempDir()
	coverPath := filepath.Join(dir, "cover.jpg")
	require.NoError(t, os.WriteFile(coverPath, []byte("jpeg"), 0600))

	deps, _ := depsWithCover(false)
	config := m4b.ProjectConfig{ProjectPath: dir, CoverPath: "cover.jpg"}
	project, err := m4b.NewProjectWithDeps(config, *deps)
	require.NoError(t, err)

	cover, err := project.Cover()
	require.NoError(t, err)

	require.Equal(t, coverPath, cover)
}

func TestCover_NoCoverAvailable_ReturnsError(t *testing.T) {
	deps, _ := depsWithCover(false)
	project, err := m4b.NewProjectWithDeps(m4b.ProjectConfig{ProjectPath: "/books/the-book"}, *deps)
	require.NoError(t, err)

	_, err = project.Cover()

	require.ErrorContains(t, err, "no cover for project /books/the-book")
}

func TestConvertToM4B_NoCoverAvailable_FailsBeforeConversion(t *testing.T) {
	deps, processor := depsWithCover(false)
	config := m4b.ProjectConfig{ProjectPath: t.TempDir(), ShouldConvert: true}
	project, err := m4b.NewProjectWithDeps(config, *deps)
	require.NoError(t, err)

	_, err = project.ConvertToM4B()

	require.ErrorContains(t, err, "has no embedded cover image")
	require.Zero(t, processor.ConcatAndEncodeCalls)
}

func TestConvertToM4B_TwoFilesWithoutConversion_EncodesThemInOnePass(t *testing.T) {
	processor := convert(t, []string{"file1.m4a", "file2.m4a"}, false)

	require.Equal(t, 1, processor.ConcatAndEncodeCalls)
	require.Zero(t, processor.CopyToM4BCalls)
}

func TestConvertToM4B_OneFileWithoutConversion_CopiesIt(t *testing.T) {
	processor := convert(t, []string{"file1.m4a"}, false)

	require.Equal(t, 1, processor.CopyToM4BCalls)
	require.Zero(t, processor.ConcatAndEncodeCalls)
}

func TestConvertToM4B_OneFileWithConversion_EncodesIt(t *testing.T) {
	processor := convert(t, []string{"file1.m4a"}, true)

	require.Equal(t, 1, processor.ConcatAndEncodeCalls)
	require.Zero(t, processor.CopyToM4BCalls)
}

func convert(t *testing.T, files []string, shouldConvert bool) *m4b.NullAudioProcessor {
	t.Helper()
	t.Setenv("HOME", t.TempDir())

	data := make(map[string]m4b.FileData)
	for i, file := range files {
		data[file] = m4b.FileData{
			Title:    "Chapter 1",
			Duration: 5000,
			HasCover: true,
			Metadata: fmt.Sprintf(";FFMETADATA1\ntitle=Chapter 1\nartist=Hans Wurst\nalbum=The Book\ntrack=%d/16", i+1),
		}
	}

	processor := &m4b.NullAudioProcessor{Data: data}
	deps := m4b.ProjectDependencies{
		AudioFileProvider: &FakeAudioFileProvider{Files: files},
		AudioProcessor:    processor,
		TrackFactory:      &m4b.FFmpegTrackFactory{AudioProcessor: processor},
	}

	config := m4b.ProjectConfig{ProjectPath: t.TempDir(), ShouldConvert: shouldConvert}
	project, err := m4b.NewProjectWithDeps(config, deps)
	require.NoError(t, err)

	_, err = project.ConvertToM4B()
	require.NoError(t, err)

	return processor
}

func depsWithCover(hasCover bool) (*m4b.ProjectDependencies, *m4b.NullAudioProcessor) {
	data := map[string]m4b.FileData{
		"file1.m4a": {
			Title:    "Chapter 1",
			Duration: 5000,
			HasCover: hasCover,
			Metadata: `;FFMETADATA1
title=Chapter 1
artist=Hans Wurst
album=The Book
track=1/16`,
		},
	}

	processor := &m4b.NullAudioProcessor{Data: data}

	return &m4b.ProjectDependencies{
		AudioFileProvider: &FakeAudioFileProvider{Files: []string{"file1.m4a"}},
		AudioProcessor:    processor,
		TrackFactory:      &m4b.FFmpegTrackFactory{AudioProcessor: processor},
	}, processor
}
