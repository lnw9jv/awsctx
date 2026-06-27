package cmd

import (
	"fmt"
	"os"
	"slices"

	awscfg "github.com/lnw9jv/awsctx/internal/aws"
	"github.com/lnw9jv/awsctx/internal/state"
)

// switchProfile validates that profile exists and emits the export line.
// Pass a non-nil profiles to reuse an already-loaded list (e.g. from the picker);
// pass nil to load from the AWS config.
func switchProfile(profile string, profiles []string) error {
	if profiles == nil {
		var err error
		if profiles, err = awscfg.LoadProfiles(awscfg.ConfigPath()); err != nil {
			return err
		}
	}
	if !slices.Contains(profiles, profile) {
		return fmt.Errorf("profile %q not found in %s", profile, awscfg.ConfigPath())
	}

	current := os.Getenv("AWS_PROFILE")
	if current != "" && current != profile {
		_ = state.SetPrevious(state.DefaultDir(), current)
	}

	fmt.Printf("export AWS_PROFILE=%s\n", profile)
	return nil
}
