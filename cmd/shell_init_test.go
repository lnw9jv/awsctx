package cmd_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lnw9jv/awsctx/cmd"
)

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

func TestShellInitPosixRejectsCommandInjection(t *testing.T) {
	shells := []struct {
		name string
		args []string
	}{
		{name: "bash", args: []string{"--noprofile", "--norc", "-c"}},
		{name: "zsh", args: []string{"-f", "-c"}},
	}

	for _, shell := range shells {
		t.Run(shell.name, func(t *testing.T) {
			shellPath, err := exec.LookPath(shell.name)
			if err != nil {
				t.Skipf("%s not installed, skipping", shell.name)
			}

			dir := t.TempDir()
			sentinel := filepath.Join(dir, "injected")
			value := "dev; touch " + sentinel
			fakeAwsctx := filepath.Join(dir, "awsctx")
			script := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' %q %q\n",
				"export AWS_PROFILE="+value,
				"export AWS_DEFAULT_REGION=us-east-1",
			)
			if err := os.WriteFile(fakeAwsctx, []byte(script), 0o755); err != nil {
				t.Fatal(err)
			}

			shellScript := cmd.ShellWrapperPosix + "\nawsctx\nprintf '%s\\n' \"$AWS_PROFILE\" \"$AWS_DEFAULT_REGION\"\n"
			args := append(shell.args, shellScript)
			c := exec.CommandContext(t.Context(), shellPath, args...)
			c.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"))
			out, err := c.CombinedOutput()
			if err != nil {
				t.Fatalf("%s exited with error: %v\noutput:\n%s", shell.name, err, out)
			}

			if got, want := strings.TrimSpace(string(out)), value+"\nus-east-1"; got != want {
				t.Errorf("shell variables = %q, want %q", got, want)
			}
			if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
				t.Errorf("injected command created %s", sentinel)
			}
		})
	}
}

func TestShellInitFishAllowsOnlyAWSContextVariables(t *testing.T) {
	fishPath, err := exec.LookPath("fish")
	if err != nil {
		t.Skip("fish not installed, skipping")
	}

	dir := t.TempDir()
	fakeAwsctx := filepath.Join(dir, "awsctx")
	script := "#!/bin/sh\n" +
		"echo 'export AWS_PROFILE=dev'\n" +
		"echo 'export AWS_DEFAULT_REGION=us-east-1'\n" +
		"echo 'export AWSCTX_UNRELATED=changed'\n" +
		"echo 'unset AWSCTX_KEEP'\n"
	if err := os.WriteFile(fakeAwsctx, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	fishScript := cmd.ShellWrapperFish + `
set -gx AWSCTX_UNRELATED original
set -gx AWSCTX_KEEP present
awsctx
printf 'VALUES:%s|%s|%s|%s\n' "$AWSCTX_UNRELATED" "$AWSCTX_KEEP" "$AWS_PROFILE" "$AWS_DEFAULT_REGION"
`
	c := exec.CommandContext(t.Context(), fishPath, "--no-config", "-c", fishScript)
	c.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"))
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("fish exited with error: %v\noutput:\n%s", err, out)
	}

	if got := string(out); !strings.Contains(got, "VALUES:original|present|dev|us-east-1\n") {
		t.Errorf("fish changed unrelated variables; output:\n%s", got)
	}
	for _, want := range []string{"export AWSCTX_UNRELATED=changed\n", "unset AWSCTX_KEEP\n"} {
		if !strings.Contains(string(out), want) {
			t.Errorf("fish did not preserve unsupported line %q; output:\n%s", want, out)
		}
	}
}

func TestShellInitFishTreatsMetacharactersAsData(t *testing.T) {
	fishPath, err := exec.LookPath("fish")
	if err != nil {
		t.Skip("fish not installed, skipping")
	}

	tests := []struct {
		name  string
		value string
	}{
		{name: "semicolon", value: "dev; touch {sentinel}"},
		{name: "command substitution", value: "dev$(touch {sentinel})"},
		{name: "quotes and glob", value: `dev * "quoted" 'single'`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			sentinel := filepath.Join(dir, "injected")
			value := strings.ReplaceAll(tt.value, "{sentinel}", sentinel)
			fakeAwsctx := filepath.Join(dir, "awsctx")
			script := "#!/bin/sh\nprintf '%s\\n' \"$AWSCTX_TEST_OUTPUT\"\n"
			if err := os.WriteFile(fakeAwsctx, []byte(script), 0o755); err != nil {
				t.Fatal(err)
			}

			fishScript := cmd.ShellWrapperFish + "\nawsctx\nprintf '%s\\n' \"$AWS_PROFILE\"\n"
			c := exec.CommandContext(t.Context(), fishPath, "--no-config", "-c", fishScript)
			c.Env = append(os.Environ(),
				"PATH="+dir+":"+os.Getenv("PATH"),
				"AWSCTX_TEST_OUTPUT=export AWS_PROFILE="+value,
			)
			out, err := c.CombinedOutput()
			if err != nil {
				t.Fatalf("fish exited with error: %v\noutput:\n%s", err, out)
			}
			if got := strings.TrimSpace(string(out)); got != value {
				t.Errorf("AWS_PROFILE = %q, want %q", got, value)
			}
			if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
				t.Errorf("injected command created %s", sentinel)
			}
		})
	}
}

