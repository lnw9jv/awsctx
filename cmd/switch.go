package cmd

import (
	"fmt"
	"io"
	"os"
	"slices"

	awscfg "github.com/lnw9jv/awsctx/internal/aws"
	"github.com/lnw9jv/awsctx/internal/state"
)

// switchProfile validates that profile exists and emits the export line.
// Pass a non-nil profiles to reuse an already-loaded list (e.g. from the picker);
// pass nil to load from the AWS config.
func switchProfile(profile string, profiles []string, stdout, stderr io.Writer) error {
	output, err := profileExport(profile, profiles)
	if err != nil {
		return err
	}
	if err := writeOutput(stdout, output); err != nil {
		return err
	}
	savePreviousProfile(profile, stderr)
	return nil
}

func profileExport(profile string, profiles []string) (string, error) {
	if err := validateExportValue("profile", profile); err != nil {
		return "", err
	}
	if profiles == nil {
		var err error
		if profiles, err = awscfg.LoadProfiles(awscfg.ConfigPath()); err != nil {
			return "", err
		}
	}
	if !slices.Contains(profiles, profile) {
		return "", fmt.Errorf("profile %q not found in %s", profile, awscfg.ConfigPath())
	}
	return fmt.Sprintf("export AWS_PROFILE=%s\n", profile), nil
}

func savePreviousProfile(profile string, stderr io.Writer) {
	current := os.Getenv("AWS_PROFILE")
	if current != "" && current != profile {
		if err := state.SetPrevious(state.DefaultDir(), current); err != nil {
			fmt.Fprintf(stderr, "warning: cannot save previous profile: %v\n", err)
		}
	}
}
