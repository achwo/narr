package m4b

import (
	"errors"
	"fmt"

	"github.com/achwo/narr/m4b"
	"github.com/achwo/narr/utils"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Convert to m4b",
	RunE: func(cmd *cobra.Command, args []string) error {
		recursive, _ := cmd.Flags().GetBool("recursive")
		multi, _ := cmd.Flags().GetBool("multi")
		if recursive && multi {
			return errors.New("cannot run both recursive and multi at the same time")
		}

		path, err := utils.GetValidFullpathFromArgs(args, 0)
		if err != nil {
			return fmt.Errorf("could not resolve path %s: %w", args[0], err)
		}

		var projects []*m4b.Project
		if recursive {
			projects, err = m4b.NewRecursiveProjectsFromPath(path, audioFileProvider, audioConverter)
		} else if multi {
			projects, err = m4b.NewMultiProjectsFromPath(path, audioFileProvider, audioConverter)
		} else {
			var project *m4b.Project
			project, err = m4b.NewProjectFromPath(path, audioFileProvider, audioConverter)
			projects = append(projects, project)
		}

		if err != nil {
			return fmt.Errorf("could not create project(s): %w", err)
		}

		outputPaths := make([]string, len(projects))

		for _, project := range projects {
			fmt.Printf("\nRunning on %s\n", project.Config.ProjectPath)
			outputPath, err := project.ConvertToM4B()
			if err != nil {
				return fmt.Errorf("could not convert to m4b: %w", err)
			}

			outputPaths = append(outputPaths, outputPath)
		}

		for _, outputPath := range outputPaths {
			fmt.Println(outputPath)
		}
		return nil
	},
}

func init() {
	M4bCmd.AddCommand(runCmd)
	runCmd.Flags().BoolP("recursive", "r", false, "Search for projects in child dirs recursively and run them all")
	runCmd.Flags().BoolP("multi", "m", false, "Use provided config for all immediate child dirs and run them all")
}
