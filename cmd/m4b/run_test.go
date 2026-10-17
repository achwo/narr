package m4b

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCmd_OutFlag_PrintsOnlyThePathBelowOutOnStdout(t *testing.T) {
	requireCommands(t, "ffmpeg", "ffprobe")
	dir := projectDir(t)
	out := t.TempDir()

	stdout, stderr, err := executeCapturingStdio(t, "run", dir, "--out", out)
	require.NoError(t, err, stderr)

	expected := filepath.Join(out, "Max Heller", "Die Naht", "Die Naht.m4b")
	assert.Equal(t, expected+"\n", stdout)
	assert.FileExists(t, expected)
	assert.Contains(t, stderr, "Running on "+dir+"\n")
	assert.Contains(t, stderr, "Concating and encoding 2 files\n")
}

func TestRunCmd_AlreadyCompleted_StillPrintsThePathOnStdout(t *testing.T) {
	requireCommands(t, "ffmpeg", "ffprobe")
	dir := projectDir(t)
	out := t.TempDir()

	_, stderr, err := executeCapturingStdio(t, "run", dir, "--out", out)
	require.NoError(t, err, stderr)

	stdout, stderr, err := executeCapturingStdio(t, "run", dir, "--out", out)
	require.NoError(t, err, stderr)

	assert.Equal(t, filepath.Join(out, "Max Heller", "Die Naht", "Die Naht.m4b")+"\n", stdout)
	assert.Contains(t, stderr, "Skipping, as already completed\n")
}

func TestRunCmd_NarrYamlOutsideTheAudioDir_TakesTheAudioFromProjectPath(t *testing.T) {
	requireCommands(t, "ffmpeg", "ffprobe")
	audioDir := audioFiles(t)
	configDir := t.TempDir()
	out := t.TempDir()
	config := "projectPath: " + audioDir + "\ncoverPath: cover.jpg\nshouldConvert: true\n"
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "narr.yaml"), []byte(config), 0600))

	stdout, stderr, err := executeCapturingStdio(t, "run", configDir, "--out", out)
	require.NoError(t, err, stderr)

	expected := filepath.Join(out, "Max Heller", "Die Naht", "Die Naht.m4b")
	assert.Equal(t, expected+"\n", stdout)
	assert.FileExists(t, expected)
	assert.Contains(t, stderr, "Running on "+audioDir+"\n")
}

func TestCheckCmd_OutFlag_ShowsThePathRunWrites(t *testing.T) {
	requireCommands(t, "ffmpeg", "ffprobe")
	dir := projectDir(t)
	out := t.TempDir()
	expected := filepath.Join(out, "Max Heller", "Die Naht", "Die Naht.m4b")

	stdout, stderr, err := executeCapturingStdio(t, "check", "filename", dir, "--out", out)
	require.NoError(t, err, stderr)
	assert.Equal(t, expected+"\n", stdout)

	stdout, stderr, err = executeCapturingStdio(t, "check", dir, "--out", out)
	require.NoError(t, err, stderr)
	assert.Contains(t, stdout, "\n## Filename\n"+expected+"\n")
}

func projectDir(t *testing.T) string {
	t.Helper()

	dir := audioFiles(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "narr.yaml"), []byte("coverPath: cover.jpg\nshouldConvert: true\n"), 0600))

	return dir
}

func audioFiles(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	for i := 1; i <= 2; i++ {
		run(t, "ffmpeg", "-v", "error", "-f", "lavfi", "-i", "sine=frequency=440:sample_rate=48000:duration=1",
			"-metadata", "title=Teil "+strconv.Itoa(i),
			"-metadata", "artist=Max Heller",
			"-metadata", "album=Die Naht",
			"-metadata", "track="+strconv.Itoa(i),
			filepath.Join(dir, "part"+strconv.Itoa(i)+".flac"))
	}
	run(t, "ffmpeg", "-v", "error", "-f", "lavfi", "-i", "color=c=red:s=64x64:d=1",
		"-frames:v", "1", filepath.Join(dir, "cover.jpg"))

	return dir
}

func executeCapturingStdio(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	stdoutFile, err := os.Create(filepath.Join(t.TempDir(), "stdout"))
	require.NoError(t, err)
	stderrFile, err := os.Create(filepath.Join(t.TempDir(), "stderr"))
	require.NoError(t, err)

	originalStdout, originalStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = stdoutFile, stderrFile
	M4bCmd.SetArgs(args)
	err = M4bCmd.Execute()
	os.Stdout, os.Stderr = originalStdout, originalStderr

	require.NoError(t, stdoutFile.Close())
	require.NoError(t, stderrFile.Close())

	stdout, readErr := os.ReadFile(stdoutFile.Name())
	require.NoError(t, readErr)
	stderr, readErr := os.ReadFile(stderrFile.Name())
	require.NoError(t, readErr)

	return string(stdout), string(stderr), err
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
