package generate

import (
	"fmt"
	"strings"
)

// Language identifies the implementation generating an OS handler artifact.
type Language string

const (
	LanguageGo     Language = "go"
	LanguageTS     Language = "ts"
	LanguagePython Language = "py"
	LanguageRust   Language = "rs"
	LanguagePHP    Language = "php"
)

// HandlerSpec describes one OS-level handler registration artifact.
type HandlerSpec struct {
	Vendor      string
	App         string
	Instance    string
	Language    Language
	Scheme      string
	Version     string
	Channel     string
	AppPath     string
	DisplayName string
}

// HandlerID returns the stable artifact identity for this handler.
func (s HandlerSpec) HandlerID() (string, error) {
	if err := s.Validate(); err != nil {
		return "", err
	}

	parts := []string{s.Vendor, s.App}
	if s.Instance != "" {
		parts = append(parts, s.Instance)
	}
	parts = append(parts, string(s.Language), s.Scheme)
	return strings.Join(parts, "."), nil
}

// Validate checks the required fields and safe artifact identity constraints.
func (s HandlerSpec) Validate() error {
	required := map[string]string{
		"vendor":   s.Vendor,
		"app":      s.App,
		"language": string(s.Language),
		"scheme":   s.Scheme,
		"app_path": s.AppPath,
	}
	for field, value := range required {
		if value == "" {
			return fmt.Errorf("generate: %s must not be empty", field)
		}
	}

	switch s.Language {
	case LanguageGo, LanguageTS, LanguagePython, LanguageRust, LanguagePHP:
	default:
		return fmt.Errorf("generate: unsupported language %q", s.Language)
	}

	for field, value := range map[string]string{
		"vendor":   s.Vendor,
		"app":      s.App,
		"instance": s.Instance,
		"language": string(s.Language),
		"scheme":   s.Scheme,
	} {
		if strings.ContainsAny(value, `/\`) {
			return fmt.Errorf("generate: %s must not contain path separators", field)
		}
	}

	return nil
}

func (s HandlerSpec) displayName() string {
	if s.DisplayName != "" {
		return s.DisplayName
	}
	id, err := s.HandlerID()
	if err != nil {
		return s.App
	}
	return id
}

// Snippet returns a platform-specific configuration snippet for registering
// the given URL scheme. Use this when runtime registration is not possible.
//
// Supported platforms: "macos", "ios", "linux", "windows".
func Snippet(platform string, spec HandlerSpec) (string, error) {
	switch platform {
	case "macos", "ios":
		return PlistSnippet(spec)
	case "linux":
		return DesktopFile(spec)
	case "windows":
		return WindowsRegSnippet(spec)
	default:
		return "", fmt.Errorf("generate: unknown platform %q", platform)
	}
}
