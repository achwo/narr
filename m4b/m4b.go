package m4b

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/achwo/narr/config"
	"github.com/achwo/narr/utils"
)

type Chapter struct {
	title           string
	previousChapter *Chapter
	files           []File
	index           int
}

func (c *Chapter) addFile(file File) {
	c.files = append(c.files, file)
}

func (c *Chapter) ChapterMarker() string {
	formatTime := func(s float64) string {
		duration := time.Duration(s * float64(time.Second))

		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		seconds := int(duration.Seconds()) % 60
		milliseconds := int(duration.Milliseconds()) % 1000
		return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)
	}

	offsetFormatted := formatTime(c.offset())
	chapterIndex := fmt.Sprintf("%02d", c.index)

	return fmt.Sprintf(
		"CHAPTER%s=%s\nCHAPTER%sNAME=%s",
		chapterIndex,
		offsetFormatted,
		chapterIndex,
		c.title,
	)
}

func (c *Chapter) duration() float64 {
	duration := 0.0

	for _, file := range c.files {
		duration += file.Duration
	}

	return duration
}

// Offset from the start in seconds
func (c *Chapter) offset() float64 {
	if c.previousChapter == nil {
		return 0
	}
	return c.previousChapter.offset() + c.previousChapter.duration()
}

type File struct {
	Name     string
	Duration float64
}

func ShowChapters(conf config.ProjectConfig) (string, error) {
	audioFiles, err := conf.AudioFiles()
	if err != nil {
		return "", fmt.Errorf("Could not load audio files: %w", err)
	}

	chapters := make(map[string]*Chapter)
	var chapterOrder []string
	var previousChapter *Chapter

	var metadataManager utils.MetadataManager = &utils.FFmpegMetadataManager{}

	for i, file := range audioFiles {
		title, duration, err := metadataManager.ReadTitleAndDuration(file)
		if err != nil {
			return "", fmt.Errorf("Could not read file data for file %s: %w", file, err)
		}
		chapterName := title

		for _, rule := range conf.ChapterRules {
			chapterName, err = rule.Apply(chapterName)
			if err != nil {
				return "", fmt.Errorf("Chapter rule invalid: %w", err)
			}
		}

		value := chapters[chapterName]
		value, exists := chapters[chapterName]

		newFile := File{Name: file, Duration: duration}
		if exists {
			value.addFile(newFile)
		} else {
			newChapter := Chapter{
				title:           chapterName,
				previousChapter: previousChapter,
				files:           []File{newFile},
				index:           i,
			}
			chapterOrder = append(chapterOrder, chapterName)
			chapters[chapterName] = &newChapter
			previousChapter = &newChapter
		}
	}

	markers := make([]string, 0, len(chapters))

	for _, name := range chapterOrder {
		chapter := chapters[name]
		markers = append(markers, chapter.ChapterMarker())
	}

	markersFileContent := strings.Join(markers, "\n\n")

	return markersFileContent, nil
}

func ShowMetadata(conf *config.ProjectConfig) (string, error) {
	audioFiles, err := conf.AudioFiles()
	if err != nil {
		return "", fmt.Errorf("could not load audio files: %w", err)
	}

	if len(audioFiles) == 0 {
		return "", errors.New("no audio files found")
	}

	referenceFile := audioFiles[0]

	var metadataManager utils.MetadataManager = &utils.FFmpegMetadataManager{}
	metadata, err := metadataManager.ReadMetadata(referenceFile)
	if err != nil {
		return "", fmt.Errorf("could not read metadata: %w", err)
	}

	return metadata, nil
}
