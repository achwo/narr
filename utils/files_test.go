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
