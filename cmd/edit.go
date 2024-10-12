package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/achwo/narr/utils"
	"github.com/spf13/cobra"
)

var dryRun bool
var tags []string
var format string
var verbose bool

var editCmd = &cobra.Command{
	Use:     "edit",
	Short:   "The metadata tags for given audio file",
	Example: "narr metadata edit [file]",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := utils.GetValidFilePathFromArgs(args, 0)
		if err != nil {
			return fmt.Errorf("could not resolve path %s: %w", args[0], err)
		}

		metadata, err := utils.ReadMetadata(path)
		if err != nil {
			return fmt.Errorf("failed to read metadata of %s: %w", path, err)
		}

		updatedMetadata, diffs := utils.UpdateMetadataTags(metadata, tags, regex, format)

		if verbose {
			fmt.Println("Metadata after update:")
			fmt.Println(updatedMetadata)
		}

		for _, diff := range diffs {
			fmt.Println(diff.Tag, ":", diff.Before, "->", diff.After)
		}

		if dryRun {
			fmt.Println("Dry run: not changing the file")
			return nil
		}

		if verbose {
			fmt.Println("Changing metadata in file", path)
		}

		if err = utils.WriteMetadata(path, updatedMetadata, verbose); err != nil {
			return fmt.Errorf("could not write metadata: %w", err)
		}

		fmt.Println("Metadata successfully changed for file", path)

		return nil
	},
}

func init() {
	metadataCmd.AddCommand(editCmd)

	editCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Skip applying the changes")
	editCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "More output")
	editCmd.Flags().StringVar(&regexStr, "regex", "", "Regular expression to apply to album titles")
	editCmd.MarkFlagRequired("regex")
	editCmd.Flags().StringSliceVarP(&tags, "tag", "t", []string{}, "Specify metadata tags to edit")
	editCmd.MarkFlagRequired("tag")
	editCmd.Flags().StringVar(&format, "format", "", "format to be used with placeholders for the capture groups from the regex")
	editCmd.MarkFlagRequired("format")

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
}
