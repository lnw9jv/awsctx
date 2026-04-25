package cmd

import (
	"fmt"
	"os"

	awscfg "github.com/lnw9jv/awsctx/internal/aws"
	"github.com/lnw9jv/awsctx/internal/state"
)

func switchProfile(profile string) error {
	st := state.New(state.DefaultDir())
	profiles, err := awscfg.LoadProfiles(awscfg.ConfigPath())
	if err != nil {
		return err
	}
	found := false
	for _, p := range profiles {
		if p == profile {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("profile %q not found in %s", profile, awscfg.ConfigPath())
	}

	current := os.Getenv("AWS_PROFILE")
	if current != "" && current != profile {
		_ = st.SetPrevious(current)
	}

	fmt.Printf("export AWS_PROFILE=%s\n", profile)
	return nil
}
