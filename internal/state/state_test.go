package state_test

import (
	"testing"

	"github.com/lnw9jv/awsctx/internal/state"
)

func TestPreviousRoundtrip(t *testing.T) {
	dir := t.TempDir()
	s := state.New(dir)
	if err := s.SetPrevious("dev"); err != nil {
		t.Fatal(err)
	}
	prev, err := s.GetPrevious()
	if err != nil {
		t.Fatal(err)
	}
	if prev != "dev" {
		t.Fatalf("expected dev, got %s", prev)
	}
}

func TestPreviousNotExist(t *testing.T) {
	s := state.New(t.TempDir())
	_, err := s.GetPrevious()
	if err == nil {
		t.Fatal("expected error when no previous profile")
	}
}
