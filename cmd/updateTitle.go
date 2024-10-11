package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var regexStr string
var regex *regexp.Regexp

var updateTitleCmd = &cobra.Command{
	Use:   "updateTitle",
	Short: "Update the metadata album title of m4bs within a given folder with a given regex",
	Long: `Update the metadata album title of m4bs within a given folder recursively with a given regex.

	The regex should have two capture groups. The first should contain the episode number,
	the second the title.
	`,
	Example: `narr updateTitle --regex "^(\\d+)/(.+)$" "Die drei ???" folderWithM4B`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("you must specify a folder")
		}

		folder := args[0]
		fullpath, err := filepath.Abs(folder)
		if err != nil {
			return fmt.Errorf("failed to get absolute path of folder %s: %w", folder, err)
		}

		file, err := os.Stat(fullpath)
		if err != nil {
			return err
		}

		if !file.IsDir() {
			return fmt.Errorf("%s is not a directory", folder)
		}

		return updateTitle(fullpath)
	},
}

func init() {
	rootCmd.AddCommand(updateTitleCmd)

	updateTitleCmd.Flags().StringVar(&regexStr, "regex", "", "Regular expression to apply to album titles")
	updateTitleCmd.MarkFlagRequired("regex")

	cobra.OnInitialize(func() {
		if regexStr != "" {
			var err error
			regex, err = regexp.Compile(regexStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error compiling regex: %v\n", err)
				os.Exit(1)
			}
		}
	})
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateTitleCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateTitleCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func updateTitle(fullpath string) error {
	files, err := getM4BFiles(fullpath)
	if err != nil {
		return err
	}
	for _, file := range files {
		metadata, err := readMetadata(file)
		if err != nil {
			continue
		}
		metadata = updateMetadataAlbum(metadata)
		outputFile := file + ".tmp.m4b"

		if err := writeMetadata(file, outputFile, metadata); err != nil {
			return fmt.Errorf("failed to write metadata to %s: %w", outputFile, err)
		}
	}
	return nil
}

func getM4BFiles(fullpath string) ([]string, error) {
	var m4bFiles []string

	err := filepath.WalkDir(fullpath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to access %s: %w", path, err)
		}

		if filepath.Ext(path) == ".m4b" {
			m4bFiles = append(m4bFiles, path)
		}
		return nil
	})

	return m4bFiles, err
}

func readMetadata(path string) (string, error) {
	extractCmd := exec.Command("ffmpeg", "-i", path, "-f", "ffmetadata", "-")

	var oldMetadata bytes.Buffer
	extractCmd.Stdout = &oldMetadata

	if err := extractCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to extract metadata for file %s: %w", path, err)
	}

	fmt.Println("Extracted metadata:")
	fmt.Println(oldMetadata.String())

	return oldMetadata.String(), nil
}

func updateMetadataAlbum(metadata string) string {
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

func writeMetadata(inputFile string, outputFile string, metadata string) error {
	writeCmd := exec.Command("ffmpeg", "-i", inputFile, "-f", "ffmetadata", "-i", "-", "-map_metadata", "1", "-c", "copy", outputFile)

	writeCmd.Stdin = bytes.NewReader([]byte(metadata))

	return writeCmd.Run()
}
