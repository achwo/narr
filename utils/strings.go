package utils

import (
	"fmt"
	"regexp"
	"strings"
)

func ApplyRegex(input string, regex *regexp.Regexp, format string) (string, error) {
	expectedMatchLen := strings.Count(format, "%s")

	if !regex.MatchString(input) {
		return input, fmt.Errorf("input '%s' does not match regex pattern '%s'", input, regex.String())
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
