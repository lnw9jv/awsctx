package aws_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lnw9jv/awsctx/internal/aws"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadProfiles(t *testing.T) {
	path := writeConfig(t, "[default]\n[profile dev]\n[profile prod]\n")

	profiles, err := aws.LoadProfiles(path)
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
	path := writeConfig(t, "[profile staging]\n[profile prod]\n")

	profiles, err := aws.LoadProfiles(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d: %v", len(profiles), profiles)
	}
}

func TestLoadProfileDetails(t *testing.T) {
	path := writeConfig(t, "[default]\nsso_account_id = 111111111111\n\n[profile dev]\nsso_account_id = 222222222222\n\n[profile prod]\naccount_id = 333333333333\n\n[profile staging]\n")

	profiles, err := aws.LoadProfileDetails(path)
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

func TestLoadProfileDetails_RoleARN(t *testing.T) {
	path := writeConfig(t, "[profile prod]\nrole_arn = arn:aws:iam::123456789012:role/MyRole\nsource_profile = default\n")

	profiles, err := aws.LoadProfileDetails(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0].AccountID != "123456789012" {
		t.Errorf("expected account ID from role_arn, got %q", profiles[0].AccountID)
	}
}

func TestLoadProfileDetails_IgnoresNonProfileSections(t *testing.T) {
	// An [sso-session] block (with its own sso_account_id) sits between two real
	// profiles: it must not become a profile, and the scanner must not carry its
	// keys over to the profile that follows.
	path := writeConfig(t, "[default]\nsso_account_id = 111111111111\n\n"+
		"[sso-session my-sso]\nsso_account_id = 999999999999\nsso_region = us-east-1\n\n"+
		"[profile dev]\nsso_account_id = 222222222222\n")

	profiles, err := aws.LoadProfileDetails(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles (sso-session ignored), got %d: %v", len(profiles), profiles)
	}

	byName := map[string]aws.Profile{}
	for _, p := range profiles {
		byName[p.Name] = p
	}
	if _, ok := byName["my-sso"]; ok {
		t.Error("[sso-session my-sso] must not be returned as a profile")
	}
	if byName["default"].AccountID != "111111111111" {
		t.Errorf("default: got %q, want 111111111111", byName["default"].AccountID)
	}
	if byName["dev"].AccountID != "222222222222" {
		t.Errorf("dev: got %q, want 222222222222 (must not inherit sso-session keys)", byName["dev"].AccountID)
	}
}

func TestLoadProfileDetails_SkipsComments(t *testing.T) {
	path := writeConfig(t, "# a hash comment\n; a semicolon comment\n[profile dev]\n"+
		"# inside the section\nsso_account_id = 222222222222\n; trailing comment\n")

	profiles, err := aws.LoadProfileDetails(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d: %v", len(profiles), profiles)
	}
	if profiles[0].Name != "dev" || profiles[0].AccountID != "222222222222" {
		t.Errorf("got %+v, want {dev 222222222222}", profiles[0])
	}
}

func TestLoadProfileDetails_SSOPreferredOverAccountID(t *testing.T) {
	path := writeConfig(t, "[profile dev]\nsso_account_id = 111111111111\naccount_id = 999999999999\n")

	profiles, err := aws.LoadProfileDetails(path)
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
