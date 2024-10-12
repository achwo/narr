package utils

import (
	"fmt"
	"regexp"
	"strings"
)

func GetMetadataTagValue(metadata string, tag string) (string, error) {
	fullTag := tag + "="
	lines := strings.Split(metadata, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, fullTag) {
			return strings.TrimPrefix(line, fullTag), nil
		}
	}
	return "", fmt.Errorf("metadata do not contain field %s", tag)
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
	for _, tag := range tags {
		currentValue, err := GetMetadataTagValue(metadata, tag)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if regex.MatchString(currentValue) {
			matches := regex.FindStringSubmatch(currentValue)

			if len(matches) == 3 {
				episode := matches[1]
				title := matches[2]
				newValue := fmt.Sprintf(format, episode, title)

				fullTag := tag + "="

				metadata = strings.ReplaceAll(metadata, fullTag+currentValue, fullTag+newValue)

				affectedLines = append(affectedLines, Diff{
					Tag:    tag,
					Before: currentValue,
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