func TestShellInitPassesThroughNonSwitchOutput(t *testing.T) {
	shells := []struct {
		name    string
		args    []string
		wrapper string
		script  string
	}{
		{
			name:    "bash",
			args:    []string{"--noprofile", "--norc", "-c"},
			wrapper: cmd.ShellWrapperPosix,
			script:  "awsctx --current\nawsctx --current=1\nawsctx -v\nawsctx __complete --profile ''\nprintf 'REGION:%s\\n' \"$AWS_DEFAULT_REGION\"\n",
		},
		{
			name:    "zsh",
			args:    []string{"-f", "-c"},
			wrapper: cmd.ShellWrapperPosix,
			script:  "awsctx --current\nawsctx --current=1\nawsctx -v\nawsctx __complete --profile ''\nprintf 'REGION:%s\\n' \"$AWS_DEFAULT_REGION\"\n",
		},
		{
			name:    "fish",
			args:    []string{"--no-config", "-c"},
			wrapper: cmd.ShellWrapperFish,
			script:  "awsctx --current\nawsctx --current=1\nawsctx -v\nawsctx __complete --profile ''\nprintf 'REGION:%s\\n' \"$AWS_DEFAULT_REGION\"\n",
		},
	}

	for _, shell := range shells {
		t.Run(shell.name, func(t *testing.T) {
			shellPath, err := exec.LookPath(shell.name)
			if err != nil {
				t.Skipf("%s not installed, skipping", shell.name)
			}

			dir := t.TempDir()
			fakeAwsctx := filepath.Join(dir, "awsctx")
			fakeScript := "#!/bin/sh\nprintf '%s\\n' 'export AWS_DEFAULT_REGION=hijacked'\n"
			if err := os.WriteFile(fakeAwsctx, []byte(fakeScript), 0o755); err != nil {
				t.Fatal(err)
			}

			shellScript := shell.wrapper + "\n" + shell.script
			args := append(shell.args, shellScript)
			c := exec.CommandContext(t.Context(), shellPath, args...)
			c.Env = append(os.Environ(),
				"PATH="+dir+":"+os.Getenv("PATH"),
				"AWS_DEFAULT_REGION=safe",
			)
			out, err := c.CombinedOutput()
			if err != nil {
				t.Fatalf("%s exited with error: %v\noutput:\n%s", shell.name, err, out)
			}

			want := "export AWS_DEFAULT_REGION=hijacked\n" +
				"export AWS_DEFAULT_REGION=hijacked\n" +
				"export AWS_DEFAULT_REGION=hijacked\n" +
				"export AWS_DEFAULT_REGION=hijacked\n" +
				"REGION:safe\n"
			if got := string(out); got != want {
				t.Errorf("non-switch output was interpreted\ngot:  %q\nwant: %q", got, want)
			}
		})
	}
}

func TestShellInitTreatsArgumentsAfterDoubleDashAsProfiles(t *testing.T) {
	shells := []struct {
		name    string
		args    []string
		wrapper string
	}{
		{name: "bash", args: []string{"--noprofile", "--norc", "-c"}, wrapper: cmd.ShellWrapperPosix},
		{name: "zsh", args: []string{"-f", "-c"}, wrapper: cmd.ShellWrapperPosix},
		{name: "fish", args: []string{"--no-config", "-c"}, wrapper: cmd.ShellWrapperFish},
	}

	for _, shell := range shells {
		t.Run(shell.name, func(t *testing.T) {
			shellPath, err := exec.LookPath(shell.name)
			if err != nil {
				t.Skipf("%s not installed, skipping", shell.name)
			}

			dir := t.TempDir()
			fakeAwsctx := filepath.Join(dir, "awsctx")
			fakeScript := "#!/bin/sh\nprintf '%s\\n' 'export AWS_PROFILE=--current'\n"
			if err := os.WriteFile(fakeAwsctx, []byte(fakeScript), 0o755); err != nil {
				t.Fatal(err)
			}

			shellScript := shell.wrapper + "\nawsctx -- --current\nprintf 'PROFILE:%s\\n' \"$AWS_PROFILE\"\n"
			args := append(shell.args, shellScript)
			c := exec.CommandContext(t.Context(), shellPath, args...)
			c.Env = append(os.Environ(),
				"PATH="+dir+":"+os.Getenv("PATH"),
				"AWS_PROFILE=original",
			)
			out, err := c.CombinedOutput()
			if err != nil {
				t.Fatalf("%s exited with error: %v\noutput:\n%s", shell.name, err, out)
			}
			if got, want := string(out), "PROFILE:--current\n"; got != want {
				t.Errorf("double-dash profile handling\ngot:  %q\nwant: %q", got, want)
			}
		})
	}
}
