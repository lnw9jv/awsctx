package state

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func DefaultDir() string {
	if d := os.Getenv("AWSCTX_STATE_DIR"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "awsctx")
}

func previousPath(dir string) string {
	return filepath.Join(dir, "previous")
}

func SetPrevious(dir, profile string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	// Replace only after the complete value is written, so readers never see a
	// truncated profile. CreateTemp uses 0600 and the same filesystem as previous.
	f, err := os.CreateTemp(dir, ".previous-*")
	if err != nil {
		return fmt.Errorf("create previous profile: %w", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(profile); err != nil {
		_ = f.Close()
		return fmt.Errorf("write previous profile: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close previous profile: %w", err)
	}
	if err := os.Rename(f.Name(), previousPath(dir)); err != nil {
		return fmt.Errorf("replace previous profile: %w", err)
	}
	return nil
}

func GetPrevious(dir string) (string, error) {
	b, err := os.ReadFile(previousPath(dir))
	if os.IsNotExist(err) {
		return "", errors.New("no previous profile — switch profiles first")
	}
	if err != nil {
		return "", err
	}
	profile := strings.TrimSpace(string(b))
	if profile == "" {
		return "", errors.New("previous profile state is empty")
	}
	return profile, nil
}
