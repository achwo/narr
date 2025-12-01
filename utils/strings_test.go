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
			input:    "01 Folge 213_ Something a little Longer.m4b",
			regex:    regexp.MustCompile(`^\d+ (.+)$`),
			format:   "%s",
			expected: "Folge 213_ Something a little Longer.m4b",
		},
		{
			name:     "replace more complex",
			input:    "01 213_Something a little Longer.m4b",
			regex:    regexp.MustCompile(`^\d+ (\d+)_(.+)$`),
			format:   "Folge %s_ %s",
			expected: "Folge 213_ Something a little Longer.m4b",
		},
		{
			name:     "no match",
			input:    "01 213_Something a little Longer.m4b",
			regex:    regexp.MustCompile(`^ \d+$`),
			format:   "Folge %s_ %s",
			expected: "01 213_Something a little Longer.m4b",
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

func TestNaturalCompare(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int // negative if a < b, 0 if equal, positive if a > b
	}{
		{name: "numbers in order", a: "file2.txt", b: "file10.txt", expected: -1},
		{name: "numbers reversed", a: "file10.txt", b: "file2.txt", expected: 1},
		{name: "equal strings", a: "file10.txt", b: "file10.txt", expected: 0},
		{name: "three digit numbers", a: "file100.txt", b: "file10.txt", expected: 1},
		{name: "leading zeros", a: "file01.txt", b: "file1.txt", expected: 1}, // numerically equal but longer string
		{name: "prefix comparison", a: "file10", b: "file10.txt", expected: -1},
		{name: "multiple numbers", a: "01-file-10.txt", b: "02-file-9.txt", expected: -1},
		{name: "chapter example", a: "10-kapitel.flac", b: "100-kapitel.flac", expected: -1},
		{name: "chapter example 2", a: "09-kapitel.flac", b: "10-kapitel.flac", expected: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NaturalCompare(tt.a, tt.b)
			if tt.expected < 0 {
				require.Less(t, result, 0, "expected %s < %s", tt.a, tt.b)
			} else if tt.expected > 0 {
				require.Greater(t, result, 0, "expected %s > %s", tt.a, tt.b)
			} else {
				require.Equal(t, 0, result, "expected %s == %s", tt.a, tt.b)
			}
		})
	}
}
