package utils

import (
	"reflect"
	"regexp"
	"testing"
)

var fullMetadata = `;FFMETADATA1
major_brand=M4A
minor_version=512
compatible_brands=M4A isomiso2
title=123/Einfache Bäumung
artist=Something With ???
album_artist=Something With ???
album=123/Einfache Bäumung
date=2002-03-11
disc=1
track=1
encoder=Lavf61.7.100`

func TestGetMetadataField(t *testing.T) {
	tests := []struct {
		name     string
		metadata string
		tags     []string
		expected []TagWithValue
		wantErr  bool
	}{
		{
			name:     "field exists",
			metadata: fullMetadata,
			tags:     []string{"title", "album", "date"},
			expected: []TagWithValue{
				{Tag: "title", Value: "123/Einfache Bäumung"},
				{Tag: "album", Value: "123/Einfache Bäumung"},
				{Tag: "date", Value: "2002-03-11"},
			},
		},
		{
			name:     "field with no content",
			metadata: "album=",
			tags:     []string{"album"},
			expected: []TagWithValue{{Tag: "album", Value: ""}},
		},
		{
			name:     "field does not exist",
			metadata: "album=Some Title",
			tags:     []string{"title"},
			expected: nil,
		},
		{
			name:     "empty metadata",
			metadata: "",
			tags:     []string{"title"},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetMetadataTagValues(tt.metadata, tt.tags)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("GetMetadataField() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestUpdateMetadataTags(t *testing.T) {
	tests := []struct {
		name     string
		metadata string
		tags     []string
		regex    *regexp.Regexp
		format   string
		expected string
		wantErr  bool
	}{
		{
			name:     "field exists",
			metadata: fullMetadata,
			tags:     []string{"album", "title"},
			regex:    regexp.MustCompile(`^(\d+)/(.+)$`),
			format:   "Folge %s: %s",
			expected: `;FFMETADATA1
major_brand=M4A
minor_version=512
compatible_brands=M4A isomiso2
title=Folge 123: Einfache Bäumung
artist=Something With ???
album_artist=Something With ???
album=Folge 123: Einfache Bäumung
date=2002-03-11
disc=1
track=1
encoder=Lavf61.7.100`,
		},
		{
			name: "field exists but does not match regex",
			metadata: `;FFMETADATA1
		title=NoMatch Bäumung
		album=NoMatch Bäumung`,
			tags:   []string{"album", "title"},
			regex:  regexp.MustCompile(`^(\d+)/(.+)$`),
			format: "Folge %s: %s",
			expected: `;FFMETADATA1
		title=NoMatch Bäumung
		album=NoMatch Bäumung`,
		},
		{
			name: "field does not exist",
			metadata: `;FFMETADATA1
		title=123/Einfache Bäumung`,
			tags:   []string{"album"},
			regex:  regexp.MustCompile(`^(\d+)/(.+)$`),
			format: "Folge %s: %s",
			expected: `;FFMETADATA1
		title=123/Einfache Bäumung`,
		},
		{
			name: "multiple fields, only one matches regex",
			metadata: `;FFMETADATA1
title=123/Einfache Bäumung
album=Other Bäumung`,
			tags:   []string{"album", "title"},
			regex:  regexp.MustCompile(`^(\d+)/(.+)$`),
			format: "Folge %s: %s",
			expected: `;FFMETADATA1
title=Folge 123: Einfache Bäumung
album=Other Bäumung`,
		},
		{
			name:     "only one format string",
			metadata: `title=Und der hunger`,
			tags:     []string{"title"},
			regex:    regexp.MustCompile(`^Und(.+)$`),
			format:   "und%s",
			expected: `title=und der hunger`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := UpdateMetadataTags(tt.metadata, tt.tags, tt.regex, tt.format)
			if got != tt.expected {
				t.Errorf("GetMetadataField() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
