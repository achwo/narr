package m4b

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const audibleMetadata = ";FFMETADATA1\ntitle=Book\nartist=Max Heller\nSERIES=Die Reihe\nSUBTITLE=Der Untertitel\n"

// what a remux by the mov muxer leaves of audibleMetadata
const remuxedMetadata = ";FFMETADATA1\ntitle=Book\nartist=Max Heller\n"

func TestFFmpegAudioProcessor_WriteMetadataO_MovMuxerDropsTags_WritesThemBackWithAtomicParsley(t *testing.T) {
	fakeCommand := &FakeCommand{Stdouts: []string{"", remuxedMetadata, audibleMetadata}}
	processor := &FFmpegAudioProcessor{Command: fakeCommand}

	err := processor.WriteMetadataO("in.m4b", "out.m4b", audibleMetadata, false)
	require.NoError(t, err)

	require.Equal(
		t,
		[][]string{{
			AtomicParsleyCommand,
			"out.m4b",
			"--rDNSatom", "Die Reihe", "name=SERIES", "domain=com.apple.iTunes",
			"--rDNSatom", "Der Untertitel", "name=SUBTITLE", "domain=com.apple.iTunes",
			"--overWrite",
		}},
		fakeCommand.CommandsNamed(AtomicParsleyCommand),
	)
}

func TestFFmpegAudioProcessor_WriteMetadataO_AllTagsSurvive_DoesNotRunAtomicParsley(t *testing.T) {
	fakeCommand := &FakeCommand{Stdouts: []string{"", audibleMetadata}}
	processor := &FFmpegAudioProcessor{Command: fakeCommand}

	err := processor.WriteMetadataO("in.m4b", "out.m4b", audibleMetadata, false)
	require.NoError(t, err)

	assert.Empty(t, fakeCommand.CommandsNamed(AtomicParsleyCommand))
}

func TestFFmpegAudioProcessor_WriteMetadataO_NonMP4File_DoesNotReadBackTheOutput(t *testing.T) {
	fakeCommand := &FakeCommand{}
	processor := &FFmpegAudioProcessor{Command: fakeCommand}

	err := processor.WriteMetadataO("in.mp3", "out.mp3", audibleMetadata, false)
	require.NoError(t, err)

	require.Len(t, fakeCommand.CreatedCommands, 1)
	assert.Empty(t, fakeCommand.CommandsNamed(AtomicParsleyCommand))
}

func TestFFmpegAudioProcessor_WriteMetadataO_TagLostAndAtomicParsleyMissing_ReturnsErrorNamingTheTags(t *testing.T) {
	fakeCommand := &FakeCommand{
		Stdouts:       []string{"", remuxedMetadata, audibleMetadata},
		MissingOnPath: []string{AtomicParsleyCommand},
	}
	processor := &FFmpegAudioProcessor{Command: fakeCommand}

	err := processor.WriteMetadataO("in.m4b", "out.m4b", audibleMetadata, false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "SERIES")
	assert.Contains(t, err.Error(), "SUBTITLE")
	assert.Contains(t, err.Error(), AtomicParsleyCommand)
	assert.Empty(t, fakeCommand.CommandsNamed(AtomicParsleyCommand))
}

func TestFFmpegAudioProcessor_WriteMetadata_AtomicParsleyMissing_LeavesTheOriginalFileAlone(t *testing.T) {
	file := filepath.Join(t.TempDir(), "book.m4b")
	require.NoError(t, os.WriteFile(file, []byte("original"), 0600))

	fakeCommand := &FakeCommand{
		Stdouts:       []string{"", remuxedMetadata, audibleMetadata},
		MissingOnPath: []string{AtomicParsleyCommand},
	}
	processor := &FFmpegAudioProcessor{Command: fakeCommand}

	err := processor.WriteMetadata(file, audibleMetadata, false)
	require.Error(t, err)

	content, readErr := os.ReadFile(file)
	require.NoError(t, readErr)
	assert.Equal(t, "original", string(content))

	_, statErr := os.Stat(file + ".tmp.m4b")
	assert.True(t, os.IsNotExist(statErr), "temp file should be cleaned up")
}

func TestFFmpegAudioProcessor_WriteMetadataO_TagLost_KeepsTheSpellingOfTheOriginalFile(t *testing.T) {
	lowercased := ";FFMETADATA1\ntitle=Book\nseries=Die Reihe\n"
	fakeCommand := &FakeCommand{Stdouts: []string{"", remuxedMetadata, audibleMetadata}}
	processor := &FFmpegAudioProcessor{Command: fakeCommand}

	err := processor.WriteMetadataO("in.m4b", "out.m4b", lowercased, false)
	require.NoError(t, err)

	require.Contains(t, fakeCommand.CommandsNamed(AtomicParsleyCommand)[0], "name=SERIES")
}

