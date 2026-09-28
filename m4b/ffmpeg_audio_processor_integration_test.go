package m4b_test

import (
	"fmt"
	"path/filepath"
	"slices"
	"testing"

	"github.com/achwo/narr/m4b"
	"github.com/stretchr/testify/require"
)

func TestFFmpegAudioProcessor_ConcatAndEncode_300Parts_HasNoGapAtTheSeams(t *testing.T) {
	requireCommands(t, "ffmpeg")

	dir := t.TempDir()
	partSamples := slices.Repeat([]int{9600}, 300)
	full, parts := equalSineParts(t, dir, len(partSamples), partSamples[0])

	result, err := m4b.NewFFmpegAudioProcessor().ConcatAndEncode(parts, dir)
	require.NoError(t, err)

	assertSeamless(t, full, result, partSamples)
}

func equalSineParts(t *testing.T, dir string, count int, samplesEach int) (string, []string) {
	t.Helper()

	source := fmt.Sprintf("sine=frequency=440:sample_rate=%d:samples_per_frame=%d", sampleRate, samplesEach)
	trim := fmt.Sprintf("atrim=end_sample=%d", count*samplesEach)
	seconds := fmt.Sprintf("%f", float64(samplesEach)/sampleRate)

	full := filepath.Join(dir, "full.wav")
	run(t, "ffmpeg", "-v", "error", "-f", "lavfi", "-i", source, "-af", trim, "-ac", "2", full)
	run(t, "ffmpeg", "-v", "error", "-f", "lavfi", "-i", source, "-af", trim, "-ac", "2",
		"-f", "segment", "-segment_time", seconds, filepath.Join(dir, "part%03d.wav"))

	parts, err := filepath.Glob(filepath.Join(dir, "part*.wav"))
	require.NoError(t, err)
	require.Len(t, parts, count)

	return full, parts
}
