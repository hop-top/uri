package handle

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetDetection() {
	supportsOnce = sync.Once{}
	supportsLinks = false
}

func TestLinkify_Supported(t *testing.T) {
	resetDetection()
	supportsLinks = true
	supportsOnce.Do(func() {}) // mark as done

	got := Linkify("https://example.com", "click")
	assert.Equal(t, "\033]8;;https://example.com\033\\click\033]8;;\033\\", got)
}

func TestLinkify_Unsupported(t *testing.T) {
	resetDetection()
	supportsLinks = false
	supportsOnce.Do(func() {}) // mark as done

	got := Linkify("https://example.com", "click")
	assert.Equal(t, "click", got)
}

func TestSupportsHyperlinks_NoColor(t *testing.T) {
	resetDetection()
	t.Setenv("NO_COLOR", "1")
	t.Setenv("TERM_PROGRAM", "iTerm.app")
	t.Setenv("TERM", "")
	t.Setenv("WT_SESSION", "")
	t.Setenv("LC_TERMINAL", "")

	assert.False(t, SupportsHyperlinks())
}

func TestSupportsHyperlinks_Dumb(t *testing.T) {
	resetDetection()
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "dumb")
	t.Setenv("TERM_PROGRAM", "iTerm.app")
	t.Setenv("WT_SESSION", "")
	t.Setenv("LC_TERMINAL", "")

	assert.False(t, SupportsHyperlinks())
}

func TestSupportsHyperlinks_ITerm(t *testing.T) {
	resetDetection()
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "iTerm.app")
	t.Setenv("WT_SESSION", "")
	t.Setenv("LC_TERMINAL", "")

	assert.True(t, SupportsHyperlinks())
}

func TestSupportsHyperlinks_WezTerm(t *testing.T) {
	resetDetection()
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "WezTerm")
	t.Setenv("WT_SESSION", "")
	t.Setenv("LC_TERMINAL", "")

	assert.True(t, SupportsHyperlinks())
}

func TestSupportsHyperlinks_Ghostty(t *testing.T) {
	resetDetection()
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "ghostty")
	t.Setenv("WT_SESSION", "")
	t.Setenv("LC_TERMINAL", "")

	assert.True(t, SupportsHyperlinks())
}

func TestSupportsHyperlinks_WindowsTerminal(t *testing.T) {
	resetDetection()
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("WT_SESSION", "some-guid")
	t.Setenv("LC_TERMINAL", "")

	assert.True(t, SupportsHyperlinks())
}

func TestSupportsHyperlinks_LCTerminal_ITerm2(t *testing.T) {
	resetDetection()
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "screen")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("WT_SESSION", "")
	t.Setenv("LC_TERMINAL", "iTerm2")

	assert.True(t, SupportsHyperlinks())
}

func TestSupportsHyperlinks_AppleTerminal(t *testing.T) {
	resetDetection()
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "Apple_Terminal")
	t.Setenv("WT_SESSION", "")
	t.Setenv("LC_TERMINAL", "")

	assert.False(t, SupportsHyperlinks())
}

func TestSupportsHyperlinks_Kitty(t *testing.T) {
	resetDetection()
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-kitty")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("WT_SESSION", "")
	t.Setenv("LC_TERMINAL", "")

	assert.True(t, SupportsHyperlinks())
}

func TestSupportsHyperlinks_Foot(t *testing.T) {
	resetDetection()
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "foot")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("WT_SESSION", "")
	t.Setenv("LC_TERMINAL", "")

	assert.True(t, SupportsHyperlinks())
}

func TestSupportsHyperlinks_PlainXterm(t *testing.T) {
	resetDetection()
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("WT_SESSION", "")
	t.Setenv("LC_TERMINAL", "")

	assert.False(t, SupportsHyperlinks())
}
