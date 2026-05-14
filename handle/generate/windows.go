package generate

import "fmt"

// WindowsRegSnippet returns .reg file content for registering the URL scheme on Windows.
func WindowsRegSnippet(spec HandlerSpec) (string, error) {
	id, err := spec.HandlerID()
	if err != nil {
		return "", err
	}
	displayName := spec.displayName()
	return fmt.Sprintf(
		"Windows Registry Editor Version 5.00\r\n\r\n[HKEY_CURRENT_USER\\Software\\Classes\\%s]\r\n@=\"URL:%s Protocol\"\r\n\"URL Protocol\"=\"\"\r\n\"FriendlyTypeName\"=\"%s\"\r\n\"HopHandlerID\"=\"%s\"\r\n\r\n[HKEY_CURRENT_USER\\Software\\Classes\\%s\\shell\\open\\command]\r\n@=\"\\\"%s\\\" \\\"%%1\\\"\"\r\n",
		spec.Scheme, displayName, displayName, id, spec.Scheme, spec.AppPath,
	), nil
}
