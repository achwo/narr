package m4b

import (
	"errors"
	"fmt"
	"path/filepath"
)

type ProjectConfig struct {
	AudioFilePath string         `yaml:"audioFilePath"`
	CoverPath     string         `yaml:"coverPath"`
	HasChapters   bool           `yaml:"hasChapters"`
	MetadataRules []MetadataRule `yaml:"metadataRules"`
	ChapterRules  []ChapterRule  `yaml:"chapterRules"`
	ProjectPath   string         `yaml:"projectPath,omitempty"`
}

func (c *ProjectConfig) Validate() error {
	if c.AudioFilePath == "" {
		return errors.New("audioFilePath must be a valid path")
	}

	for _, rule := range c.MetadataRules {
		err := rule.Validate()
		if err != nil {
			return fmt.Errorf("metadata rule invalid: %w", err)
		}
	}

	for _, rule := range c.ChapterRules {
		err := rule.Validate()
		if err != nil {
			return fmt.Errorf("chapter rule invalid: %w", err)
		}
	}

	return nil
}

func (c *ProjectConfig) FullAudioFilePath() (string, error) {
	audioFilePath, err := filepath.Abs(c.AudioFilePath)
	if err != nil {
		return "", fmt.Errorf("could not get absolute path %s, %w", c.AudioFilePath, err)
	}
	return audioFilePath, nil
}
