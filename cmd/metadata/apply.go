package metadata

import (
	"fmt"

	"github.com/achwo/narr/m4b"
	"github.com/achwo/narr/utils"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:          "apply [path]",
	Short:        "Apply the metadataRules of a narr project to its audio files",
	SilenceUsage: true,
	Long: `Apply writes the metadataRules of a narr project config directly to the
audio files of the project, without converting them to m4b.

Files with more than one hard link are skipped, as writing them would also
change the other copies.`,
	Example: "narr metadata apply [path]",
	RunE: func(cmd *cobra.Command, args []string) error {
		recursive, _ := cmd.Flags().GetBool("recursive")
		dryRun, _ := cmd.Flags().GetBool("dryRun")
		verbose, _ := cmd.Flags().GetBool("verbose")

		path, err := utils.GetValidFullpathFromArgs(args, 0)
		if err != nil {
			return fmt.Errorf("could not resolve path: %w", err)
		}

		projects, err := m4b.NewProjectsByArgs(path, recursive)
		if err != nil {
			return fmt.Errorf("could not create project(s): %w", err)
		}

		applier := m4b.NewMetadataApplier()
		var results []m4b.FileResult

		for _, project := range projects {
			fmt.Printf("\nRunning on %s\n", project.Config.ProjectPath)

			projectResults, err := applier.ApplyToProject(project.Config, m4b.ApplyOptions{
				DryRun:  dryRun,
				Verbose: verbose,
			})
			results = append(results, projectResults...)

			if err != nil {
				return fmt.Errorf("could not apply metadata rules: %w", err)
			}
		}

		printSummary(results)

		return nil
	},
}

func printSummary(results []m4b.FileResult) {
	counts := make(map[m4b.ApplyStatus]int)
	for _, result := range results {
		counts[result.Status]++
	}

	fmt.Printf(
		"\n%d files: %d updated, %d unchanged, %d with changes not written (dry run), %d skipped (hard links)\n",
		len(results),
		counts[m4b.StatusUpdated],
		counts[m4b.StatusUnchanged],
		counts[m4b.StatusDryRun],
		counts[m4b.StatusSkippedHardLink],
	)
}

func init() {
	MetadataCmd.AddCommand(applyCmd)
	applyCmd.Flags().BoolP("recursive", "r", false, "Find and process all projects below the path")
	applyCmd.Flags().Bool("dryRun", false, "Skip applying the changes")
	applyCmd.Flags().BoolP("verbose", "v", false, "More output")
}
