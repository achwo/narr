package utils

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// ApplyRegex applies a regular expression pattern to an input string and formats the captured groups
// using the provided format string. The format string should contain %s placeholders that match
// the number of capture groups in the regex pattern.
//
// Parameters:
//   - input: The input string to apply the regex pattern to
//   - regex: A compiled regular expression pattern with capture groups
//   - format: A format string containing %s placeholders for each capture group
//
// Returns:
//   - string: The formatted result using captured groups, or the original input if there's an error
//   - error: An error if the input doesn't match the pattern or if the number of capture groups
//     doesn't match the format string placeholders
func ApplyRegex(input string, regex *regexp.Regexp, format string) (string, error) {
	expectedMatchLen := strings.Count(format, "%s")

	if !regex.MatchString(input) {
		return input, nil
	}

	matches := regex.FindStringSubmatch(input)
	if len(matches) != expectedMatchLen+1 {
		return input, fmt.Errorf(
			"expected %d matches based on the format string, but got %d matches for input '%s' using regex '%s'",
			expectedMatchLen,
			len(matches)-1,
			input,
			regex.String(),
		)
	}

	captureGroups := matches[1:]
	args := make([]interface{}, len(captureGroups))
	for i, v := range captureGroups {
		args[i] = v
	}
	newValue := fmt.Sprintf(format, args...)

	return newValue, nil
}

// SanitizePathComponent replaces unallowed symbols with _ from a path component
func SanitizePathComponent(s string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9 _.\-']`)
	return re.ReplaceAllString(s, "_")
}

func ReplaceDirAndExt(file string, dir string, ext string) string {
	fileName := filepath.Base(file)
	fileName = strings.TrimSuffix(fileName, filepath.Ext(file)) + ".m4a"
	return path.Join(dir, fileName)
}
