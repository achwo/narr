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

		path, err := utils.GetValidFilePathFromArgs(args, 0)
		if err != nil {
			return fmt.Errorf("could not resolve path %s: %w", args[0], err)
		}

		metadata, err := utils.ReadMetadata(path)
		if err != nil {
			return fmt.Errorf("failed to read metadata of %s: %w", path, err)
		}

		if len(tags) > 0 {
			tagValues := utils.GetMetadataTagValues(metadata, tags)

			for _, value := range tagValues {
				fmt.Println(value.String())
			}

			return nil
		}

		fmt.Println(metadata)

		return nil
	},
}

func init() {
	metadataCmd.AddCommand(showCmd)
	showCmd.Flags().StringSliceP("tag", "t", []string{}, "Specify metadata tags to show")
}
