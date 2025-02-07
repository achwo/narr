package m4b

import (
	"bytes"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFFmpegAudioProcessor_ToM4A(t *testing.T) {
	fakeCommand := FakeCommand{}
	processor := &FFmpegAudioProcessor{
		Command: &fakeCommand,
	}
	inputFiles := []string{"filepath1.m4a", "filepath2.m4a"}
	output := "./output"

	files, err := processor.ToM4A(inputFiles, output)
	require.NoError(t, err)

	require.ElementsMatch(
		t,
		[][]string{
			{"ffmpeg", "-i", "filepath1.m4a", "-c", "copy", "-c:a", "aac_at", "output/filepath1.m4a"},
			{"ffmpeg", "-i", "filepath2.m4a", "-c", "copy", "-c:a", "aac_at", "output/filepath2.m4a"},
		},
		fakeCommand.CreatedCommands,
	)
	require.ElementsMatch(
		t,
		[]string{"output/filepath1.m4a", "output/filepath2.m4a"},
		files,
	)

	require.True(t, fakeCommand.Cmd.Executed)
}

func TestFFmpegAudioProcessor_Concat(t *testing.T) {
	fakeCommand := FakeCommand{}
	processor := &FFmpegAudioProcessor{
		Command: &fakeCommand,
	}
	inputFiles := []string{"filepath'1.m4a", "filepath2.m4a"}
	outputPath := "./output"

	filelistFile, err := os.CreateTemp("", "filelist")
	require.NoError(t, err)
	defer os.Remove(filelistFile.Name())

	result, err := processor.Concat(inputFiles, filelistFile.Name(), outputPath)
	require.NoError(t, err)

	expectedFilelistContent := "file 'filepath'\\''1.m4a'\nfile 'filepath2.m4a'\n"

	actualContent, err := os.ReadFile(filelistFile.Name())
	require.NoError(t, err)
	require.Equal(t, expectedFilelistContent, string(actualContent))

	require.Equal(t, "output/concat.m4b", result)

	require.Equal(
		t,
		[]string{"ffmpeg", "-f", "concat", "-safe", "0", "-i", filelistFile.Name(), "-c", "copy", "-vn", "output/concat.m4b"},
		fakeCommand.CreatedCommands[0],
	)
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

	require.Len(t, fakeCommand.CreatedCommands, 1)
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

	require.Len(t, fakeCommand.CreatedCommands, 1)
	require.Equal(
		t,
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
		fakeCommand.CreatedCommands[0],
	)
	require.True(t, fakeCommand.Cmd.Executed)

	_, err = os.Stat(outputFile)
	require.True(t, os.IsNotExist(err), "Output file should not exist")
}

type FakeCommand struct {
	mu              sync.Mutex
	CreatedCommands [][]string
	Cmd             *FakeCmd
}

func (c *FakeCommand) Create(name string, args ...string) Cmd {
	c.mu.Lock()
	defer c.mu.Unlock()

	fullArgs := append([]string{name}, args...)
	c.CreatedCommands = append(c.CreatedCommands, fullArgs)
	c.Cmd = &FakeCmd{Stdout: "", Stderr: "", Executed: false}
	return c.Cmd
}

type FakeCmd struct {
	Stdout   string
	Stderr   string
	Executed bool
}

func (c *FakeCmd) Run(_, _ *bytes.Buffer) error {
	c.Executed = true
	return nil
}

func (c *FakeCmd) RunI(_ *bytes.Reader, _, _ *bytes.Buffer) error {
	c.Executed = true
	return nil
}
