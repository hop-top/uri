package handle

import (
	"os"
	"strings"
	"sync"
)

// SupportsHyperlinks reports whether the current terminal supports
// OSC 8 hyperlinks. The result is cached after the first call.
func SupportsHyperlinks() bool {
	supportsOnce.Do(detectHyperlinks)
	return supportsLinks
}

// Linkify wraps a URI and label in OSC 8 hyperlink escape
// sequences if the terminal supports it. Returns plain label
// if terminal doesn't support OSC 8.
func Linkify(uri, label string) string {
	if !SupportsHyperlinks() {
		return label
	}
	return "\033]8;;" + uri + "\033\\" + label + "\033]8;;\033\\"
}

var (
	supportsOnce  sync.Once
	supportsLinks bool
)

func detectHyperlinks() {
	if os.Getenv("NO_COLOR") != "" {
		return
	}
	term := os.Getenv("TERM")
	if term == "dumb" {
		return
	}

	// Windows Terminal
	if os.Getenv("WT_SESSION") != "" {
		supportsLinks = true
		return
	}

	termProg := os.Getenv("TERM_PROGRAM")
	switch termProg {
	case "iTerm.app", "WezTerm", "ghostty":
		supportsLinks = true
		return
	case "Apple_Terminal":
		return
	}

	lcTerm := os.Getenv("LC_TERMINAL")
	if lcTerm == "iTerm2" {
		supportsLinks = true
		return
	}

	// xterm-256color with known capable terminals via TERM_PROGRAM
	// already handled above; check for kitty/foot in TERM itself
	if strings.HasPrefix(term, "xterm") {
		lower := strings.ToLower(term)
		for _, t := range []string{"kitty", "foot"} {
			if strings.Contains(lower, t) {
				supportsLinks = true
				return
			}
		}
	}

	// kitty sets TERM=xterm-kitty
	if strings.Contains(term, "kitty") {
		supportsLinks = true
		return
	}

	// foot sets TERM=foot or TERM=foot-extra
	if strings.HasPrefix(term, "foot") {
		supportsLinks = true
		return
	}
}
