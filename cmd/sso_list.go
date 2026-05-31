package cmd

import (
	"awsm/internal/aws"
	"awsm/internal/tui"
	"awsm/internal/util"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	ssoFilterRegion string
	ssoNameFilter   string
	ssoSortBy       string
	ssoOutputJSON   bool
)

var ssoListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all SSO sessions",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		sessions, err := aws.ListSSOSessions()
		if err != nil {
			return err
		}

		if len(sessions) == 0 {
			if !ssoOutputJSON {
				tui.PrintWarning("No SSO sessions found.")
			} else {
				fmt.Println("[]")
			}
			return nil
		}

		// Apply filters
		var filtered []aws.SSOSessionInfo
		for _, s := range sessions {
			if ssoFilterRegion != "" && !strings.EqualFold(s.Region, ssoFilterRegion) {
				continue
			}
			if ssoNameFilter != "" && !strings.Contains(strings.ToLower(s.Name), strings.ToLower(ssoNameFilter)) {
				continue
			}
			filtered = append(filtered, s)
		}

		if len(filtered) == 0 {
			if !ssoOutputJSON {
				tui.PrintWarning("No SSO sessions match the specified filters.")
			} else {
				fmt.Println("[]")
			}
			return nil
		}

		// Sort sessions
		switch strings.ToLower(ssoSortBy) {
		case "region":
			util.SortBy(filtered, func(s1, s2 aws.SSOSessionInfo) bool {
				return s1.Region < s2.Region
			})
		default:
			util.SortBy(filtered, func(s1, s2 aws.SSOSessionInfo) bool {
				return s1.Name < s2.Name
			})
		}

		if ssoOutputJSON {
			return outputSSOSessionsJSON(filtered)
		}

		printDetailedSSOSessions(filtered)
		return nil
	},
}

func outputSSOSessionsJSON(sessions []aws.SSOSessionInfo) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(sessions)
}

func printDetailedSSOSessions(sessions []aws.SSOSessionInfo) {
	tui.PrintHeader("🔐 SSO Sessions")
	fmt.Fprintln(os.Stderr)

	for i, s := range sessions {
		fmt.Fprintf(os.Stderr, "  %s %s\n", tui.ProfileSSO.Render("●"), tui.SubheaderStyle.Render(s.Name))
		tui.PrintKeyValue("Start URL", s.StartURL)
		tui.PrintKeyValue("Region", s.Region)
		tui.PrintKeyValue("Scopes", s.Scopes)
		if i < len(sessions)-1 {
			fmt.Fprintln(os.Stderr)
		}
	}
	fmt.Fprintln(os.Stderr)
}

func init() {
	ssoListCmd.Flags().StringVarP(&ssoFilterRegion, "region", "r", "", "Filter by region")
	ssoListCmd.Flags().StringVarP(&ssoNameFilter, "name", "n", "", "Filter by session name (case-insensitive)")
	ssoListCmd.Flags().StringVarP(&ssoSortBy, "sort", "s", "name", "Sort by field (name, region)")
	ssoListCmd.Flags().BoolVarP(&ssoOutputJSON, "json", "j", false, "Output sessions in JSON format")
	ssoCmd.AddCommand(ssoListCmd)
}
