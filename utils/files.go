package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

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
