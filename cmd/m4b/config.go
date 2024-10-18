package m4b

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/achwo/narr/config"
	"github.com/achwo/narr/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage m4b project config",
	Run:   func(cmd *cobra.Command, args []string) {},
}

var generateCmd = &cobra.Command{
	Use:   "generate <dir>",
	Short: "Generate an empty config",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := utils.GetValidFullpathFromArgs(args, 0)
		if err != nil {
			return fmt.Errorf("could not resolve path %s: %w", args[0], err)
		}

		emptyConfig := config.ProjectConfig{
			AudioFilePath: "",
			CoverPath:     "",
			HasChapters:   false,
			MetadataRules: []config.RegexRule{{}},
			ChapterRules:  []config.RegexRule{{}},
			OutputRules:   []config.RegexRule{{}},
		}

		jsonBytes, err := yaml.Marshal(emptyConfig)
		if err != nil {
			return fmt.Errorf("could not marshal empty config, %w", err)
		}

		fullpath := filepath.Join(path, "narr.yaml")
		fmt.Println("Writing config to", fullpath)
		os.WriteFile(fullpath, jsonBytes, 0644)
		return nil
	},
}

func init() {
	M4bCmd.AddCommand(configCmd)
	configCmd.AddCommand(generateCmd)
}
