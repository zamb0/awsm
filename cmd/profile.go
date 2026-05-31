package cmd

import (
	"awsm/internal/aws"
	"awsm/internal/tui"
	"awsm/internal/util"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	listDetailed bool
	filterType   string
	filterRegion string
	nameFilter   string
	sortBy       string
	showHelp     bool
	outputJSON   bool
)

// JSONProfileInfo represents the profile information in a scripting-friendly format
type JSONProfileInfo struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Region        string `json:"region"`
	AccountID     string `json:"account_id,omitempty"`
	RoleARN       string `json:"role_arn,omitempty"`
	SourceProfile string `json:"source_profile,omitempty"`
	SSOStartURL   string `json:"sso_start_url,omitempty"`
	SSORegion     string `json:"sso_region,omitempty"`
	SSOAccountID  string `json:"sso_account_id,omitempty"`
	SSORoleName   string `json:"sso_role_name,omitempty"`
	SSOSession    string `json:"sso_session,omitempty"`
	MFASerial     string `json:"mfa_serial,omitempty"`
	IsActive      bool   `json:"is_active"`
}

// Profile type descriptions
var profileTypeDescriptions = map[aws.ProfileType]string{
	aws.ProfileTypeSSO: "AWS IAM Identity Center (SSO) profile - Uses SSO for authentication",
	aws.ProfileTypeIAM: "IAM profile with role assumption - Requires MFA and assumes a role",
	aws.ProfileTypeKey: "Static credentials - Uses AWS access key and secret",
}

var profileCmd = &cobra.Command{
	Use:     "profile",
	Short:   "Manage AWS profiles",
	Aliases: []string{"p"},
}

var profileListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all available AWS profiles",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if showHelp {
			printProfileTypeHelp()
			return nil
		}

		profiles, err := aws.ListProfilesDetailed()
		if err != nil {
			return err
		}

		if len(profiles) == 0 {
			if !outputJSON {
				tui.PrintWarning("No profiles found.")
			} else {
				fmt.Println("[]")
			}
			return nil
		}

		// Apply filters
		var filtered []aws.ProfileInfo
		for _, p := range profiles {
			if filterType != "" && !strings.EqualFold(string(p.Type), filterType) {
				continue
			}
			if filterRegion != "" && !strings.EqualFold(p.Region, filterRegion) {
				continue
			}
			if nameFilter != "" && !strings.Contains(strings.ToLower(p.Name), strings.ToLower(nameFilter)) {
				continue
			}
			filtered = append(filtered, p)
		}

		if len(filtered) == 0 {
			if !outputJSON {
				tui.PrintWarning("No profiles match the specified filters.")
			} else {
				fmt.Println("[]")
			}
			return nil
		}

		// Sort profiles based on sortBy flag
		switch strings.ToLower(sortBy) {
		case "type":
			util.SortBy(filtered, func(p1, p2 aws.ProfileInfo) bool {
				return string(p1.Type) < string(p2.Type)
			})
		case "region":
			util.SortBy(filtered, func(p1, p2 aws.ProfileInfo) bool {
				return p1.Region < p2.Region
			})
		default:
			util.SortBy(filtered, func(p1, p2 aws.ProfileInfo) bool {
				return p1.Name < p2.Name
			})
		}

		if outputJSON {
			return outputProfilesJSON(filtered)
		}

		if listDetailed {
			printDetailedProfiles(filtered)
		} else {
			printSimpleProfiles(filtered)
		}

		return nil
	},
}

