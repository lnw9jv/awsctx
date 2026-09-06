package state_test

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lnw9jv/awsctx/internal/state"
)

func TestPreviousReplacementKeepsOpenReadersIntact(t *testing.T) {
	dir := t.TempDir()
	if err := state.SetPrevious(dir, "development"); err != nil {
		t.Fatal(err)
	}
	reader, err := os.Open(filepath.Join(dir, "previous"))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	if err := state.SetPrevious(dir, "prod"); err != nil {
		t.Fatal(err)
	}
	old, err := io.ReadAll(reader)
	if err != nil || string(old) != "development" {
		t.Fatalf("existing reader = %q, %v; want development", old, err)
	}
	if got, err := state.GetPrevious(dir); err != nil || got != "prod" {
		t.Fatalf("new reader = %q, %v; want prod", got, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("state directory = %v, %v; want only previous", entries, err)
	}
}

func TestPreviousFailedReplacementCleansTemporaryFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "previous")
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := state.SetPrevious(dir, "prod"); err == nil {
		t.Fatal("expected error replacing a directory")
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 || !entries[0].IsDir() {
		t.Fatalf("state directory = %v, %v; want original directory only", entries, err)
	}
}

func TestPreviousRoundtrip(t *testing.T) {
	dir := t.TempDir()
	if err := state.SetPrevious(dir, "dev"); err != nil {
		t.Fatal(err)
	}
	prev, err := state.GetPrevious(dir)
	if err != nil {
		t.Fatal(err)
	}
	if prev != "dev" {
		t.Fatalf("expected dev, got %s", prev)
	}
}

func TestPreviousNotExist(t *testing.T) {
	_, err := state.GetPrevious(t.TempDir())
	if err == nil {
		t.Fatal("expected error when no previous profile")
	}
}

func TestPreviousEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := state.SetPrevious(dir, " \n\t"); err != nil {
		t.Fatal(err)
	}

	_, err := state.GetPrevious(dir)
	if err == nil {
		t.Fatal("GetPrevious() error = nil, want empty-state error")
	}
	if !strings.Contains(err.Error(), "previous profile state is empty") {
		t.Errorf("GetPrevious() error = %q, want empty-state error", err)
	}
}