func TestFFmpegAudioProcessor_WriteMetadataO_DeletedTag_IsNotWrittenBack(t *testing.T) {
	withoutSeries := ";FFMETADATA1\ntitle=Book\nartist=Max Heller\n"
	fakeCommand := &FakeCommand{Stdouts: []string{"", remuxedMetadata}}
	processor := &FFmpegAudioProcessor{Command: fakeCommand}

	err := processor.WriteMetadataO("in.m4b", "out.m4b", withoutSeries, false)
	require.NoError(t, err)

	assert.Empty(t, fakeCommand.CommandsNamed(AtomicParsleyCommand))
}

func TestFFmpegAudioProcessor_AddMetadata_MovMuxerDropsTags_WritesThemBackWithAtomicParsley(t *testing.T) {
	dir := t.TempDir()
	m4bFile := filepath.Join(dir, "book.m4b")
	require.NoError(t, os.WriteFile(m4bFile, []byte("audio"), 0600))

	tempFile := filepath.Join(dir, "book.withMetadata.m4b")
	require.NoError(t, os.WriteFile(tempFile, []byte("remuxed"), 0600))

	fakeCommand := &FakeCommand{Stdouts: []string{"", remuxedMetadata, remuxedMetadata, ""}}
	processor := &FFmpegAudioProcessor{Command: fakeCommand}

	err := processor.AddMetadata(m4bFile, audibleMetadata, "Book")
	require.NoError(t, err)

	require.Len(t, fakeCommand.CommandsNamed(AtomicParsleyCommand), 1)
	assert.Contains(t, fakeCommand.CommandsNamed(AtomicParsleyCommand)[0], "name=SERIES")
	assert.Contains(
		t,
		fakeCommand.CommandsNamed(AtomicParsleyCommand)[0],
		tempFile,
	)
}

func TestFFmpegAudioProcessor_AddCover_MovMuxerDropsTags_WritesThemBackWithAtomicParsley(t *testing.T) {
	dir := t.TempDir()
	m4bFile := filepath.Join(dir, "book.m4b")
	require.NoError(t, os.WriteFile(m4bFile, []byte("audio"), 0600))

	require.NoError(t, os.WriteFile(filepath.Join(dir, "book.withCover.m4b"), []byte("remuxed"), 0600))

	fakeCommand := &FakeCommand{Stdouts: []string{audibleMetadata, "", remuxedMetadata, audibleMetadata, ""}}
	processor := &FFmpegAudioProcessor{Command: fakeCommand}

	err := processor.AddCover(m4bFile, filepath.Join(dir, "cover.jpg"))
	require.NoError(t, err)

	require.Len(t, fakeCommand.CommandsNamed(AtomicParsleyCommand), 1)
	assert.Contains(t, fakeCommand.CommandsNamed(AtomicParsleyCommand)[0], "name=SERIES")
}

func TestFFmpegAudioProcessor_AddCover_AtomicParsleyMissing_LeavesTheM4bAlone(t *testing.T) {
	dir := t.TempDir()
	m4bFile := filepath.Join(dir, "book.m4b")
	require.NoError(t, os.WriteFile(m4bFile, []byte("audio"), 0600))

	fakeCommand := &FakeCommand{
		Stdouts:       []string{audibleMetadata, "", remuxedMetadata, audibleMetadata},
		MissingOnPath: []string{AtomicParsleyCommand},
	}
	processor := &FFmpegAudioProcessor{Command: fakeCommand}

	err := processor.AddCover(m4bFile, filepath.Join(dir, "cover.jpg"))
	require.Error(t, err)

	content, readErr := os.ReadFile(m4bFile)
	require.NoError(t, readErr)
	assert.Equal(t, "audio", string(content))
}

func TestLostTags_TagPresentWithDifferentValue_IsNotReported(t *testing.T) {
	wanted := ";FFMETADATA1\nencoder=narr\n"
	written := ";FFMETADATA1\nencoder=Lavf61.7.100\n"

	assert.Empty(t, lostTags(wanted, written))
}

func TestLostTags_EmptyValue_IsNotReported(t *testing.T) {
	wanted := ";FFMETADATA1\nSERIES=\n"

	assert.Empty(t, lostTags(wanted, ";FFMETADATA1\n"))
}

func TestIsMP4(t *testing.T) {
	assert.True(t, isMP4("a.m4b"))
	assert.True(t, isMP4("a.M4A"))
	assert.True(t, isMP4("a.mp4"))
	assert.False(t, isMP4("a.mp3"))
	assert.False(t, isMP4("a.flac"))
}
