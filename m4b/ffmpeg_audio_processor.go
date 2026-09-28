package m4b

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

// FFmpegAudioProcessor handles audio file processing operations using FFmpeg
type FFmpegAudioProcessor struct {
	Command Command
}

// NewFFmpegAudioProcessor creates a FFmpegAudioProcessor that runs ffmpeg as an
// external command.
func NewFFmpegAudioProcessor() *FFmpegAudioProcessor {
	return &FFmpegAudioProcessor{Command: &ExecCommand{}}
}

// CopyToM4B copies the audio of a single file into an M4B file without encoding it
// It takes the input file and an output directory
// Returns the path to the created M4B file or an error
func (p *FFmpegAudioProcessor) CopyToM4B(file string, outputPath string) (string, error) {
	outputFilepath := filepath.Join(outputPath, "copy.m4b")

	cmd := p.Command.Create(
		"ffmpeg",
		"-i",
		file,
		"-map",
		"0:a:0",
		"-map_chapters",
		"-1",
		"-c",
		"copy",
		outputFilepath,
	)
	var outBuf bytes.Buffer
	if err := cmd.Run(&outBuf, &outBuf); err != nil {
		fmt.Println(outBuf.String())
		return "", fmt.Errorf("could not copy file: %w", err)
	}

	return outputFilepath, nil
}

// ConcatAndEncode decodes all files, joins them and encodes the result once
// It takes the input files and an output directory
// Returns the path to the created M4B file or an error
func (p *FFmpegAudioProcessor) ConcatAndEncode(files []string, outputPath string) (string, error) {
	if err := passOpenFileLimitToChildren(); err != nil {
		return "", err
	}

	outputFilepath := filepath.Join(outputPath, "concat.m4b")

	args := make([]string, 0, 2*len(files)+9)
	var inputLabels strings.Builder
	for i, file := range files {
		args = append(args, "-i", file)
		fmt.Fprintf(&inputLabels, "[%d:a:0]", i)
	}

	args = append(
		args,
		"-filter_complex",
		fmt.Sprintf("%sconcat=n=%d:v=0:a=1[a]", inputLabels.String(), len(files)),
		"-map",
		"[a]",
		"-map_chapters",
		"-1",
		"-c:a",
		"aac_at",
		outputFilepath,
	)

	cmd := p.Command.Create("ffmpeg", args...)
	var outBuf bytes.Buffer
	if err := cmd.Run(&outBuf, &outBuf); err != nil {
		fmt.Println(outBuf.String())
		return "", fmt.Errorf("could not concat and encode files: %w", err)
	}

	return outputFilepath, nil
}

func passOpenFileLimitToChildren() error {
	var limit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return fmt.Errorf("could not read open file limit: %w", err)
	}

	// Without an explicit Setrlimit, Go hands child processes the original, lower limit.
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return fmt.Errorf("could not set open file limit: %w", err)
	}

	return nil
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
// Freeform MP4 tags that the remux drops are written back afterwards
func (p *FFmpegAudioProcessor) AddCover(m4bFile string, coverFile string) error {
	tempFile := p.ChangeFileExtension(m4bFile, ".withCover.m4b")

	tagsBefore, err := p.mp4TagsOf(m4bFile)
	if err != nil {
		return fmt.Errorf("could not read metadata of %s: %w", m4bFile, err)
	}

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
	err = cmd.Run(&outBuf, &outBuf)
	if err != nil {
		fmt.Println(outBuf.String())
		return fmt.Errorf("could not add cover: %w", err)
	}

	if err := p.preserveFreeformTags(m4bFile, tempFile, tagsBefore); err != nil {
		os.Remove(tempFile)
		return err
	}

	err = os.Rename(tempFile, m4bFile)
	if err != nil {
		return fmt.Errorf("could not rename m4b file: %w", err)
	}

	return nil
}

