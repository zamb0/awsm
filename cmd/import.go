package cmd

import (
	"awsm/internal/aws"
	"awsm/internal/tui"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

type ExportData struct {
	ExportedAt      time.Time            `json:"exported_at"`
	Version         string               `json:"version"`
	Profiles        []aws.ProfileInfo    `json:"profiles"`
	SSOSessions     []aws.SSOSessionInfo `json:"sso_sessions"`
	ConfigFile      string               `json:"config_file,omitempty"`
	CredentialsFile string               `json:"credentials_file,omitempty"`
}

var (
	importForce        bool
	importRestoreFiles bool
)

var importCmd = &cobra.Command{
	Use:   "import <export-file>",
	Short: "Import profiles and SSO sessions from an export file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		importFile := args[0]

		tui.PrintInfo(fmt.Sprintf("Importing AWS configuration from: %s", tui.FormatBold(importFile)))

		// Read import file
		file, err := os.Open(importFile)
		if err != nil {
			return fmt.Errorf("failed to open import file: %w", err)
		}
		defer file.Close()

		var exportData ExportData
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&exportData); err != nil {
			return fmt.Errorf("failed to parse import file: %w", err)
		}

		tui.PrintInfo(fmt.Sprintf("Import file contains: %d profiles, %d SSO sessions",
			len(exportData.Profiles), len(exportData.SSOSessions)))

		if !importForce {
			confirmed, err := tui.Confirm("Continue with import? This may overwrite existing configurations")
			if err != nil {
				return err
			}
			if !confirmed {
				tui.PrintMuted("Import cancelled")
				return nil
			}
		}

		// Check for restore mode
		if importRestoreFiles {
			if !importForce {
				confirmed, err := tui.ConfirmDanger("--restore-files will OVERWRITE your local config and credentials files entirely. Continue?")
				if err != nil {
					return err
				}
				if !confirmed {
					tui.PrintMuted("Import cancelled")
					return nil
				}
			}

			if err := aws.RestoreConfigFiles(exportData.ConfigFile, exportData.CredentialsFile); err != nil {
				return fmt.Errorf("failed to restore files: %w", err)
			}

			tui.PrintSuccess("AWS configuration files restored successfully (original files moved to .bak)")
			return nil
		}

		// Import SSO sessions first (Merge Mode)
		importedSessions := 0
		for _, session := range exportData.SSOSessions {
			if err := aws.ImportSSOSession(session); err != nil {
				tui.PrintError(fmt.Sprintf("Failed to import SSO session '%s': %v", session.Name, err))
			} else {
				tui.PrintSuccess(fmt.Sprintf("Imported SSO session '%s'", session.Name))
				importedSessions++
			}
		}

		// Import profiles (Merge Mode)
		importedProfiles := 0
		for _, profile := range exportData.Profiles {
			if err := aws.ImportProfile(profile); err != nil {
				tui.PrintError(fmt.Sprintf("Failed to import profile '%s': %v", profile.Name, err))
			} else {
				tui.PrintSuccess(fmt.Sprintf("Imported profile '%s'", profile.Name))
				importedProfiles++
			}
		}

		tui.PrintSuccess(fmt.Sprintf("Import complete: %d profiles, %d SSO sessions imported (Merge Mode)",
			importedProfiles, importedSessions))
		return nil
	},
}

func init() {
	importCmd.Flags().BoolVarP(&importForce, "force", "f", false, "Force import without confirmation")
	importCmd.Flags().BoolVar(&importRestoreFiles, "restore-files", false, "Restore exact config and credentials files (overwrites everything)")
	rootCmd.AddCommand(importCmd)
}
