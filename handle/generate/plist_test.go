package generate_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"hop.top/uri/handle/generate"
)

func TestPlistSnippet_ContainsScheme(t *testing.T) {
	snippet, err := generate.PlistSnippet(testSpec())
	require.NoError(t, err)
	assert.Contains(t, snippet, "<key>CFBundleURLSchemes</key>")
	assert.Contains(t, snippet, "<key>CFBundleURLName</key>")
	assert.Contains(t, snippet, "<string>hop-top.scheme.go.task</string>")
	assert.Contains(t, snippet, "<string>task</string>")
}

func TestPatchPlist_InjectsFragment(t *testing.T) {
	input := "<?xml version=\"1.0\"?>\n<plist version=\"1.0\">\n<dict>\n\t<key>CFBundleIdentifier</key>\n\t<string>com.example.app</string>\n</dict>\n</plist>"
	out, err := generate.PatchPlist(strings.NewReader(input), testSpec())
	require.NoError(t, err)
	assert.Contains(t, out, "CFBundleURLSchemes")
	assert.Contains(t, out, "<string>task</string>")
	assert.Contains(t, out, "<string>hop-top.scheme.go.task</string>")
	assert.Contains(t, out, "</plist>")
}

func TestPatchPlist_EmptyScheme(t *testing.T) {
	spec := testSpec()
	spec.Scheme = ""
	_, err := generate.PatchPlist(strings.NewReader("<plist/>"), spec)
	require.Error(t, err)
}
