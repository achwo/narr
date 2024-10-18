package m4b

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/achwo/narr/utils"
	"gopkg.in/yaml.v3"
)

const configFileName = "narr.yaml"

type Project interface {
	AudioFiles() ([]string, error)
	ShowChapters() (string, error)
}

type M4bProject struct {
	Config ProjectConfig
}

func (c *M4bProject) AudioFiles() ([]string, error) {
	fullpath, err := c.Config.FullAudioFilePath()
	if err != nil {
		return nil, err
	}

	return utils.GetFilesByExtension(fullpath, ".m4a")
}

func (p *M4bProject) ShowChapters() (string, error) {
	audioFiles, err := p.AudioFiles()
	if err != nil {
		return "", fmt.Errorf("Could not load audio files: %w", err)
	}

	chapters := make(map[string]*Chapter)
	var chapterOrder []string
	var previousChapter *Chapter

	var metadataManager utils.MetadataManager = &utils.FFmpegMetadataManager{}

	for i, file := range audioFiles {
		title, duration, err := metadataManager.ReadTitleAndDuration(file)
		if err != nil {
			return "", fmt.Errorf("Could not read file data for file %s: %w", file, err)
		}
		chapterName := title

		for _, rule := range p.Config.ChapterRules {
			chapterName, err = rule.Apply(chapterName)
			if err != nil {
				return "", fmt.Errorf("Chapter rule invalid: %w", err)
			}
		}

		value := chapters[chapterName]
		value, exists := chapters[chapterName]

		newFile := File{Name: file, Duration: duration}
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

func NewProject(config ProjectConfig) (*M4bProject, error) {
	err := config.Validate()
	if err != nil {
		return nil, err
	}

	return &M4bProject{Config: config}, nil
}

func NewProjectFromPath(path string) (*M4bProject, error) {
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

	config.ProjectPath = filepath.Base(fullpath)

	return NewProject(config)
}

func (p *M4bProject) ShowMetadata() (string, error) {
	audioFiles, err := p.AudioFiles()
	if err != nil {
		return "", fmt.Errorf("could not load audio files: %w", err)
	}

	if len(audioFiles) == 0 {
		return "", errors.New("no audio files found")
	}

	referenceFile := audioFiles[0]

	var metadataManager utils.MetadataManager = &utils.FFmpegMetadataManager{}
	metadata, err := metadataManager.ReadMetadata(referenceFile)
	if err != nil {
		return "", fmt.Errorf("could not read metadata: %w", err)
	}

	return metadata, nil
}
