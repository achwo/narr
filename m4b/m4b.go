package m4b

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/achwo/narr/config"
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

func ShowChapters(projectPath string, conf config.ProjectConfig) error {
	audioFiles, err := conf.AudioFiles()
	if err != nil {
		return err
	}

	chapters := make(map[string]*Chapter)
	var chapterOrder []string
	var previousChapter *Chapter

	for i, file := range audioFiles {
		title, duration, err := readTitleAndDuration(file)
		if err != nil {
			return fmt.Errorf("Could not read file data for file %s: %w", file, err)
		}
		chapterName := title

		for _, rule := range conf.ChapterRules {
			chapterName, err = rule.Apply(chapterName)
			if err != nil {
				return fmt.Errorf("Chapter rule invalid: %w", err)
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

	fmt.Println(markersFileContent)

	// outputFilePath := filepath.Join(projectPath, "chapters.txt")

	// err = os.WriteFile(outputFilePath, []byte(markersFileContent), 0664)
	// if err != nil {
	// 	return fmt.Errorf("Could not write chapters file: %w", err)
	// }
	return nil
}

func readTitleAndDuration(file string) (string, float64, error) {
	dataCmd := exec.Command(
		"ffprobe",
		"-v",
		"error",
		"-select_streams",
		"a:0",
		"-show_entries",
		"format=duration:format_tags=title",
		file,
	)

	var data bytes.Buffer
	dataCmd.Stdout = &data

	if err := dataCmd.Run(); err != nil {
		return "", 0, fmt.Errorf("failed to extract metadata for file %s: %w", file, err)
	}

	probeContent := data.String()

	durationRegex := regexp.MustCompile(`duration=([0-9]+\.[0-9]+)`)
	titleRegex := regexp.MustCompile(`TAG:title=(.+)`)

	titleMatch := titleRegex.FindStringSubmatch(probeContent)
	if len(titleMatch) < 2 {
		return "", 0, fmt.Errorf("title not found")
	}

	title := titleMatch[1]

	durationMatch := durationRegex.FindStringSubmatch(probeContent)
	if len(durationMatch) < 2 {
		return "", 0, fmt.Errorf("duration not found")
	}

	duration, err := strconv.ParseFloat(durationMatch[1], 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid duration value")
	}

	return title, duration, nil
}
