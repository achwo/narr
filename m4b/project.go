package m4b

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const configFileName = "narr.yaml"

type AudioFileProvider interface {
	AudioFiles(fullPath string) ([]string, error)
}

type MetadataProvider interface {
	ReadTitleAndDuration(file string) (string, float64, error)
	ReadMetadata(file string) (string, error)
}

type Project interface {
	AudioFiles() ([]string, error)
	ShowChapters() (string, error)
	ShowMetadata() (string, error)
}

type M4bProject struct {
	Config           ProjectConfig
	AudioProvider    AudioFileProvider
	MetadataProvider MetadataProvider
}

type Track struct {
	File     string
	Metadata map[string]string
	TagOrder []string
}

func (t *Track) DiscNumber() (int, bool) {
	disc, exists := t.Metadata["disc"]
	if !exists || disc == "" {
		return 0, false
	}

	parts := strings.Split(disc, "/")

	discNumber, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, false
	}
	return discNumber, true
}

func (t *Track) TrackNumber() (int, bool) {
	track, exists := t.Metadata["track"]
	if !exists || track == "" {
		return 0, false
	}

	parts := strings.Split(track, "/")

	trackNumber, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, false
	}
	return trackNumber, true
}

func (p *M4bProject) Tracks() ([]Track, error) {
	fullpath, err := p.Config.FullAudioFilePath()
	if err != nil {
		return nil, err
	}

	audioFiles, err := p.AudioProvider.AudioFiles(fullpath)
	if err != nil {
		return nil, err
	}

	tracks := make([]Track, 0, len(audioFiles))

	for _, file := range audioFiles {
		metadata, tagOrder, err := p.getUpdatedFileMetadata(file)
		if err != nil {
			return nil, fmt.Errorf("could not read metadata for '%s': %w", file, err)
		}
		track := Track{File: file, Metadata: metadata, TagOrder: tagOrder}
		tracks = append(tracks, track)

		if _, exists := track.DiscNumber(); !exists {
			return nil, fmt.Errorf("track '%s' has no disc number", track.File)
		}
	}

	slices.SortFunc(tracks, func(a, b Track) int {
		discI, discIExists := a.DiscNumber()
		discJ, discJExists := b.DiscNumber()

		if discIExists && discJExists && discI != discJ {
			return cmp.Compare(discI, discJ)
		}

		trackNumberI, trackNumberIExists := a.TrackNumber()
		trackNumberJ, trackNumberJExists := b.TrackNumber()

		if trackNumberIExists && trackNumberJExists && trackNumberI != trackNumberJ {
			return cmp.Compare(trackNumberI, trackNumberJ)
		}

		return strings.Compare(a.File, b.File)
	})

	return tracks, nil
}

func NewProject(
	config ProjectConfig,
	audioProvider AudioFileProvider,
	metadataProvider MetadataProvider,
) (*M4bProject, error) {
	err := config.Validate()
	if err != nil {
		return nil, err
	}

	if config.ProjectPath == "" {
		config.ProjectPath = "."
	}

	return &M4bProject{
		Config:           config,
		AudioProvider:    audioProvider,
		MetadataProvider: metadataProvider,
	}, nil
}

func NewProjectFromPath(
	path string,
	audioProvider AudioFileProvider,
	metadataProvider MetadataProvider,
) (*M4bProject, error) {
	var fullpath string

	if strings.HasSuffix(path, configFileName) {
		fullpath = path
	} else {
		fullpath = filepath.Join(path, configFileName)
	}

	bytes, err := os.ReadFile(fullpath)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", fullpath, err)
	}

	var config ProjectConfig
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal file %s: %w", fullpath, err)
	}

	config.ProjectPath = filepath.Dir(fullpath)

	return NewProject(config, audioProvider, metadataProvider)
}

