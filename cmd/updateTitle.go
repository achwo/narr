package cmd

import (
	"errors"
	"fmt"
	"github.com/achwo/narr/utils"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

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
		metadata, err := utils.ReadMetadata(file)
		if err != nil {
			continue
		}
		metadata = utils.UpdateMetadataTags(metadata, []string{"album", "title"}, regex, "Folge %s: %s")
		outputFile := file + ".tmp.m4b"

		if err := utils.WriteMetadata(file, outputFile, metadata); err != nil {
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
