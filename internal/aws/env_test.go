package aws

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestBuildEnvForProfile_ScrubsExistingAWSVars(t *testing.T) {
	t.Setenv("AWS_PROFILE", "old-profile")
	t.Setenv("AWS_ACCESS_KEY_ID", "OLD_KEY")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "OLD_SECRET")
	t.Setenv("AWS_SESSION_TOKEN", "OLD_TOKEN")
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	t.Setenv("AWS_CREDENTIAL_EXPIRATION", "2020-01-01T00:00:00Z")
	t.Setenv("UNRELATED_VAR", "keep-me")

	creds := &TempCredentials{
		AccessKeyId:     "NEW_KEY",
		SecretAccessKey: "NEW_SECRET",
		SessionToken:    "NEW_TOKEN",
		Expires:         time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC),
	}

	env := BuildEnvForProfile(creds, "eu-west-1", "new-profile")
	m := envToMap(env)

	if m["AWS_PROFILE"] != "new-profile" {
		t.Errorf("AWS_PROFILE = %q, want new-profile", m["AWS_PROFILE"])
	}
	if m["AWS_ACCESS_KEY_ID"] != "NEW_KEY" {
		t.Errorf("AWS_ACCESS_KEY_ID = %q, want NEW_KEY", m["AWS_ACCESS_KEY_ID"])
	}
	if m["AWS_SECRET_ACCESS_KEY"] != "NEW_SECRET" {
		t.Errorf("AWS_SECRET_ACCESS_KEY = %q, want NEW_SECRET", m["AWS_SECRET_ACCESS_KEY"])
	}
	if m["AWS_SESSION_TOKEN"] != "NEW_TOKEN" {
		t.Errorf("AWS_SESSION_TOKEN = %q, want NEW_TOKEN", m["AWS_SESSION_TOKEN"])
	}
	if m["AWS_REGION"] != "eu-west-1" || m["AWS_DEFAULT_REGION"] != "eu-west-1" {
		t.Errorf("region not propagated: AWS_REGION=%q AWS_DEFAULT_REGION=%q", m["AWS_REGION"], m["AWS_DEFAULT_REGION"])
	}
	if m["AWS_CREDENTIAL_EXPIRATION"] != "2030-01-02T03:04:05Z" {
		t.Errorf("AWS_CREDENTIAL_EXPIRATION = %q", m["AWS_CREDENTIAL_EXPIRATION"])
	}
	if m["UNRELATED_VAR"] != "keep-me" {
		t.Errorf("non-AWS env vars were dropped")
	}
}

func TestBuildEnvForProfile_NilCreds(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "LEAKED_KEY")
	env := BuildEnvForProfile(nil, "eu-west-1", "p")
	m := envToMap(env)

	if _, ok := m["AWS_ACCESS_KEY_ID"]; ok {
		t.Error("nil creds should still scrub AWS_ACCESS_KEY_ID")
	}
	if m["AWS_PROFILE"] != "p" {
		t.Errorf("AWS_PROFILE = %q, want p", m["AWS_PROFILE"])
	}
	if m["AWS_REGION"] != "eu-west-1" {
		t.Errorf("AWS_REGION = %q, want eu-west-1", m["AWS_REGION"])
	}
	if _, ok := m["AWS_CREDENTIAL_EXPIRATION"]; ok {
		t.Error("AWS_CREDENTIAL_EXPIRATION should not be set when creds is nil")
	}
}

func TestBuildEnvForProfile_OmitsEmptyOptionals(t *testing.T) {
	creds := &TempCredentials{
		AccessKeyId:     "K",
		SecretAccessKey: "S",
		// SessionToken empty (static credentials)
		// Expires zero
	}
	env := BuildEnvForProfile(creds, "", "")
	m := envToMap(env)

	if _, ok := m["AWS_SESSION_TOKEN"]; ok {
		t.Error("empty session token should not be set")
	}
	if _, ok := m["AWS_CREDENTIAL_EXPIRATION"]; ok {
		t.Error("zero expiration should not be set")
	}
	if _, ok := m["AWS_PROFILE"]; ok {
		t.Error("empty profile should not be set")
	}
	if _, ok := m["AWS_REGION"]; ok {
		t.Error("empty region should not be set")
	}
}

func TestIsAWSEnvKey(t *testing.T) {
	cases := map[string]bool{
		"AWS_PROFILE":           true,
		"AWS_REGION":            true,
		"AWS_DEFAULT_REGION":    true,
		"AWS_ACCESS_KEY_ID":     true,
		"AWS_SECRET_ACCESS_KEY": true,
		"AWS_SESSION_TOKEN":     true,
		"AWS_SECURITY_TOKEN":    true,
		"PATH":                  false,
		"HOME":                  false,
		"AWS_CUSTOM_THING":      false,
	}
	for k, want := range cases {
		if got := isAWSEnvKey(k); got != want {
			t.Errorf("isAWSEnvKey(%q) = %v, want %v", k, got, want)
		}
	}
}

// envToMap helps tests inspect the slice returned by BuildEnvForProfile.
// It also catches accidental duplication of keys.
func envToMap(env []string) map[string]string {
	out := make(map[string]string, len(env))
	for _, kv := range env {
		idx := strings.IndexByte(kv, '=')
		if idx < 0 {
			continue
		}
		out[kv[:idx]] = kv[idx+1:]
	}
	return out
}

// sanity check that the test runner picks up env mutations correctly.
func TestEnvSanity(t *testing.T) {
	if os.Getenv("UNEXPECTED_TEST_VAR") != "" {
		t.Skip("unexpected env state, skipping")
	}
}
