// Package utils provides utility functions for file operations and audio file handling
package utils

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
)

// OSAudioFileProvider implements audio file discovery functionality using the OS filesystem
type OSAudioFileProvider struct{}

// AudioFiles returns a list of audio files found at the given path
func (p *OSAudioFileProvider) AudioFiles(fullPath string) ([]string, error) {
	return GetFilesByExtensions(fullPath, []string{".m4a", ".mp3", ".flac"})
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

// HardLinkCount returns the number of hard links pointing to the file at path.
// A count greater than one means that writing to the file also changes every
// other path linked to the same inode.
func HardLinkCount(path string) (uint64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("could not stat %s: %w", path, err)
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, fmt.Errorf("could not determine hard link count of %s", path)
	}

	return uint64(stat.Nlink), nil
}

// MoveFile moves src to dst. When src and dst are on different volumes, where a
// rename fails with EXDEV, it copies src next to dst, renames the copy to dst and
// removes src, so that dst never holds a partial file.
func MoveFile(src string, dst string) error {
	err := os.Rename(src, dst)
	if !errors.Is(err, syscall.EXDEV) {
		return err
	}

	if err := copyFile(src, dst); err != nil {
		return err
	}

	if err := os.Remove(src); err != nil {
		return fmt.Errorf("could not remove %s after copying it to %s: %w", src, dst, err)
	}
	return nil
}

func copyFile(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("could not open %s: %w", src, err)
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return fmt.Errorf("could not stat %s: %w", src, err)
	}

	out, err := os.CreateTemp(filepath.Dir(dst), "."+filepath.Base(dst)+".*")
	if err != nil {
		return fmt.Errorf("could not create a copy next to %s: %w", dst, err)
	}

	if err := writeCopy(out, in, info.Mode().Perm()); err != nil {
		os.Remove(out.Name())
		return fmt.Errorf("could not copy %s to %s: %w", src, out.Name(), err)
	}

	if err := os.Rename(out.Name(), dst); err != nil {
		os.Remove(out.Name())
		return fmt.Errorf("could not rename %s to %s: %w", out.Name(), dst, err)
	}
	return nil
}

func writeCopy(out *os.File, in io.Reader, perm fs.FileMode) error {
	_, err := io.Copy(out, in)
	if err == nil {
		err = out.Chmod(perm)
	}
	if err == nil {
		err = out.Sync()
	}
	return errors.Join(err, out.Close())
}
