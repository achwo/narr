package m4b_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/achwo/narr/m4b"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetadataApplier_ApplyToFile_RealAudibleStyleM4b_KeepsFreeformAtoms runs the
// real ffmpeg and AtomicParsley against a real m4b, because the whole point of
// the freeform handling is what the mov muxer does to a file on disk.
func TestMetadataApplier_ApplyToFile_RealAudibleStyleM4b_KeepsFreeformAtoms(t *testing.T) {
	requireCommands(t, "ffmpeg", "ffprobe", m4b.AtomicParsleyCommand)

	file := filepath.Join(t.TempDir(), "book.m4b")

	run(t, "ffmpeg", "-v", "error", "-f", "lavfi", "-i", "sine=frequency=440:duration=2",
		"-c:a", "aac", "-b:a", "32k",
		"-metadata", "title=Folge 3: Der Fall",
		"-metadata", "artist=Max Heller",
		"-metadata", "album=Die Reihe",
		file,
	)
	run(t, m4b.AtomicParsleyCommand, file,
		"--rDNSatom", "Die Reihe", "name=SERIES", "domain=com.apple.iTunes",
		"--rDNSatom", "3", "name=PART", "domain=com.apple.iTunes",
		"--rDNSatom", "B01ABCDEFG", "name=AUDIBLE_ASIN", "domain=com.apple.iTunes",
		"--overWrite",
	)

	before := formatTags(t, file)
	require.Contains(t, before, "TAG:SERIES=Die Reihe")

	applier := m4b.NewMetadataApplier()
	applier.Out = nil

	rules := []m4b.MetadataRule{
		{Type: "regex", Tag: "title", Regex: `^Folge (\d+): (.+)$`, Format: "%s. %s"},
	}

	result, err := applier.ApplyToFile(file, rules, m4b.ApplyOptions{})
	require.NoError(t, err)
	require.Equal(t, m4b.StatusUpdated, result.Status)

	after := formatTags(t, file)

	assert.Contains(t, after, "TAG:title=3. Der Fall", "the rule should have been applied")
	assert.Contains(t, after, "TAG:SERIES=Die Reihe")
	assert.Contains(t, after, "TAG:PART=3")
	assert.Contains(t, after, "TAG:AUDIBLE_ASIN=B01ABCDEFG")
}

// TestMetadataApplier_ApplyToFile_RealM4bWithCoverAndChapters_PatchesInPlace
// checks that the tags are patched into the file instead of being remuxed into
// a new one. The freeform atom in a foreign domain is the evidence: the mov
// muxer drops it, and narr can only write it back under com.apple.iTunes.
func TestMetadataApplier_ApplyToFile_RealM4bWithCoverAndChapters_PatchesInPlace(t *testing.T) {
	requireCommands(t, "ffmpeg", "ffprobe", m4b.AtomicParsleyCommand)

	dir := t.TempDir()
	file := filepath.Join(dir, "book.m4b")
	plain := filepath.Join(dir, "plain.m4b")
	tagged := filepath.Join(dir, "tagged.m4b")
	cover := filepath.Join(dir, "cover.jpg")
	metadataFile := filepath.Join(dir, "metadata.txt")

	require.NoError(t, os.WriteFile(metadataFile, []byte(`;FFMETADATA1
title=Folge 3: Der Fall
artist=Max Heller
album=Die Reihe
[CHAPTER]
TIMEBASE=1/1000
START=0
END=1000
title=Eins
[CHAPTER]
TIMEBASE=1/1000
START=1000
END=2000
title=Zwei
`), 0600))

	run(t, "ffmpeg", "-v", "error", "-f", "lavfi", "-i", "sine=frequency=440:duration=2",
		"-c:a", "aac", "-b:a", "32k", plain)
	run(t, "ffmpeg", "-v", "error", "-f", "lavfi", "-i", "color=c=red:s=64x64:d=1",
		"-frames:v", "1", cover)
	run(t, "ffmpeg", "-v", "error", "-i", plain, "-i", cover, "-map", "0", "-map", "1",
		"-c", "copy", "-disposition:v", "attached_pic", tagged)
	run(t, "ffmpeg", "-v", "error", "-i", tagged, "-i", metadataFile, "-map", "0",
		"-map_metadata", "1", "-map_chapters", "1", "-c", "copy",
		"-disposition:v", "attached_pic", file)
	run(t, m4b.AtomicParsleyCommand, file,
		"--rDNSatom", "Die Reihe", "name=SERIES", "domain=com.apple.iTunes",
		"--rDNSatom", "bleib hier", "name=NARRTEST", "domain=org.narr",
		"--overWrite",
	)

	applier := m4b.NewMetadataApplier()
	applier.Out = nil

	rules := []m4b.MetadataRule{
		{Type: "regex", Tag: "title", Regex: `^Folge (\d+): (.+)$`, Format: "%s. %s"},
	}

	result, err := applier.ApplyToFile(file, rules, m4b.ApplyOptions{})
	require.NoError(t, err)
	require.Equal(t, m4b.StatusUpdated, result.Status)

	assert.Contains(t, formatTags(t, file), "TAG:title=3. Der Fall")
	assert.Contains(t, formatTags(t, file), "TAG:SERIES=Die Reihe")
	assert.Contains(
		t,
		run(t, m4b.AtomicParsleyCommand, file, "-t"),
		"[org.narr;NARRTEST]",
		"the atom in the foreign domain proves the file was patched, not remuxed",
	)

	chapters := run(t, "ffprobe", "-v", "error", "-show_chapters", "-of", "csv", file)
	assert.Contains(t, chapters, "Eins")
	assert.Contains(t, chapters, "Zwei")

	streams := run(t, "ffprobe", "-v", "error", "-select_streams", "v",
		"-show_entries", "stream=codec_name", "-of", "csv=p=0", file)
	assert.Contains(t, streams, "mjpeg", "the cover should still be there")
}

func requireCommands(t *testing.T, names ...string) {
	t.Helper()
	for _, name := range names {
		if _, err := exec.LookPath(name); err != nil {
			t.Skipf("%s not on the PATH", name)
		}
	}
}

func run(t *testing.T, name string, args ...string) string {
	t.Helper()
	output, err := exec.Command(name, args...).CombinedOutput()
	require.NoError(t, err, "%s failed: %s", name, string(output))
	return string(output)
}

func formatTags(t *testing.T, file string) string {
	t.Helper()
	output, err := exec.Command(
		"ffprobe", "-v", "error", "-show_entries", "format_tags", "-of", "default", file,
	).Output()
	require.NoError(t, err)
	return strings.TrimSpace(string(output))
}
