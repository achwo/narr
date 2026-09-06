package utils

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
)

// TagWithValue represents a metadata tag and its associated value
type TagWithValue struct {
	Tag   string // The metadata tag name
	Value string // The value associated with the tag
}

// Prefix returns the tag name with an equals sign appended
func (t TagWithValue) Prefix() string {
	return fmt.Sprintf("%s=", t.Tag)
}

// String returns the tag and value formatted as "tag=value"
func (t TagWithValue) String() string {
	return fmt.Sprintf("%s=%s", t.Tag, t.Value)
}

// GetMetadataTagValues returns metadata tag values from a string
// Deprecated: Use Project#GetMetadataTags instead
func GetMetadataTagValues(metadata string, tags []string) []TagWithValue {
	var tagValues []TagWithValue

	lines := strings.Split(metadata, "\n")
	for _, line := range lines {
		for _, tag := range tags {
			fullTag := tag + "="
			if strings.HasPrefix(line, fullTag) {

				withValue := TagWithValue{
					Tag:   tag,
					Value: strings.TrimPrefix(line, fullTag),
				}

				tagValues = append(tagValues, withValue)
			}
		}
	}

	return tagValues
}

// UpdateMetadataTags updates the metadata fields provided in the tags slice,
// replacing their values based on the provided regular expression and format.
//
// The function looks for the current value of each tag in the metadata. If the
// current value matches the provided regular expression, it constructs a new
// value using the format string and the capture groups from the regex.
//
// Parameters:
//   - metadata: The full metadata string where fields are located.
//   - tags: A list of tags on which the substitution is applied.
//   - regex: A regex with capture groups
//   - format: A format string for constructing the new tag value, with placeholders
//     for the capture groups from the regex. (in go syntax)
//
// Returns: The updated metadata and diffs for each change.
func UpdateMetadataTags(
	metadata string,
	tags []string,
	regex *regexp.Regexp,
	format string,
) (string, []Diff) {
	var affectedLines []Diff
	tagsWithValue := GetMetadataTagValues(metadata, tags)

	for _, currentValue := range tagsWithValue {
		newValue, err := ApplyRegex(currentValue.Value, regex, format)
		if err != nil {
			continue
		}

		metadata = strings.ReplaceAll(
			metadata,
			currentValue.String(),
			currentValue.Prefix()+newValue,
		)

		affectedLines = append(affectedLines, Diff{
			Tag:    currentValue.Tag,
			Before: currentValue.Value,
			After:  newValue,
		})
	}
	return metadata, affectedLines
}

// Diff represents a difference between two metadata values
type Diff struct {
	// Tag represents the metadata tag name that was modified
	Tag string
	// Before contains the original value before modification
	Before string
	// After contains the new value after modification
	After string
}

// FFMetadataHeader is the first line of every ffmetadata document.
const FFMetadataHeader = ";FFMETADATA1"

// FFMetadata is a parsed ffmetadata document. Tags contains the global tags with
// unescaped values, TagOrder preserves the order in which they appeared, and
// Sections keeps everything from the first section header ([CHAPTER], [STREAM])
// onwards verbatim.
type FFMetadata struct {
	Tags     map[string]string
	TagOrder []string
	Sections string
}

// ParseFFMetadata parses a metadata string in ffmetadata format, as produced by
// "ffmpeg -f ffmetadata". Tag names are lowercased and values are unescaped.
func ParseFFMetadata(metadata string) *FFMetadata {
	parsed := &FFMetadata{Tags: make(map[string]string)}

	lines := strings.Split(metadata, "\n")

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if strings.HasPrefix(trimmed, "[") {
			parsed.Sections = strings.Join(lines[i:], "\n")
			break
		}

		for endsWithEscape(line) && i+1 < len(lines) {
			i++
			line = line + "\n" + lines[i]
		}

		separator := indexUnescaped(line, '=')
		if separator < 0 {
			continue
		}

		tag := strings.ToLower(UnescapeFFMetadataValue(line[:separator]))
		parsed.SetTag(tag, UnescapeFFMetadataValue(line[separator+1:]))
	}

	return parsed
}

// String renders the metadata back into ffmetadata format, escaping values and
// appending the preserved sections. Tags removed from the Tags map are omitted.
func (m *FFMetadata) String() string {
	lines := make([]string, 0, len(m.TagOrder)+2)
	lines = append(lines, FFMetadataHeader)

	for _, tag := range m.TagOrder {
		value, exists := m.Tags[tag]
		if !exists {
			continue
		}
		lines = append(lines, EscapeFFMetadataValue(tag)+"="+EscapeFFMetadataValue(value))
	}

	if m.Sections != "" {
		lines = append(lines, strings.TrimRight(m.Sections, "\n"))
	}

	return strings.Join(lines, "\n") + "\n"
}

// SetTag sets the value of a tag, appending it to TagOrder if it is new.
func (m *FFMetadata) SetTag(tag string, value string) {
	if m.Tags == nil {
		m.Tags = make(map[string]string)
	}
	if _, exists := m.Tags[tag]; !exists {
		m.TagOrder = append(m.TagOrder, tag)
	}
	m.Tags[tag] = value
}

// SyncTagOrder aligns TagOrder with Tags after the map was modified directly:
// removed tags are dropped and newly added tags are appended in sorted order.
func (m *FFMetadata) SyncTagOrder() {
	order := make([]string, 0, len(m.Tags))
	known := make(map[string]bool, len(m.Tags))

	for _, tag := range m.TagOrder {
		if _, exists := m.Tags[tag]; !exists || known[tag] {
			continue
		}
		known[tag] = true
		order = append(order, tag)
	}

	added := make([]string, 0)
	for tag := range m.Tags {
		if !known[tag] {
			added = append(added, tag)
		}
	}
	slices.Sort(added)

	m.TagOrder = append(order, added...)
}

// Clone returns a deep copy of the metadata.
func (m *FFMetadata) Clone() *FFMetadata {
	return &FFMetadata{
		Tags:     maps.Clone(m.Tags),
		TagOrder: slices.Clone(m.TagOrder),
		Sections: m.Sections,
	}
}

// EscapeFFMetadataValue escapes the characters that are special in ffmetadata
// documents, so that the value survives a write/read round trip.
func EscapeFFMetadataValue(value string) string {
	var sb strings.Builder
	for _, r := range value {
		if strings.ContainsRune("=;#\\\n", r) {
			sb.WriteRune('\\')
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

// UnescapeFFMetadataValue reverses EscapeFFMetadataValue.
func UnescapeFFMetadataValue(value string) string {
	var sb strings.Builder
	escaped := false

	for _, r := range value {
		if !escaped && r == '\\' {
			escaped = true
			continue
		}
		escaped = false
		sb.WriteRune(r)
	}

	if escaped {
		sb.WriteRune('\\')
	}

	return sb.String()
}

func endsWithEscape(line string) bool {
	backslashes := 0
	for i := len(line) - 1; i >= 0 && line[i] == '\\'; i-- {
		backslashes++
	}
	return backslashes%2 == 1
}

func indexUnescaped(line string, char byte) int {
	for i := 0; i < len(line); i++ {
		if line[i] == '\\' {
			i++
			continue
		}
		if line[i] == char {
			return i
		}
	}
	return -1
}
