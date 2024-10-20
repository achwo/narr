package m4b

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/achwo/narr/utils"
)

type MetadataRule struct {
	Type   string `yaml:"type"`
	Tag    string `yaml:"tag,omitempty"`
	Value  string `yaml:"value,omitempty"`
	Regex  string `yaml:"regex,omitempty"`
	Format string `yaml:"format,omitempty"`
}

func (r *MetadataRule) Apply(tags map[string]string) error {
	value, exists := tags[r.Tag]

	if !exists {
		return fmt.Errorf("Tag %s does not exist", r.Tag)
	}

	// TODO: implement delete, set

	switch r.Type {
	case "regex":
		regex, err := regexp.Compile(r.Regex)
		if err != nil {
			// TODO: might be better in construction (want to know validity in config check also)
			return fmt.Errorf("Metadata rule regex '%s' is invalid: %w", r.Regex, err)
		}
		newValue, err := utils.ApplyRegex(value, regex, r.Format)
		tags[r.Tag] = newValue
	default:
		return errors.ErrUnsupported
	}

	return nil
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

type ChapterRule struct {
	Regex  string `yaml:"regex"`
	Format string `yaml:"format"`
}

func (r *ChapterRule) Validate() error {
	if r.Regex == "" || r.Format == "" {
		return errors.New("regex rule requires both regex and format")
	}
	return nil
}

func (r *ChapterRule) Apply(chapter string) (string, error) {
	regex, err := regexp.Compile(r.Regex)
	if err != nil {
		// TODO: might be better in construction (want to know validity in config check also)
		return "", fmt.Errorf("Chapter rule regex '%s' is invalid: %w", r.Regex, err)
	}
	return utils.ApplyRegex(chapter, regex, r.Format)
}
