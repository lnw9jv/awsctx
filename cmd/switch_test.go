package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSwitchProfileWarnsWhenPreviousCannotBeSaved(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "state-parent")
	if err := os.WriteFile(blocker, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AWSCTX_STATE_DIR", filepath.Join(blocker, "awsctx"))
	t.Setenv("AWS_PROFILE", "dev")

	var stdout, stderr bytes.Buffer
	err := switchProfile("prod", []string{"dev", "prod"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("switchProfile() error = %v", err)
	}
	if got, want := stdout.String(), "export AWS_PROFILE=prod\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got := stderr.String(); !strings.Contains(got, "warning: cannot save previous profile:") {
		t.Errorf("stderr = %q, want previous-profile warning", got)
	}
}

func TestSwitchProfileRejectsUnsafeValue(t *testing.T) {
	profile := "dev\nunset AWS_PROFILE"
	var stdout, stderr bytes.Buffer
	err := switchProfile(profile, []string{profile}, &stdout, &stderr)
	if err == nil {
		t.Fatal("switchProfile() error = nil, want unsafe-value error")
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
}

func TestSwitchProfilePropagatesOutputError(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("AWSCTX_STATE_DIR", stateDir)
	t.Setenv("AWS_PROFILE", "dev")
	wantErr := errors.New("write failed")
	var stderr bytes.Buffer
	err := switchProfile("prod", []string{"prod"}, errorWriter{err: wantErr}, &stderr)
	if !errors.Is(err, wantErr) {
		t.Fatalf("switchProfile() error = %v, want %v", err, wantErr)
	}
	if _, err := os.Stat(filepath.Join(stateDir, "previous")); !os.IsNotExist(err) {
		t.Fatalf("previous state exists after output failure: %v", err)
	}
}
