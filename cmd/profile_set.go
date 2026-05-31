package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"awsm/internal/aws"
	"awsm/internal/tui"

	"github.com/spf13/cobra"
)

// --- Command Definitions ---
var profileSetCmd = &cobra.Command{
	Use:               "set <profile>",
	Short:             "Set credentials for a profile in the default AWS credentials file",
	Long:              `Updates the default profile in ~/.aws/credentials with the specified profile's credentials.`,
	Args:              cobra.ExactArgs(1),
	RunE:              runProfileSet,
	ValidArgsFunction: completeProfiles,
}

// --- Main Logic ---
func runProfileSet(cmd *cobra.Command, args []string) error {
	profileName := args[0]

	// Get profile region first
	region, err := aws.GetProfileRegion(profileName)
	if err != nil {
		// Region is optional, continue without it
		region = ""
	}
	if region != "" && !aws.IsValidRegion(region) {
		return fmt.Errorf("invalid region for profile '%s': %s", profileName, region)
	}

	creds, isStatic, err := ensureCredentialsWithLogin(context.Background(), profileName)
	if err != nil {
		fmt.Fprintln(os.Stderr, tui.ErrorStyle.Render("✗ Error: "+err.Error()))
		return err
	}

	if isStatic {
		err = aws.UpdateStaticProfile(profileName)
		if err != nil {
			fmt.Fprintln(os.Stderr, tui.ErrorStyle.Render("✗ Error updating credentials file: "+err.Error()))
			return fmt.Errorf("failed to update credentials file")
		}
		fmt.Fprintln(os.Stderr, tui.SuccessStyle.Render("✓ Switched to profile '"+profileName+"' in default credentials."))
		return nil
	}

	if creds == nil {
		return fmt.Errorf("unexpected error: credentials are nil")
	}

	err = aws.UpdateCredentialsFile(creds, region, profileName)
	if err != nil {
		fmt.Fprintln(os.Stderr, tui.ErrorStyle.Render("✗ Error updating credentials file: "+err.Error()))
		return fmt.Errorf("failed to update credentials file")
	}

	fmt.Fprintln(os.Stderr, tui.SuccessStyle.Render("✓ Credentials for profile '"+profileName+"' are set."))
	return nil
}

// --- Autocompletion Logic ---
// completeProfiles provides completion for profile arguments, excluding sso-session profiles
var completeProfiles = aws.CompleteProfilesFiltered(func(profile string) bool {
	return !strings.HasPrefix(profile, "sso-session")
})

// --- Initialization ---
func init() {
	// Command will be added to profile subcommand in profile.go
}
