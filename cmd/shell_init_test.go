package cmd_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lnw9jv/awsctx/cmd"
)

func TestShellInitZsh(t *testing.T) {
	if !strings.Contains(cmd.ShellWrapperPosix, "command awsctx") {
		t.Error("zsh wrapper missing 'command awsctx'")
	}
	if !strings.Contains(cmd.ShellWrapperPosix, "eval") {
		t.Error("zsh wrapper missing eval")
	}
}

func TestShellInitFish(t *testing.T) {
	if !strings.Contains(cmd.ShellWrapperFish, "command awsctx") {
		t.Error("fish wrapper missing 'command awsctx'")
	}
	if !strings.Contains(cmd.ShellWrapperFish, "set -gx") {
		t.Error("fish wrapper missing 'set -gx' for export translation")
	}
}

// TestShellInitFishTranslatesExports actually runs the wrapper under fish to
// prove it sets real shell variables, rather than just checking the source
// text. This is what would have caught the "=" vs space-separated set bug.
func TestShellInitFishTranslatesExports(t *testing.T) {
	fishPath, err := exec.LookPath("fish")
	if err != nil {
		t.Skip("fish not installed, skipping")
	}

	dir := t.TempDir()
	fakeAwsctx := filepath.Join(dir, "awsctx")
	script := "#!/bin/sh\necho 'export AWS_PROFILE=dev'\necho 'export AWS_DEFAULT_REGION=us-east-1'\n"
	if err := os.WriteFile(fakeAwsctx, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	fishScript := cmd.ShellWrapperFish + "\nawsctx\necho $AWS_PROFILE\necho $AWS_DEFAULT_REGION\n"
	c := exec.CommandContext(t.Context(), fishPath, "--no-config", "-c", fishScript)
	c.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"))
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("fish exited with error: %v\noutput:\n%s", err, out)
	}

	got := strings.TrimSpace(string(out))
	want := "dev\nus-east-1"
	if got != want {
		t.Errorf("fish wrapper did not set variables correctly\ngot:  %q\nwant: %q", got, want)
	}
}
