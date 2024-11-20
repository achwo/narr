package files

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/achwo/narr/utils"
	"github.com/spf13/cobra"
)

// renameCmd represents the rename command
var renameCmd = &cobra.Command{
	Use:   "rename",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dryRun")
		// verbose, _ := cmd.Flags().GetBool("verbose")
		// dir, _ := cmd.Flags().GetBool("dir")
		regexStr, _ := cmd.Flags().GetString("regex")
		format, _ := cmd.Flags().GetString("format")

		path, err := utils.GetValidFullpathFromArgs(args, 0)
		if err != nil {
			return fmt.Errorf("could not resolve path %s: %w", args[0], err)
		}

		regex, err := regexp.Compile(regexStr)
		if err != nil {
			return fmt.Errorf("error compiling regex: %w", err)
		}

		files, err := utils.GetFilesByExtensions(path, []string{".m4b"})
		if err != nil {
			return fmt.Errorf("failed to read files within %s: %w", path, err)
		}

		for _, file := range files {
			basename := filepath.Base(file)
			newName, err := utils.ApplyRegex(basename, regex, format)
			if err != nil {
				continue
			}

			newPath := filepath.Join(filepath.Dir(file), newName)

			println(basename, "->", newName)
			if !dryRun {
				os.Rename(file, newPath)
			}
		}

		return nil
	},
}

func init() {
	FilesCmd.AddCommand(renameCmd)
	renameCmd.Flags().Bool("dryRun", false, "Skip applying the changes")
	renameCmd.Flags().BoolP("verbose", "v", false, "More output")
	// renameCmd.Flags().Bool("dir", false, "Rename directories instead of files")
	renameCmd.Flags().String("regex", "", "Regular expression to apply to album titles")
	renameCmd.MarkFlagRequired("regex")
	renameCmd.Flags().String("format", "", "Format to be used with placeholders for the capture groups from the regex")
	renameCmd.MarkFlagRequired("format")
}
