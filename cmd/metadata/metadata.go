package metadata

import (
	"github.com/spf13/cobra"
)

var MetadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "Work with audio file metadata",
	Run: func(cmd *cobra.Command, args []string) {
	},
}
