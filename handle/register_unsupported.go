//go:build !darwin && !linux && !windows

package handle

import "errors"

var ErrUnsupported = errors.New("handle: URL scheme registration not supported on this platform")

func Register(scheme, appPath string) error {
	return ErrUnsupported
}
