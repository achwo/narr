package m4b

import (
	"fmt"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [dir]",
	Short: "Convert to m4b",
	Long: `Convert to m4b

Prints the m4b path of each project to stdout, one per line, and nothing else,
also for projects skipped as already completed. Progress and diagnostics go to
stderr.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		projects, err := projectsFromArgs(cmd, args)
		if err != nil {
			return err
		}

		for _, project := range projects {
			fmt.Fprintf(cmd.ErrOrStderr(), "\nRunning on %s\n", project.Config.ProjectPath)
			outputPath, err := project.ConvertToM4B()
			if err != nil {
				return fmt.Errorf("could not convert to m4b: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), outputPath)
		}

		return nil
	},
}

func init() {
	M4bCmd.AddCommand(runCmd)
	addOutFlag(runCmd)
}
