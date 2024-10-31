package m4b

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFFmpegAudioProcessor_ToM4A(t *testing.T) {
	fakeCommand := FakeCommand{}
	processor := &FFmpegAudioProcessor{
		Command: &fakeCommand,
	}
	inputFiles := []string{"filepath1.m4a", "filepath2.m4a"}
	output := "./output"

	files, err := processor.ToM4A(inputFiles, output)

	if err != nil {
		t.Errorf("Convert failed: %v", err)
	}

	assert.Equal(
		t,
		[]string{"ffmpeg", "-i", "filepath1.m4a", "-c", "copy", "-c:a", "aac_at", "output/filepath1.m4a"},
		fakeCommand.CreatedCommands[0],
	)

	assert.Equal(
		t,
		[]string{"ffmpeg", "-i", "filepath2.m4a", "-c", "copy", "-c:a", "aac_at", "output/filepath2.m4a"},
		fakeCommand.CreatedCommands[1],
	)

	assert.Equal(
		t,
		[]string{"output/filepath1.m4a", "output/filepath2.m4a"},
		files,
	)

	assert.True(t, fakeCommand.Cmd.Executed)
}

type FakeCommand struct {
	CreatedCommands [][]string
	Cmd             *FakeCmd
}

func (c *FakeCommand) Create(name string, args ...string) Cmd {
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

func (c *FakeCmd) Run(_ *bytes.Buffer, _ *bytes.Buffer) error {
	c.Executed = true
	return nil
}
