package aws

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// accountID resolves a profile's account ID from its keys, in order:
// sso_account_id → account_id → the account segment of role_arn
// (arn:aws:iam::ACCOUNT:role/NAME).
func accountID(keys map[string]string) string {
	for _, k := range []string{"sso_account_id", "account_id"} {
		if v := keys[k]; v != "" {
			return v
		}
	}
	if arn := keys["role_arn"]; arn != "" {
		parts := strings.SplitN(arn, ":", 6)
		if len(parts) >= 5 && parts[4] != "" {
			return parts[4]
		}
	}
	return ""
}

// LoadProfileDetails returns all profiles from the given config file path, including account IDs.
// [default] → Profile{Name:"default"}, [profile foo] → Profile{Name:"foo"}.
// Other sections (e.g. [sso-session foo]) are ignored.
func LoadProfileDetails(configPath string) ([]Profile, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read AWS config %s: %w", configPath, err)
	}
	defer f.Close()

	var profiles []Profile
	var name string             // current profile name, "" when current section isn't a profile
	var keys map[string]string  // keys of the current profile section
	flush := func() {
		if name != "" {
			profiles = append(profiles, Profile{Name: name, AccountID: accountID(keys)})
		}
	}

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' || line[0] == ';' {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			flush()
			header := strings.TrimSpace(line[1 : len(line)-1])
			switch {
			case header == "default":
				name = "default"
			case strings.HasPrefix(header, "profile "):
				name = strings.TrimSpace(strings.TrimPrefix(header, "profile "))
			default:
				name = "" // not a profile section
			}
			keys = map[string]string{}
			continue
		}
		// key = value (inline comments are kept as part of the value, matching prior behavior)
		if name != "" {
			if k, v, ok := strings.Cut(line, "="); ok {
				keys[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("cannot read AWS config %s: %w", configPath, err)
	}
	flush()
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
