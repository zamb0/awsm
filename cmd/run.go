package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"awsm/internal/aws"

	"github.com/spf13/cobra"
)

var (
	runProfile string
	runRegion  string
)

var runCmd = &cobra.Command{
	Use:   "run [--profile <name>] [--region <region>] -- <command> [args...]",
	Short: "Run a command with AWS credentials injected as environment variables",
	Long: `Resolve AWS credentials for the given profile (or the active one) and execute
the specified command with AWS_* environment variables set. The parent shell
environment is NOT modified.

This is useful for running tools like terraform, aws-cli, kubectl (with aws-iam-
authenticator), etc. against a specific profile without exporting credentials
globally.

Use a literal -- separator to delimit awsm flags from the command to execute.

Examples:
  awsm run -- aws s3 ls
  awsm run --profile prod -- terraform plan
  awsm run --profile dev --region eu-west-1 -- env | grep AWS_`,
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: false,
	RunE:               runRun,
}

func init() {
	runCmd.Flags().StringVarP(&runProfile, "profile", "p", "", "Profile to use (defaults to active profile)")
	runCmd.Flags().StringVarP(&runRegion, "region", "r", "", "Override the region for the child process")
	rootCmd.AddCommand(runCmd)
}

func runRun(cmd *cobra.Command, args []string) error {
	profileName, err := resolveProfileName(runProfile)
	if err != nil {
		if errors.Is(err, aws.ErrNoActiveProfile) {
			return fmt.Errorf("no active profile; pass --profile or run `awsm profile set <name>` first")
		}
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	creds, isStatic, err := ensureCredentialsWithLogin(ctx, profileName)
	if err != nil {
		return err
	}

	// Determine region: explicit flag wins, then profile region, then leave empty.
	region := runRegion
	if region == "" {
		if r, rerr := aws.GetProfileRegion(profileName); rerr == nil {
			region = r
		}
	}

	// For static profiles GetCredentialsForProfile may return non-nil creds too,
	// but if it returned only isStatic=true with nil creds we still want AWS_PROFILE
	// to be honored by the child process.
	envCreds := creds
	_ = isStatic

	env := aws.BuildEnvForProfile(envCreds, region, profileName)

	bin, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("command not found: %s", args[0])
	}

	child := exec.Command(bin, args[1:]...)
	child.Env = env
	child.Stdin = os.Stdin
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr

	// Forward common termination signals to the child so it can clean up.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer signal.Stop(sigCh)

	if err := child.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", args[0], err)
	}

	go func() {
		for sig := range sigCh {
			if child.Process != nil {
				_ = child.Process.Signal(sig)
			}
		}
	}()

	if err := child.Wait(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// Mimic the child's exit code.
			os.Exit(exitErr.ExitCode())
		}
		return err
	}

	return nil
}
