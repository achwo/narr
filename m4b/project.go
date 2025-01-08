package m4b

import (
	"cmp"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/achwo/narr/utils"
	"gopkg.in/yaml.v3"
)

const configFileName = "narr.yaml"

func NewProjectsByArgs(path string, recursive bool) ([]*Project, error) {
	var projects []*Project

	var err error
	if recursive {
		projects, err = NewRecursiveProjectsFromPath(path)
	} else {
		projects, err = NewProjectsFromPath(path)
	}

	if err != nil {
		return nil, fmt.Errorf("could not create project(s): %w", err)
	}

	return projects, nil
}

func NewRecursiveProjectsFromPath(path string) ([]*Project, error) {
	projectConfigs, err := utils.GetAllFilesByName(path, "narr.yaml")
	if err != nil {
		return nil, fmt.Errorf("could not get project directories: %w", err)
	}

	var projects []*Project

	for _, config := range projectConfigs {
		project, err := NewProjectsFromPath(config)
		if err != nil {
			return nil, fmt.Errorf("could not create project for path '%s': %w", config, err)
		}
		projects = slices.Concat(projects, project)
	}

	return projects, nil
}

// NewProjectsFromPath returns a slice of Projects
// Depending on the config it might be:
// - a single Project
// - multiple Projects (when config Multi is true)
func NewProjectsFromPath(path string) ([]*Project, error) {
	var fullpath string

	if strings.HasSuffix(path, configFileName) {
		fullpath = path
	} else {
		fullpath = filepath.Join(path, configFileName)
	}

	config, err := readConfig(fullpath)
	if err != nil {
		return nil, fmt.Errorf("could not read config file %s: %w", fullpath, err)
	}

	if config.Multi {
		baseDir := filepath.Dir(fullpath)
		projectDirEntries, err := os.ReadDir(baseDir)
		if err != nil {
			return nil, fmt.Errorf("could not get multi project directories: %w", err)
		}

		var projects []*Project

		for _, dirEntry := range projectDirEntries {
			if !dirEntry.IsDir() {
				continue
			}

			projectConfig := *config
			projectConfig.ProjectPath = filepath.Join(baseDir, dirEntry.Name())

			project, err := NewProject(projectConfig)
			if err != nil {
				return nil, fmt.Errorf("could not create project for path '%s': %w", dirEntry, err)
			}

			projects = append(projects, project)
		}

		return projects, nil

	} else {
		config.ProjectPath = filepath.Dir(fullpath)

		project, err := NewProject(*config)
		if err != nil {
			return nil, err
		}

		return []*Project{project}, nil
	}
}

func readConfig(fullpath string) (*ProjectConfig, error) {
	bytes, err := os.ReadFile(fullpath)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", fullpath, err)
	}

	var config ProjectConfig
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal file %s: %w", fullpath, err)
	}

	return &config, nil
}

// NewProjectWithDeps creates a new Project with the given configuration and providers.
// It validates the configuration before creating the project.
// Returns an error if the configuration is invalid.
func NewProjectWithDeps(config ProjectConfig, deps ProjectDependencies) (*Project, error) {
	err := config.Validate()
	if err != nil {
		return nil, err
	}

	return &Project{Config: config, deps: deps}, nil
}

// NewProject creates a new Project with the given configuration.
// It validates the configuration before creating the project.
// Returns an error if the configuration is invalid.
func NewProject(config ProjectConfig) (*Project, error) {
	audioFileProvider := &utils.OSAudioFileProvider{}
	audioProcessor := &FFmpegAudioProcessor{Command: &ExecCommand{}}
	trackFactory := &FFmpegTrackFactory{AudioProcessor: audioProcessor}

	deps := ProjectDependencies{
		AudioFileProvider: audioFileProvider,
		AudioProcessor:    audioProcessor,
		TrackFactory:      trackFactory,
	}
	return NewProjectWithDeps(config, deps)
}

// audioFileProvider defines the interface for providing audio files from a directory.
type audioFileProvider interface {
	AudioFiles(fullPath string) ([]string, error)
}