// AddMetadata adds metadata tags to an M4B file
// It takes the M4B file path, metadata content, and book title
// Freeform MP4 tags that ffmpeg cannot write are written back afterwards
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

	if err := p.preserveFreeformTags(m4bFile, tempFile, metadata); err != nil {
		os.Remove(tempFile)
		return err
	}

	err = os.Rename(tempFile, m4bFile)
	if err != nil {
		return fmt.Errorf("could not rename m4b file: %w", err)
	}

	return nil
}

// ExtractCover extracts cover artwork from an audio file (M4A, MP3, FLAC, etc.)
// It takes the audio file path and returns the path to the extracted cover image
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

// HasCoverStream reports whether the audio file contains an embedded cover image
// It takes the audio file path and returns true if the file has a video stream
func (p *FFmpegAudioProcessor) HasCoverStream(file string) (bool, error) {
	cmd := p.Command.Create(
		"ffprobe",
		"-v",
		"error",
		"-select_streams",
		"v",
		"-show_entries",
		"stream=codec_name",
		"-of",
		"csv=p=0",
		file,
	)

	var out, errOut bytes.Buffer
	if err := cmd.Run(&out, &errOut); err != nil {
		fmt.Println(errOut.String())
		return false, fmt.Errorf("could not probe streams of file %s: %w", file, err)
	}

	return strings.TrimSpace(out.String()) != "", nil
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
		"format=duration:format_tags=title:stream_tags=title",
		file,
	)

	var data bytes.Buffer

	if err := dataCmd.Run(&data, &data); err != nil {
		return "", 0, fmt.Errorf("failed to extract title and duration for file %s: %w", file, err)
	}

	probeContent := data.String()

	durationRegex := regexp.MustCompile(`duration=([0-9]+\.?[0-9]*)`)
	titleRegex := regexp.MustCompile(`(?i)TAG:title=(.+)`)

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

// ReadDecodedDurations decodes the first audio stream of each file, five at a time
// Returns the decoded durations in seconds, in the order of the files
// They can differ from the durations the containers state
func (p *FFmpegAudioProcessor) ReadDecodedDurations(files []string) ([]float64, error) {
	durations := make([]float64, len(files))
	errs := make([]error, len(files))

	var wg sync.WaitGroup
	slots := make(chan struct{}, 5)

	for i, file := range files {
		wg.Add(1)
		slots <- struct{}{}
		go func() {
			defer wg.Done()
			durations[i], errs[i] = p.readDecodedDuration(file)
			<-slots
		}()
	}
	wg.Wait()

	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return durations, nil
}

func (p *FFmpegAudioProcessor) readDecodedDuration(file string) (float64, error) {
	cmd := p.Command.Create("ffmpeg", "-v", "error", "-progress", "pipe:1", "-i", file, "-map", "0:a:0", "-f", "null", "-")

	var out, errOut bytes.Buffer
	if err := cmd.Run(&out, &errOut); err != nil {
		fmt.Println(errOut.String())
		return 0, fmt.Errorf("could not decode %s: %w", file, err)
	}

	outTimeRegex := regexp.MustCompile(`out_time_us=(\d+)`)
	matches := outTimeRegex.FindAllStringSubmatch(out.String(), -1)
	if len(matches) == 0 {
		return 0, fmt.Errorf("no decoded duration for %s", file)
	}

	microseconds, err := strconv.ParseInt(matches[len(matches)-1][1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid decoded duration for %s: %w", file, err)
	}

	return float64(microseconds) / 1e6, nil
}

// WriteMetadata updates the metadata in the file
// WriteMetadata updates the metadata in the media file
// Creates a temporary file during the process and replaces the original file
// If verbose is true, prints FFmpeg command and output
func (p *FFmpegAudioProcessor) WriteMetadata(file string, metadata string, verbose bool) error {
	tmpFile := file + ".tmp" + filepath.Ext(file)

	err := p.WriteMetadataO(file, tmpFile, metadata, verbose)
	if err != nil {
		os.Remove(tmpFile)
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
// Freeform MP4 tags that ffmpeg cannot write are written back afterwards
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

	return p.preserveFreeformTags(inputFile, outputFile, metadata)
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
