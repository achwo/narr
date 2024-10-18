package config

import (
	"errors"
	"fmt"
)

type ProjectConfig struct {
	AudioFilePath string         `yaml:"audioFilePath"`
	CoverPath     string         `yaml:"coverPath"`
	HasChapters   bool           `yaml:"hasChapters"`
	MetadataRules []MetadataRule `yaml:"metadataRules"`
	ChapterRules  []ChapterRule  `yaml:"chapterRules"`
	OutputRules   []OutputRule   `yaml:"outputRules"`
}

type MetadataRule struct {
	Type   string `yaml:"type"`
	Tag    string `yaml:"tag,omitempty"`
	Value  string `yaml:"value,omitempty"`
	Regex  string `yaml:"regex,omitempty"`
	Format string `yaml:"format,omitempty"`
}

type ChapterRule struct {
	Regex  string `yaml:"regex"`
	Format string `yaml:"format"`
}

type OutputRule struct {
	Type   string `yaml:"type"`
	Regex  string `yaml:"regex,omitempty"`
	Format string `yaml:"format,omitempty"`
	Value  string `yaml:"value,omitempty"`
}

func (r *MetadataRule) Validate() error {
	if r.Tag == "" {
		return errors.New("rule must have a tag")
	}
	switch r.Type {
	case "delete":
		if r.Value != "" || r.Regex != "" || r.Format != "" {
			return errors.New("delete rule cannot have value, regex, or format")
		}
	case "set":
		if r.Value == "" {
			return errors.New("set rule requires a value")
		}
		if r.Regex != "" || r.Format != "" {
			return errors.New("set rule cannot have regex or format")
		}
	case "regex":
		if r.Regex == "" || r.Format == "" {
			return errors.New("regex rule requires both regex and format")
		}
		if r.Value != "" {
			return errors.New("regex rule cannot have value")
		}
	default:
		return fmt.Errorf("unknown rule type: %s", r.Type)
	}
	return nil
}

func (r *OutputRule) Validate() error {
	switch r.Type {
	case "delete":
		if r.Value != "" || r.Regex != "" || r.Format != "" {
			return errors.New("delete rule cannot have value, regex, or format")
		}
	case "set":
		if r.Value == "" {
			return errors.New("set rule requires a value")
		}
		if r.Regex != "" || r.Format != "" {
			return errors.New("set rule cannot have regex or format")
		}
	case "regex":
		if r.Regex == "" || r.Format == "" {
			return errors.New("regex rule requires both regex and format")
		}
		if r.Value != "" {
			return errors.New("regex rule cannot have value")
		}
	default:
		return fmt.Errorf("unknown rule type: %s", r.Type)
	}
	return nil
}

func (r *ChapterRule) Validate() error {
	if r.Regex == "" || r.Format == "" {
		return errors.New("regex rule requires both regex and format")
	}
	return nil
}

func (c *ProjectConfig) Validate() error {
	if c.AudioFilePath == "" {
		return errors.New("audioFilePath must be a valid path")
	}

	for _, rule := range c.MetadataRules {
		err := rule.Validate()
		if err != nil {
			return fmt.Errorf("Metadata rule invalid: %w", err)
		}
	}

	for _, rule := range c.ChapterRules {
		err := rule.Validate()
		if err != nil {
			return fmt.Errorf("Chapter rule invalid: %w", err)
		}
	}

	for _, rule := range c.OutputRules {
		err := rule.Validate()
		if err != nil {
			return fmt.Errorf("Output rule invalid: %w", err)
		}
	}

	return nil
}
