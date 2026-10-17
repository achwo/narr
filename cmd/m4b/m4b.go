package m4b

import (
	"fmt"
	"path/filepath"

	"github.com/achwo/narr/m4b"
	"github.com/achwo/narr/utils"
	"github.com/spf13/cobra"
)

// M4bCmd represents the m4b command that creates m4b files from a list of m4a files
var M4bCmd = &cobra.Command{
	Use:   "m4b",
	Short: "Create m4b files from a list of m4a",
}

func init() {
	M4bCmd.PersistentFlags().BoolP("recursive", "r", false, "Search for projects in child dirs recursively and run them all")
}

func addOutFlag(cmd *cobra.Command) {
	cmd.Flags().String("out", "", "Directory below which <album_artist>/<album>/<album>.m4b is written (default ~/narr)")
}

func projectsFromArgs(cmd *cobra.Command, args []string) ([]*m4b.Project, error) {
	recursive, _ := cmd.Flags().GetBool("recursive")

	path, err := utils.GetValidFullpathFromArgs(args, 0)
	if err != nil {
		return nil, fmt.Errorf("could not resolve path %s: %w", args[0], err)
	}

	projects, err := m4b.NewProjectsByArgs(path, recursive)
	if err != nil {
		return nil, fmt.Errorf("could not create project(s): %w", err)
	}

	outputRoot, err := outputRootFromFlag(cmd)
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		project.Config.OutputRoot = outputRoot
	}

	return projects, nil
}

func outputRootFromFlag(cmd *cobra.Command) (string, error) {
	out, _ := cmd.Flags().GetString("out")
	if out == "" {
		return "", nil
	}

	outputRoot, err := filepath.Abs(out)
	if err != nil {
		return "", fmt.Errorf("could not resolve --out %s: %w", out, err)
	}
	return outputRoot, nil
}
