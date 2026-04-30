//go:build integration

package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "awsctx")
	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", bin, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build failed: %v", err)
	}
	return bin
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "aws-config-*")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(content)
	f.Close()
	return f.Name()
}

func TestSwitchByName(t *testing.T) {
	bin := buildBinary(t)
	cfg := writeConfig(t, "[default]\n[profile dev]\n[profile prod]\n")
	stateDir := t.TempDir()

	cmd := exec.Command(bin, "dev")
	cmd.Env = append(os.Environ(),
		"AWS_CONFIG_FILE="+cfg,
		"AWSCTX_STATE_DIR="+stateDir,
	)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("switch failed: %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != "export AWS_PROFILE=dev" {
		t.Errorf("expected 'export AWS_PROFILE=dev', got %q", got)
	}
}

func TestSwitchPrevious(t *testing.T) {
	bin := buildBinary(t)
	cfg := writeConfig(t, "[profile dev]\n[profile prod]\n")
	stateDir := t.TempDir()

	run := func(profile string, args ...string) string {
		cmd := exec.Command(bin, args...)
		cmd.Env = append(os.Environ(),
			"AWS_CONFIG_FILE="+cfg,
			"AWSCTX_STATE_DIR="+stateDir,
			"AWS_PROFILE="+profile,
		)
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("cmd %v failed: %v", args, err)
		}
		return strings.TrimSpace(string(out))
	}

	run("dev", "prod") // switch dev→prod, writes previous=dev
	got := run("prod", "-")
	if got != "export AWS_PROFILE=dev" {
		t.Errorf("expected 'export AWS_PROFILE=dev', got %q", got)
	}
}

func TestUnset(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "-u")
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unset failed: %v", err)
	}
	got := strings.TrimSpace(string(out))
	want := "unset AWS_PROFILE\nunset AWS_DEFAULT_REGION"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestRegionFlag(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "-r", "us-east-1")
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("region flag failed: %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != "export AWS_DEFAULT_REGION=us-east-1" {
		t.Errorf("expected 'export AWS_DEFAULT_REGION=us-east-1', got %q", got)
	}
}

func TestRegionFlagWithProfile(t *testing.T) {
	bin := buildBinary(t)
	cfg := writeConfig(t, "[default]\n[profile dev]\n")
	stateDir := t.TempDir()

	cmd := exec.Command(bin, "dev", "-r", "ap-southeast-1")
	cmd.Env = append(os.Environ(),
		"AWS_CONFIG_FILE="+cfg,
		"AWSCTX_STATE_DIR="+stateDir,
	)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("region+profile failed: %v", err)
	}
	got := strings.TrimSpace(string(out))
	want := "export AWS_PROFILE=dev\nexport AWS_DEFAULT_REGION=ap-southeast-1"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestCurrent(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "-c")
	cmd.Env = append(os.Environ(), "AWS_PROFILE=staging")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("current failed: %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != "staging" {
		t.Errorf("expected 'staging', got %q", got)
	}
}
