package aws

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

// Profile holds a profile name and its AWS account ID (empty when absent).
type Profile struct {
	Name      string
	AccountID string
}

func ConfigPath() string {
	if p := os.Getenv("AWS_CONFIG_FILE"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, ".aws", "config")
}

// hasDefaultSection reports whether the config file contains an explicit [default] section.
// ini.v1 stores [default] as a named section "default" (not the synthetic root at index 0),
// but does not include it in sections iterable via CutPrefix("profile "), so we detect it separately.
func hasDefaultSection(configPath string) bool {
	f, err := os.Open(configPath)
	if err != nil {
		return false
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.EqualFold(line, "[default]") {
			return true
		}
	}
	_ = sc.Err() // scanner errors are treated as "not found"
	return false
}

func sectionAccountID(sec *ini.Section) string {
	for _, key := range []string{"sso_account_id", "account_id"} {
		if v := sec.Key(key).String(); v != "" {
			return v
		}
	}
	return ""
}

// LoadProfileDetails returns all profiles from the given config file path, including account IDs.
// [default] → Profile{Name:"default"}, [profile foo] → Profile{Name:"foo"}.
func LoadProfileDetails(configPath string) ([]Profile, error) {
	cfg, err := ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
	}, configPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read AWS config %s: %w", configPath, err)
	}

	var profiles []Profile
	if hasDefaultSection(configPath) {
		// ini.v1 stores [default] as a named section "default" (not the synthetic root).
		profiles = append(profiles, Profile{
			Name:      "default",
			AccountID: sectionAccountID(cfg.Section("default")),
		})
	}

	for i, s := range cfg.Sections() {
		if i == 0 {
			continue // always skip synthetic root
		}
		name := s.Name()
		if after, ok := strings.CutPrefix(name, "profile "); ok {
			profiles = append(profiles, Profile{
				Name:      after,
				AccountID: sectionAccountID(s),
			})
		}
	}
	return profiles, nil
}

// LoadProfiles returns all profile names from the given config file path.
// [default] → "default", [profile foo] → "foo".
func LoadProfiles(configPath string) ([]string, error) {
	profiles, err := LoadProfileDetails(configPath)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(profiles))
	for i, p := range profiles {
		names[i] = p.Name
	}
	return names, nil
}
