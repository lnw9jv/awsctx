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

func TestProfileFlag(t *testing.T) {
	bin := buildBinary(t)
	cfg := writeConfig(t, "[default]\n[profile dev]\n[profile prod]\n")
	stateDir := t.TempDir()

	cmd := exec.Command(bin, "--profile", "dev")
	cmd.Env = append(os.Environ(),
		"AWS_CONFIG_FILE="+cfg,
		"AWSCTX_STATE_DIR="+stateDir,
	)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("--profile flag failed: %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != "export AWS_PROFILE=dev" {
		t.Errorf("expected 'export AWS_PROFILE=dev', got %q", got)
	}
}

func TestProfileShortFlag(t *testing.T) {
	bin := buildBinary(t)
	cfg := writeConfig(t, "[default]\n[profile dev]\n[profile prod]\n")
	stateDir := t.TempDir()

	cmd := exec.Command(bin, "-p", "prod")
	cmd.Env = append(os.Environ(),
		"AWS_CONFIG_FILE="+cfg,
		"AWSCTX_STATE_DIR="+stateDir,
	)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("-p flag failed: %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != "export AWS_PROFILE=prod" {
		t.Errorf("expected 'export AWS_PROFILE=prod', got %q", got)
	}
}

func TestProfileFlagWithRegion(t *testing.T) {
	bin := buildBinary(t)
	cfg := writeConfig(t, "[default]\n[profile dev]\n")
	stateDir := t.TempDir()

	cmd := exec.Command(bin, "--profile", "dev", "-r", "ap-southeast-1")
	cmd.Env = append(os.Environ(),
		"AWS_CONFIG_FILE="+cfg,
		"AWSCTX_STATE_DIR="+stateDir,
	)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("--profile + -r failed: %v", err)
	}
	got := strings.TrimSpace(string(out))
	want := "export AWS_PROFILE=dev\nexport AWS_DEFAULT_REGION=ap-southeast-1"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestProfileFlagPrevious(t *testing.T) {
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

	run("dev", "prod")
	got := run("prod", "--profile", "-")
	if got != "export AWS_PROFILE=dev" {
		t.Errorf("expected 'export AWS_PROFILE=dev', got %q", got)
	}
}

func TestProfileFlagCompletion(t *testing.T) {
	bin := buildBinary(t)
	cfg := writeConfig(t, "[default]\n[profile dev]\n[profile prod]\n")

	for _, flag := range []string{"--profile", "-p"} {
		t.Run(flag, func(t *testing.T) {
			cmd := exec.Command(bin, "__complete", flag, "")
			cmd.Env = append(os.Environ(), "AWS_CONFIG_FILE="+cfg)
			out, err := cmd.Output()
			if err != nil {
				t.Fatalf("completion for %s failed: %v", flag, err)
			}
			got := string(out)
			for _, want := range []string{"dev", "prod", "-"} {
				if !strings.Contains(got, want) {
					t.Errorf("completion for %s: expected %q in output, got:\n%s", flag, want, got)
				}
			}
		})
	}
}

func TestPositionalArgNoCompletion(t *testing.T) {
	bin := buildBinary(t)
	cfg := writeConfig(t, "[profile dev]\n[profile prod]\n")

	cmd := exec.Command(bin, "__complete", "")
	cmd.Env = append(os.Environ(), "AWS_CONFIG_FILE="+cfg)
	out, _ := cmd.Output()
	got := string(out)
	for _, unwanted := range []string{"dev", "prod"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("positional completion should not list profiles, got %q in output:\n%s", unwanted, got)
		}
	}
}

func TestList(t *testing.T) {
	bin := buildBinary(t)
	cfg := writeConfig(t, "[default]\nsso_account_id = 111122223333\n\n[profile dev]\nsso_account_id = 444455556666\n\n[profile prod]\n")

	cmd := exec.Command(bin, "list-profile")
	cmd.Env = append(os.Environ(),
		"AWS_CONFIG_FILE="+cfg,
		"AWS_PROFILE=dev",
	)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	got := string(out)
	if !strings.Contains(got, "* dev") {
		t.Errorf("current profile not marked with *, got:\n%s", got)
	}
	if !strings.Contains(got, "111122223333") {
		t.Errorf("expected default account ID in output, got:\n%s", got)
	}
	if !strings.Contains(got, "444455556666") {
		t.Errorf("expected dev account ID in output, got:\n%s", got)
	}
	if !strings.Contains(got, "-") {
		t.Errorf("expected '-' for prod (no account ID), got:\n%s", got)
	}
}

func TestListAlias(t *testing.T) {
	bin := buildBinary(t)
	cfg := writeConfig(t, "[profile dev]\n[profile prod]\n")

	cmd := exec.Command(bin, "ls")
	cmd.Env = append(os.Environ(), "AWS_CONFIG_FILE="+cfg)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("ls alias failed: %v", err)
	}
	got := string(out)
	for _, want := range []string{"dev", "prod"} {
		if !strings.Contains(got, want) {
			t.Errorf("ls alias: expected %q in output, got:\n%s", want, got)
		}
	}
}

func TestListNoCurrentMark(t *testing.T) {
	bin := buildBinary(t)
	cfg := writeConfig(t, "[profile dev]\n[profile prod]\n")

	cmd := exec.Command(bin, "list-profile")
	cmd.Env = append(os.Environ(),
		"AWS_CONFIG_FILE="+cfg,
		"AWS_PROFILE=",
	)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if strings.Contains(string(out), "*") {
		t.Errorf("expected no '*' marker when no profile active, got:\n%s", out)
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
