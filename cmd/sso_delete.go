package cmd

import (
	"awsm/internal/aws"
	"awsm/internal/tui"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	ssoForceDelete bool
)

var ssoDeleteCmd = &cobra.Command{
	Use:               "delete <sso-session>",
	Short:             "Delete an SSO session and optionally its associated profiles",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeSSOSessions,
	RunE: func(cmd *cobra.Command, args []string) error {
		ssoSession := args[0]

		// Check if SSO session exists
		sessions, err := aws.ListSSOSessions()
		if err != nil {
			return err
		}

		var sessionExists bool
		for _, session := range sessions {
			if session.Name == ssoSession {
				sessionExists = true
				break
			}
		}

		if !sessionExists {
			tui.PrintWarning(fmt.Sprintf("SSO session '%s' does not exist", ssoSession))
			return nil
		}

		// Get associated profiles
		profiles, err := aws.GetProfilesBySSO(ssoSession)
		if err != nil {
			return err
		}

		// Show what will be deleted
		tui.PrintInfo(fmt.Sprintf("SSO session '%s' will be deleted", ssoSession))
		if len(profiles) > 0 {
			tui.PrintInfo(fmt.Sprintf("This will also delete %d associated profiles:", len(profiles)))
			for _, profile := range profiles {
				tui.PrintBullet(profile)
			}
		}

		// Confirm deletion unless forced
		if !ssoForceDelete {
			totalItems := 1 + len(profiles)
			confirmed, err := tui.ConfirmDanger(fmt.Sprintf("Delete SSO session and %d total items?", totalItems))
			if err != nil {
				return err
			}
			if !confirmed {
				tui.PrintMuted("Deletion cancelled")
				return nil
			}
		}

		// Delete associated profiles first
		for _, profile := range profiles {
			if err := aws.DeleteProfile(profile); err != nil {
				tui.PrintError(fmt.Sprintf("Failed to delete profile '%s': %v", profile, err))
			} else {
				tui.PrintSuccess(fmt.Sprintf("Deleted profile '%s'", profile))
			}
		}

		// Delete SSO session
		if err := aws.DeleteSSOSession(ssoSession); err != nil {
			return fmt.Errorf("failed to delete SSO session: %w", err)
		}

		tui.PrintSuccess(fmt.Sprintf("SSO session '%s' deleted successfully", ssoSession))
		return nil
	},
}

func init() {
	ssoDeleteCmd.Flags().BoolVarP(&ssoForceDelete, "force", "f", false, "Force deletion without confirmation")
	ssoCmd.AddCommand(ssoDeleteCmd)
}
