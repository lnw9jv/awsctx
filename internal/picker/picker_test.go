package picker_test

import (
	"testing"

	"github.com/lnw9jv/awsctx/internal/picker"
)

func TestPickRejectsEmptyItems(t *testing.T) {
	_, err := picker.Pick(nil, "")
	if err == nil {
		t.Fatal("Pick() error = nil, want empty-items error")
	}
	if got, want := err.Error(), "no AWS profiles found"; got != want {
		t.Errorf("Pick() error = %q, want %q", got, want)
	}
}
