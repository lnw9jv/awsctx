package aws_test

import (
	"os"
	"testing"

	"github.com/lnw9jv/awsctx/internal/aws"
)

func TestLoadProfiles(t *testing.T) {
	f, _ := os.CreateTemp("", "aws-config-*")
	f.WriteString("[default]\n[profile dev]\n[profile prod]\n")
	f.Close()

	profiles, err := aws.LoadProfiles(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 3 {
		t.Fatalf("expected 3 profiles, got %d: %v", len(profiles), profiles)
	}
	found := map[string]bool{}
	for _, p := range profiles {
		found[p] = true
	}
	for _, want := range []string{"default", "dev", "prod"} {
		if !found[want] {
			t.Errorf("missing profile %q in %v", want, profiles)
		}
	}
}

func TestLoadProfilesNoDefault(t *testing.T) {
	f, _ := os.CreateTemp("", "aws-config-*")
	f.WriteString("[profile staging]\n[profile prod]\n")
	f.Close()

	profiles, err := aws.LoadProfiles(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d: %v", len(profiles), profiles)
	}
}
