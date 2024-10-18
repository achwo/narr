package m4b

import (
	"github.com/spf13/cobra"
)

var M4bCmd = &cobra.Command{
	Use:   "m4b",
	Short: "Create m4b files from a list of m4a",
	Run: func(cmd *cobra.Command, args []string) {
	},
}
