package m4b

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/achwo/narr/m4b"
	"github.com/achwo/narr/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var generateCmd = &cobra.Command{
	Use:   "generate <dir>",
	Short: "Generate an empty config",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := utils.GetValidFullpathFromArgs(args, 0)
		if err != nil {
			return fmt.Errorf("could not resolve path %s: %w", args[0], err)
		}

		emptyConfig := m4b.ProjectConfig{
			CoverPath:     "",
			HasChapters:   false,
			MetadataRules: []m4b.MetadataRule{},
			ChapterRules:  []m4b.ChapterRule{},
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
		recursive, _ := cmd.Flags().GetBool("recursive")

		path, err := utils.GetValidFullpathFromArgs(args, 0)
		if err != nil {
			return fmt.Errorf("could not resolve path %s: %w", args[0], err)
		}

		projects, err := m4b.NewProjectsByArgs(path, recursive)

		if err != nil {
			return fmt.Errorf("could not create project(s): %w", err)
		}

		for _, project := range projects {
			fmt.Printf("\n# Project %s\n", project.Config.ProjectPath)
			fmt.Println("## Tracks")
			tracks, err := project.Tracks()
			if err != nil {
				return fmt.Errorf("could not get tracks: %w", err)
			}

			for _, track := range tracks {
				fmt.Println(track.File)
			}

			if project.Config.HasChapters {
				fmt.Println("\n## Chapters")
				chaptersContent, err := project.Chapters()
				if err != nil {
					return fmt.Errorf("could not get chapters: %w", err)
				}
				fmt.Println(chaptersContent)
			}

			fmt.Println("\n## Metadata")
			metadata, err := project.Metadata()
			if err != nil {
				return fmt.Errorf("could not get metadata: %w", err)
			}
			fmt.Println(metadata)

			fmt.Println("\n## Filename")
			filename, err := project.Filename()
			if err != nil {
				return fmt.Errorf("could not get filename: %w", err)
			}
			fmt.Println(filename)
		}

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

		projects, err := m4b.NewProjectsFromPath(path)
		if err != nil {
			return err
		}
		if len(projects) < 1 {
			return fmt.Errorf("got no projects for path")
		}

		chaptersContent, err := projects[0].Chapters()
		if err != nil {
			return fmt.Errorf("could not get chapters: %w", err)
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

		projects, err := m4b.NewProjectsFromPath(path)
		if err != nil {
			return fmt.Errorf("could not create project: %w", err)
		}

		metadata, err := projects[0].Metadata()
		if err != nil {
			return fmt.Errorf("could not get metadata: %w", err)
		}

		fmt.Println(metadata)

		return nil
	},
}

var filenameCmd = &cobra.Command{
	Use:   "filename <dir>",
	Short: "Show filename with applied rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := utils.GetValidFullpathFromArgs(args, 0)
		if err != nil {
			return fmt.Errorf("could not resolve path %s: %w", args[0], err)
		}

		projects, err := m4b.NewProjectsFromPath(path)
		if err != nil {
			return fmt.Errorf("could not load config %s: %w", path, err)
		}
		if len(projects) < 1 {
			return fmt.Errorf("got no projects for path")
		}

		filename, err := projects[0].Filename()
		if err != nil {
			return fmt.Errorf("could not get filename: %w", err)
		}

		fmt.Println(filename)

		return nil
	},
}

var filesCmd = &cobra.Command{
	Use:   "files <dir>",
	Short: "Show input files in processing order",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := utils.GetValidFullpathFromArgs(args, 0)
		if err != nil {
			return fmt.Errorf("could not resolve path %s: %w", args[0], err)
		}

		projects, err := m4b.NewProjectsFromPath(path)
		if err != nil {
			return fmt.Errorf("could not load config %s: %w", path, err)
		}
		if len(projects) < 1 {
			return fmt.Errorf("got no projects for path")
		}

		tracks, err := projects[0].Tracks()
		if err != nil {
			return fmt.Errorf("could not get tracks: %w", err)
		}

		for _, track := range tracks {
			fmt.Println(track.File)
		}

		return nil
	},
}

func init() {
	M4bCmd.AddCommand(generateCmd)
	M4bCmd.AddCommand(checkCmd)
	checkCmd.AddCommand(chaptersCmd)
	checkCmd.AddCommand(metadataCmd)
	checkCmd.AddCommand(filenameCmd)
	checkCmd.AddCommand(filesCmd)
}
