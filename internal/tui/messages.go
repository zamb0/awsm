package tui

import (
	"fmt"
	"os"
	"strings"
)

// ─── Status Messages ───────────────────────────────────────────────────────────
// These provide consistent, beautiful output across all commands.

// PrintSuccess prints a success message with a checkmark
func PrintSuccess(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", SuccessStyle.Render("✓"), msg)
}

// PrintError prints an error message with an X
func PrintError(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", ErrorStyle.Render("✗"), msg)
}

// PrintWarning prints a warning message
func PrintWarning(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", WarningStyle.Render("⚠"), msg)
}

// PrintInfo prints an informational message
func PrintInfo(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", InfoStyle.Render("›"), msg)
}

// PrintStep prints a step in a multi-step process
func PrintStep(msg string) {
	fmt.Fprintf(os.Stderr, "  %s %s\n", InfoStyle.Render("→"), msg)
}

// PrintKeyValue prints a labeled value with consistent formatting
func PrintKeyValue(label, value string) {
	fmt.Fprintf(os.Stderr, "  %s %s\n", LabelStyle.Render(label+":"), ValueStyle.Render(value))
}

// PrintHeader prints a section header
func PrintHeader(msg string) {
	fmt.Fprintf(os.Stderr, "\n%s\n", HeaderStyle.Render(msg))
}

// PrintBullet prints a bullet point item
func PrintBullet(msg string) {
	fmt.Fprintf(os.Stderr, "  %s %s\n", InfoStyle.Render("•"), msg)
}

// PrintMuted prints muted/secondary text
func PrintMuted(msg string) {
	fmt.Fprintf(os.Stderr, "  %s\n", MutedStyle.Render(msg))
}

// ─── Formatters (return string, don't print) ──────────────────────────────────

// FormatSuccess returns a formatted success string
func FormatSuccess(msg string) string {
	return fmt.Sprintf("%s %s", SuccessStyle.Render("✓"), msg)
}

// FormatError returns a formatted error string
func FormatError(msg string) string {
	return fmt.Sprintf("%s %s", ErrorStyle.Render("✗"), msg)
}

// FormatWarning returns a formatted warning string
func FormatWarning(msg string) string {
	return fmt.Sprintf("%s %s", WarningStyle.Render("⚠"), msg)
}

// FormatProfileType returns a styled profile type badge
func FormatProfileType(profileType string) string {
	switch strings.ToUpper(profileType) {
	case "SSO":
		return ProfileSSO.Render("SSO")
	case "IAM":
		return ProfileIAM.Render("IAM")
	case "KEY":
		return ProfileKey.Render("Key")
	default:
		return MutedStyle.Render(profileType)
	}
}

// FormatBold returns bold-styled text
func FormatBold(msg string) string {
	return SubheaderStyle.Render(msg)
}
