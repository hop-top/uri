//go:build darwin

package handle_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"hop.top/uri/handle"
)

func TestRegister_Darwin_EmptyScheme(t *testing.T) {
	err := handle.Register("", "/Applications/Foo.app")
	require.Error(t, err)
	require.Contains(t, err.Error(), "scheme")
}

func TestRegister_Darwin_EmptyApp(t *testing.T) {
	err := handle.Register("ctxt", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "app")
}
