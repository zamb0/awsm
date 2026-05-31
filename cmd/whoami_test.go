package cmd

import (
	"strings"
	"testing"
	"time"
)

func TestHumanizeDuration(t *testing.T) {
	cases := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"expired-zero", 0, "expired"},
		{"expired-negative", -5 * time.Second, "expired"},
		{"seconds", 42 * time.Second, "42s"},
		{"minutes", 5*time.Minute + 30*time.Second, "5m30s"},
		{"hours-with-minutes", 2*time.Hour + 15*time.Minute, "2h15m"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := humanizeDuration(c.d)
			if got != c.want {
				t.Errorf("humanizeDuration(%v) = %q, want %q", c.d, got, c.want)
			}
		})
	}
}

func TestFormatTTL_ContainsTimestampAndDuration(t *testing.T) {
	// formatTTL applies lipgloss styling; in tests styles render plain text by
	// default when there's no TTY, but ANSI escape codes may still wrap the
	// output. We assert on substrings only.
	expiresAt := "2030-01-01T00:00:00Z"
	out := formatTTL(expiresAt, 2*time.Hour+10*time.Minute)
	if !strings.Contains(out, expiresAt) {
		t.Errorf("formatTTL output missing timestamp: %q", out)
	}
	if !strings.Contains(out, "2h10m") {
		t.Errorf("formatTTL output missing human duration: %q", out)
	}
}

func TestFormatTTL_ExpiredShowsExpiredLabel(t *testing.T) {
	out := formatTTL("2020-01-01T00:00:00Z", 0)
	if !strings.Contains(out, "expired") {
		t.Errorf("formatTTL with 0 TTL should contain 'expired': %q", out)
	}
}
