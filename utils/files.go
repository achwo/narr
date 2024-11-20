// Package utils provides utility functions for file operations and audio file handling
package utils

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// OSAudioFileProvider implements audio file discovery functionality using the OS filesystem
type OSAudioFileProvider struct{}

// AudioFiles returns a list of M4A audio files found at the given path
func (p *OSAudioFileProvider) AudioFiles(fullPath string) ([]string, error) {
	return GetFilesByExtensions(fullPath, []string{".m4a", ".mp3"})
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
		ex, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current path: %w", err)
		}

		return ex, nil
	}

	path := args[0]
	fullpath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path of %s: %w", path, err)
	}

	return fullpath, nil
}

// GetFilesByExtensions walks through a directory tree and returns all files with any of the
// specified extensions.
// It returns an error if there are any issues accessing the filesystem during the walk.
func GetFilesByExtensions(fullpath string, extensions []string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(fullpath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to access %s: %w", path, err)
		}

		if slices.Contains(extensions, filepath.Ext(path)) {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

func GetAllFilesByName(basepath string, name string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(basepath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to access %s: %w", path, err)
		}

		if d.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, name) {
			files = append(files, path)
		}

		return nil
	})
	return files, err
}