func outputProfilesJSON(profiles []aws.ProfileInfo) error {
	var jsonProfiles []JSONProfileInfo

	for _, p := range profiles {
		accountID := p.SSOAccountID
		if accountID == "" && p.RoleARN != "" {
			parts := strings.Split(p.RoleARN, ":")
			if len(parts) >= 5 {
				accountID = parts[4]
			}
		}

		jsonProfile := JSONProfileInfo{
			Name:          p.Name,
			Type:          string(p.Type),
			Region:        p.Region,
			AccountID:     accountID,
			RoleARN:       p.RoleARN,
			SourceProfile: p.SourceProfile,
			SSOStartURL:   p.SSOStartURL,
			SSORegion:     p.SSORegion,
			SSOAccountID:  p.SSOAccountID,
			SSORoleName:   p.SSORoleName,
			SSOSession:    p.SSOSession,
			MFASerial:     p.MFASerial,
			IsActive:      p.IsActive,
		}
		jsonProfiles = append(jsonProfiles, jsonProfile)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jsonProfiles)
}

func printProfileTypeHelp() {
	tui.PrintHeader("AWS Profile Types")
	fmt.Fprintln(os.Stderr)
	for pType, desc := range profileTypeDescriptions {
		fmt.Fprintf(os.Stderr, "  %s  %s\n", tui.FormatProfileType(string(pType)), tui.MutedStyle.Render(desc))
	}
	fmt.Fprintln(os.Stderr)
}

func printSimpleProfiles(profiles []aws.ProfileInfo) {
	regionStyle := lipgloss.NewStyle().Foreground(tui.Warning)
	accountStyle := lipgloss.NewStyle().Foreground(tui.Primary)
	activeStyle := lipgloss.NewStyle().Foreground(tui.Success).Bold(true)
	ssoStyle := lipgloss.NewStyle().Foreground(tui.Success).Bold(true)
	iamStyle := lipgloss.NewStyle().Foreground(tui.Secondary).Bold(true)
	keyStyle := lipgloss.NewStyle().Foreground(tui.Warning).Bold(true)
	sessionStyle := lipgloss.NewStyle().Foreground(tui.Secondary).Bold(true)

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, tui.HeaderStyle.Render("🚀 AWS Profiles"))
	fmt.Fprintln(os.Stderr)

	var ssoProfiles, iamProfiles, keyProfiles []aws.ProfileInfo
	for _, p := range profiles {
		switch p.Type {
		case aws.ProfileTypeSSO:
			ssoProfiles = append(ssoProfiles, p)
		case aws.ProfileTypeIAM:
			iamProfiles = append(iamProfiles, p)
		case aws.ProfileTypeKey:
			keyProfiles = append(keyProfiles, p)
		}
	}

	// Group SSO profiles by session
	if len(ssoProfiles) > 0 {
		ssoSessions := make(map[string][]aws.ProfileInfo)
		for _, p := range ssoProfiles {
			session := p.SSOSession
			if session == "" {
				session = "Unknown Session"
			}
			ssoSessions[session] = append(ssoSessions[session], p)
		}

		fmt.Fprintln(os.Stderr, ssoStyle.Render("● SSO Profiles"))

		for session, sessionProfiles := range ssoSessions {
			fmt.Fprintf(os.Stderr, "  %s\n", sessionStyle.Render("📁 "+session))
			for _, p := range sessionProfiles {
				fmt.Fprint(os.Stderr, "    ")
				if p.IsActive {
					fmt.Fprint(os.Stderr, activeStyle.Render("▶ "))
				} else {
					fmt.Fprint(os.Stderr, "  ")
				}
				fmt.Fprintf(os.Stderr, "%s ", p.Name)
				if p.SSOAccountID != "" {
					fmt.Fprintf(os.Stderr, "%s ", accountStyle.Render("("+p.SSOAccountID+")"))
				}
				fmt.Fprintf(os.Stderr, "%s\n", regionStyle.Render("["+p.Region+"]"))
			}
			fmt.Fprintln(os.Stderr)
		}
	}

	// IAM Profiles
	if len(iamProfiles) > 0 {
		fmt.Fprintln(os.Stderr, iamStyle.Render("● IAM Role Profiles"))
		for _, p := range iamProfiles {
			fmt.Fprint(os.Stderr, "  ")
			if p.IsActive {
				fmt.Fprint(os.Stderr, activeStyle.Render("▶ "))
			} else {
				fmt.Fprint(os.Stderr, "  ")
			}
			fmt.Fprintf(os.Stderr, "%s ", p.Name)
			if p.RoleARN != "" {
				parts := strings.Split(p.RoleARN, ":")
				if len(parts) >= 5 {
					fmt.Fprintf(os.Stderr, "%s ", accountStyle.Render("("+parts[4]+")"))
				}
			}
			fmt.Fprintf(os.Stderr, "%s\n", regionStyle.Render("["+p.Region+"]"))
		}
		fmt.Fprintln(os.Stderr)
	}

	// Key Profiles
	if len(keyProfiles) > 0 {
		fmt.Fprintln(os.Stderr, keyStyle.Render("● Static Key Profiles"))
		for _, p := range keyProfiles {
			fmt.Fprint(os.Stderr, "  ")
			if p.IsActive {
				fmt.Fprint(os.Stderr, activeStyle.Render("▶ "))
			} else {
				fmt.Fprint(os.Stderr, "  ")
			}
			fmt.Fprintf(os.Stderr, "%s ", p.Name)
			fmt.Fprintf(os.Stderr, "%s\n", regionStyle.Render("["+p.Region+"]"))
		}
		fmt.Fprintln(os.Stderr)
	}

	// Legend
	fmt.Fprintln(os.Stderr, tui.MutedStyle.Render("  ▶ active • (123456789012) account • [region]"))
}

