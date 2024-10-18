package m4b

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/achwo/narr/m4b"
	"github.com/achwo/narr/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const configFileName = "narr.yaml"

var metadataProvider m4b.MetadataProvider = &utils.FFmpegMetadataProvider{}
var audioFileProvider m4b.AudioFileProvider = &utils.OSAudioFileProvider{}

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

		emptyConfig := m4b.ProjectConfig{
			AudioFilePath: "",
			CoverPath:     "",
			HasChapters:   false,
			MetadataRules: []m4b.MetadataRule{},
			ChapterRules:  []m4b.ChapterRule{},
			OutputRules:   []m4b.OutputRule{},
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

var checkCmd = &cobra.Command{
	Use:   "check <dir>",
	Short: "Check config for validity",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := utils.GetValidFullpathFromArgs(args, 0)
		if err != nil {
			return fmt.Errorf("could not resolve path %s: %w", args[0], err)
		}

		var fullpath string
		if strings.HasSuffix(path, configFileName) {
			fullpath = path
		} else {
			fullpath = filepath.Join(path, configFileName)
		}

		bytes, err := os.ReadFile(fullpath)
		if err != nil {
			return fmt.Errorf("could not read file %s: %w", fullpath, err)
		}

		var config m4b.ProjectConfig
		err = yaml.Unmarshal(bytes, &config)
		if err != nil {
			return fmt.Errorf("could not unmarshal file %s: %w", fullpath, err)
		}

		fmt.Println(config)

		return nil
	},
}

var chaptersCmd = &cobra.Command{
	Use:   "chapters <dir>",
	Short: "Show chapters with applied rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := utils.GetValidFullpathFromArgs(args, 0)
		if err != nil {
			return fmt.Errorf("could not resolve path %s: %w", args[0], err)
		}

		project, err := m4b.NewProjectFromPath(path, audioFileProvider, metadataProvider)
		if err != nil {
			return err
		}

		chaptersContent, err := project.ShowChapters()
		if err != nil {
			return fmt.Errorf("Could not get chapters: %w", err)
		}

		fmt.Println(chaptersContent)

		return nil
	},
}

var metadataCmd = &cobra.Command{
	Use:   "metadata <dir>",
	Short: "Show metadata with applied rules",
	Long: `Show metadata with applied rules
	
	Uses the first file for metadata.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := utils.GetValidFullpathFromArgs(args, 0)
		if err != nil {
			return fmt.Errorf("could not resolve path %s: %w", args[0], err)
		}

		project, err := m4b.NewProjectFromPath(path, audioFileProvider, metadataProvider)
		if err != nil {
			return fmt.Errorf("could not load config %s: %w", path, err)
		}

		metadata, err := project.ShowMetadata()
		if err != nil {
			return fmt.Errorf("Could not get metadata: %w", err)
		}

		fmt.Println(metadata)

		return nil
	},
}

func init() {
	M4bCmd.AddCommand(configCmd)
	configCmd.AddCommand(generateCmd)
	configCmd.AddCommand(checkCmd)
	checkCmd.AddCommand(chaptersCmd)
	checkCmd.AddCommand(metadataCmd)
}
