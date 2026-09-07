package m4b_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/achwo/narr/m4b"
	"github.com/achwo/narr/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fileMetadata = `;FFMETADATA1
title=Chapter 01-02: Star dust
artist=Hans Wurst
album=The Book
comment=some note
`

func TestMetadataApplier_ApplyToFile_RegexRule_WritesUpdatedMetadata(t *testing.T) {
	file, applier, processor := setupApplier(t, "file1.m4b", fileMetadata)

	rules := []m4b.MetadataRule{
		{Type: "regex", Tag: "title", Regex: `^Chapter (\d+)-\d+: (.+)$`, Format: "%s - %s"},
	}

	result, err := applier.ApplyToFile(file, rules, m4b.ApplyOptions{})
	require.NoError(t, err)

	assert.Equal(t, m4b.StatusUpdated, result.Status)
	assert.Equal(
		t,
		[]m4b.TagChange{{
			Tag:    "title",
			Before: "Chapter 01-02: Star dust",
			After:  "01 - Star dust",
			Kind:   m4b.ChangeUpdated,
		}},
		result.Changes,
	)
	assert.Equal(
		t,
		`;FFMETADATA1
title=01 - Star dust
artist=Hans Wurst
album=The Book
comment=some note
`,
		processor.Written[file],
	)
}

func TestMetadataApplier_ApplyToFile_SetRuleWithNewTag_AppendsTag(t *testing.T) {
	file, applier, processor := setupApplier(t, "file1.mp3", fileMetadata)

	rules := []m4b.MetadataRule{{Type: "set", Tag: "genre", Value: "Hoerspiel"}}

	result, err := applier.ApplyToFile(file, rules, m4b.ApplyOptions{})
	require.NoError(t, err)

	assert.Equal(t, m4b.StatusUpdated, result.Status)
	assert.Equal(
		t,
		[]m4b.TagChange{{Tag: "genre", After: "Hoerspiel", Kind: m4b.ChangeAdded}},
		result.Changes,
	)
	assert.Contains(t, processor.Written[file], "\ngenre=Hoerspiel\n")
}

func TestMetadataApplier_ApplyToFile_DeleteRule_RemovesTagFromMetadata(t *testing.T) {
	file, applier, processor := setupApplier(t, "file1.flac", fileMetadata)

	rules := []m4b.MetadataRule{{Type: "delete", Tag: "comment"}}

	result, err := applier.ApplyToFile(file, rules, m4b.ApplyOptions{})
	require.NoError(t, err)

	assert.Equal(
		t,
		[]m4b.TagChange{{Tag: "comment", Before: "some note", Kind: m4b.ChangeDeleted}},
		result.Changes,
	)
	assert.NotContains(t, processor.Written[file], "comment")
}

func TestMetadataApplier_ApplyToFile_NoMatchingRule_LeavesFileUntouched(t *testing.T) {
	file, applier, processor := setupApplier(t, "file1.m4a", fileMetadata)

	rules := []m4b.MetadataRule{{Type: "set", Tag: "album", Value: "The Book"}}

	result, err := applier.ApplyToFile(file, rules, m4b.ApplyOptions{})
	require.NoError(t, err)

	assert.Equal(t, m4b.StatusUnchanged, result.Status)
	assert.Empty(t, result.Changes)
	assert.Empty(t, processor.Written)
}

func TestMetadataApplier_ApplyToFile_DryRun_DoesNotWrite(t *testing.T) {
	file, applier, processor := setupApplier(t, "file1.m4b", fileMetadata)

	rules := []m4b.MetadataRule{{Type: "set", Tag: "album", Value: "Another Book"}}

	result, err := applier.ApplyToFile(file, rules, m4b.ApplyOptions{DryRun: true})
	require.NoError(t, err)

	assert.Equal(t, m4b.StatusDryRun, result.Status)
	assert.Len(t, result.Changes, 1)
	assert.Empty(t, processor.Written)
}

func TestMetadataApplier_ApplyToFile_HardLinkedFile_SkipsFile(t *testing.T) {
	file, applier, processor := setupApplier(t, "file1.m4b", fileMetadata)
	require.NoError(t, os.Link(file, filepath.Join(filepath.Dir(file), "testcopy.m4b")))

	rules := []m4b.MetadataRule{{Type: "set", Tag: "album", Value: "Another Book"}}

	result, err := applier.ApplyToFile(file, rules, m4b.ApplyOptions{})
	require.NoError(t, err)

	assert.Equal(t, m4b.StatusSkippedHardLink, result.Status)
	assert.Empty(t, processor.Written)
}

func TestMetadataApplier_ApplyToFile_ChaptersInMetadata_KeepsChapters(t *testing.T) {
	metadata := ";FFMETADATA1\ntitle=Book\n[CHAPTER]\nTIMEBASE=1/1000\nSTART=0\nEND=500\ntitle=Ch1\n"
	file, applier, processor := setupApplier(t, "file1.m4b", metadata)

	rules := []m4b.MetadataRule{{Type: "set", Tag: "title", Value: "Other Book"}}

	_, err := applier.ApplyToFile(file, rules, m4b.ApplyOptions{})
	require.NoError(t, err)

	assert.Equal(
		t,
		";FFMETADATA1\ntitle=Other Book\n[CHAPTER]\nTIMEBASE=1/1000\nSTART=0\nEND=500\ntitle=Ch1\n",
		processor.Written[file],
	)
}

func TestMetadataApplier_ApplyToFile_Changes_PrintsDiffBeforeWriting(t *testing.T) {
	file, applier, _ := setupApplier(t, "file1.m4b", fileMetadata)
	out := &bytes.Buffer{}
	applier.Out = out

	rules := []m4b.MetadataRule{{Type: "set", Tag: "album", Value: "Another Book"}}

	_, err := applier.ApplyToFile(file, rules, m4b.ApplyOptions{})
	require.NoError(t, err)

	assert.Contains(t, out.String(), file)
	assert.Contains(t, out.String(), "album : The Book -> Another Book")
}

func TestMetadataApplier_ApplyToFile_MissingTagForRegexRule_ReturnsError(t *testing.T) {
	file, applier, _ := setupApplier(t, "file1.m4b", fileMetadata)

	rules := []m4b.MetadataRule{
		{Type: "regex", Tag: "publisher", Regex: `^(.+)$`, Format: "%s"},
	}

	_, err := applier.ApplyToFile(file, rules, m4b.ApplyOptions{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "publisher")
}

func TestMetadataApplier_ApplyToFile_WriteFails_ReturnsError(t *testing.T) {
	file, applier, processor := setupApplier(t, "file1.m4b", fileMetadata)
	processor.ErrWrite = errors.New("ffmpeg failed")

	rules := []m4b.MetadataRule{{Type: "set", Tag: "album", Value: "Another Book"}}

	_, err := applier.ApplyToFile(file, rules, m4b.ApplyOptions{})

	require.Error(t, err)
	assert.ErrorIs(t, err, processor.ErrWrite)
}

func TestMetadataApplier_ApplyToProject_SupportedFormats_AppliesToAllAudioFiles(t *testing.T) {
	dir := t.TempDir()
	audioFiles := []string{"a.m4b", "b.m4a", "c.mp3", "d.flac"}
	data := make(map[string]m4b.FileData)

	for _, name := range audioFiles {
		file := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(file, []byte{}, 0644))
		data[file] = m4b.FileData{Metadata: fileMetadata}
	}

	require.NoError(t, os.WriteFile(filepath.Join(dir, "cover.jpg"), []byte{}, 0644))

	processor := &m4b.NullAudioProcessor{Data: data}
	applier := &m4b.MetadataApplier{AudioProcessor: processor}

	config := m4b.ProjectConfig{
		ProjectPath:   dir,
		MetadataRules: []m4b.MetadataRule{{Type: "set", Tag: "album", Value: "Another Book"}},
	}

	results, err := applier.ApplyToProject(config, m4b.ApplyOptions{})
	require.NoError(t, err)

	require.Len(t, results, len(audioFiles))
	for _, result := range results {
		assert.Equal(t, m4b.StatusUpdated, result.Status)
	}
	assert.Len(t, processor.Written, len(audioFiles))
}

func TestMetadataApplier_ApplyToFile_ProcessorCanPatch_PatchesOnlyTheChangedTags(t *testing.T) {
	metadata := ";FFMETADATA1\ntitle=Book\nSERIES=Die Reihe\ncomment=some note\n"
	file, applier, processor := setupApplier(t, "file1.m4b", metadata)
	patcher := &patchingProcessor{NullAudioProcessor: processor}
	applier.AudioProcessor = patcher

	rules := []m4b.MetadataRule{
		{Type: "set", Tag: "series", Value: "Andere Reihe"},
		{Type: "delete", Tag: "comment"},
	}

	result, err := applier.ApplyToFile(file, rules, m4b.ApplyOptions{})
	require.NoError(t, err)

	assert.Equal(t, m4b.StatusUpdated, result.Status)
	assert.Equal(
		t,
		[]utils.TagWithValue{
			{Tag: "SERIES", Value: "Andere Reihe"},
			{Tag: "comment", Value: ""},
		},
		patcher.Patched[file],
	)
	assert.Empty(t, processor.Written, "a patched file must not be remuxed")
}

func TestMetadataApplier_ApplyToFile_PatchRefused_WritesTheWholeMetadata(t *testing.T) {
	file, applier, processor := setupApplier(t, "file1.m4b", fileMetadata)
	applier.AudioProcessor = &patchingProcessor{NullAudioProcessor: processor, Refuse: true}

	rules := []m4b.MetadataRule{{Type: "set", Tag: "album", Value: "Another Book"}}

	result, err := applier.ApplyToFile(file, rules, m4b.ApplyOptions{})
	require.NoError(t, err)

	assert.Equal(t, m4b.StatusUpdated, result.Status)
	assert.Contains(t, processor.Written[file], "album=Another Book")
}

// patchingProcessor is a NullAudioProcessor that also patches tags in place.
type patchingProcessor struct {
	*m4b.NullAudioProcessor
	Patched map[string][]utils.TagWithValue
	Refuse  bool
}

func (p *patchingProcessor) PatchMP4Tags(file string, tags []utils.TagWithValue, _ bool) bool {
	if p.Patched == nil {
		p.Patched = make(map[string][]utils.TagWithValue)
	}
	p.Patched[file] = tags

	return !p.Refuse
}

func setupApplier(t *testing.T, name string, metadata string) (string, *m4b.MetadataApplier, *m4b.NullAudioProcessor) {
	t.Helper()

	file := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(file, []byte{}, 0644))

	processor := &m4b.NullAudioProcessor{
		Data: map[string]m4b.FileData{file: {Metadata: metadata}},
	}

	return file, &m4b.MetadataApplier{AudioProcessor: processor}, processor
}
