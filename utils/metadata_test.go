package utils

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"
)

var fullMetadata = `;FFMETADATA1
major_brand=M4A
minor_version=512
compatible_brands=M4A isomiso2
title=102/Doppelte Bäumung
artist=Something With ???
album_artist=Something With ???
album=102/Doppelte Bäumung
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
			tags:     []string{"title", "album"},
			expected: []TagWithValue{
				{Tag: "title", Value: "102/Doppelte Bäumung"},
				{Tag: "album", Value: "102/Doppelte Bäumung"},
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
			if got == nil {
				fmt.Println("foll null")
			}
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
title=Folge 102: Doppelte Bäumung
artist=Something With ???
album_artist=Something With ???
album=Folge 102: Doppelte Bäumung
date=2002-03-11
disc=1
copyright=℗ 2002 Sony Music Entertainment GmbH
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
title=102/Doppelte Bäumung`,
			tags:   []string{"album"},
			regex:  regexp.MustCompile(`^(\d+)/(.+)$`),
			format: "Folge %s: %s",
			expected: `;FFMETADATA1
title=102/Doppelte Bäumung`,
		},
		{
			name: "multiple fields, only one matches regex",
			metadata: `;FFMETADATA1
title=102/Doppelte Bäumung
album=Other Bäumung`,
			tags:   []string{"album", "title"},
			regex:  regexp.MustCompile(`^(\d+)/(.+)$`),
			format: "Folge %s: %s",
			expected: `;FFMETADATA1
title=Folge 102: Doppelte Bäumung
album=Other Bäumung`,
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
