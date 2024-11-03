// Package utils provides utility functions for file operations and audio file handling
package utils

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// OSAudioFileProvider implements audio file discovery functionality using the OS filesystem
type OSAudioFileProvider struct{}

// AudioFiles returns a list of M4A audio files found at the given path
func (p *OSAudioFileProvider) AudioFiles(fullPath string) ([]string, error) {
	return GetFilesByExtension(fullPath, ".m4a")
}

// GetValidFilePathFromArgs retrieves and validates a file path from command line arguments.
// It returns an error if the path at the given index doesn't exist or is a directory.
func GetValidFilePathFromArgs(args []string, index int) (string, error) {
	path, err := GetValidFullpathFromArgs(args, index)
	if err != nil {
		return "", err
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if fileInfo.IsDir() {
		return "", fmt.Errorf("%s is not a directory", path)
	}

	return path, nil
}

// GetValidDirPathFromArgs retrieves and validates a directory path from command line arguments.
// It returns an error if the path at the given index doesn't exist or is not a directory.
func GetValidDirPathFromArgs(args []string, index int) (string, error) {
	path, err := GetValidFullpathFromArgs(args, index)
	if err != nil {
		return "", err
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if !fileInfo.IsDir() {
		return "", fmt.Errorf("%s is not a directory", path)
	}

	return path, nil
}

// GetValidFullpathFromArgs retrieves a path from command line arguments and converts it to an absolute path.
// It returns an error if the index is out of bounds or the path conversion fails.
func GetValidFullpathFromArgs(args []string, index int) (string, error) {
	if len(args) < index+1 {
		return "", errors.New("you must specify a path")
	}

	path := args[0]
	fullpath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path of %s: %w", path, err)
	}

	return fullpath, nil
}

// GetFilesByExtension walks through a directory tree and returns all files with the specified extension.
// It returns an error if there are any issues accessing the filesystem during the walk.
func GetFilesByExtension(fullpath string, extension string) ([]string, error) {
	var m4bFiles []string

	err := filepath.WalkDir(fullpath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to access %s: %w", path, err)
		}

		if filepath.Ext(path) == extension {
			m4bFiles = append(m4bFiles, path)
		}
		return nil
	})

	return m4bFiles, err
}
