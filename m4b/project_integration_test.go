package m4b_test

import (
	"encoding/binary"
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/achwo/narr/m4b"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sampleRate      = 48000
	samplesPerMilli = sampleRate / 1000
	seamReach       = 50 * samplesPerMilli
)

func TestProject_ConvertToM4B_SineCutIntoWavParts_HasNoGapAtTheSeams(t *testing.T) {
	requireCommands(t, "ffmpeg", "ffprobe")
	t.Setenv("HOME", t.TempDir())

	dir := t.TempDir()
	partSamples := []int{96000, 96000, 96000, 96000, 96000}
	full, parts := sineParts(t, dir, partSamples)

	project := realProject(t, dir, parts, m4b.ProjectConfig{ShouldConvert: true})

	result, err := project.ConvertToM4B()
	require.NoError(t, err)

	assertSeamless(t, full, result, partSamples)
}

func TestProject_ConvertToM4B_M4aPartsWithoutConversion_HasNoGapAtTheSeams(t *testing.T) {
	requireCommands(t, "ffmpeg", "ffprobe")
	t.Setenv("HOME", t.TempDir())

	dir := t.TempDir()
	partSamples := []int{96000, 96000, 96000, 96000, 96000}
	full, wavParts := sineParts(t, dir, partSamples)
	parts := encodeParts(t, wavParts, ".m4a", "-c:a", "aac")

	project := realProject(t, dir, parts, m4b.ProjectConfig{ShouldConvert: false})

	result, err := project.ConvertToM4B()
	require.NoError(t, err)

	assertSeamless(t, full, result, partSamples)
}

func TestProject_ConvertToM4B_OneChapterPerMp3WithoutXingHeader_MarksSitWhereThePartsStart(t *testing.T) {
	requireCommands(t, "ffmpeg", "ffprobe", "mp4chaps")
	t.Setenv("HOME", t.TempDir())

	dir := t.TempDir()
	_, wavParts := sineParts(t, dir, []int{96000, 144000, 120000})
	parts := encodeParts(t, wavParts, ".mp3", "-c:a", "libmp3lame", "-q:a", "2", "-write_xing", "0")

	project := realProject(t, dir, parts, m4b.ProjectConfig{ShouldConvert: true, HasChapters: true})

	result, err := project.ConvertToM4B()
	require.NoError(t, err)

	starts := chapterStarts(t, result)
	require.Len(t, starts, len(parts))

	decodedBefore := 0
	for i, part := range parts {
		assert.InDelta(t, float64(decodedBefore)/sampleRate, starts[i], 0.021, "start of chapter %d", i+1)
		decodedBefore += len(decodeSamples(t, part))
	}
}

func encodeParts(t *testing.T, wavParts []string, ext string, codecArgs ...string) []string {
	t.Helper()

	parts := make([]string, 0, len(wavParts))
	for _, wavPart := range wavParts {
		part := strings.TrimSuffix(wavPart, ".wav") + ext
		args := append([]string{"-v", "error", "-i", wavPart}, codecArgs...)
		run(t, "ffmpeg", append(args, part)...)
		parts = append(parts, part)
	}
	return parts
}

func chapterStarts(t *testing.T, file string) []float64 {
	t.Helper()

	output := run(t, "ffprobe", "-v", "error", "-show_entries", "chapter=start_time", "-of", "csv=p=0", file)

	var starts []float64
	for _, line := range strings.Fields(output) {
		start, err := strconv.ParseFloat(line, 64)
		require.NoError(t, err)
		starts = append(starts, start)
	}
	return starts
}

func assertSeamless(t *testing.T, original string, joined string, partSamples []int) {
	t.Helper()

	originalSamples := decodeSamples(t, original)
	joinedSamples := decodeSamples(t, joined)

	assert.InDelta(t, len(originalSamples), len(joinedSamples), 2048, "sample count of the result")

	halfAmplitude := peak(originalSamples) / 2
	for _, seam := range seams(partSamples) {
		assert.GreaterOrEqual(
			t,
			quietestMillisecond(joinedSamples, seam),
			halfAmplitude,
			"level of the quietest 1 ms window within 50 ms of the seam at sample %d",
			seam,
		)
	}
}

func sineParts(t *testing.T, dir string, partSamples []int) (string, []string) {
	t.Helper()

	total := 0
	for _, samples := range partSamples {
		total += samples
	}

	full := filepath.Join(dir, "full.wav")
	run(t, "ffmpeg", "-v", "error", "-f", "lavfi",
		"-i", fmt.Sprintf("sine=frequency=440:sample_rate=%d", sampleRate),
		"-af", fmt.Sprintf("atrim=end_sample=%d", total),
		"-ac", "2", full)

	parts := make([]string, 0, len(partSamples))
	start := 0
	for i, samples := range partSamples {
		part := filepath.Join(dir, fmt.Sprintf("part%d.wav", i+1))
		run(t, "ffmpeg", "-v", "error", "-i", full,
			"-af", fmt.Sprintf("atrim=start_sample=%d:end_sample=%d,asetpts=N/SR/TB", start, start+samples),
			"-metadata", "title=Teil "+strconv.Itoa(i+1),
			"-metadata", "artist=Max Heller",
			"-metadata", "album=Die Naht",
			part)
		parts = append(parts, part)
		start += samples
	}

	return full, parts
}

func realProject(t *testing.T, dir string, parts []string, config m4b.ProjectConfig) *m4b.Project {
	t.Helper()

	cover := filepath.Join(dir, "cover.jpg")
	run(t, "ffmpeg", "-v", "error", "-f", "lavfi", "-i", "color=c=red:s=64x64:d=1",
		"-frames:v", "1", cover)

	config.ProjectPath = dir
	config.CoverPath = cover

	processor := m4b.NewFFmpegAudioProcessor()
	project, err := m4b.NewProjectWithDeps(config, m4b.ProjectDependencies{
		AudioFileProvider: &FakeAudioFileProvider{Files: parts},
		AudioProcessor:    processor,
		TrackFactory:      &m4b.FFmpegTrackFactory{AudioProcessor: processor},
	})
	require.NoError(t, err)

	return project
}

func seams(partSamples []int) []int {
	var seams []int
	position := 0
	for _, samples := range partSamples[:len(partSamples)-1] {
		position += samples
		seams = append(seams, position)
	}
	return seams
}

func decodeSamples(t *testing.T, file string) []int16 {
	t.Helper()

	raw, err := exec.Command("ffmpeg", "-v", "error", "-i", file, "-map", "0:a:0",
		"-ac", "1", "-ar", strconv.Itoa(sampleRate), "-f", "s16le", "-").Output()
	require.NoError(t, err, "could not decode %s", file)

	samples := make([]int16, len(raw)/2)
	for i := range samples {
		samples[i] = int16(binary.LittleEndian.Uint16(raw[2*i:]))
	}
	return samples
}

func peak(samples []int16) float64 {
	highest := 0.0
	for _, sample := range samples {
		highest = math.Max(highest, math.Abs(float64(sample)))
	}
	return highest
}

func quietestMillisecond(samples []int16, around int) float64 {
	quietest := math.Inf(1)
	first := max(0, around-seamReach)
	last := min(len(samples)-samplesPerMilli, around+seamReach)

	for start := first; start <= last; start++ {
		quietest = math.Min(quietest, rms(samples[start:start+samplesPerMilli]))
	}
	return quietest
}

func rms(samples []int16) float64 {
	sum := 0.0
	for _, sample := range samples {
		sum += float64(sample) * float64(sample)
	}
	return math.Sqrt(sum / float64(len(samples)))
}
