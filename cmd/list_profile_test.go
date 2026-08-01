package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunListPropagatesOutputError(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(configPath, []byte("[profile dev]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AWS_CONFIG_FILE", configPath)

	wantErr := errors.New("write failed")
	testCmd := &cobra.Command{}
	testCmd.SetOut(errorWriter{err: wantErr})
	err := runList(testCmd, nil)
	if !errors.Is(err, wantErr) {
		t.Fatalf("runList() error = %v, want %v", err, wantErr)
	}
}
