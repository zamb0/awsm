package cmd

import (
	"fmt"
	"strings"

	"awsm/internal/aws"
	"awsm/internal/mfa"
	"awsm/internal/tui"

	"github.com/spf13/cobra"
)

var mfaCmd = &cobra.Command{
	Use:   "mfa",
	Short: "Manage TOTP secrets for automatic MFA token generation",
	Long: `Store and manage TOTP secrets so awsm can generate MFA tokens automatically,
without prompting you each time your credentials expire.

Secrets are stored in the OS keychain (macOS Keychain, Windows Credential Manager,
Linux Secret Service). On headless Linux without a Secret Service, they fall back
to ~/.awsm/mfa/<profile>.totp with 0600 permissions.`,
}

var mfaSetCmd = &cobra.Command{
	Use:   "set [profile]",
	Short: "Store a TOTP secret for a profile",
	Long: `Store the TOTP secret (seed) for an AWS profile.
You can find this secret when setting up a virtual MFA device in the AWS console
— it's the string shown alongside the QR code.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName, err := resolveProfileName(stringArg(args, 0))
		if err != nil {
			return err
		}

		secret, err := tui.PromptInput(
			fmt.Sprintf("TOTP secret for profile %s", tui.FormatBold(profileName)),
			tui.WithEchoNone(),
			tui.WithRequired(),
		)
		if err != nil {
			return err
		}

		if err := mfa.SetTOTPSecret(profileName, secret); err != nil {
			return fmt.Errorf("failed to store TOTP secret: %w", err)
		}

		// Optionally update mfa_serial in ~/.aws/config
		serial, err := tui.PromptInput(
			"MFA Serial ARN (leave empty to keep current)",
			tui.WithPlaceholder("arn:aws:iam::123456789012:mfa/username"),
		)
		if err != nil {
			return err
		}
		serial = strings.TrimSpace(serial)
		if serial != "" {
			if err := aws.UpdateMFASerial(profileName, serial); err != nil {
				tui.PrintWarning(fmt.Sprintf("TOTP secret stored but failed to update mfa_serial: %v", err))
			}
		}

		tui.PrintSuccess(fmt.Sprintf("TOTP secret stored for profile '%s'", profileName))
		return nil
	},
}

var mfaRemoveCmd = &cobra.Command{
	Use:     "remove [profile]",
	Aliases: []string{"delete", "rm"},
	Short:   "Remove the stored TOTP secret for a profile",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName, err := resolveProfileName(stringArg(args, 0))
		if err != nil {
			return err
		}

		if err := mfa.DeleteTOTPSecret(profileName); err != nil {
			return fmt.Errorf("failed to remove TOTP secret: %w", err)
		}

		tui.PrintSuccess(fmt.Sprintf("TOTP secret removed for profile '%s'", profileName))
		return nil
	},
}

var mfaTestCmd = &cobra.Command{
	Use:   "test [profile]",
	Short: "Generate and display the current TOTP code for a profile",
	Long:  `Generate the current 6-digit TOTP code. Use this to verify that your stored secret is correct by comparing it with your authenticator app.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName, err := resolveProfileName(stringArg(args, 0))
		if err != nil {
			return err
		}

		code, err := mfa.GenerateTOTP(profileName)
		if err != nil {
			return fmt.Errorf("failed to generate TOTP: %w\n\nRun 'awsm mfa set %s' to store a TOTP secret first", err, profileName)
		}

		fmt.Printf("Current TOTP for '%s': %s\n", profileName, code)
		return nil
	},
}

// stringArg returns args[i] or "" if out of bounds.
func stringArg(args []string, i int) string {
	if i < len(args) {
		return args[i]
	}
	return ""
}

func init() {
	mfaCmd.AddCommand(mfaSetCmd)
	mfaCmd.AddCommand(mfaRemoveCmd)
	mfaCmd.AddCommand(mfaTestCmd)

	// Completion for profile argument
	mfaSetCmd.RegisterFlagCompletionFunc("profile", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	rootCmd.AddCommand(mfaCmd)
}
