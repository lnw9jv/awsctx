package aws

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gopkg.in/ini.v1"
)

func ConfigPath() string {
	if p := os.Getenv("AWS_CONFIG_FILE"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return home + "/.aws/config"
}

// hasDefaultSection reports whether the config file contains an explicit [default] section.
// ini.v1 merges [default] into its synthetic root section (index 0), so we detect it separately.
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
	return false
}

// LoadProfiles returns all profile names from the given config file path.
// [default] → "default", [profile foo] → "foo".
func LoadProfiles(configPath string) ([]string, error) {
	cfg, err := ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
	}, configPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read AWS config %s: %w", configPath, err)
	}

	var profiles []string
	if hasDefaultSection(configPath) {
		profiles = append(profiles, "default")
	}

	for i, s := range cfg.Sections() {
		if i == 0 {
			continue // always skip synthetic root
		}
		name := s.Name()
		if after, ok := strings.CutPrefix(name, "profile "); ok {
			profiles = append(profiles, after)
		}
	}
	return profiles, nil
}
