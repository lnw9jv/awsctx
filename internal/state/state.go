package state

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var ErrNoPrevious = errors.New("no previous profile — switch profiles first")

type State struct{ dir string }

func New(dir string) *State { return &State{dir: dir} }

func DefaultDir() string {
	if d := os.Getenv("AWSCTX_STATE_DIR"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "awsctx")
}

func (s *State) previousPath() string {
	return filepath.Join(s.dir, "previous")
}

func (s *State) SetPrevious(profile string) error {
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(s.previousPath(), []byte(profile), 0o600)
}

func (s *State) GetPrevious() (string, error) {
	b, err := os.ReadFile(s.previousPath())
	if os.IsNotExist(err) {
		return "", ErrNoPrevious
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}
