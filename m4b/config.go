// Package m4b provides functionality for handling M4B audiobook files,
// including configuration, metadata management, and chapter handling.
package m4b

import (
	"errors"
	"fmt"
	"path/filepath"
)

// ProjectConfig represents the configuration for an M4B audiobook project,
// including paths to required files and rules for metadata and chapters.
type ProjectConfig struct {
	AudioFilePath string         `yaml:"audioFilePath"`
	CoverPath     string         `yaml:"coverPath"`
	HasChapters   bool           `yaml:"hasChapters"`
	MetadataRules []MetadataRule `yaml:"metadataRules"`
	ChapterRules  []ChapterRule  `yaml:"chapterRules"`
	ProjectPath   string         `yaml:"projectPath,omitempty"`
}

// Validate checks if the ProjectConfig is valid by ensuring required fields
// are present and all rules are valid. Returns an error if validation fails.
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

// FullAudioFilePath returns the absolute path to the audio file.
// Returns an error if the absolute path cannot be determined.
func (c *ProjectConfig) FullAudioFilePath() (string, error) {
	if filepath.IsAbs(c.AudioFilePath) {
		return c.AudioFilePath, nil
	}

	audioFilePath, err := filepath.Abs(filepath.Join(c.ProjectPath, c.AudioFilePath))
	if err != nil {
		return "", fmt.Errorf("could not get absolute path %s, %w", c.AudioFilePath, err)
	}
	return audioFilePath, nil
}
