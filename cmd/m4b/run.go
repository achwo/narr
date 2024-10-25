package m4b

import (
	"fmt"

	"github.com/achwo/narr/m4b"
	"github.com/achwo/narr/utils"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Convert to m4b",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := utils.GetValidFullpathFromArgs(args, 0)
		if err != nil {
			return fmt.Errorf("could not resolve path %s: %w", args[0], err)
		}

		project, err := m4b.NewProjectFromPath(path, audioFileProvider, metadataProvider, audioConverter)

		outputPath, err := project.ConvertToM4B()
		if err != nil {
			return fmt.Errorf("could not convert to m4b: %w", err)
		}

		fmt.Println(outputPath)
		return nil
	},
}

func init() {
	M4bCmd.AddCommand(runCmd)
}
