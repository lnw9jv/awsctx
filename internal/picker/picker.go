package picker

import (
	"errors"
	"fmt"
	"os"

	fzf "github.com/junegunn/fzf/src"
)

// Pick runs an embedded fzf (the github.com/junegunn/fzf/src library, compiled
// into the binary) so users don't need a separately installed fzf.
//
// stdout invariant: shell wrappers interpret machine-readable lines on stdout,
// so the picker must never let fzf write there. The selection comes back via
// the Output channel, and fzf's UI is drawn to the terminal — but we still
// redirect os.Stdout to /dev/tty for the duration of Run() to guarantee nothing
// fzf might print can corrupt the shell environment.
func Pick(items []string, currentProfile string) (string, error) {
	if len(items) == 0 {
		return "", errors.New("no AWS profiles found")
	}

	// Color the current profile green, matching kubectx's style.
	// fzf strips ANSI codes from output, so the returned string is plain.
	lines := make([]string, len(items))
	for i, item := range items {
		if item == currentProfile {
			lines[i] = "\033[32m" + item + "\033[0m"
		} else {
			lines[i] = item
		}
	}

	// Global fzf options can enable filtering or change selection/output semantics.
	// Keep this picker interactive and single-select regardless of shell defaults.
	opts, err := fzf.ParseOptions(false, []string{"--ansi", "--no-preview"})
	if err != nil {
		return "", fmt.Errorf("fzf: %w", err)
	}

	inputChan := make(chan string)
	go func() {
		for _, l := range lines {
			inputChan <- l
		}
		close(inputChan)
	}()

	outputChan := make(chan string, len(items))
	opts.Input = inputChan
	opts.Output = outputChan

	restore := guardStdout()
	code, runErr := fzf.Run(opts)
	close(outputChan)
	restore()

	switch code {
	case fzf.ExitInterrupt:
		return "", errors.New("cancelled")
	case fzf.ExitNoMatch:
		return "", errors.New("no profile selected")
	}
	if runErr != nil {
		return "", fmt.Errorf("fzf: %w", runErr)
	}

	var selected string
	for s := range outputChan {
		selected = s // single-select: keep the one (last) line
	}
	if selected == "" {
		return "", errors.New("no profile selected")
	}
	return selected, nil
}

// guardStdout points os.Stdout away from the machine-readable stdout for the
// duration of fzf.Run, preferring /dev/tty and falling back to stderr. It
// returns a function that restores the original os.Stdout.
func guardStdout() func() {
	saved := os.Stdout
	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		os.Stdout = os.Stderr
		return func() { os.Stdout = saved }
	}
	os.Stdout = tty
	return func() {
		os.Stdout = saved
		tty.Close()
	}
}
