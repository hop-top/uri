//go:build windows

package handle_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"hop.top/uri/handle"
)

func TestRegister_Windows_EmptyScheme(t *testing.T) {
	err := handle.Register("", `C:\Program Files\ctxt\ctxt.exe`)
	require.Error(t, err)
	require.Contains(t, err.Error(), "scheme")
}

func TestRegister_Windows_EmptyApp(t *testing.T) {
	err := handle.Register("ctxt", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "app")
}
