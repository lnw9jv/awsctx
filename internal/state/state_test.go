package state_test

import (
	"testing"

	"github.com/lnw9jv/awsctx/internal/state"
)

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
