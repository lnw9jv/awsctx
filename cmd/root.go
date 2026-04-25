package cmd

import (
	"fmt"
	"os"

	awscfg "github.com/lnw9jv/awsctx/internal/aws"
	"github.com/lnw9jv/awsctx/internal/picker"
	"github.com/lnw9jv/awsctx/internal/state"
	"github.com/spf13/cobra"
)

var (
	unsetFlag   bool
	currentFlag bool
)

var rootCmd = &cobra.Command{
	Use:   "awsctx [profile]",
	Short: "Switch AWS profiles",
	Long:  "awsctx — switch AWS_PROFILE like kubectx.\nRun 'awsctx shell-init zsh|bash|fish' to set up shell integration.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if unsetFlag {
			fmt.Println("unset AWS_PROFILE")
			return nil
		}
		if currentFlag {
			profile := os.Getenv("AWS_PROFILE")
			if profile == "" {
				profile = "none"
			}
			fmt.Println(profile)
			return nil
		}
		if len(args) == 1 && args[0] == "-" {
			st := state.New(state.DefaultDir())
			prev, err := st.GetPrevious()
			if err != nil {
				return err
			}
			if current := os.Getenv("AWS_PROFILE"); current != "" {
				_ = st.SetPrevious(current)
			}
			fmt.Printf("export AWS_PROFILE=%s\n", prev)
			return nil
		}
		if len(args) == 1 {
			return switchProfile(args[0])
		}
		profiles, err := awscfg.LoadProfiles(awscfg.ConfigPath())
		if err != nil {
			return err
		}
		selected, err := picker.Pick(profiles, os.Getenv("AWS_PROFILE"))
		if err != nil {
			return err
		}
		return switchProfile(selected)
	},
}

func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&unsetFlag, "unset", "u", false, "Unset AWS_PROFILE")
	rootCmd.Flags().BoolVarP(&currentFlag, "current", "c", false, "Print current AWS profile")
}
