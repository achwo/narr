package m4b

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type FFmpegAudioProcessor struct {
	Command Command
}

func (p *FFmpegAudioProcessor) ToM4A(files []string, outputPath string) ([]string, error) {
	outputFiles := make([]string, 0, len(files))
	for _, file := range files {
		fileName := filepath.Base(file)
		out := path.Join(outputPath, fileName)
		outputFiles = append(outputFiles, out)
		cmd := p.Command.Create("ffmpeg", "-i", file, "-c", "copy", "-c:a", "aac_at", out)

		var outBuf bytes.Buffer
		err := cmd.Run(&outBuf, &outBuf)
		if err != nil {
			fmt.Println(outBuf.String())
			return nil, fmt.Errorf("could not convert file %s:, %w", out, err)
		}
		fmt.Print(".")
	}
	fmt.Println()
	return outputFiles, nil
}

func (p *FFmpegAudioProcessor) Concat(files []string, filelistFile string, outputPath string) (string, error) {
	err := os.WriteFile(filelistFile, []byte(p.filelistFileContent(files)), 0600)
	if err != nil {
		return "", fmt.Errorf("could not write filelist file: %w", err)
	}

	outputFilepath := filepath.Join(outputPath, "concat.m4b")

	cmd := p.Command.Create(
		"ffmpeg",
		"-f",
		"concat",
		"-safe",
		"0",
		"-i",
		filelistFile,
		"-c",
		"copy",
		"-vn",
		outputFilepath,
	)
	var outBuf bytes.Buffer
	err = cmd.Run(&outBuf, &outBuf)
	if err != nil {
		fmt.Println(outBuf.String())
		return "", fmt.Errorf("could not concat files: %w", err)
	}

	return outputFilepath, nil
}

func (p *FFmpegAudioProcessor) filelistFileContent(files []string) string {
	var sb strings.Builder
	for _, file := range files {
		fmt.Fprintf(&sb, "file '%s'\n", file)
	}
	return sb.String()
}

func (p *FFmpegAudioProcessor) AddChapters(m4bFile string, chapters string) error {
	err := p.createChaptersFile(m4bFile, chapters)
	if err != nil {
		return fmt.Errorf("Could not create chapters file: %w", err)
	}

	cmd := p.Command.Create("mp4chaps", "--import", m4bFile)
	var outBuf bytes.Buffer
	err = cmd.Run(&outBuf, &outBuf)
	if err != nil {
		fmt.Println(outBuf.String())
		return fmt.Errorf("could not import chapters: %w", err)
	}

	// m4bfilename.chapters.txt
	return nil
}

func (p *FFmpegAudioProcessor) createChaptersFile(m4bFile string, chapters string) error {
	withoutExtension := strings.TrimSuffix(m4bFile, filepath.Ext(m4bFile))
	chaptersFileName := withoutExtension + ".chapters.txt"
	if err := os.MkdirAll(filepath.Dir(chaptersFileName), 0755); err != nil {
		return err
	}
	return os.WriteFile(chaptersFileName, []byte(chapters), 0600)
}

func (p *FFmpegAudioProcessor) AddCover(m4bFile string, coverFile string) error {
	return nil
}

func (p *FFmpegAudioProcessor) AddMetadata(m4bFile string, metadata string) error {
	return nil
}
