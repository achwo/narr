package m4b

import (
	"errors"
	"strings"
	"testing"

	"github.com/achwo/narr/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatchMP4Tags_StandardAndFreeformTags_CallsAtomicParsleyOnce(t *testing.T) {
	written := ";FFMETADATA1\nalbum=Die Reihe\nSERIES=Die Reihe\n"
	fakeCommand := &FakeCommand{Stdouts: []string{"", written}}
	processor := &FFmpegAudioProcessor{Command: fakeCommand}

	patched := processor.PatchMP4Tags("book.m4b", []utils.TagWithValue{
		{Tag: "album", Value: "Die Reihe"},
		{Tag: "SERIES", Value: "Die Reihe"},
	}, false)

	assert.True(t, patched)
	require.Equal(
		t,
		[][]string{{
			AtomicParsleyCommand,
			"book.m4b",
			"--album", "Die Reihe",
			"--rDNSatom", "Die Reihe", "name=SERIES", "domain=com.apple.iTunes",
			"--overWrite",
		}},
		fakeCommand.CommandsNamed(AtomicParsleyCommand),
	)
}

func TestPatchMP4Tags_DeletedTag_PassesTheEmptyValueThatClearsTheAtom(t *testing.T) {
	fakeCommand := &FakeCommand{Stdouts: []string{"", ";FFMETADATA1\ntitle=Book\n"}}
	processor := &FFmpegAudioProcessor{Command: fakeCommand}

	patched := processor.PatchMP4Tags("book.m4b", []utils.TagWithValue{{Tag: "comment"}}, false)

	assert.True(t, patched)
	assert.Equal(
		t,
		[]string{AtomicParsleyCommand, "book.m4b", "--comment", "", "--overWrite"},
		fakeCommand.CommandsNamed(AtomicParsleyCommand)[0],
	)
}

func TestPatchMP4Tags_NonMP4File_DoesNotPatch(t *testing.T) {
	fakeCommand := &FakeCommand{}
	processor := &FFmpegAudioProcessor{Command: fakeCommand}

	patched := processor.PatchMP4Tags("book.mp3", []utils.TagWithValue{{Tag: "album", Value: "Die Reihe"}}, false)

	assert.False(t, patched)
	assert.Empty(t, fakeCommand.CreatedCommands)
}

func TestPatchMP4Tags_StandardTagWithoutAtomicParsleySwitch_DoesNotPatch(t *testing.T) {
	fakeCommand := &FakeCommand{}
	processor := &FFmpegAudioProcessor{Command: fakeCommand}

	patched := processor.PatchMP4Tags("book.m4b", []utils.TagWithValue{
		{Tag: "album", Value: "Die Reihe"},
		{Tag: "sort_album", Value: "Reihe, Die"},
	}, false)

	assert.False(t, patched)
	assert.Empty(t, fakeCommand.CreatedCommands)
}

func TestPatchMP4Tags_AtomicParsleyMissing_DoesNotPatch(t *testing.T) {
	fakeCommand := &FakeCommand{MissingOnPath: []string{AtomicParsleyCommand}}
	processor := &FFmpegAudioProcessor{Command: fakeCommand}

	patched := processor.PatchMP4Tags("book.m4b", []utils.TagWithValue{{Tag: "album", Value: "Die Reihe"}}, false)

	assert.False(t, patched)
	assert.Empty(t, fakeCommand.CreatedCommands)
}

func TestPatchMP4Tags_AtomicParsleyFails_DoesNotPatch(t *testing.T) {
	fakeCommand := &FakeCommand{
		Stdout: "AtomicParsley error: too many atoms",
		RunErr: errors.New("exit status 1"),
	}
	processor := &FFmpegAudioProcessor{Command: fakeCommand}

	patched := processor.PatchMP4Tags("book.m4b", []utils.TagWithValue{{Tag: "album", Value: "Die Reihe"}}, false)

	assert.False(t, patched)
}

func TestPatchMP4Tags_ValueTruncatedByAtomicParsley_DoesNotPatch(t *testing.T) {
	long := strings.Repeat("a", 300)
	truncated := ";FFMETADATA1\ncomment=" + strings.Repeat("a", 255) + "\n"
	fakeCommand := &FakeCommand{Stdouts: []string{"", truncated}}
	processor := &FFmpegAudioProcessor{Command: fakeCommand}

	patched := processor.PatchMP4Tags("book.m4b", []utils.TagWithValue{{Tag: "comment", Value: long}}, false)

	assert.False(t, patched)
}
