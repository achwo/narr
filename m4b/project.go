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

// NewProjectFromPath creates a new BookProject from a configuration file at the given path.
// It reads and parses the narr.yaml configuration file and initializes the project with the
// provided audio file provider, metadata provider, and audio converter.
// Returns an error if the configuration file cannot be read or parsed.
func NewProjectFromPath(
	path string,
	audioProvider AudioFileProvider,
	metadataProvider MetadataProvider,
	audioConverter AudioProcessor,
) (*BookProject, error) {
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

// NewProject creates a new BookProject with the given configuration and providers.
// It validates the configuration before creating the project.
// Returns an error if the configuration is invalid.
func NewProject(
	config ProjectConfig,
	audioProvider AudioFileProvider,
	metadataProvider MetadataProvider,
	audioConverter AudioProcessor,
) (*BookProject, error) {
	err := config.Validate()
	if err != nil {
		return nil, err
	}

	return &BookProject{
		Config:            config,
		AudioFileProvider: audioProvider,
		MetadataProvider:  metadataProvider,
		AudioProcessor:    audioConverter,
	}, nil
}

// AudioFileProvider defines the interface for providing audio files from a directory.
type AudioFileProvider interface {
	AudioFiles(fullPath string) ([]string, error)
}

// AudioProcessor defines the interface for processing audio files, including
// conversion, concatenation, and metadata manipulation operations.
type AudioProcessor interface {
	ToM4A(files []string, outputPath string) ([]string, error)
	Concat(m4aFiles []string, templateFilePath string, outputPath string) (string, error)
	AddMetadata(m4bFile string, metadata string, bookTitle string) error
	AddCover(m4bFile string, coverFile string) error
	ExtractCover(m4aFile string) (string, error)
	AddChapters(m4bFile string, chapters string) error
}

// MetadataProvider defines the interface for reading metadata from audio files.
type MetadataProvider interface {
	ReadTitleAndDuration(file string) (string, float64, error)
	ReadMetadata(file string) (string, error)
}

// Project defines the interface for working with audiobook projects.
type Project interface {
	AudioFiles() ([]string, error)
	ShowChapters() (string, error)
	ShowMetadata() (string, error)
}

// BookProject represents an audiobook project that can be converted to M4B format.
// It contains configuration, providers for audio files and metadata, and an audio processor
// for handling audio file conversions and manipulations.
type BookProject struct {
	Config            ProjectConfig
	AudioFileProvider AudioFileProvider
	MetadataProvider  MetadataProvider
	AudioProcessor    AudioProcessor
	tracks            []Track
	workDir           string
}

// ConvertToM4B processes all audio files in the project and creates a single M4B audiobook file.
// It handles conversion to M4A, concatenation, and addition of metadata, cover art, and chapters.
// Returns the path to the created M4B file and any error encountered during the process.
func (p *BookProject) ConvertToM4B() (string, error) {
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
		return "", fmt.Errorf("could not get metadata for m4b: %w", err)
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

// Cover returns the path to the cover image for the audiobook.
// It first checks for a cover specified in the configuration, then attempts to
// extract a cover from the first audio file if no configuration cover exists.
func (p *BookProject) Cover() (string, error) {
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

// Tracks returns a sorted list of all audio tracks in the project.
// Tracks are sorted by disc number and track number, with filename as a fallback.
// Results are cached after the first call.
func (p *BookProject) Tracks() ([]Track, error) {
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

// Chapters generates chapter markers for the audiobook based on the track metadata
// and configured chapter rules. Returns the chapter markers in FFmpeg metadata format.
func (p *BookProject) Chapters() (string, error) {
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

	for i, name := range chapterOrder {
		chapter := chapters[name]
		markers = append(markers, chapter.ChapterMarker(i))
	}

	markersFileContent := strings.Join(markers, "\n\n")

	return markersFileContent, nil
}

// Metadata returns the audiobook metadata in FFmpeg metadata format.
// The metadata is derived from the first track and processed according to the
// configured metadata rules.
func (p *BookProject) Metadata() (string, error) {
	tags, tagOrder, err := p.getUpdatedMetadata()
	if err != nil {
		return "", err
	}

	firstLine := ";FFMETADATA1"

	lines := []string{firstLine}

	for _, tag := range tagOrder {
		switch tag {
		case "track":
			continue
		case "disc":
			continue
		default:
			lines = append(lines, tag+"="+tags[tag])
		}
	}

	return strings.Join(lines, "\n"), nil
}

// Filename returns the output file name for the project.
// It takes artist and book title from the first file as a basis.
func (p *BookProject) Filename() (string, error) {
	artist, album, err := p.ArtistAndBookTitle()
	if err != nil {
		return "", err
	}
	filename := filepath.Join(p.Config.ProjectPath, artist, album+".m4b")

	return filename, nil
}

// ArtistAndBookTitle reads the metadata from the first track and returns the
// artist and book title.
func (p *BookProject) ArtistAndBookTitle() (string, string, error) {
	audioFiles, err := p.Tracks()
	if err != nil {
		return "", "", fmt.Errorf("could not load audio files: %w", err)
	}

	if len(audioFiles) == 0 {
		return "", "", errors.New("no audio files found")
	}

	tags, _, err := p.getUpdatedMetadata()

	if err != nil {
		return "", "", fmt.Errorf("could not get metadata for artist and book title: %w", err)
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

func (p *BookProject) getUpdatedMetadata() (map[string]string, []string, error) {
	audioFiles, err := p.Tracks()
	if err != nil {
		return nil, nil, fmt.Errorf("could not load audio files: %w", err)
	}

	if len(audioFiles) == 0 {
		return nil, nil, errors.New("no audio files found")
	}

	return audioFiles[0].Metadata, audioFiles[0].TagOrder, nil
}

func (p *BookProject) getUpdatedFileMetadata(file string) (map[string]string, []string, error) {
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

func (p *BookProject) m4aPath() (string, error) {
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

func (p *BookProject) filelistFile() string {
	return filepath.Join(p.workDir, "filelist.txt")
}

func (p *BookProject) getMetadataTags(metadata string) (map[string]string, []string) {
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
