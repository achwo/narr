package files

import (
	"fmt"
	"path/filepath"

	"github.com/achwo/narr/utils"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List files within or below given path",
	Example: "narr file list [path]",
	RunE: func(cmd *cobra.Command, args []string) error {
		noPath, _ := cmd.Flags().GetBool("noPath")
		path, err := utils.GetValidFullpathFromArgs(args, 0)
		if err != nil {
			return fmt.Errorf("could not resolve path %s: %w", args[0], err)
		}

		files, err := utils.GetFilesByExtensions(path, []string{".m4b"})
		if err != nil {
			return fmt.Errorf("failed to read files within %s: %w", path, err)
		}

		for _, file := range files {
			if noPath {
				fmt.Println(filepath.Base(file))
			} else {
				fmt.Println(file)
			}
		}

		return nil
	},
}

func init() {
	FilesCmd.AddCommand(listCmd)

	listCmd.Flags().Bool("noPath", false, "Print only the file name")
}
