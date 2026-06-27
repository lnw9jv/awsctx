package state

import (
	"errors"
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
	return os.WriteFile(previousPath(dir), []byte(profile), 0o600)
}

func GetPrevious(dir string) (string, error) {
	b, err := os.ReadFile(previousPath(dir))
	if os.IsNotExist(err) {
		return "", errors.New("no previous profile — switch profiles first")
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}
