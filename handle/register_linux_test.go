//go:build linux

package handle_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"hop.top/uri/handle"
)

func TestRegister_Linux_EmptyScheme(t *testing.T) {
	err := handle.Register("", "/usr/bin/ctxt")
	require.Error(t, err)
	require.Contains(t, err.Error(), "scheme")
}

func TestRegister_Linux_EmptyApp(t *testing.T) {
	err := handle.Register("ctxt", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "app")
}
