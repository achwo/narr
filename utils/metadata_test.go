package utils

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
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
			name: "two tags with the same value, only one is requested",
			metadata: `;FFMETADATA1
artist=Helge Schneider
album_artist=Helge Schneider`,
			tags:   []string{"artist"},
			regex:  regexp.MustCompile(`^(.+)$`),
			format: "%s (Autor)",
			expected: `;FFMETADATA1
artist=Helge Schneider (Autor)
album_artist=Helge Schneider`,
		},
		{
			name: "chapter with the same title as the requested tag",
			metadata: `;FFMETADATA1
title=Folge 1
[CHAPTER]
TIMEBASE=1/1000
START=0
END=500
title=Folge 1`,
			tags:   []string{"title"},
			regex:  regexp.MustCompile(`^Folge (\d+)$`),
			format: "%s. Folge",
			expected: `;FFMETADATA1
title=1. Folge
[CHAPTER]
TIMEBASE=1/1000
START=0
END=500
title=Folge 1`,
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

func TestParseFFMetadata_GlobalTags_ReturnsTagsInFileOrder(t *testing.T) {
	parsed := ParseFFMetadata(fullMetadata)

	assert.Equal(t, "123/Einfache Bäumung", parsed.Tags["title"])
	assert.Equal(t, "Something With ???", parsed.Tags["album_artist"])
	assert.Equal(
		t,
		[]string{
			"major_brand", "minor_version", "compatible_brands", "title", "artist",
			"album_artist", "album", "date", "disc", "track", "encoder",
		},
		parsed.TagOrder,
	)
	assert.Empty(t, parsed.Sections)
}

func TestParseFFMetadata_UppercaseTag_LowercasesTagName(t *testing.T) {
	parsed := ParseFFMetadata(";FFMETADATA1\nTITLE=Some Title")

	assert.Equal(t, "Some Title", parsed.Tags["title"])
	assert.Equal(t, []string{"title"}, parsed.TagOrder)
}

func TestFFMetadata_OriginalName_UppercaseTag_ReturnsTheSpellingFromTheDocument(t *testing.T) {
	parsed := ParseFFMetadata(";FFMETADATA1\nSERIES=Die Reihe")

	assert.Equal(t, "SERIES", parsed.OriginalName("series"))
}

func TestFFMetadata_OriginalName_UnknownTag_ReturnsTagUnchanged(t *testing.T) {
	parsed := ParseFFMetadata(";FFMETADATA1\ntitle=Book")

	assert.Equal(t, "series", parsed.OriginalName("series"))
}

func TestFFMetadata_OriginalName_AfterClone_IsPreserved(t *testing.T) {
	parsed := ParseFFMetadata(";FFMETADATA1\nSERIES=Die Reihe").Clone()

	assert.Equal(t, "SERIES", parsed.OriginalName("series"))
}

func TestParseFFMetadata_EscapedValue_ReturnsUnescapedValue(t *testing.T) {
	parsed := ParseFFMetadata(";FFMETADATA1\ntitle=a\\=b\\;c\\#d\\\\e")

	assert.Equal(t, `a=b;c#d\e`, parsed.Tags["title"])
}

func TestParseFFMetadata_EscapedLineBreak_JoinsLines(t *testing.T) {
	parsed := ParseFFMetadata(";FFMETADATA1\ntitle=first\\\nsecond\nalbum=Book")

	assert.Equal(t, "first\nsecond", parsed.Tags["title"])
	assert.Equal(t, "Book", parsed.Tags["album"])
}

func TestParseFFMetadata_ChapterSection_KeepsSectionVerbatim(t *testing.T) {
	metadata := ";FFMETADATA1\ntitle=Book\n[CHAPTER]\nTIMEBASE=1/1000\nSTART=0\nEND=500\ntitle=Ch1\n"

	parsed := ParseFFMetadata(metadata)

	assert.Equal(t, map[string]string{"title": "Book"}, parsed.Tags)
	assert.Equal(t, "[CHAPTER]\nTIMEBASE=1/1000\nSTART=0\nEND=500\ntitle=Ch1\n", parsed.Sections)
}

func TestParseFFMetadata_EmptyLines_AreIgnored(t *testing.T) {
	parsed := ParseFFMetadata(";FFMETADATA1\ntitle=Book\n\n")

	assert.Equal(t, []string{"title"}, parsed.TagOrder)
}

func TestFFMetadata_String_ParsedMetadata_RoundTrips(t *testing.T) {
	metadata := ";FFMETADATA1\ntitle=a\\=b\nalbum=Book\n[CHAPTER]\nTIMEBASE=1/1000\nSTART=0\nEND=500\ntitle=Ch1\n"

	parsed := ParseFFMetadata(metadata)

	assert.Equal(t, metadata, parsed.String())
}

func TestFFMetadata_String_RemovedTag_OmitsTag(t *testing.T) {
	parsed := ParseFFMetadata(";FFMETADATA1\ntitle=Book\nalbum=Album\n")
	delete(parsed.Tags, "album")

	assert.Equal(t, ";FFMETADATA1\ntitle=Book\n", parsed.String())
}

func TestFFMetadata_String_SpecialCharacters_EscapesValue(t *testing.T) {
	parsed := ParseFFMetadata(";FFMETADATA1\n")
	parsed.SetTag("title", "Folge 1; Teil 2 = Ende")

	assert.Equal(t, ";FFMETADATA1\ntitle=Folge 1\\; Teil 2 \\= Ende\n", parsed.String())
}

func TestFFMetadata_SetTag_ExistingTag_KeepsOrder(t *testing.T) {
	parsed := ParseFFMetadata(";FFMETADATA1\ntitle=Book\nalbum=Album\n")

	parsed.SetTag("title", "Other")

	assert.Equal(t, []string{"title", "album"}, parsed.TagOrder)
	assert.Equal(t, "Other", parsed.Tags["title"])
}

func TestFFMetadata_SyncTagOrder_ChangedTags_DropsRemovedAndAppendsNew(t *testing.T) {
	parsed := ParseFFMetadata(";FFMETADATA1\ntitle=Book\nalbum=Album\n")
	delete(parsed.Tags, "album")
	parsed.Tags["genre"] = "Hoerspiel"
	parsed.Tags["comment"] = "Note"

	parsed.SyncTagOrder()

	assert.Equal(t, []string{"title", "comment", "genre"}, parsed.TagOrder)
}

func TestFFMetadata_Clone_ModifiedClone_LeavesOriginalUntouched(t *testing.T) {
	parsed := ParseFFMetadata(";FFMETADATA1\ntitle=Book\n")

	clone := parsed.Clone()
	clone.SetTag("title", "Other")
	clone.SetTag("album", "Album")

	assert.Equal(t, "Book", parsed.Tags["title"])
	assert.Equal(t, []string{"title"}, parsed.TagOrder)
}

func TestUnescapeFFMetadataValue_TrailingBackslash_KeepsBackslash(t *testing.T) {
	assert.Equal(t, `abc\`, UnescapeFFMetadataValue(`abc\`))
}
