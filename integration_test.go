//go:build integration

package main_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var integrationBinary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "awsctx-integration-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create integration temp directory: %v\n", err)
		os.Exit(1)
	}

	integrationBinary = filepath.Join(dir, "awsctx")
	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", integrationBinary, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "build integration binary: %v\n", err)
		_ = os.RemoveAll(dir)
		os.Exit(1)
	}

	code := m.Run()
	if err := os.RemoveAll(dir); err != nil && code == 0 {
		fmt.Fprintf(os.Stderr, "remove integration temp directory: %v\n", err)
		code = 1
	}
	os.Exit(code)
}

func buildBinary(t *testing.T) string {
	t.Helper()
	if integrationBinary == "" {
		t.Fatal("integration binary was not built")
	}
	return integrationBinary
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestSwitchByName(t *testing.T) {
	bin := buildBinary(t)
	cfg := writeConfig(t, "[default]\n[profile dev]\n[profile prod]\n")
	stateDir := t.TempDir()

	cmd := exec.CommandContext(t.Context(), bin, "dev")
	cmd.Env = append(os.Environ(),
		"AWS_CONFIG_FILE="+cfg,
		"AWSCTX_STATE_DIR="+stateDir,
	)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("switch failed: %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != "export AWS_PROFILE=dev" {
		t.Errorf("expected 'export AWS_PROFILE=dev', got %q", got)
	}
	if got, want := stderr.String(), "Switched to profile dev\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestSwitchPrevious(t *testing.T) {
	bin := buildBinary(t)
	cfg := writeConfig(t, "[profile dev]\n[profile prod]\n")
	stateDir := t.TempDir()

	run := func(profile string, args ...string) string {
		cmd := exec.CommandContext(t.Context(), bin, args...)
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

func TestSwitchPreviousDeleted(t *testing.T) {
	bin := buildBinary(t)
	cfg := writeConfig(t, "[profile dev]\n")
	stateDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(stateDir, "previous"), []byte("gone"), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := exec.CommandContext(t.Context(), bin, "-")
	cmd.Env = append(os.Environ(),
		"AWS_CONFIG_FILE="+cfg,
		"AWSCTX_STATE_DIR="+stateDir,
	)
	out, err := cmd.Output()
	if err == nil {
		t.Fatal("expected error when previous profile no longer exists in config")
	}
	if len(out) != 0 {
		t.Errorf("expected empty stdout, got %q", out)
	}
}

func TestSwitchPreviousStateErrors(t *testing.T) {
	bin := buildBinary(t)
	cfg := writeConfig(t, "[profile dev]\n")
	tests := []struct {
		name      string
		setup     func(*testing.T, string)
		wantError string
	}{
		{
			name:      "missing",
			setup:     func(*testing.T, string) {},
			wantError: "no previous profile",
		},
		{
			name: "empty",
			setup: func(t *testing.T, stateDir string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(stateDir, "previous"), []byte(" \n\t"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantError: "previous profile state is empty",
		},
		{
			name: "unreadable",
			setup: func(t *testing.T, stateDir string) {
				t.Helper()
				if err := os.Mkdir(filepath.Join(stateDir, "previous"), 0o700); err != nil {
					t.Fatal(err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			tt.setup(t, stateDir)
			cmd := exec.CommandContext(t.Context(), bin, "-")
			cmd.Env = append(os.Environ(),
				"AWS_CONFIG_FILE="+cfg,
				"AWSCTX_STATE_DIR="+stateDir,
			)
			var stdout, stderr strings.Builder
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err == nil {
				t.Fatal("previous-profile state error succeeded, want error")
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want empty", stdout.String())
			}
			if tt.wantError != "" && !strings.Contains(stderr.String(), tt.wantError) {
				t.Errorf("stderr = %q, want %q", stderr.String(), tt.wantError)
			}
		})
	}
}

func TestUnset(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.CommandContext(t.Context(), bin, "-u")
	cmd.Env = os.Environ()
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unset failed: %v", err)
	}
	got := strings.TrimSpace(string(out))
	want := "unset AWS_PROFILE\nunset AWS_DEFAULT_REGION"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
	if got, want := stderr.String(), "Cleared AWS profile and region\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestRegionFlag(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.CommandContext(t.Context(), bin, "-r", "us-east-1")
	cmd.Env = os.Environ()
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("region flag failed: %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != "export AWS_DEFAULT_REGION=us-east-1" {
		t.Errorf("expected 'export AWS_DEFAULT_REGION=us-east-1', got %q", got)
	}
	if got, want := stderr.String(), "Switched to region us-east-1\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestRegionFlagWithProfile(t *testing.T) {
	bin := buildBinary(t)
	cfg := writeConfig(t, "[default]\n[profile dev]\n")
	stateDir := t.TempDir()

	cmd := exec.CommandContext(t.Context(), bin, "dev", "-r", "ap-southeast-1")
	cmd.Env = append(os.Environ(),
		"AWS_CONFIG_FILE="+cfg,
		"AWSCTX_STATE_DIR="+stateDir,
	)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("region+profile failed: %v", err)
	}
	got := strings.TrimSpace(string(out))
	want := "export AWS_PROFILE=dev\nexport AWS_DEFAULT_REGION=ap-southeast-1"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
	if got, want := stderr.String(), "Switched to profile dev and region ap-southeast-1\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestProfileFlag(t *testing.T) {
	bin := buildBinary(t)
	cfg := writeConfig(t, "[default]\n[profile dev]\n[profile prod]\n")
	stateDir := t.TempDir()

	cmd := exec.CommandContext(t.Context(), bin, "--profile", "dev")
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

	cmd := exec.CommandContext(t.Context(), bin, "-p", "prod")
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

	cmd := exec.CommandContext(t.Context(), bin, "--profile", "dev", "-r", "ap-southeast-1")
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
		cmd := exec.CommandContext(t.Context(), bin, args...)
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
			cmd := exec.CommandContext(t.Context(), bin, "__complete", flag, "")
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

	cmd := exec.CommandContext(t.Context(), bin, "__complete", "")
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

	cmd := exec.CommandContext(t.Context(), bin, "list-profile")
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

	cmd := exec.CommandContext(t.Context(), bin, "ls")
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

	cmd := exec.CommandContext(t.Context(), bin, "list-profile")
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
	cmd := exec.CommandContext(t.Context(), bin, "-c")
	cmd.Env = append(os.Environ(), "AWS_PROFILE=staging")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("current failed: %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != "staging" {
		t.Errorf("expected 'staging', got %q", got)
	}
}

func TestRejectsAmbiguousRootArguments(t *testing.T) {
	bin := buildBinary(t)
	cfg := writeConfig(t, "[profile dev]\n[profile prod]\n")

	tests := []struct {
		name string
		args []string
	}{
		{name: "positional and profile flag", args: []string{"dev", "--profile", "prod"}},
		{name: "current and region", args: []string{"--current", "--region", "us-east-1"}},
		{name: "current and unset", args: []string{"--current", "--unset"}},
		{name: "unset and region", args: []string{"--unset", "--region", "us-east-1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.CommandContext(t.Context(), bin, tt.args...)
			cmd.Env = append(os.Environ(),
				"AWS_CONFIG_FILE="+cfg,
				"AWSCTX_STATE_DIR="+t.TempDir(),
			)
			out, err := cmd.Output()
			if err == nil {
				t.Fatalf("awsctx %v succeeded, want error", tt.args)
			}
			if len(out) != 0 {
				t.Errorf("stdout = %q, want empty", out)
			}
		})
	}
}

func TestCompletionRejectsUnsupportedShell(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.CommandContext(t.Context(), bin, "completion", "powershell")
	out, err := cmd.Output()
	if err == nil {
		t.Fatal("unsupported completion shell succeeded, want error")
	}
	if len(out) != 0 {
		t.Errorf("stdout = %q, want empty", out)
	}
}

func TestRegionRejectsControlCharacters(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.CommandContext(t.Context(), bin, "--region", "us-east-1\nunset AWS_PROFILE")
	out, err := cmd.Output()
	if err == nil {
		t.Fatal("region with control characters succeeded, want error")
	}
	if len(out) != 0 {
		t.Errorf("stdout = %q, want empty", out)
	}
}

func TestFalseBooleanFlagsDoNotConflict(t *testing.T) {
	bin := buildBinary(t)
	for _, flag := range []string{"--current=false", "--unset=false"} {
		t.Run(flag, func(t *testing.T) {
			cmd := exec.CommandContext(t.Context(), bin, flag, "--region", "us-east-1")
			out, err := cmd.Output()
			if err != nil {
				t.Fatalf("awsctx %s failed: %v", flag, err)
			}
			if got, want := strings.TrimSpace(string(out)), "export AWS_DEFAULT_REGION=us-east-1"; got != want {
				t.Errorf("stdout = %q, want %q", got, want)
			}
		})
	}
}

func TestEmptyStringFlagsFail(t *testing.T) {
	bin := buildBinary(t)
	cfg := writeConfig(t, "[profile dev]\n")
	tests := []struct {
		name string
		args []string
	}{
		{name: "profile", args: []string{"--profile", "", "--region", "us-east-1"}},
		{name: "region", args: []string{"--profile", "dev", "--region", ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.CommandContext(t.Context(), bin, tt.args...)
			cmd.Env = append(os.Environ(), "AWS_CONFIG_FILE="+cfg)
			out, err := cmd.Output()
			if err == nil {
				t.Fatalf("awsctx %v succeeded, want empty-value error", tt.args)
			}
			if len(out) != 0 {
				t.Errorf("stdout = %q, want empty", out)
			}
		})
	}
}

func TestEmptyPositionalProfileFails(t *testing.T) {
	bin := buildBinary(t)
	cfg := writeConfig(t, "[profile dev]\n")
	cmd := exec.CommandContext(t.Context(), bin, "")
	cmd.Env = append(os.Environ(),
		"AWS_CONFIG_FILE="+cfg,
		"AWSCTX_STATE_DIR="+t.TempDir(),
	)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("empty positional profile succeeded, want error")
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "profile cannot be empty") {
		t.Errorf("stderr = %q, want empty-profile error", stderr.String())
	}
}
