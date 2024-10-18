package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
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

type FFmpegMetadataProvider struct{}

func (m *FFmpegMetadataProvider) ReadMetadata(path string) (string, error) {
	extractCmd := exec.Command("ffmpeg", "-i", path, "-f", "ffmetadata", "-")

	var metadata bytes.Buffer
	extractCmd.Stdout = &metadata

	if err := extractCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to extract metadata for file %s: %w", path, err)
	}

	return metadata.String(), nil
}

// WriteMetadata updates the metadata in the file
func (m *FFmpegMetadataProvider) WriteMetadata(file string, metadata string, verbose bool) error {
	tmpFile := file + ".tmp" + filepath.Ext(file)

	err := m.WriteMetadataO(file, tmpFile, metadata, verbose)
	if err != nil {
		return fmt.Errorf("could not write metadata: %w", err)
	}

	err = os.Rename(tmpFile, file)
	if err != nil {
		return fmt.Errorf("could not rename temp file to output file: %w", err)
	}

	return nil
}

// WriteMetadataO is like WriteMetadata with explicit output file
func (m *FFmpegMetadataProvider) WriteMetadataO(inputFile string, outputFile string, metadata string, verbose bool) error {
	writeCmd := exec.Command("ffmpeg", "-i", inputFile, "-f", "ffmetadata", "-i", "-", "-map_metadata", "1", "-c", "copy", outputFile)

	writeCmd.Stdin = bytes.NewReader([]byte(metadata))
	var outBuf, errBuf bytes.Buffer
	writeCmd.Stdout = &outBuf
	writeCmd.Stderr = &errBuf

	if verbose {
		fmt.Println(writeCmd.String())
	}

	err := writeCmd.Run()

	if verbose {
		fmt.Printf("Command output:\n%s\n", outBuf.String())
		fmt.Printf("Command error output:\n%s\n", errBuf.String())
	}

	if err != nil {
		return fmt.Errorf("ffmpeg command failed: %v\n%s", err, errBuf.String())
	}
	return nil
}

func (m *FFmpegMetadataProvider) ReadTitleAndDuration(file string) (string, float64, error) {
	dataCmd := exec.Command(
		"ffprobe",
		"-v",
		"error",
		"-select_streams",
		"a:0",
		"-show_entries",
		"format=duration:format_tags=title",
		file,
	)

	var data bytes.Buffer
	dataCmd.Stdout = &data

	if err := dataCmd.Run(); err != nil {
		return "", 0, fmt.Errorf("failed to extract metadata for file %s: %w", file, err)
	}

	probeContent := data.String()

	durationRegex := regexp.MustCompile(`duration=([0-9]+\.[0-9]+)`)
	titleRegex := regexp.MustCompile(`TAG:title=(.+)`)

	titleMatch := titleRegex.FindStringSubmatch(probeContent)
	if len(titleMatch) < 2 {
		return "", 0, fmt.Errorf("title not found")
	}

	title := titleMatch[1]

	durationMatch := durationRegex.FindStringSubmatch(probeContent)
	if len(durationMatch) < 2 {
		return "", 0, fmt.Errorf("duration not found")
	}

	duration, err := strconv.ParseFloat(durationMatch[1], 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid duration value")
	}

	return title, duration, nil
}