// audioProcessor defines the interface for processing audio files, including
// conversion, concatenation, and metadata manipulation operations.
type audioProcessor interface {
	Concat(m4aFiles []string, templateFilePath string, outputPath string) (string, error)
	AddMetadata(m4bFile string, metadata string, bookTitle string) error
	AddCover(m4bFile string, coverFile string) error
	ExtractCover(m4aFile string, workDir string) (string, error)
	AddChapters(m4bFile string, chapters string) error
	ReadTitleAndDuration(file string) (string, float64, error)
	ReadMetadata(file string) (string, error)
	ToM4A(files []string, outputPath string) ([]string, error)
}

type trackFactory interface {
	LoadTracks(file []string, metadataRules []MetadataRule) ([]Track, error)
	LoadTrack(file string, metadataRules []MetadataRule) (Track, error)
}

// Project represents an audiobook project that can be converted to M4B format.
// It contains configuration, providers for audio files and metadata, and an audio processor
// for handling audio file conversions and manipulations.
type Project struct {
	Config  ProjectConfig
	tracks  []Track
	workDir string
	deps    ProjectDependencies
}

type ProjectDependencies struct {
	AudioFileProvider audioFileProvider
	AudioProcessor    audioProcessor
	TrackFactory      trackFactory
}

// ConvertToM4B processes all audio files in the project and creates a single M4B audiobook file.
// It handles conversion to M4A, concatenation, and addition of metadata, cover art, and chapters.
// Returns the path to the created M4B file and any error encountered during the process.
func (p *Project) ConvertToM4B() (string, error) {
	if workDir, err := os.MkdirTemp("", "convert"); err == nil {
		p.workDir = workDir
	} else {
		return "", fmt.Errorf("could not create work dir: %w", err)
	}
	defer os.RemoveAll(p.workDir)

	tracks, err := p.Tracks()
	if err != nil {
		return "", fmt.Errorf("could not load audio files: %w", err)
	}

	finalFilename, err := p.Filename()
	if err != nil {
		return "", fmt.Errorf("could not get filename: %w", err)
	}

	if completed := p.AlreadyCompleted(); completed {
		fmt.Println("Skipping, as already completed")
		return finalFilename, nil
	}
	files := make([]string, 0, len(tracks))

	for _, track := range tracks {
		files = append(files, track.File)
	}

	// running chapters before conversion to prevent long wait before error
	chapters, err := p.Chapters()
	if err != nil {
		return "", fmt.Errorf("could not get chapters: %w", err)
	}

	m4aFiles := files
	if p.Config.ShouldConvert {
		fmt.Printf("Converting %d files to m4a\n", len(files))
		m4aPath, err := p.m4aPath()
		if err != nil {
			return "", fmt.Errorf("could not create m4a path: %w", err)
		}

		m4aFiles, err = p.deps.AudioProcessor.ToM4A(files, m4aPath)

		if err != nil {
			return "", fmt.Errorf("could not convert files to m4a: %w", err)
		}
	}

	fmt.Println("Concating files")
	m4bFile, err := p.deps.AudioProcessor.Concat(m4aFiles, p.filelistFile(), p.workDir)
	if err != nil {
		return "", err
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

	if err = p.deps.AudioProcessor.AddMetadata(m4bFile, metadata, bookTitle); err != nil {
		return "", fmt.Errorf("could not add metadata to %s: %w", m4bFile, err)
	}

	cover, err := p.Cover()
	if err != nil {
		return "", fmt.Errorf("could not get cover: %w", err)
	}

	fmt.Println("Adding cover to m4b")
	if err = p.deps.AudioProcessor.AddCover(m4bFile, cover); err != nil {
		return "", fmt.Errorf("could not add cover to %s: %w", m4bFile, err)
	}

	if p.Config.HasChapters {
		fmt.Println("Adding chapters to m4b")
		if err = p.deps.AudioProcessor.AddChapters(m4bFile, chapters); err != nil {
			return "", fmt.Errorf("could not add chapters to %s: %w", m4bFile, err)
		}
	}

	if err = os.MkdirAll(filepath.Dir(finalFilename), 0755); err != nil {
		return "", fmt.Errorf("could not create target dir: %w", err)
	}

	if err = os.Rename(m4bFile, finalFilename); err != nil {
		return "", fmt.Errorf("could not rename file: %w", err)
	}

	return finalFilename, nil
}

// Cover returns the path to the cover image for the audiobook.
// It first checks for a cover specified in the configuration, then attempts to
// extract a cover from the first audio file if no configuration cover exists.
func (p *Project) Cover() (string, error) {
	coverFromConfig := p.Config.CoverPath
	if !filepath.IsAbs(coverFromConfig) {
		coverFromConfig = filepath.Join(p.Config.ProjectPath, coverFromConfig)
	}

	if info, err := os.Stat(coverFromConfig); !errors.Is(err, os.ErrNotExist) && !info.IsDir() {
		return coverFromConfig, nil
	}

	tracks, err := p.Tracks()
	if err != nil {
		return "", err
	}
	firstFile := tracks[0].File
	return p.deps.AudioProcessor.ExtractCover(firstFile, p.workDir)
}

// Tracks returns a sorted list of all audio tracks in the project.
// Tracks are sorted by disc number and track number, with filename as a fallback.
// Results are cached after the first call.
func (p *Project) Tracks() ([]Track, error) {
	if p.tracks != nil {
		return p.tracks, nil
	}

	fullpath, err := p.Config.FullAudioFilePath()
	if err != nil {
		return nil, err
	}

	audioFiles, err := p.deps.AudioFileProvider.AudioFiles(fullpath)
	if err != nil {
		return nil, err
	}

	tracks, err := p.deps.TrackFactory.LoadTracks(audioFiles, p.Config.MetadataRules)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(tracks, sortTracks)
	p.tracks = tracks
	return tracks, nil
}

func sortTracks(a, b Track) int {
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
}

// Chapters generates chapter markers for the audiobook based on the track metadata
// and configured chapter rules. Returns the chapter markers in FFmpeg metadata format.
func (p *Project) Chapters() (string, error) {
	tracks, err := p.Tracks()
	if err != nil {
		return "", fmt.Errorf("could not load audio files: %w", err)
	}

	chapters := make(map[string]*Chapter)
	var chapterOrder []string
	var previousChapter *Chapter

	for i, track := range tracks {
		title, duration, err := track.TitleAndDuration()
		if err != nil {
			return "", fmt.Errorf("could not read file data for file %s: %w", track.File, err)
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
func (p *Project) Metadata() (string, error) {
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
func (p *Project) Filename() (string, error) {
	artist, album, err := p.ArtistAndBookTitle()
	if err != nil {
		return "", err
	}

	filename := filepath.Join(
		p.Config.OutputPath(),
		utils.SanitizePathComponent(artist),
		utils.SanitizePathComponent(album),
		utils.SanitizePathComponent(album)+".m4b",
	)

	return filename, nil
}

// ArtistAndBookTitle reads the metadata from the first track and returns the
// artist and book title.
func (p *Project) ArtistAndBookTitle() (string, string, error) {
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

func (p *Project) getUpdatedMetadata() (map[string]string, []string, error) {
	tracks, err := p.Tracks()
	if err != nil {
		return nil, nil, fmt.Errorf("could not load audio files: %w", err)
	}

	if len(tracks) == 0 {
		return nil, nil, errors.New("no audio files found")
	}

	metadata, tagOrder, err := tracks[0].Metadata()
	if err != nil {
		return nil, nil, err
	}

	return metadata, tagOrder, nil
}

func (p *Project) m4aPath() (string, error) {
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

func (p *Project) filelistFile() string {
	return filepath.Join(p.workDir, "filelist.txt")
}

func (p *Project) AlreadyCompleted() bool {
	tracks, err := p.Tracks()
	if err != nil {
		return false
	}

	total := 0.0

	for _, track := range tracks {
		total = total + track.duration
	}

	outputFile, err := p.Filename()
	if err != nil {
		return false
	}

	if _, err := os.Stat(outputFile); err != nil {
		return false
	}

	outputTrack, err := p.deps.TrackFactory.LoadTrack(outputFile, nil)
	if err != nil {
		return false
	}

	difference := math.Abs(outputTrack.duration - total)
	threshold := total * 0.05
	return difference <= threshold
}
