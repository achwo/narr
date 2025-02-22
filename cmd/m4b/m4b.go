package m4b

import (
	"github.com/spf13/cobra"
)

// M4bCmd represents the m4b command that creates m4b files from a list of m4a files
var M4bCmd = &cobra.Command{
	Use:   "m4b",
	Short: "Create m4b files from a list of m4a",
}

func init() {
	M4bCmd.AddCommand(runCmd)
	M4bCmd.PersistentFlags().BoolP("recursive", "r", false, "Search for projects in child dirs recursively and run them all")
}
