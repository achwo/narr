package m4b

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

const configFileName = "narr.yaml"

func NewProjectFromPath(
	path string,
	audioProvider AudioFileProvider,
	metadataProvider MetadataProvider,
	audioConverter AudioProcessor,
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

	return NewProject(config, audioProvider, metadataProvider, audioConverter)
}

func NewProject(
	config ProjectConfig,
	audioProvider AudioFileProvider,
	metadataProvider MetadataProvider,
	audioConverter AudioProcessor,
) (*M4bProject, error) {
	err := config.Validate()
	if err != nil {
		return nil, err
	}

	if config.ProjectPath == "" {
		config.ProjectPath = "."
	}

	return &M4bProject{
		Config:            config,
		AudioFileProvider: audioProvider,
		MetadataProvider:  metadataProvider,
		AudioProcessor:    audioConverter,
	}, nil
}

type AudioFileProvider interface {
	AudioFiles(fullPath string) ([]string, error)
}

type AudioProcessor interface {
	ToM4A(files []string, outputPath string) ([]string, error)
	Concat(m4aFiles []string, templateFilePath string, outputPath string) (string, error)
	AddMetadata(m4bFile string, metadata string, bookTitle string) error
	AddCover(m4bFile string, coverFile string) error
	ExtractCover(m4aFile string) (string, error)
	AddChapters(m4bFile string, chapters string) error
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
	Config            ProjectConfig
	AudioFileProvider AudioFileProvider
	MetadataProvider  MetadataProvider
	AudioProcessor    AudioProcessor
	tracks            []Track
	workDir           string
}

func (p *M4bProject) ConvertToM4B() (string, error) {
	p.workDir = filepath.Join(p.Config.ProjectPath, "temp")
	_ = os.Mkdir(p.workDir, 0755)
	defer os.RemoveAll(p.workDir)

	tracks, err := p.Tracks()
	if err != nil {
		return "", fmt.Errorf("could not load audio files: %w", err)
	}

	files := make([]string, 0, len(tracks))

	for _, track := range tracks {
		files = append(files, track.File)
	}

	fmt.Printf("Converting %d files to m4a\n", len(files))
	m4aPath, err := p.m4aPath()
	if err != nil {
		return "", fmt.Errorf("could not create m4a path: %w", err)
	}

	m4aFiles, err := p.AudioProcessor.ToM4A(files, m4aPath)
	if err != nil {
		return "", fmt.Errorf("could not convert files to m4a: %w", err)
	}

	fmt.Println("Concating files")
	m4bFile, err := p.AudioProcessor.Concat(m4aFiles, p.filelistFile(), p.workDir)
	if err != nil {
		return "", fmt.Errorf("could not concat files: %w", err)
	}

	fmt.Println("Adding metadata to m4b")
	metadata, err := p.Metadata()
	if err != nil {
		return "", fmt.Errorf("could not get metadata: %w", err)
	}

	_, bookTitle, err := p.ArtistAndBookTitle()
	if err != nil {
		return "", fmt.Errorf("could not read book title: %w", err)
	}

	if err = p.AudioProcessor.AddMetadata(m4bFile, metadata, bookTitle); err != nil {
		return "", fmt.Errorf("could not add metadata to %s: %w", m4bFile, err)
	}

	cover, err := p.Cover()
	if err != nil {
		return "", fmt.Errorf("could not get cover: %w", err)
	}

	fmt.Println("Adding cover to m4b")
	if err = p.AudioProcessor.AddCover(m4bFile, cover); err != nil {
		return "", fmt.Errorf("could not add cover to %s: %w", m4bFile, err)
	}

	fmt.Println("Adding chapters to m4b")
	chapters, err := p.Chapters()
	if err != nil {
		return "", fmt.Errorf("could not get chapters: %w", err)
	}

	if err = p.AudioProcessor.AddChapters(m4bFile, chapters); err != nil {
		return "", fmt.Errorf("could not add chapters to %s: %w", m4bFile, err)
	}

	finalFilename, err := p.Filename()
	if err != nil {
		return "", fmt.Errorf("could not get filename: %w", err)
	}

	if err = os.MkdirAll(filepath.Dir(finalFilename), 0755); err != nil {
		return "", fmt.Errorf("could not create target dir: %w", err)
	}

	if err = os.Rename(m4bFile, finalFilename); err != nil {
		return "", fmt.Errorf("could not rename file: %w", err)
	}

	return m4bFile, nil
}

func (p *M4bProject) Cover() (string, error) {
	coverFromConfig := p.Config.CoverPath

	if _, err := os.Stat(coverFromConfig); !errors.Is(err, os.ErrNotExist) {
		return coverFromConfig, nil
	}

	tracks, err := p.Tracks()
	if err != nil {
		return "", err
	}
	firstFile := tracks[0].File
	return p.AudioProcessor.ExtractCover(firstFile)
}

func (p *M4bProject) Tracks() ([]Track, error) {
	if p.tracks != nil {
		return p.tracks, nil
	}

	fullpath, err := p.Config.FullAudioFilePath()
	if err != nil {
		return nil, err
	}

	audioFiles, err := p.AudioFileProvider.AudioFiles(fullpath)
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
	p.tracks = tracks
	return tracks, nil
}

func (p *M4bProject) Chapters() (string, error) {
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

func (p *M4bProject) Metadata() (string, error) {
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

func (p *M4bProject) Filename() (string, error) {
	artist, album, err := p.ArtistAndBookTitle()
	if err != nil {
		return "", err
	}
	filename := filepath.Join(p.Config.ProjectPath, artist, album+".m4b")

	return filename, nil
}

func (p *M4bProject) ArtistAndBookTitle() (string, string, error) {
	audioFiles, err := p.Tracks()
	if err != nil {
		return "", "", fmt.Errorf("could not load audio files: %w", err)
	}

	if len(audioFiles) == 0 {
		return "", "", errors.New("no audio files found")
	}

	tags, _, err := p.getUpdatedMetadata()

	if err != nil {
		return "", "", fmt.Errorf("could not get metadata: %w", err)
	}

	artist, exists := tags["artist"]
	if !exists {
		return "", "", errors.New("no artist found in metadata")
	}

	album, exists := tags["album"]
	if !exists {
		return "", "", errors.New("no album found in metadata")
	}

	return artist, album, nil
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

func (p *M4bProject) m4aPath() (string, error) {
	m4aPath := filepath.Join(p.workDir, "m4a")

	if _, err := os.Stat(m4aPath); !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("m4a directory already exists: %s", m4aPath)
	}

	err := os.Mkdir(m4aPath, 0755)
	if err != nil {
		return "", err
	}
	return m4aPath, nil
}

func (p *M4bProject) filelistFile() string {
	return filepath.Join(p.workDir, "filelist.txt")
}

func (p *M4bProject) getMetadataTags(metadata string) (map[string]string, []string) {
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
