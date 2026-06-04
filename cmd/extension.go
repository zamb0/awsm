/*
AWSM - AWS Manager
Copyright (c) 2024 Alessandro Gallo. All rights reserved.

Licensed under the Business Source License 1.1.
See LICENSE file for full terms.
*/

package cmd

import (
	"fmt"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

const (
	amoListingURL    = "https://addons.mozilla.org/firefox/addon/awsm-container-opener/"
	extensionRepoURL = "https://github.com/AleG03/awsm-firefox-container"
)

var extensionCmd = &cobra.Command{
	Use:   "extension",
	Short: "Manage the AWSM Firefox container extension",
	Long: `Manage the AWSM Container Opener Firefox extension.

The extension handles ext+container: URLs (the protocol used by 'awsm console
--firefox-container') and is published on addons.mozilla.org for permanent,
auto-updating installation.`,
}

var extensionInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Open the AMO listing to install the extension",
	RunE: func(cmd *cobra.Command, args []string) error {
		w := cmd.OutOrStdout()
		fmt.Fprintln(w, "Opening the AWSM Container Opener listing on addons.mozilla.org...")
		fmt.Fprintln(w, "")
		fmt.Fprintf(w, "  %s\n\n", amoListingURL)
		fmt.Fprintln(w, "Click \"Add to Firefox\" to install. The extension will update automatically.")

		if err := browser.OpenURL(amoListingURL); err != nil {
			fmt.Fprintf(w, "\nCould not open browser: %v\n", err)
			fmt.Fprintf(w, "Please visit the URL above manually.\n")
		}
		return nil
	},
}

var extensionStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show extension info and links",
	Run: func(cmd *cobra.Command, args []string) {
		w := cmd.OutOrStdout()
		fmt.Fprintf(w, "AMO listing:    %s\n", amoListingURL)
		fmt.Fprintf(w, "Source code:    %s\n", extensionRepoURL)
		fmt.Fprintf(w, "Extension ID:   awsm-container@awsm.io\n")
	},
}

func init() {
	extensionCmd.AddCommand(extensionInstallCmd, extensionStatusCmd)
	rootCmd.AddCommand(extensionCmd)
}
