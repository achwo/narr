package m4b

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecCmd_RunI_WithStdin_PipesStdinToCommand(t *testing.T) {
	command := &ExecCommand{}
	cmd := command.Create("cat")

	var stdout, stderr bytes.Buffer
	err := cmd.RunI(bytes.NewReader([]byte("some metadata")), &stdout, &stderr)

	require.NoError(t, err)
	assert.Equal(t, "some metadata", stdout.String())
}
