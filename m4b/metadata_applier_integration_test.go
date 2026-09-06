package m4b_test

import (
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

func requireCommands(t *testing.T, names ...string) {
	t.Helper()
	for _, name := range names {
		if _, err := exec.LookPath(name); err != nil {
			t.Skipf("%s not on the PATH", name)
		}
	}
}

func run(t *testing.T, name string, args ...string) {
	t.Helper()
	output, err := exec.Command(name, args...).CombinedOutput()
	require.NoError(t, err, "%s failed: %s", name, string(output))
}

func formatTags(t *testing.T, file string) string {
	t.Helper()
	output, err := exec.Command(
		"ffprobe", "-v", "error", "-show_entries", "format_tags", "-of", "default", file,
	).Output()
	require.NoError(t, err)
	return strings.TrimSpace(string(output))
}
