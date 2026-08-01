package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

type errorWriter struct {
	err error
}

func (w errorWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestRootUsesCommandOutputWriter(t *testing.T) {
	oldUnset, oldCurrent := unsetFlag, currentFlag
	oldRegion, oldProfile := regionFlag, profileFlag
	t.Cleanup(func() {
		unsetFlag, currentFlag = oldUnset, oldCurrent
		regionFlag, profileFlag = oldRegion, oldProfile
	})

	unsetFlag = true
	currentFlag = false
	regionFlag = ""
	profileFlag = ""

	var stdout bytes.Buffer
	testCmd := &cobra.Command{}
	testCmd.SetOut(&stdout)
	if err := rootCmd.RunE(testCmd, nil); err != nil {
		t.Fatalf("root RunE() error = %v", err)
	}
	if got, want := stdout.String(), "unset AWS_PROFILE\nunset AWS_DEFAULT_REGION\n"; got != want {
		t.Errorf("command stdout = %q, want %q", got, want)
	}
}

func TestRootPropagatesOutputError(t *testing.T) {
	oldUnset, oldCurrent := unsetFlag, currentFlag
	oldRegion, oldProfile := regionFlag, profileFlag
	t.Cleanup(func() {
		unsetFlag, currentFlag = oldUnset, oldCurrent
		regionFlag, profileFlag = oldRegion, oldProfile
	})

	unsetFlag = true
	currentFlag = false
	regionFlag = ""
	profileFlag = ""

	wantErr := errors.New("write failed")
	var stderr bytes.Buffer
	testCmd := &cobra.Command{}
	testCmd.SetOut(errorWriter{err: wantErr})
	testCmd.SetErr(&stderr)
	err := rootCmd.RunE(testCmd, nil)
	if !errors.Is(err, wantErr) {
		t.Fatalf("root RunE() error = %v, want %v", err, wantErr)
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty", stderr.String())
	}
}

func TestRootDoesNotSavePreviousWhenOutputFails(t *testing.T) {
	oldUnset, oldCurrent := unsetFlag, currentFlag
	oldRegion, oldProfile := regionFlag, profileFlag
	t.Cleanup(func() {
		unsetFlag, currentFlag = oldUnset, oldCurrent
		regionFlag, profileFlag = oldRegion, oldProfile
	})

	configPath := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(configPath, []byte("[profile dev]\n[profile prod]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	stateDir := t.TempDir()
	t.Setenv("AWS_CONFIG_FILE", configPath)
	t.Setenv("AWSCTX_STATE_DIR", stateDir)
	t.Setenv("AWS_PROFILE", "dev")
	unsetFlag = false
	currentFlag = false
	regionFlag = ""
	profileFlag = ""

	wantErr := errors.New("write failed")
	var stderr bytes.Buffer
	testCmd := &cobra.Command{}
	testCmd.SetOut(errorWriter{err: wantErr})
	testCmd.SetErr(&stderr)
	err := rootCmd.RunE(testCmd, []string{"prod"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("root RunE() error = %v, want %v", err, wantErr)
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(stateDir, "previous")); !os.IsNotExist(err) {
		t.Fatalf("previous state exists after output failure: %v", err)
	}
}
