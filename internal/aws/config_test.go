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

func TestLoadProfileDetails(t *testing.T) {
	f, _ := os.CreateTemp("", "aws-config-*")
	f.WriteString("[default]\nsso_account_id = 111111111111\n\n[profile dev]\nsso_account_id = 222222222222\n\n[profile prod]\naccount_id = 333333333333\n\n[profile staging]\n")
	f.Close()

	profiles, err := aws.LoadProfileDetails(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 4 {
		t.Fatalf("expected 4 profiles, got %d: %v", len(profiles), profiles)
	}

	byName := map[string]aws.Profile{}
	for _, p := range profiles {
		byName[p.Name] = p
	}

	cases := []struct{ name, wantID string }{
		{"default", "111111111111"},
		{"dev", "222222222222"},
		{"prod", "333333333333"},
		{"staging", ""},
	}
	for _, tc := range cases {
		p, ok := byName[tc.name]
		if !ok {
			t.Errorf("missing profile %q", tc.name)
			continue
		}
		if p.AccountID != tc.wantID {
			t.Errorf("profile %q: got AccountID %q, want %q", tc.name, p.AccountID, tc.wantID)
		}
	}
}

func TestLoadProfileDetails_SSOPreferredOverAccountID(t *testing.T) {
	f, _ := os.CreateTemp("", "aws-config-*")
	f.WriteString("[profile dev]\nsso_account_id = 111111111111\naccount_id = 999999999999\n")
	f.Close()

	profiles, err := aws.LoadProfileDetails(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0].AccountID != "111111111111" {
		t.Errorf("sso_account_id should take precedence, got %q", profiles[0].AccountID)
	}
}
