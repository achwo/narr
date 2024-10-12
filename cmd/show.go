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
		path, err := utils.GetValidFilePathFromArgs(args, 0)
		if err != nil {
			return fmt.Errorf("could not resolve path %s: %w", args[0], err)
		}

		metadata, err := utils.ReadMetadata(path)
		if err != nil {
			return fmt.Errorf("failed to read metadata of %s: %w", path, err)
		}

		fmt.Println(metadata)

		return nil
	},
}

func init() {
	metadataCmd.AddCommand(showCmd)

	// showCmd.Flags().StringSliceVarP(&tags, "tag", "t", []string{}, "Specify metadata tags to edit")

}
