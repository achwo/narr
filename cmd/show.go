package cmd

import (
	"fmt"

	"github.com/achwo/narr/utils"

	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:     "show",
	Short:   "Show metadata for given file",
	Example: "narr metadata show [file]",
	RunE: func(cmd *cobra.Command, args []string) error {
		tags, _ := cmd.Flags().GetStringSlice("tag")

		path, err := utils.GetValidFullpathFromArgs(args, 0)
		if err != nil {
			return fmt.Errorf("could not resolve path %s: %w", args[0], err)
		}

		files, err := utils.GetFilesByExtension(path, ".m4b")
		if err != nil {
			return fmt.Errorf("failed to read files within %s: %w", path, err)
		}

		for _, file := range files {
			metadata, err := utils.ReadMetadata(file)
			if err != nil {
				return fmt.Errorf("failed to read metadata of %s: %w", file, err)
			}

			fmt.Println("#", file)
			if len(tags) > 0 {
				tagValues := utils.GetMetadataTagValues(metadata, tags)

				for _, value := range tagValues {
					fmt.Println(value.String())
				}
			} else {
				fmt.Println(metadata)
			}
			fmt.Println("")
		}

		return nil
	},
}

func init() {
	metadataCmd.AddCommand(showCmd)
	showCmd.Flags().StringSliceP("tag", "t", []string{}, "Specify metadata tags to show")
}
