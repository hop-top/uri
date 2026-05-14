package generate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"hop.top/uri/handle/generate"
)

func TestDesktopFile_ContainsRequired(t *testing.T) {
	spec := testSpec()
	out, err := generate.DesktopFile(spec)
	require.NoError(t, err)
	assert.Contains(t, out, "[Desktop Entry]")
	assert.Contains(t, out, "MimeType=x-scheme-handler/task")
	assert.Contains(t, out, "Exec=/usr/bin/task %u")
	assert.Contains(t, out, "Name=Hop Task Handler")
	assert.Contains(t, out, "X-Hop-Handler-ID=hop-top.scheme.go.task")
	assert.Contains(t, out, "Type=Application")
}

func TestDesktopFilename_UniqueByHandlerID(t *testing.T) {
	got, err := generate.DesktopFilename(testSpec())
	require.NoError(t, err)
	assert.Equal(t, "hop-top.scheme.go.task.desktop", got)
}
