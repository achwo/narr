// Package m4b provides functionality for handling M4B audiobook files,
// including configuration, metadata management, and chapter handling.
package m4b

import (
	"fmt"
	"os"
	"path/filepath"
)

// ProjectConfig represents the configuration for an M4B audiobook project,
// including paths to required files and rules for metadata and chapters.
type ProjectConfig struct {
	CoverPath     string         `yaml:"coverPath"`
	HasChapters   bool           `yaml:"hasChapters"`
	MetadataRules []MetadataRule `yaml:"metadataRules"`
	ChapterRules  []ChapterRule  `yaml:"chapterRules"`
	ShouldConvert bool           `yaml:"shouldConvert"`
	Multi         bool           `yaml:"multi"`
	ProjectPath   string         `yaml:"projectPath,omitempty"`
	OutputRoot    string         `yaml:"-"`
}

// Validate checks if the ProjectConfig is valid by ensuring required fields
// are present and all rules are valid. Returns an error if validation fails.
func (c *ProjectConfig) Validate() error {
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
	audioFilePath, err := filepath.Abs(c.ProjectPath)
	if err != nil {
		return "", fmt.Errorf("could not get absolute path %s, %w", c.ProjectPath, err)
	}
	return audioFilePath, nil
}

// OutputPath returns the directory below which the m4b is written: OutputRoot,
// or ~/narr when OutputRoot is empty.
func (c *ProjectConfig) OutputPath() (string, error) {
	if c.OutputRoot != "" {
		return c.OutputRoot, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get user home dir: %w", err)
	}

	return filepath.Join(home, "narr"), nil
}
