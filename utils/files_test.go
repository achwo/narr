package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHardLinkCount_FileWithoutLinks_ReturnsOne(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file.m4b")
	require.NoError(t, os.WriteFile(file, []byte{}, 0644))

	count, err := HardLinkCount(file)

	require.NoError(t, err)
	assert.Equal(t, uint64(1), count)
}

func TestHardLinkCount_HardLinkedFile_ReturnsTwo(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file.m4b")
	require.NoError(t, os.WriteFile(file, []byte{}, 0644))
	require.NoError(t, os.Link(file, filepath.Join(dir, "copy.m4b")))

	count, err := HardLinkCount(file)

	require.NoError(t, err)
	assert.Equal(t, uint64(2), count)
}

func TestHardLinkCount_MissingFile_ReturnsError(t *testing.T) {
	_, err := HardLinkCount(filepath.Join(t.TempDir(), "missing.m4b"))

	assert.Error(t, err)
}

func TestMoveFile_SameVolume_RenamesTheFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "concat.m4b")
	dst := filepath.Join(dir, "book.m4b")
	require.NoError(t, os.WriteFile(src, []byte("audio"), 0644))

	require.NoError(t, MoveFile(src, dst))

	assert.NoFileExists(t, src)
	content, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, "audio", string(content))
}

func TestCopyFile_TargetDirExists_CopiesContentAndModeWithoutLeftovers(t *testing.T) {
	src := filepath.Join(t.TempDir(), "concat.m4b")
	require.NoError(t, os.WriteFile(src, []byte("audio"), 0640))
	targetDir := t.TempDir()
	dst := filepath.Join(targetDir, "book.m4b")

	require.NoError(t, copyFile(src, dst))

	content, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, "audio", string(content))

	info, err := os.Stat(dst)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0640), info.Mode().Perm())

	entries, err := os.ReadDir(targetDir)
	require.NoError(t, err)
	assert.Len(t, entries, 1, "only the copy itself should be in the target dir")
	assert.FileExists(t, src, "removing the source is left to MoveFile")
}

func TestCopyFile_TargetIsADirectory_FailsAndLeavesNoTempFile(t *testing.T) {
	src := filepath.Join(t.TempDir(), "concat.m4b")
	require.NoError(t, os.WriteFile(src, []byte("audio"), 0644))
	targetDir := t.TempDir()
	dst := filepath.Join(targetDir, "book.m4b")
	require.NoError(t, os.Mkdir(dst, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dst, "keep"), nil, 0644))

	require.Error(t, copyFile(src, dst))

	entries, err := os.ReadDir(targetDir)
	require.NoError(t, err)
	assert.Len(t, entries, 1, "the temp copy should have been removed")
}
