package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestStatusBadge(t *testing.T) {
	cases := map[checkStatus]string{
		statusOK:   "OK",
		statusWarn: "WARN",
		statusFail: "FAIL",
		statusInfo: "INFO",
	}
	for s, want := range cases {
		got := statusBadge(s)
		if !strings.Contains(got, want) {
			t.Errorf("statusBadge(%q) = %q, want substring %q", s, got, want)
		}
	}
}

func TestSummarize(t *testing.T) {
	res := []checkResult{
		{Status: statusOK}, {Status: statusOK},
		{Status: statusWarn},
		{Status: statusFail},
		{Status: statusInfo}, {Status: statusInfo}, {Status: statusInfo},
	}
	got := summarize(res)
	want := "Summary: 2 ok, 1 warn, 1 fail, 3 info"
	if got != want {
		t.Errorf("summarize() = %q, want %q", got, want)
	}
}

func TestCheckVersions_AlwaysReturnsThreeInfoEntries(t *testing.T) {
	got := checkVersions()
	if len(got) != 3 {
		t.Fatalf("checkVersions() returned %d entries, want 3", len(got))
	}
	for _, r := range got {
		if r.Category != "Versions" {
			t.Errorf("unexpected category: %q", r.Category)
		}
		if r.Status != statusInfo {
			t.Errorf("expected statusInfo, got %q", r.Status)
		}
	}
}

func TestFileCheck_Missing(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "does-not-exist")
	r := fileCheck("cat", "label", tmp, 0o600)
	if r.Status != statusWarn {
		t.Errorf("missing file should warn, got %q (msg=%q)", r.Status, r.Message)
	}
}

func TestFileCheck_GoodPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("perm bits not enforced on windows")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "f")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	r := fileCheck("cat", "label", path, 0o600)
	if r.Status != statusOK {
		t.Errorf("0600 file should be OK, got %q (msg=%q)", r.Status, r.Message)
	}
}

func TestFileCheck_TooOpenPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("perm bits not enforced on windows")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "f")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := fileCheck("cat", "label", path, 0o600)
	if r.Status != statusWarn {
		t.Errorf("0644 file should warn, got %q (msg=%q)", r.Status, r.Message)
	}
	if !strings.Contains(r.Message, "permissions too open") {
		t.Errorf("warn message should mention permissions, got %q", r.Message)
	}
}

func TestCollectDoctorResults_RunsAllCategories(t *testing.T) {
	// Smoke test: ensure the aggregator runs without panicking and produces
	// at least one entry per known category.
	results := collectDoctorResults()
	if len(results) == 0 {
		t.Fatal("collectDoctorResults returned nothing")
	}
	seen := make(map[string]bool)
	for _, r := range results {
		seen[r.Category] = true
	}
	for _, cat := range []string{"Versions", "AWS files", "External tools", "awsm config", "Profiles", "Browsers"} {
		if !seen[cat] {
			t.Errorf("missing category in doctor output: %q", cat)
		}
	}
}
