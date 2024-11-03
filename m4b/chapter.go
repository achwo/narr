package m4b

import (
	"fmt"
	"time"
)

// File represents an audio file with its name and duration
type File struct {
	Name     string
	Duration float64
}

// Chapter represents a section in an audiobook with timing information and contained files
type Chapter struct {
	title           string
	previousChapter *Chapter
	files           []File
	index           int
}

func (c *Chapter) addFile(file File) {
	c.files = append(c.files, file)
}

// ChapterMarker returns a formatted string containing chapter timing and title information
// in the format used by common audiobook chapter markers
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
