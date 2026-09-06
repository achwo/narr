package metadata

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/achwo/narr/m4b"
	"github.com/achwo/narr/utils"
	"github.com/stretchr/testify/require"
)

func TestMetadataEdit_MatchingAlbumTag_RewritesTagInFile(t *testing.T) {
	dir := t.TempDir()
	file := createM4BFile(t, dir, "book.m4b", "Folge 5: Der Test")

	require.Equal(t, "Folge 5: Der Test", readTag(t, file, "album"))

	err := runMetadataCmd(
		t,
		"edit", dir,
		"--regex", `Folge (\d+): (.*)`,
		"--format", "%s. %s",
		"--tag", "album",
	)
	require.NoError(t, err)

	require.Equal(t, "5. Der Test", readTag(t, file, "album"))
}

func TestMetadataShow_ExistingFile_Succeeds(t *testing.T) {
	dir := t.TempDir()
	createM4BFile(t, dir, "book.m4b", "Folge 5: Der Test")

	err := runMetadataCmd(t, "show", dir, "--tag", "album")

	require.NoError(t, err)
}

// Cobra keeps flag state on the shared command instances, so run each subcommand only once per test binary.
func runMetadataCmd(t *testing.T, args ...string) error {
	t.Helper()

	var out bytes.Buffer
	MetadataCmd.SetOut(&out)
	MetadataCmd.SetErr(&out)
	MetadataCmd.SetArgs(args)

	return MetadataCmd.Execute()
}

func createM4BFile(t *testing.T, dir string, name string, album string) string {
	t.Helper()

	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available")
	}

	file := filepath.Join(dir, name)
	cmd := exec.Command(
		"ffmpeg", "-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "anullsrc=r=44100:cl=mono",
		"-t", "1", "-c:a", "aac",
		"-metadata", "album="+album,
		file,
	)

	output, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "could not create test file: %s", output)

	return file
}

func readTag(t *testing.T, file string, tag string) string {
	t.Helper()

	metadata, err := m4b.NewFFmpegAudioProcessor().ReadMetadata(file)
	require.NoError(t, err)

	values := utils.GetMetadataTagValues(metadata, []string{tag})
	require.Len(t, values, 1)

	return values[0].Value
}