func (p *M4bProject) ShowChapters() (string, error) {
	tracks, err := p.Tracks()
	if err != nil {
		return "", fmt.Errorf("could not load audio files: %w", err)
	}

	chapters := make(map[string]*Chapter)
	var chapterOrder []string
	var previousChapter *Chapter

	for i, track := range tracks {
		title, duration, err := p.MetadataProvider.ReadTitleAndDuration(track.File)
		if err != nil {
			return "", fmt.Errorf("could not read file data for file %s: %w", track, err)
		}
		chapterName := title

		for _, rule := range p.Config.ChapterRules {
			chapterName, err = rule.Apply(chapterName)
			if err != nil {
				return "", fmt.Errorf("chapter rule invalid: %w", err)
			}
		}

		value, exists := chapters[chapterName]

		newFile := File{Name: track.File, Duration: duration}
		if exists {
			value.addFile(newFile)
		} else {
			newChapter := Chapter{
				title:           chapterName,
				previousChapter: previousChapter,
				files:           []File{newFile},
				index:           i,
			}
			chapterOrder = append(chapterOrder, chapterName)
			chapters[chapterName] = &newChapter
			previousChapter = &newChapter
		}
	}

	markers := make([]string, 0, len(chapters))

	for _, name := range chapterOrder {
		chapter := chapters[name]
		markers = append(markers, chapter.ChapterMarker())
	}

	markersFileContent := strings.Join(markers, "\n\n")

	return markersFileContent, nil
}

func (p *M4bProject) ShowMetadata() (string, error) {
	tags, tagOrder, err := p.getUpdatedMetadata()
	if err != nil {
		return "", err
	}

	firstLine := ";FFMETADATA1"

	lines := []string{firstLine}

	for _, tag := range tagOrder {
		lines = append(lines, tag+"="+tags[tag])
	}

	return strings.Join(lines, "\n"), nil
}

func (p *M4bProject) ShowFilename() (string, error) {
	audioFiles, err := p.Tracks()
	if err != nil {
		return "", fmt.Errorf("could not load audio files: %w", err)
	}

	if len(audioFiles) == 0 {
		return "", errors.New("no audio files found")
	}

	tags, _, err := p.getUpdatedMetadata()

	if err != nil {
		return "", fmt.Errorf("could not get metadata: %w", err)
	}

	artist, exists := tags["artist"]
	if !exists {
		return "", errors.New("no artist found in metadata")
	}

	album, exists := tags["album"]
	if !exists {
		return "", errors.New("no album found in metadata")
	}

	filename := filepath.Join(p.Config.ProjectPath, artist, album+".m4b")

	return filename, nil
}

func (p *M4bProject) getUpdatedMetadata() (map[string]string, []string, error) {
	audioFiles, err := p.Tracks()
	if err != nil {
		return nil, nil, fmt.Errorf("could not load audio files: %w", err)
	}

	if len(audioFiles) == 0 {
		return nil, nil, errors.New("no audio files found")
	}

	return audioFiles[0].Metadata, audioFiles[0].TagOrder, nil
}

func (p *M4bProject) getUpdatedFileMetadata(file string) (map[string]string, []string, error) {
	metadata, err := p.MetadataProvider.ReadMetadata(file)
	if err != nil {
		return nil, nil, fmt.Errorf("could not read metadata: %w", err)
	}

	tags, tagOrder := p.getMetadataTags(metadata)

	for _, rule := range p.Config.MetadataRules {
		err = rule.Apply(tags)
		if err != nil {
			return nil, nil, err
		}
	}

	return tags, tagOrder, nil
}

func (m *M4bProject) getMetadataTags(metadata string) (map[string]string, []string) {
	var tags = make(map[string]string)

	lines := strings.Split(metadata, "\n")[1:]
	tagOrder := make([]string, 0, len(lines))
	for _, line := range lines {
		split := strings.SplitN(line, "=", 2)
		if len(split) < 1 {
			continue
		}

		tagOrder = append(tagOrder, split[0])

		if len(split) == 1 {
			tags[split[0]] = ""
			continue
		}

		tags[split[0]] = split[1]
	}

	return tags, tagOrder
}
