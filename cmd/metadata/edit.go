package metadata

import (
	"fmt"
	"regexp"

	"github.com/achwo/narr/m4b"
	"github.com/achwo/narr/utils"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:     "edit",
	Short:   "The metadata tags for given audio file",
	Example: "narr metadata edit [file]",
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dryRun")
		verbose, _ := cmd.Flags().GetBool("verbose")
		regexStr, _ := cmd.Flags().GetString("regex")
		tags, _ := cmd.Flags().GetStringSlice("tag")
		format, _ := cmd.Flags().GetString("format")

		path, err := utils.GetValidFullpathFromArgs(args, 0)
		if err != nil {
			return fmt.Errorf("could not resolve path %s: %w", args[0], err)
		}

		regex, err := regexp.Compile(regexStr)
		if err != nil {
			return fmt.Errorf("error compiling regex: %w", err)
		}

		files, err := utils.GetFilesByExtensions(path, []string{".m4b"})
		if err != nil {
			return fmt.Errorf("could not get files: %w", err)
		}

		audioProcesoor := &m4b.FFmpegAudioProcessor{}

		for _, file := range files {
			metadata, err := audioProcesoor.ReadMetadata(file)
			if err != nil {
				return fmt.Errorf("failed to read metadata of %s: %w", file, err)
			}

			updatedMetadata, diffs := utils.UpdateMetadataTags(metadata, tags, regex, format)

			if len(diffs) == 0 {
				if verbose {
					fmt.Println("#", file)
					fmt.Println("Nothing to do.")
				}
				continue
			}

			fmt.Println("#", file)
			if verbose {
				fmt.Println("Metadata after update:")
				fmt.Println(updatedMetadata)
			}

			for _, diff := range diffs {
				fmt.Println(diff.Tag, ":", diff.Before, "->", diff.After)
			}

			if dryRun {
				fmt.Println("")
				continue
			}

			if verbose {
				fmt.Println("Changing metadata in file", file)
			}

			if err = audioProcesoor.WriteMetadata(file, updatedMetadata, verbose); err != nil {
				return fmt.Errorf("could not write metadata: %w", err)
			}

			fmt.Println("Metadata successfully changed for file", file)
			fmt.Println("")
		}

		return nil
	},
}

func init() {
	MetadataCmd.AddCommand(editCmd)
	editCmd.Flags().Bool("dryRun", false, "Skip applying the changes")
	editCmd.Flags().BoolP("verbose", "v", false, "More output")
	editCmd.Flags().String("regex", "", "Regular expression to apply to album titles")
	editCmd.MarkFlagRequired("regex")
	editCmd.Flags().StringSliceP("tag", "t", []string{}, "Specify metadata tags to edit")
	editCmd.MarkFlagRequired("tag")
	editCmd.Flags().String("format", "", "Format to be used with placeholders for the capture groups from the regex")
	editCmd.MarkFlagRequired("format")
}
