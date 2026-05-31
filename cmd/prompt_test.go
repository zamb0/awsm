package cmd

import (
	"strings"
	"testing"

	"awsm/internal/aws"
)

func TestIconFor(t *testing.T) {
	cases := map[aws.ProfileType]string{
		aws.ProfileTypeSSO: "☁",
		aws.ProfileTypeIAM: "🔐",
		aws.ProfileTypeKey: "🔑",
		"":                 "",
	}
	for in, want := range cases {
		if got := iconFor(in); got != want {
			t.Errorf("iconFor(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestColorize_DisabledReturnsRaw(t *testing.T) {
	out := colorize("hello", "#FF0000", true)
	if out != "hello" {
		t.Errorf("colorize disabled should return raw, got %q", out)
	}
}

func TestColorize_EmptyReturnsEmpty(t *testing.T) {
	if got := colorize("", "#FF0000", false); got != "" {
		t.Errorf("colorize empty should return empty, got %q", got)
	}
}

func TestColorize_EnabledRendersText(t *testing.T) {
	out := colorize("hello", "#FF0000", false)
	// Whatever wrapping lipgloss applies, the literal text must survive.
	if !strings.Contains(out, "hello") {
		t.Errorf("colorize output missing original text: %q", out)
	}
}
