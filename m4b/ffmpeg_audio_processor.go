package m4b

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
)

type FFmpegAudioProcessor struct {
	Command Command
}

func (c *FFmpegAudioProcessor) ToM4A(files []string, outputPath string) ([]string, error) {
	outputFiles := make([]string, 0, len(files))
	for _, file := range files {
		fileName := filepath.Base(file)
		out := path.Join(outputPath, fileName)
		outputFiles = append(outputFiles, out)
		cmd := c.Command.Create("ffmpeg", "-i", file, "-c", "copy", "-c:a", "aac_at", out)

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

func (f *FFmpegAudioProcessor) Concat(files []string, output string) error {
	return nil
}

func (f *FFmpegAudioProcessor) AddChapters(m4bFile string, chapters string) error {
	return nil
}

func (f *FFmpegAudioProcessor) AddCover(m4bFile string, coverFile string) error {
	return nil
}

func (f *FFmpegAudioProcessor) AddMetadata(m4bFile string, metadata string) error {
	return nil
}
