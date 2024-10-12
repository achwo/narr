package cmd

import (
	"fmt"
	"regexp"

	"github.com/achwo/narr/utils"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:     "edit",
	Short:   "The metadata tags for given audio file",
	Example: "narr metadata edit [file]",
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		verbose, _ := cmd.Flags().GetBool("verbose")
		regexStr, _ := cmd.Flags().GetString("regex")
		tags, _ := cmd.Flags().GetStringSlice("tag")
		format, _ := cmd.Flags().GetString("format")

		path, err := utils.GetValidFilePathFromArgs(args, 0)
		if err != nil {
			return fmt.Errorf("could not resolve path %s: %w", args[0], err)
		}

		metadata, err := utils.ReadMetadata(path)
		if err != nil {
			return fmt.Errorf("failed to read metadata of %s: %w", path, err)
		}

		regex, err := regexp.Compile(regexStr)
		if err != nil {
			return fmt.Errorf("error compiling regex: %v", err)
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

	editCmd.Flags().Bool("dry-run", false, "Skip applying the changes")
	editCmd.Flags().BoolP("verbose", "v", false, "More output")
	editCmd.Flags().String("regex", "", "Regular expression to apply to album titles")
	editCmd.MarkFlagRequired("regex")
	editCmd.Flags().StringSliceP("tag", "t", []string{}, "Specify metadata tags to edit")
	editCmd.MarkFlagRequired("tag")
	editCmd.Flags().String("format", "", "Format to be used with placeholders for the capture groups from the regex")
	editCmd.MarkFlagRequired("format")
}
