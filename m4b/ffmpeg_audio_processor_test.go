package m4b

import (
	"bytes"
	"fmt"
	"os"
	"slices"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFFmpegAudioProcessor_Always_ReturnsProcessorWithCommand(t *testing.T) {
	processor := NewFFmpegAudioProcessor()

	require.Equal(t, &ExecCommand{}, processor.Command)
}

func TestFFmpegAudioProcessor_ConcatAndEncode_ThreeFiles_EncodesThemInOnePass(t *testing.T) {
	fakeCommand := FakeCommand{}
	processor := &FFmpegAudioProcessor{Command: &fakeCommand}

	result, err := processor.ConcatAndEncode([]string{"a.mp3", "b's.flac", "c.m4a"}, "./output")
	require.NoError(t, err)

	require.Equal(t, "output/concat.m4b", result)
	require.Equal(
		t,
		[][]string{{
			"ffmpeg",
			"-i", "a.mp3",
			"-i", "b's.flac",
			"-i", "c.m4a",
			"-filter_complex", "[0:a:0][1:a:0][2:a:0]concat=n=3:v=0:a=1[a]",
			"-map", "[a]",
			"-map_chapters", "-1",
			"-c:a", "aac_at",
			"output/concat.m4b",
		}},
		fakeCommand.CreatedCommands,
	)
	require.True(t, fakeCommand.Cmd.Executed)
}

func TestFFmpegAudioProcessor_CopyToM4B_OneFile_CopiesItsAudioWithoutEncoding(t *testing.T) {
	fakeCommand := FakeCommand{}
	processor := &FFmpegAudioProcessor{Command: &fakeCommand}

	result, err := processor.CopyToM4B("a.m4a", "./output")
	require.NoError(t, err)

	require.Equal(t, "output/copy.m4b", result)
	require.Equal(
		t,
		[][]string{{"ffmpeg", "-i", "a.m4a", "-map", "0:a:0", "-map_chapters", "-1", "-c", "copy", "output/copy.m4b"}},
		fakeCommand.CreatedCommands,
	)
	require.True(t, fakeCommand.Cmd.Executed)
}

func TestFFmpegAudioProcessor_ReadDecodedDurations_ProgressOutput_ReturnsTheLastOutTime(t *testing.T) {
	fakeCommand := FakeCommand{
		Stdout: "out_time_us=1000000\nprogress=continue\nout_time_us=2005333\nprogress=end\n",
	}
	processor := &FFmpegAudioProcessor{Command: &fakeCommand}

	durations, err := processor.ReadDecodedDurations([]string{"a.mp3", "b.flac"})
	require.NoError(t, err)

	require.Equal(t, []float64{2.005333, 2.005333}, durations)
	require.ElementsMatch(
		t,
		[][]string{
			{"ffmpeg", "-v", "error", "-progress", "pipe:1", "-i", "a.mp3", "-map", "0:a:0", "-f", "null", "-"},
			{"ffmpeg", "-v", "error", "-progress", "pipe:1", "-i", "b.flac", "-map", "0:a:0", "-f", "null", "-"},
		},
		fakeCommand.CreatedCommands,
	)
}

func TestFFmpegAudioProcessor_ReadDecodedDurations_NoOutTime_ReturnsError(t *testing.T) {
	fakeCommand := FakeCommand{Stdout: "out_time_us=N/A\nprogress=end\n"}
	processor := &FFmpegAudioProcessor{Command: &fakeCommand}

	_, err := processor.ReadDecodedDurations([]string{"a.mp3"})

	require.EqualError(t, err, "no decoded duration for a.mp3")
}

func TestFFmpegAudioProcessor_AddChapters(t *testing.T) {
	fakeCommand := FakeCommand{}
	processor := &FFmpegAudioProcessor{Command: &fakeCommand}
	inputFile := "filepath1.m4b"
	chaptersContent := "chapters"
	chaptersFile := "filepath1.chapters.txt"
	defer os.Remove(chaptersFile)

	err := processor.AddChapters(inputFile, chaptersContent)
	require.NoError(t, err)

	actualContent, err := os.ReadFile(chaptersFile)
	require.NoError(t, err)

	require.Equal(t, chaptersContent, string(actualContent))

	require.Len(t, fakeCommand.CreatedCommands, 1)
	require.Equal(
		t,
		[]string{"mp4chaps", "--import", inputFile},
		fakeCommand.CreatedCommands[0],
	)
	require.True(t, fakeCommand.Cmd.Executed)
}

func TestFFmpegAudioProcessor_AddMetadata(t *testing.T) {
	fakeCommand := FakeCommand{}
	processor := &FFmpegAudioProcessor{Command: &fakeCommand}
	inputFile := "filepath1.m4b"
	defer os.Remove(inputFile)
	metadataContent := "metadata"
	bookTitle := "booktitle"
	metadataFile := "filepath1.metadata"
	defer os.Remove(metadataFile)
	outputFile := "filepath1.withMetadata.m4b"
	// create output file to check that it is deleted
	os.WriteFile(outputFile, []byte{}, 0600)
	t.Cleanup(func() {
		_ = os.Remove(outputFile)
	})

	err := processor.AddMetadata(inputFile, metadataContent, bookTitle)
	require.NoError(t, err)

	actualContent, err := os.ReadFile(metadataFile)
	require.NoError(t, err)
	require.Equal(t, metadataContent, string(actualContent))

	require.Equal(
		t,
		[]string{
			"ffmpeg",
			"-i",
			inputFile,
			"-i",
			metadataFile,
			"-map_metadata",
			"1",
			"-c",
			"copy",
			"-metadata",
			"title=booktitle",
			outputFile,
		},
		fakeCommand.CreatedCommands[0],
	)
	require.Empty(t, fakeCommand.CommandsNamed(AtomicParsleyCommand))
	require.True(t, fakeCommand.Cmd.Executed)

	_, err = os.Stat(outputFile)
	require.True(t, os.IsNotExist(err), "Output file should not exist")
}

func TestFFmpegAudioProcessor_ExtractCover(t *testing.T) {
	fakeCommand := FakeCommand{}
	processor := &FFmpegAudioProcessor{Command: &fakeCommand}
	inputFile := "filepath1.m4a"

	coverFile, err := processor.ExtractCover(inputFile, ".")
	require.NoError(t, err)

	require.Equal(t, "cover.jpg", coverFile)

	require.Len(t, fakeCommand.CreatedCommands, 1)
	require.Equal(
		t,
		[]string{"ffmpeg", "-i", inputFile, "-an", "-vcodec", "copy", coverFile},
		fakeCommand.CreatedCommands[0],
	)
	require.True(t, fakeCommand.Cmd.Executed)
}

func TestFFmpegAudioProcessor_HasCoverStream_FileWithVideoStream_ReturnsTrue(t *testing.T) {
	fakeCommand := FakeCommand{Stdout: "mjpeg\n"}
	processor := &FFmpegAudioProcessor{Command: &fakeCommand}
	inputFile := "filepath1.mp3"

	hasCover, err := processor.HasCoverStream(inputFile)
	require.NoError(t, err)

	require.True(t, hasCover)

	require.Len(t, fakeCommand.CreatedCommands, 1)
	require.Equal(
		t,
		[]string{
			"ffprobe",
			"-v",
			"error",
			"-select_streams",
			"v",
			"-show_entries",
			"stream=codec_name",
			"-of",
			"csv=p=0",
			inputFile,
		},
		fakeCommand.CreatedCommands[0],
	)
	require.True(t, fakeCommand.Cmd.Executed)
}

func TestFFmpegAudioProcessor_HasCoverStream_FileWithoutVideoStream_ReturnsFalse(t *testing.T) {
	fakeCommand := FakeCommand{Stdout: "\n"}
	processor := &FFmpegAudioProcessor{Command: &fakeCommand}

	hasCover, err := processor.HasCoverStream("filepath1.mp3")
	require.NoError(t, err)

	require.False(t, hasCover)
}

func TestFFmpegAudioProcessor_AddCover(t *testing.T) {
	fakeCommand := FakeCommand{}
	processor := &FFmpegAudioProcessor{Command: &fakeCommand}
	inputFile := "filepath1.m4b"
	defer os.Remove(inputFile)
	coverFile := "cover.jpg"
	outputFile := "filepath1.withCover.m4b"
	// create output file to check that it is deleted
	os.WriteFile(outputFile, []byte{}, 0600)
	t.Cleanup(func() {
		_ = os.Remove(outputFile)
	})

	err := processor.AddCover(inputFile, coverFile)
	require.NoError(t, err)

	require.Contains(
		t,
		fakeCommand.CreatedCommands,
		[]string{
			"ffmpeg",
			"-i",
			inputFile,
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
			outputFile,
		},
	)
	require.Empty(t, fakeCommand.CommandsNamed(AtomicParsleyCommand))
	require.True(t, fakeCommand.Cmd.Executed)

	_, err = os.Stat(outputFile)
	require.True(t, os.IsNotExist(err), "Output file should not exist")
}

type FakeCommand struct {
	mu              sync.Mutex
	CreatedCommands [][]string
	Cmd             *FakeCmd
	Stdout          string
	Stderr          string
	Stdouts         []string
	MissingOnPath   []string
	RunErr          error
}

func (c *FakeCommand) Create(name string, args ...string) Cmd {
	c.mu.Lock()
	defer c.mu.Unlock()

	fullArgs := append([]string{name}, args...)
	c.CreatedCommands = append(c.CreatedCommands, fullArgs)

	stdout := c.Stdout
	if len(c.Stdouts) > 0 {
		stdout = c.Stdouts[0]
		c.Stdouts = c.Stdouts[1:]
	}

	c.Cmd = &FakeCmd{Stdout: stdout, Stderr: c.Stderr, Err: c.RunErr, Executed: false}
	return c.Cmd
}

func (c *FakeCommand) LookPath(name string) error {
	if slices.Contains(c.MissingOnPath, name) {
		return fmt.Errorf("%s: executable file not found in $PATH", name)
	}
	return nil
}

// CommandsNamed returns all created commands that invoke the given executable.
func (c *FakeCommand) CommandsNamed(name string) [][]string {
	c.mu.Lock()
	defer c.mu.Unlock()

	var matching [][]string
	for _, command := range c.CreatedCommands {
		if command[0] == name {
			matching = append(matching, command)
		}
	}
	return matching
}

type FakeCmd struct {
	Stdout   string
	Stderr   string
	Err      error
	Executed bool
}

func (c *FakeCmd) Run(stdout, stderr *bytes.Buffer) error {
	c.Executed = true
	stdout.WriteString(c.Stdout)
	stderr.WriteString(c.Stderr)
	return c.Err
}

func (c *FakeCmd) RunI(_ *bytes.Reader, stdout, stderr *bytes.Buffer) error {
	c.Executed = true
	stdout.WriteString(c.Stdout)
	stderr.WriteString(c.Stderr)
	return c.Err
}
