package cmd

import (
	"awsm/internal/aws"
	"awsm/internal/tui"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	deleteAllSSO bool
	forceDelete  bool
)

var profileDeleteCmd = &cobra.Command{
	Use:               "delete <profile-name>",
	Short:             "Delete an AWS profile",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeProfiles,
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]

		if deleteAllSSO {
			return deleteAllSSOProfiles(profileName)
		}

		// Check if profile exists
		exists, err := aws.ProfileExists(profileName)
		if err != nil {
			return err
		}
		if !exists {
			tui.PrintWarning(fmt.Sprintf("Profile '%s' does not exist", profileName))
			return nil
		}

		// Confirm deletion unless forced
		if !forceDelete {
			confirmed, err := tui.ConfirmDanger(fmt.Sprintf("Delete profile '%s'?", profileName))
			if err != nil {
				return err
			}
			if !confirmed {
				tui.PrintMuted("Deletion cancelled")
				return nil
			}
		}

		if err := aws.DeleteProfile(profileName); err != nil {
			return fmt.Errorf("failed to delete profile: %w", err)
		}

		tui.PrintSuccess(fmt.Sprintf("Profile '%s' deleted successfully", profileName))
		return nil
	},
}

func deleteAllSSOProfiles(ssoSession string) error {
	profiles, err := aws.GetProfilesBySSO(ssoSession)
	if err != nil {
		return err
	}

	if len(profiles) == 0 {
		tui.PrintWarning(fmt.Sprintf("No profiles found for SSO session '%s'", ssoSession))
		return nil
	}

	tui.PrintInfo(fmt.Sprintf("Found %d profiles for SSO session '%s':", len(profiles), ssoSession))
	for _, profile := range profiles {
		tui.PrintBullet(profile)
	}

	if !forceDelete {
		confirmed, err := tui.ConfirmDanger(fmt.Sprintf("Delete all %d profiles?", len(profiles)))
		if err != nil {
			return err
		}
		if !confirmed {
			tui.PrintMuted("Deletion cancelled")
			return nil
		}
	}

	for _, profile := range profiles {
		if err := aws.DeleteProfile(profile); err != nil {
			tui.PrintError(fmt.Sprintf("Failed to delete profile '%s': %v", profile, err))
		} else {
			tui.PrintSuccess(fmt.Sprintf("Deleted profile '%s'", profile))
		}
	}

	return nil
}

func init() {
	profileDeleteCmd.Flags().BoolVar(&deleteAllSSO, "all-sso", false, "Delete all profiles for the specified SSO session")
	profileDeleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Force deletion without confirmation")
	profileCmd.AddCommand(profileDeleteCmd)
}
