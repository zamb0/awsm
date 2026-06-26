package mfa

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/zalando/go-keyring"
)

// validSecret is a known-good base32 TOTP seed usable in tests.
const validSecret = "JBSWY3DPEHPK3PXP"

func init() {
	// Redirect all keyring operations to an in-memory mock so tests never
	// touch the real OS keychain.
	keyring.MockInit()
}

// setTempHome redirects UserHomeDir by overriding HOME so file-fallback paths
// land in a temporary directory. Returns a cleanup func.
func setTempHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	// USERPROFILE is used on Windows
	t.Setenv("USERPROFILE", dir)
	return dir
}

// TestSetTOTPSecret_StoresAndRetrieves verifies round-trip via keyring mock.
func TestSetTOTPSecret_StoresAndRetrieves(t *testing.T) {
	setTempHome(t)
	const profile = "test-profile"

	if err := SetTOTPSecret(profile, validSecret); err != nil {
		t.Fatalf("SetTOTPSecret: %v", err)
	}

	got, err := GetTOTPSecret(profile)
	if err != nil {
		t.Fatalf("GetTOTPSecret: %v", err)
	}
	if got != validSecret {
		t.Errorf("got %q, want %q", got, validSecret)
	}
}

// TestSetTOTPSecret_NormalizesSpacesAndCase ensures secrets with spaces or
// lowercase are stored as clean uppercase base32.
func TestSetTOTPSecret_NormalizesSpacesAndCase(t *testing.T) {
	setTempHome(t)
	const profile = "norm-profile"
	withSpaces := "jbsw y3dp ehpk 3pxp" // same secret, lowercase + spaces

	if err := SetTOTPSecret(profile, withSpaces); err != nil {
		t.Fatalf("SetTOTPSecret: %v", err)
	}

	got, err := GetTOTPSecret(profile)
	if err != nil {
		t.Fatalf("GetTOTPSecret: %v", err)
	}
	if got != validSecret {
		t.Errorf("got %q, want %q", got, validSecret)
	}
}

// TestSetTOTPSecret_RejectsEmpty ensures empty or whitespace-only input fails.
func TestSetTOTPSecret_RejectsEmpty(t *testing.T) {
	setTempHome(t)
	for _, input := range []string{"", "   "} {
		if err := SetTOTPSecret("p", input); err == nil {
			t.Errorf("expected error for empty secret %q", input)
		}
	}
}

// TestSetTOTPSecret_RejectsInvalidBase32 ensures garbage secrets are rejected.
func TestSetTOTPSecret_RejectsInvalidBase32(t *testing.T) {
	setTempHome(t)
	if err := SetTOTPSecret("p", "NOT!VALID@BASE32#"); err == nil {
		t.Error("expected error for invalid base32 secret")
	}
}

// TestDeleteTOTPSecret_RemovesEntry verifies the secret is gone after deletion.
func TestDeleteTOTPSecret_RemovesEntry(t *testing.T) {
	setTempHome(t)
	const profile = "del-profile"

	if err := SetTOTPSecret(profile, validSecret); err != nil {
		t.Fatalf("SetTOTPSecret: %v", err)
	}
	if err := DeleteTOTPSecret(profile); err != nil {
		t.Fatalf("DeleteTOTPSecret: %v", err)
	}
	if _, err := GetTOTPSecret(profile); err == nil {
		t.Error("expected error after deletion, got nil")
	}
}

// TestDeleteTOTPSecret_NonExistent returns an error when nothing is stored.
func TestDeleteTOTPSecret_NonExistent(t *testing.T) {
	setTempHome(t)
	if err := DeleteTOTPSecret("ghost-profile"); err == nil {
		t.Error("expected error deleting non-existent secret")
	}
}

// TestGetTOTPSecret_NotFound returns an error for unknown profile.
func TestGetTOTPSecret_NotFound(t *testing.T) {
	setTempHome(t)
	if _, err := GetTOTPSecret("no-such-profile"); err == nil {
		t.Error("expected error, got nil")
	}
}

// TestGenerateTOTP_ProducesValidCode verifies the generated code is 6 digits
// and matches what pquerna/otp would produce for the same instant.
func TestGenerateTOTP_ProducesValidCode(t *testing.T) {
	setTempHome(t)
	const profile = "gen-profile"

	if err := SetTOTPSecret(profile, validSecret); err != nil {
		t.Fatalf("SetTOTPSecret: %v", err)
	}

	now := time.Now()
	code, err := GenerateTOTP(profile)
	if err != nil {
		t.Fatalf("GenerateTOTP: %v", err)
	}

	if len(code) != 6 {
		t.Errorf("expected 6-digit code, got %q (len %d)", code, len(code))
	}
	for _, ch := range code {
		if ch < '0' || ch > '9' {
			t.Errorf("non-digit in TOTP code: %q", code)
		}
	}

	expected, _ := totp.GenerateCode(validSecret, now)
	if code != expected {
		t.Errorf("code mismatch: got %q, want %q", code, expected)
	}
}

// TestGenerateTOTP_NoSecret returns an error when no secret is stored.
func TestGenerateTOTP_NoSecret(t *testing.T) {
	setTempHome(t)
	if _, err := GenerateTOTP("no-secret-profile"); err == nil {
		t.Error("expected error, got nil")
	}
}

// TestFileFallback_RoundTrip exercises the file-based storage path directly.
func TestFileFallback_RoundTrip(t *testing.T) {
	dir := setTempHome(t)

	const profile = "file-profile"
	if err := writeSecretFile(profile, validSecret); err != nil {
		t.Fatalf("writeSecretFile: %v", err)
	}

	got, err := readSecretFile(profile)
	if err != nil {
		t.Fatalf("readSecretFile: %v", err)
	}
	if got != validSecret {
		t.Errorf("got %q, want %q", got, validSecret)
	}

	// Verify file permissions and location
	path := filepath.Join(dir, ".awsm", "mfa", profile+".totp")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("file permissions: got %04o, want 0600", perm)
	}

	if err := deleteSecretFile(profile); err != nil {
		t.Fatalf("deleteSecretFile: %v", err)
	}
	if _, err := readSecretFile(profile); err == nil {
		t.Error("expected error after delete")
	}
}

// TestFileFallback_StripsWhitespace ensures trailing newlines in file are handled.
func TestFileFallback_StripsWhitespace(t *testing.T) {
	setTempHome(t)
	const profile = "ws-profile"

	withNewline := validSecret + "\n"
	if err := writeSecretFile(profile, withNewline); err != nil {
		t.Fatalf("writeSecretFile: %v", err)
	}

	got, err := readSecretFile(profile)
	if err != nil {
		t.Fatalf("readSecretFile: %v", err)
	}
	if strings.Contains(got, "\n") {
		t.Errorf("readSecretFile did not strip newline: %q", got)
	}
}
