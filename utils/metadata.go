package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func GetMetadataField(metadata string, field string) (string, error) {
	lines := strings.Split(metadata, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, field) {
			return line, nil
		}
	}
	return "", fmt.Errorf("metadata do not contain field %s", field)
}

func ReadMetadata(path string) (string, error) {
	extractCmd := exec.Command("ffmpeg", "-i", path, "-f", "ffmetadata", "-")

	var metadata bytes.Buffer
	extractCmd.Stdout = &metadata

	if err := extractCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to extract metadata for file %s: %w", path, err)
	}

	fmt.Println("Extracted metadata:")
	fmt.Println(metadata.String())

	return metadata.String(), nil
}

func UpdateMetadataAlbum(metadata string, regex *regexp.Regexp) string {
	if !strings.Contains(metadata, "album=") {
		fmt.Println("Warning: No album tag found.")
		return metadata
	}
	lines := strings.Split(metadata, "\n")

	tags := []string{"album", "title"}

	for i, line := range lines {
		for _, tag := range tags {
			fullTag := tag + "="
			if strings.HasPrefix(line, fullTag) {
				currentValue := strings.TrimPrefix(line, fullTag)

				if regex.MatchString(currentValue) {
					matches := regex.FindStringSubmatch(currentValue)

					if len(matches) == 3 {
						episode := matches[1]
						title := matches[2]
						currentValue = fmt.Sprintf("Folge %s: %s", episode, title)
					}
				}

				lines[i] = fullTag + currentValue
				break
			}
		}
	}
	newMetadata := strings.Join(lines, "\n")

	fmt.Println("Modified metadata:")
	fmt.Println(newMetadata)

	return newMetadata
}

func WriteMetadata(inputFile string, outputFile string, metadata string) error {
	writeCmd := exec.Command("ffmpeg", "-i", inputFile, "-f", "ffmetadata", "-i", "-", "-map_metadata", "1", "-c", "copy", outputFile)

	writeCmd.Stdin = bytes.NewReader([]byte(metadata))

	return writeCmd.Run()
}
