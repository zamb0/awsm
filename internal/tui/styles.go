package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// ─── Color Palette ─────────────────────────────────────────────────────────
	Primary   = lipgloss.Color("#00D9FF") // Cyan - main brand
	Secondary = lipgloss.Color("#7C3AED") // Purple - accents
	Success   = lipgloss.Color("#10B981") // Green - success states
	Warning   = lipgloss.Color("#F59E0B") // Amber - warnings
	Error     = lipgloss.Color("#EF4444") // Red - errors
	Muted     = lipgloss.Color("#6B7280") // Gray - secondary text
	Subtle    = lipgloss.Color("#374151") // Dark gray - borders, dividers
	White     = lipgloss.Color("#F9FAFB") // White - high contrast text
	Rose      = lipgloss.Color("#F43F5E") // Rose - destructive actions

	// ─── Base Styles ───────────────────────────────────────────────────────────
	BaseStyle = lipgloss.NewStyle().
			Padding(0, 1)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			Padding(1, 0)

	SubheaderStyle = lipgloss.NewStyle().
			Foreground(White).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(Primary)

	MutedStyle = lipgloss.NewStyle().
			Foreground(Muted)

	// ─── Profile Type Styles ───────────────────────────────────────────────────
	ProfileActiveStyle = lipgloss.NewStyle().
				Foreground(Success).
				Bold(true)

	ProfileSSO = lipgloss.NewStyle().
			Foreground(Primary)

	ProfileIAM = lipgloss.NewStyle().
			Foreground(Secondary)

	ProfileKey = lipgloss.NewStyle().
			Foreground(Warning)

	// ─── Box / Container Styles ────────────────────────────────────────────────
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(1, 2).
			Margin(1, 0)

	// ─── Input Styles ──────────────────────────────────────────────────────────
	InputPromptStyle = lipgloss.NewStyle().
				Foreground(Primary).
				Bold(true)

	InputTextStyle = lipgloss.NewStyle().
			Foreground(White)

	InputPlaceholderStyle = lipgloss.NewStyle().
				Foreground(Muted)

	InputCursorStyle = lipgloss.NewStyle().
				Foreground(Primary)

	// ─── Spinner Style ─────────────────────────────────────────────────────────
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(Primary)

	// ─── Label/Value Pairs ─────────────────────────────────────────────────────
	LabelStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Width(14)

	ValueStyle = lipgloss.NewStyle().
			Foreground(White)

	// ─── Divider ───────────────────────────────────────────────────────────────
	DividerStyle = lipgloss.NewStyle().
			Foreground(Subtle)
)

// Divider returns a styled horizontal rule
func Divider(width int) string {
	line := ""
	for i := 0; i < width; i++ {
		line += "─"
	}
	return DividerStyle.Render(line)
}
