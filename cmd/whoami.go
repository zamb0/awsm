package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"awsm/internal/aws"
	"awsm/internal/tui"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	whoamiProfile string
	whoamiJSON    bool
	whoamiNoCall  bool
)

type whoamiOutput struct {
	Profile        string `json:"profile"`
	Type           string `json:"type,omitempty"`
	Region         string `json:"region,omitempty"`
	Account        string `json:"account,omitempty"`
	Arn            string `json:"arn,omitempty"`
	UserID         string `json:"user_id,omitempty"`
	ExpiresAt      string `json:"expires_at,omitempty"`
	ExpiresInSecs  int64  `json:"expires_in_seconds,omitempty"`
	Static         bool   `json:"static,omitempty"`
	CachedOnly     bool   `json:"cached_only,omitempty"`
	CallerIDFailed string `json:"caller_id_error,omitempty"`
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show the active AWS identity (account, ARN, region, credentials TTL)",
	Long: `Display information about the currently active AWS profile and the identity
behind it. Calls sts:GetCallerIdentity to verify the credentials are valid and
shows when temporary credentials will expire.

Use --profile to inspect a profile other than the active one. Use --json for
machine-readable output (useful for scripting).`,
	RunE: runWhoami,
}

func init() {
	whoamiCmd.Flags().StringVarP(&whoamiProfile, "profile", "p", "", "Profile to inspect (defaults to active profile)")
	whoamiCmd.Flags().BoolVar(&whoamiJSON, "json", false, "Output as JSON")
	whoamiCmd.Flags().BoolVar(&whoamiNoCall, "no-call", false, "Don't call sts:GetCallerIdentity (offline mode)")
	rootCmd.AddCommand(whoamiCmd)
}

func runWhoami(cmd *cobra.Command, args []string) error {
	profileName, err := resolveProfileName(whoamiProfile)
	if err != nil {
		if errors.Is(err, aws.ErrNoActiveProfile) {
			if whoamiJSON {
				return jsonEncode(whoamiOutput{})
			}
			tui.PrintWarning("No active profile. Use `awsm profile set <name>` or pass --profile.")
			return nil
		}
		return err
	}

	out := whoamiOutput{Profile: profileName}

	// Best-effort region + type lookup from config.
	if region, rerr := aws.GetProfileRegion(profileName); rerr == nil {
		out.Region = region
	}
	out.Type = lookupProfileType(profileName)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Resolve credentials (handles SSO auto-login + MFA prompt).
	creds, isStatic, err := ensureCredentialsWithLogin(ctx, profileName)
	if err != nil {
		if whoamiJSON {
			out.CallerIDFailed = err.Error()
			return jsonEncode(out)
		}
		return err
	}
	out.Static = isStatic
	if creds != nil && !creds.Expires.IsZero() {
		out.ExpiresAt = creds.Expires.UTC().Format(time.RFC3339)
		out.ExpiresInSecs = int64(time.Until(creds.Expires).Seconds())
	}

	if !whoamiNoCall {
		ident, idErr := aws.GetCallerIdentity(ctx, profileName, creds, out.Region)
		if idErr != nil {
			out.CallerIDFailed = idErr.Error()
		} else {
			out.Account = ident.Account
			out.Arn = ident.Arn
			out.UserID = ident.UserID
		}
	} else {
		out.CachedOnly = true
	}

	if whoamiJSON {
		return jsonEncode(out)
	}
	printWhoami(out)
	return nil
}

func lookupProfileType(name string) string {
	profiles, err := aws.ListProfilesDetailed()
	if err != nil {
		return ""
	}
	for _, p := range profiles {
		if p.Name == name {
			return string(p.Type)
		}
	}
	return ""
}

func jsonEncode(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func printWhoami(o whoamiOutput) {
	tui.PrintHeader("🪪  AWS Identity")

	tui.PrintKeyValue("Profile", o.Profile)
	if o.Type != "" {
		tui.PrintKeyValue("Type", tui.FormatProfileType(o.Type))
	}
	if o.Region != "" {
		tui.PrintKeyValue("Region", o.Region)
	}
	if o.Account != "" {
		tui.PrintKeyValue("Account", o.Account)
	}
	if o.Arn != "" {
		tui.PrintKeyValue("ARN", o.Arn)
	}
	if o.UserID != "" {
		tui.PrintKeyValue("User ID", o.UserID)
	}

	if o.Static {
		tui.PrintKeyValue("Credentials", lipgloss.NewStyle().Foreground(tui.Warning).Render("static (no expiration)"))
	} else if o.ExpiresAt != "" {
		ttl := time.Duration(o.ExpiresInSecs) * time.Second
		tui.PrintKeyValue("Expires", formatTTL(o.ExpiresAt, ttl))
	}

	if o.CallerIDFailed != "" {
		fmt.Fprintln(os.Stderr)
		tui.PrintWarning("sts:GetCallerIdentity failed: " + o.CallerIDFailed)
	}
	fmt.Fprintln(os.Stderr)
}

func formatTTL(expiresAt string, ttl time.Duration) string {
	var color lipgloss.Color
	switch {
	case ttl <= 5*time.Minute:
		color = tui.Error
	case ttl <= 30*time.Minute:
		color = tui.Warning
	default:
		color = tui.Success
	}
	style := lipgloss.NewStyle().Foreground(color).Bold(true)
	humanTTL := humanizeDuration(ttl)
	return fmt.Sprintf("%s %s", style.Render(humanTTL), tui.MutedStyle.Render("("+expiresAt+")"))
}

func humanizeDuration(d time.Duration) string {
	if d <= 0 {
		return "expired"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	if mins == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh%dm", hours, mins)
}
