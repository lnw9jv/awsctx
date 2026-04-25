package picker

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/term"
)

// Pick tries fzf first; falls back to the built-in TUI if fzf is not installed.
func Pick(items []string, currentProfile string) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no AWS profiles found")
	}
	if _, err := exec.LookPath("fzf"); err == nil {
		return pickWithFzf(items, currentProfile)
	}
	fmt.Fprintln(os.Stderr, "tip: install fzf for a better picker (brew install fzf)")
	return pickLegacy(items, currentProfile)
}

func pickWithFzf(items []string, currentProfile string) (string, error) {
	// Mark the current profile so the user can see it at a glance.
	lines := make([]string, len(items))
	for i, item := range items {
		if item == currentProfile {
			lines[i] = item + " ✓"
		} else {
			lines[i] = item
		}
	}

	header := "Select AWS profile"
	if currentProfile != "" {
		header = "current: " + currentProfile
	}

	var stdout bytes.Buffer
	cmd := exec.Command("fzf",
		"--height=40%",
		"--layout=reverse",
		"--header="+header,
		"--prompt=AWS Profile> ",
		"--pointer=▶",
	)
	cmd.Stdin = strings.NewReader(strings.Join(lines, "\n"))
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 130 {
			return "", fmt.Errorf("cancelled")
		}
		return "", fmt.Errorf("fzf: %w", err)
	}

	selected := strings.TrimSuffix(strings.TrimSpace(stdout.String()), " ✓")
	if selected == "" {
		return "", fmt.Errorf("no profile selected")
	}
	return selected, nil
}

// pickLegacy is the original from-scratch TUI used when fzf is not available.
func pickLegacy(items []string, currentProfile string) (string, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("cannot open terminal: %w", err)
	}
	defer tty.Close()

	fd := int(tty.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer term.Restore(fd, oldState)

	query := ""
	cursor := 0

	for {
		filtered := filter(items, query)
		if cursor >= len(filtered) {
			cursor = max(0, len(filtered)-1)
		}

		render(tty, query, filtered, cursor, currentProfile)

		b := make([]byte, 4)
		n, err := tty.Read(b)
		if err != nil || n == 0 {
			return "", fmt.Errorf("read error")
		}

		switch {
		case b[0] == 13: // Enter
			fmt.Fprintf(tty, "\r\n")
			if len(filtered) == 0 {
				return "", fmt.Errorf("no profile selected")
			}
			return filtered[cursor], nil
		case b[0] == 3 || b[0] == 27 && n == 1: // Ctrl-C or Escape
			fmt.Fprintf(tty, "\r\n")
			return "", fmt.Errorf("cancelled")
		case n >= 3 && b[0] == 27 && b[1] == 91 && b[2] == 65: // Up
			if cursor > 0 {
				cursor--
			}
		case n >= 3 && b[0] == 27 && b[1] == 91 && b[2] == 66: // Down
			if cursor < len(filtered)-1 {
				cursor++
			}
		case b[0] == 127 || b[0] == 8: // Backspace
			if len(query) > 0 {
				query = query[:len(query)-1]
				cursor = 0
			}
		default:
			if b[0] >= 32 && b[0] < 127 {
				query += string(b[:1])
				cursor = 0
			}
		}
	}
}

func filter(items []string, query string) []string {
	if query == "" {
		return items
	}
	q := strings.ToLower(query)
	var out []string
	for _, item := range items {
		if strings.Contains(strings.ToLower(item), q) {
			out = append(out, item)
		}
	}
	return out
}

func render(tty *os.File, query string, filtered []string, cursor int, current string) {
	fmt.Fprintf(tty, "\r\033[K> %s\r\n", query)
	for i, p := range filtered {
		label := p
		if p == current {
			label += " ✓"
		}
		if i == cursor {
			fmt.Fprintf(tty, "\r\033[K\033[7m  %s\033[0m\r\n", label)
		} else {
			fmt.Fprintf(tty, "\r\033[K  %s\r\n", label)
		}
	}
	lines := len(filtered) + 1
	fmt.Fprintf(tty, "\033[%dA", lines)
	fmt.Fprintf(tty, "\r\033[%dC", len(query)+2)
}

