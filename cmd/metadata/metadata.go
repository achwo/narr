package metadata

import (
	"github.com/spf13/cobra"
)

// MetadataCmd allows viewing and modifying audio file metadata
var MetadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "Work with audio file metadata",
	Long: `The metadata command provides functionality to work with audio file metadata.
It allows viewing and modifying metadata tags like title, artist, album, etc.
for various audio file formats.`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}
