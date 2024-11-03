package m4b

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/achwo/narr/utils"
)

// MetadataRule defines a rule for modifying metadata tags in an M4B file.
// Rules can be of different types like "regex", "delete", or "set" and operate
// on specific metadata tags.
type MetadataRule struct {
	Type   string `yaml:"type"`
	Tag    string `yaml:"tag,omitempty"`
	Value  string `yaml:"value,omitempty"`
	Regex  string `yaml:"regex,omitempty"`
	Format string `yaml:"format,omitempty"`
}

// Apply executes the rule on the provided tags map, modifying the tags according
// to the rule's type and parameters. Returns an error if the rule application fails.
func (r *MetadataRule) Apply(tags map[string]string) error {
	value, exists := tags[r.Tag]

	if !exists {
		return fmt.Errorf("tag %s does not exist", r.Tag)
	}

	// TODO: implement delete, set

	switch r.Type {
	case "regex":
		regex, err := regexp.Compile(r.Regex)
		if err != nil {
			// TODO: might be better in construction (want to know validity in config check also)
			return fmt.Errorf("metadata rule regex '%s' is invalid: %w", r.Regex, err)
		}
		newValue, err := utils.ApplyRegex(value, regex, r.Format)
		if err != nil {
			// TODO: might be better in construction (want to know validity in config check also)
			return fmt.Errorf("could not apply rule '%s': %w", r.Regex, err)
		}
		tags[r.Tag] = newValue
	default:
		return errors.ErrUnsupported
	}

	return nil
}

// Validate checks if the rule is properly configured with all required fields
// based on its type. Returns an error if the configuration is invalid.
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

// ChapterRule defines a rule for modifying chapter titles in an M4B file
// using regex pattern matching and formatting.
type ChapterRule struct {
	Regex  string `yaml:"regex"`
	Format string `yaml:"format"`
}

// Validate checks if the chapter rule has both required regex and format fields.
// Returns an error if either field is missing.
func (r *ChapterRule) Validate() error {
	if r.Regex == "" || r.Format == "" {
		return errors.New("regex rule requires both regex and format")
	}
	return nil
}

// Apply executes the chapter rule on the provided chapter title string.
// Returns the modified chapter title and any error that occurred during processing.
func (r *ChapterRule) Apply(chapter string) (string, error) {
	regex, err := regexp.Compile(r.Regex)
	if err != nil {
		// TODO: might be better in construction (want to know validity in config check also)
		return "", fmt.Errorf("Chapter rule regex '%s' is invalid: %w", r.Regex, err)
	}
	return utils.ApplyRegex(chapter, regex, r.Format)
}
