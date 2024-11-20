package m4b

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/achwo/narr/utils"
)

// FFmpegAudioProcessor handles audio file processing operations using FFmpeg
type FFmpegAudioProcessor struct {
	Command Command
}

// ToM4A converts audio files to M4A format using FFmpeg
// It takes a slice of input file paths and an output directory path
// Returns a slice of converted file paths or an error
func (p *FFmpegAudioProcessor) ToM4A(files []string, outputPath string) ([]string, error) {
	outInOrder := make([]string, 0, len(files))

	for _, file := range files {
		outFile := utils.ReplaceDirAndExt(file, outputPath, ".m4a")
		outInOrder = append(outInOrder, outFile)
	}

	const numWorkers = 5

	in := make(chan string, numWorkers)
	out := make(chan string, numWorkers)
	errs := make(chan error, numWorkers)

	for i := 0; i < numWorkers; i++ {
		go p.convertToM4AWorker(in, out, errs, outputPath)
	}

	go func() {
		for _, file := range files {
			in <- file
		}
		close(in)
	}()

	for i := 0; i < len(files); i++ {
		select {
		case <-out:
			fmt.Print(".")
		case err := <-errs:
			return nil, fmt.Errorf("could not convert track: %w", err)
		}
	}

	fmt.Println()
	return outInOrder, nil
}

func (p *FFmpegAudioProcessor) convertToM4AWorker(
	in <-chan string,
	out chan<- string,
	error chan<- error,
	outputPath string,
) {
	for file := range in {
		outFile := utils.ReplaceDirAndExt(file, outputPath, ".m4a")
		cmd := p.Command.Create("ffmpeg", "-i", file, "-c", "copy", "-c:a", "aac_at", outFile)

		var outBuf bytes.Buffer
		err := cmd.Run(&outBuf, &outBuf)
		if err != nil {
			fmt.Println(outBuf.String())
			error <- fmt.Errorf("could not convert file %s:, %w", outFile, err)
			continue
		}

		out <- outFile
	}
}

// Concat concatenates multiple audio files into a single M4B file
// It takes input files, a temporary filelist path, and an output directory
// Returns the path to the concatenated file or an error
func (p *FFmpegAudioProcessor) Concat(files []string, filelistFile string, outputPath string) (string, error) {
	fileListContent := p.filelistFileContent(files)
	err := os.WriteFile(filelistFile, []byte(fileListContent), 0600)
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

// AddChapters adds chapter markers to an M4B file using mp4chaps
// It takes the M4B file path and a string containing chapter information
func (p *FFmpegAudioProcessor) AddChapters(m4bFile string, chapters string) error {
	if err := p.createChaptersFile(m4bFile, chapters); err != nil {
		return fmt.Errorf("could not create chapters file: %w", err)
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

// AddCover adds cover artwork to an M4B file
// It takes the M4B file path and the cover image file path
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
		return fmt.Errorf("could not add cover: %w", err)
	}

	err = os.Rename(tempFile, m4bFile)
	if err != nil {
		return fmt.Errorf("could not rename m4b file: %w", err)
	}

	return nil
}

// AddMetadata adds metadata tags to an M4B file
// It takes the M4B file path, metadata content, and book title
func (p *FFmpegAudioProcessor) AddMetadata(m4bFile string, metadata string, bookTitle string) error {
	metadataFile, err := p.createMetadataFile(m4bFile, metadata)
	if err != nil {
		return fmt.Errorf("could not create metadata file: %w", err)
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
		return fmt.Errorf("could not add metadata: %w", err)
	}

	err = os.Rename(tempFile, m4bFile)
	if err != nil {
		return fmt.Errorf("could not rename m4b file: %w", err)
	}

	return nil
}

// ExtractCover extracts cover artwork from an M4A file
// It takes the M4A file path and returns the path to the extracted cover image
func (p *FFmpegAudioProcessor) ExtractCover(m4aFile string, workDir string) (string, error) {
	coverFile := filepath.Join(workDir, "cover.jpg")
	cmd := p.Command.Create("ffmpeg", "-i", m4aFile, "-an", "-vcodec", "copy", coverFile)

	var outBuf bytes.Buffer
	err := cmd.Run(&outBuf, &outBuf)
	if err != nil {
		fmt.Println(outBuf.String())
		return "", fmt.Errorf("could not extract cover: %w", err)
	}

	return coverFile, nil
}

func (p *FFmpegAudioProcessor) createMetadataFile(m4bFile string, metadata string) (string, error) {
	metadataFile := p.ChangeFileExtension(m4bFile, ".metadata")
	if err := os.MkdirAll(filepath.Dir(metadataFile), 0755); err != nil {
		return "", err
	}
	return metadataFile, os.WriteFile(metadataFile, []byte(metadata), 0600)
}

// ChangeFileExtension changes the extension of a file path
// It takes the original file path and new extension, returns the modified path
func (p *FFmpegAudioProcessor) ChangeFileExtension(file string, ext string) string {
	withoutExt := strings.TrimSuffix(file, filepath.Ext(file))
	return withoutExt + ext
}

// ReadTitleAndDuration extracts the title and duration from a media file
// Returns the title string and duration in seconds
func (p *FFmpegAudioProcessor) ReadTitleAndDuration(file string) (string, float64, error) {
	dataCmd := p.Command.Create(
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

	if err := dataCmd.Run(&data, &data); err != nil {
		return "", 0, fmt.Errorf("failed to extract title and duration for file %s: %w", file, err)
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

// WriteMetadata updates the metadata in the file
// WriteMetadata updates the metadata in the media file
// Creates a temporary file during the process and replaces the original file
// If verbose is true, prints FFmpeg command and output
func (p *FFmpegAudioProcessor) WriteMetadata(file string, metadata string, verbose bool) error {
	tmpFile := file + ".tmp" + filepath.Ext(file)

	err := p.WriteMetadataO(file, tmpFile, metadata, verbose)
	if err != nil {
		return fmt.Errorf("could not write metadata: %w", err)
	}

	err = os.Rename(tmpFile, file)
	if err != nil {
		return fmt.Errorf("could not rename temp file to output file: %w", err)
	}

	return nil
}

// WriteMetadataO is like WriteMetadata with explicit output file
// WriteMetadataO writes metadata to a new output file instead of modifying the input file
// If verbose is true, prints FFmpeg command and output
func (p *FFmpegAudioProcessor) WriteMetadataO(inputFile string, outputFile string, metadata string, verbose bool) error {
	writeCmd := p.Command.Create("ffmpeg", "-i", inputFile, "-f", "ffmetadata", "-i", "-", "-map_metadata", "1", "-c", "copy", outputFile)

	var outBuf bytes.Buffer

	err := writeCmd.RunI(bytes.NewReader([]byte(metadata)), &outBuf, &outBuf)

	if verbose {
		fmt.Printf("Command output:\n%s\n", outBuf.String())
	}

	if err != nil {
		return fmt.Errorf("ffmpeg command failed: %v\n%s", err, outBuf.String())
	}
	return nil
}

// ReadMetadata extracts metadata from a media file at the given path
// Returns the metadata as a string in FFmpeg metadata format
func (p *FFmpegAudioProcessor) ReadMetadata(path string) (string, error) {
	extractCmd := p.Command.Create("ffmpeg", "-i", path, "-f", "ffmetadata", "-")

	var metadata, errout bytes.Buffer

	if err := extractCmd.Run(&metadata, &errout); err != nil {
		fmt.Println(errout.String())
		return "", fmt.Errorf("failed to extract metadata for file %s: %w", path, err)
	}

	return metadata.String(), nil
}
