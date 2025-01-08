package utils

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyRegex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		regex    *regexp.Regexp
		format   string
		expected string
		wantErr  bool
	}{
		{
			name:     "partial replace",
			input:    "01 Folge 213_ Der Fluch der Medusa.m4b",
			regex:    regexp.MustCompile(`^\d+ (.+)$`),
			format:   "%s",
			expected: "Folge 213_ Der Fluch der Medusa.m4b",
		},
		{
			name:     "replace more complex",
			input:    "01 213_Der Fluch der Medusa.m4b",
			regex:    regexp.MustCompile(`^\d+ (\d+)_(.+)$`),
			format:   "Folge %s_ %s",
			expected: "Folge 213_ Der Fluch der Medusa.m4b",
		},
		{
			name:     "no match",
			input:    "01 213_Der Fluch der Medusa.m4b",
			regex:    regexp.MustCompile(`^ \d+$`),
			format:   "Folge %s_ %s",
			expected: "01 213_Der Fluch der Medusa.m4b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := ApplyRegex(tt.input, tt.regex, tt.format)
			if got != tt.expected {
				t.Errorf("GetMetadataField() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestSanitizePathComponent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "replaces ?", input: "Something With ???", expected: "Something With ___"},
		{name: "does not replace .", input: "J.R.R. Tolkien", expected: "J.R.R. Tolkien"},
		{name: "does not replace '", input: "Stanislawa d'Asp", expected: "Stanislawa d'Asp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizePathComponent(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}
