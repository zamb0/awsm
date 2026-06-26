package mfa

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/zalando/go-keyring"
)

const keyringSvc = "awsm"

// SetTOTPSecret stores the TOTP secret for a profile.
// Tries OS keyring first; falls back to a plain-text file (~/.awsm/mfa/<profile>.totp)
// with 0600 permissions when keyring is unavailable (e.g. headless Linux).
func SetTOTPSecret(profileName, secret string) error {
	secret = strings.ToUpper(strings.ReplaceAll(secret, " ", ""))
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return fmt.Errorf("TOTP secret cannot be empty")
	}

	// Validate it's a real TOTP secret before storing
	if _, err := totp.GenerateCode(secret, time.Now()); err != nil {
		return fmt.Errorf("invalid TOTP secret: %w", err)
	}

	if err := keyring.Set(keyringSvc, profileName, secret); err != nil {
		// Keyring unavailable — fall back to file
		if fileErr := writeSecretFile(profileName, secret); fileErr != nil {
			return fmt.Errorf("keyring unavailable (%v) and file fallback failed: %w", err, fileErr)
		}
		fmt.Fprintf(os.Stderr, "Warning: OS keyring unavailable. TOTP secret stored in ~/.awsm/mfa/%s.totp (mode 0600).\n", profileName)
		return nil
	}
	return nil
}

// GetTOTPSecret retrieves the TOTP secret for a profile.
func GetTOTPSecret(profileName string) (string, error) {
	secret, err := keyring.Get(keyringSvc, profileName)
	if err == nil {
		return secret, nil
	}

	// Try file fallback
	secret, fileErr := readSecretFile(profileName)
	if fileErr != nil {
		return "", fmt.Errorf("no TOTP secret found for profile %q", profileName)
	}
	return secret, nil
}

// DeleteTOTPSecret removes the stored TOTP secret for a profile.
func DeleteTOTPSecret(profileName string) error {
	keyringErr := keyring.Delete(keyringSvc, profileName)
	fileErr := deleteSecretFile(profileName)

	if keyringErr != nil && fileErr != nil {
		return fmt.Errorf("no TOTP secret found for profile %q", profileName)
	}
	return nil
}

// GenerateTOTP generates the current 6-digit TOTP code for a profile.
// Returns an error if no secret is configured for the profile.
func GenerateTOTP(profileName string) (string, error) {
	secret, err := GetTOTPSecret(profileName)
	if err != nil {
		return "", err
	}
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		return "", fmt.Errorf("failed to generate TOTP for profile %q: %w", profileName, err)
	}
	return code, nil
}

// secretFilePath returns the path for the file-based fallback secret.
func secretFilePath(profileName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".awsm", "mfa", profileName+".totp"), nil
}

func writeSecretFile(profileName, secret string) error {
	path, err := secretFilePath(profileName)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(secret), 0600)
}

func readSecretFile(profileName string) (string, error) {
	path, err := secretFilePath(profileName)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func deleteSecretFile(profileName string) error {
	path, err := secretFilePath(profileName)
	if err != nil {
		return err
	}
	return os.Remove(path)
}