func printDetailedProfiles(profiles []aws.ProfileInfo) {
	fmt.Fprintln(os.Stderr)
	tui.PrintHeader("🚀 AWS Profiles (Detailed)")
	fmt.Fprintln(os.Stderr)

	for i, p := range profiles {
		if p.IsActive {
			fmt.Fprintf(os.Stderr, "  %s ", tui.SuccessStyle.Render("▶"))
		} else {
			fmt.Fprint(os.Stderr, "    ")
		}
		fmt.Fprintf(os.Stderr, "%s  %s\n",
			tui.SubheaderStyle.Render(p.Name),
			tui.FormatProfileType(string(p.Type)))

		switch p.Type {
		case aws.ProfileTypeSSO:
			tui.PrintKeyValue("Account", p.SSOAccountID)
			tui.PrintKeyValue("Region", p.Region)
			if p.SSOSession != "" {
				tui.PrintKeyValue("Session", p.SSOSession)
			}
			if p.SSORoleName != "" {
				tui.PrintKeyValue("Role", p.SSORoleName)
			}

		case aws.ProfileTypeIAM:
			if p.RoleARN != "" {
				tui.PrintKeyValue("Role ARN", p.RoleARN)
			}
			tui.PrintKeyValue("Region", p.Region)
			if p.MFASerial != "" {
				tui.PrintKeyValue("MFA", p.MFASerial)
			}
			if p.SourceProfile != "" {
				tui.PrintKeyValue("Source", p.SourceProfile)
			}

		case aws.ProfileTypeKey:
			tui.PrintKeyValue("Region", p.Region)
		}

		if i < len(profiles)-1 {
			fmt.Fprintln(os.Stderr)
		}
	}
	fmt.Fprintln(os.Stderr)
}

func init() {
	profileListCmd.Flags().BoolVarP(&listDetailed, "detailed", "d", false, "Show detailed profile information")
	profileListCmd.Flags().StringVarP(&filterType, "type", "t", "", "Filter by profile type (SSO, IAM, Key)")
	profileListCmd.Flags().StringVarP(&filterRegion, "region", "r", "", "Filter by region")
	profileListCmd.Flags().StringVarP(&nameFilter, "name", "n", "", "Filter by profile name (case-insensitive)")
	profileListCmd.Flags().StringVarP(&sortBy, "sort", "s", "name", "Sort by field (name, type, region)")
	profileListCmd.Flags().BoolVarP(&showHelp, "help-types", "H", false, "Show help about profile types")
	profileListCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "Output profiles in JSON format")

	profileCmd.AddCommand(profileSetCmd)
	profileCmd.AddCommand(profileListCmd)
	rootCmd.AddCommand(profileCmd)
}
