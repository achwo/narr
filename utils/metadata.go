package utils

import (
	"fmt"
	"regexp"
	"strings"
)

type TagWithValue struct {
	Tag   string
	Value string
}

func (t TagWithValue) Prefix() string {
	return fmt.Sprintf("%s=", t.Tag)
}

func (t TagWithValue) String() string {
	return fmt.Sprintf("%s=%s", t.Tag, t.Value)
}

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
		if regex.MatchString(currentValue.Value) {
			matches := regex.FindStringSubmatch(currentValue.Value)

			if len(matches) == 3 {
				episode := matches[1]
				title := matches[2]
				newValue := fmt.Sprintf(format, episode, title)

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
		}
	}
	return metadata, affectedLines
}

type Diff struct {
	Tag    string
	Before string
	After  string
}
