package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func ReadMetadata(path string) (string, error) {
	extractCmd := exec.Command("ffmpeg", "-i", path, "-f", "ffmetadata", "-")

	var metadata bytes.Buffer
	extractCmd.Stdout = &metadata

	if err := extractCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to extract metadata for file %s: %w", path, err)
	}

	return metadata.String(), nil
}

// WriteMetadata updates the metadata in the file
func WriteMetadata(file string, metadata string, verbose bool) error {
	tmpFile := file + ".tmp" + filepath.Ext(file)

	err := WriteMetadataO(file, tmpFile, metadata, verbose)
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
func WriteMetadataO(inputFile string, outputFile string, metadata string, verbose bool) error {
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
