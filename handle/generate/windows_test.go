package generate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"hop.top/uri/handle/generate"
)

func TestWindowsRegSnippet_ContainsRequired(t *testing.T) {
	spec := testSpec()
	spec.AppPath = `C:\Program Files\task\task.exe`
	out, err := generate.WindowsRegSnippet(spec)
	require.NoError(t, err)
	assert.Contains(t, out, "Windows Registry Editor Version 5.00")
	assert.Contains(t, out, `Software\Classes\task`)
	assert.Contains(t, out, "URL Protocol")
	assert.Contains(t, out, "HopHandlerID")
	assert.Contains(t, out, "hop-top.scheme.go.task")
	assert.Contains(t, out, `task.exe`)
}

func TestWindowsRegSnippet_EmptyScheme(t *testing.T) {
	spec := testSpec()
	spec.Scheme = ""
	_, err := generate.WindowsRegSnippet(spec)
	require.Error(t, err)
}
