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
	if err := p.createChaptersFile(m4bFile, chapters); err != nil {
		return fmt.Errorf("Could not create chapters file: %w", err)
	}

	cmd := p.Command.Create("mp4chaps", "--import", m4bFile)
	var outBuf bytes.Buffer
	err := cmd.Run(&outBuf, &outBuf)
	if err != nil {
		fmt.Println(outBuf.String())
		return fmt.Errorf("could not import chapters: %w", err)
	}

	// m4bfilename.chapters.txt
	return nil
}

func (p *FFmpegAudioProcessor) createChaptersFile(m4bFile string, chapters string) error {
	chaptersFile := p.ChangeFileExtension(m4bFile, ".chapters.txt")
	if err := os.MkdirAll(filepath.Dir(chaptersFile), 0755); err != nil {
		return err
	}
	return os.WriteFile(chaptersFile, []byte(chapters), 0600)
}

func (p *FFmpegAudioProcessor) AddCover(m4bFile string, coverFile string) error {
	tempFile := p.ChangeFileExtension(m4bFile, ".withCover.m4b")

	cmd := p.Command.Create(
		"ffmpeg",
		"-i",
		m4bFile,
		"-i",
		coverFile,
		"-map",
		"0",
		"-map",
		"1",
		"-c",
		"copy",
		"-disposition:v",
		"attached_pic",
		tempFile,
	)
	var outBuf bytes.Buffer
	err := cmd.Run(&outBuf, &outBuf)
	if err != nil {
		fmt.Println(outBuf.String())
		return fmt.Errorf("could not import chapters: %w", err)
	}

	err = os.Rename(tempFile, m4bFile)
	if err != nil {
		return fmt.Errorf("Could not rename m4b file: %w", err)
	}

	return nil
}

func (p *FFmpegAudioProcessor) AddMetadata(m4bFile string, metadata string, bookTitle string) error {
	metadataFile, err := p.createMetadataFile(m4bFile, metadata)
	if err != nil {
		return fmt.Errorf("Could not create metadata file: %w", err)
	}

	tempFile := p.ChangeFileExtension(m4bFile, ".withMetadata.m4b")

	cmd := p.Command.Create(
		"ffmpeg",
		"-i",
		m4bFile,
		"-i",
		metadataFile,
		"-map_metadata",
		"1",
		"-c",
		"copy",
		"-metadata",
		"title="+bookTitle,
		tempFile,
	)
	var outBuf bytes.Buffer
	err = cmd.Run(&outBuf, &outBuf)
	if err != nil {
		fmt.Println(outBuf.String())
		return fmt.Errorf("could not import chapters: %w", err)
	}

	err = os.Rename(tempFile, m4bFile)
	if err != nil {
		return fmt.Errorf("Could not rename m4b file: %w", err)
	}

	return nil
}

func (p *FFmpegAudioProcessor) ExtractCover(m4aFile string) (string, error) {
	cmd := p.Command.Create("ffmpeg", "-i", m4aFile, "-an", "-vcodec", "copy", "cover.jpg")

	var outBuf bytes.Buffer
	err := cmd.Run(&outBuf, &outBuf)
	if err != nil {
		fmt.Println(outBuf.String())
		return "", fmt.Errorf("could not import chapters: %w", err)
	}

	return "cover.jpg", nil
}

func (p *FFmpegAudioProcessor) createMetadataFile(m4bFile string, metadata string) (string, error) {
	metadataFile := p.ChangeFileExtension(m4bFile, ".metadata")
	if err := os.MkdirAll(filepath.Dir(metadataFile), 0755); err != nil {
		return "", err
	}
	return metadataFile, os.WriteFile(metadataFile, []byte(metadata), 0600)
}

func (p *FFmpegAudioProcessor) ChangeFileExtension(file string, ext string) string {
	withoutExt := strings.TrimSuffix(file, filepath.Ext(file))
	return withoutExt + ext
}
