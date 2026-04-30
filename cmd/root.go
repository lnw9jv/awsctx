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
	regionFlag  string
)

var rootCmd = &cobra.Command{
	Use:   "awsctx [profile]",
	Short: "Switch AWS profiles",
	Long:  "awsctx — switch AWS_PROFILE like kubectx.\nRun 'awsctx shell-init zsh|bash|fish' to set up shell integration.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if unsetFlag {
			fmt.Println("unset AWS_PROFILE")
			fmt.Println("unset AWS_DEFAULT_REGION")
			if regionFlag != "" {
				fmt.Printf("export AWS_DEFAULT_REGION=%s\n", regionFlag)
			}
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
			if regionFlag != "" {
				fmt.Printf("export AWS_DEFAULT_REGION=%s\n", regionFlag)
			}
			return nil
		}
		if len(args) == 0 && regionFlag != "" {
			fmt.Printf("export AWS_DEFAULT_REGION=%s\n", regionFlag)
			return nil
		}
		if len(args) == 1 {
			if err := switchProfile(args[0]); err != nil {
				return err
			}
			if regionFlag != "" {
				fmt.Printf("export AWS_DEFAULT_REGION=%s\n", regionFlag)
			}
			return nil
		}
		profiles, err := awscfg.LoadProfiles(awscfg.ConfigPath())
		if err != nil {
			return err
		}
		selected, err := picker.Pick(profiles, os.Getenv("AWS_PROFILE"))
		if err != nil {
			return err
		}
		if err := switchProfile(selected); err != nil {
			return err
		}
		if regionFlag != "" {
			fmt.Printf("export AWS_DEFAULT_REGION=%s\n", regionFlag)
		}
		return nil
	},
}

func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&unsetFlag, "unset", "u", false, "Unset AWS_PROFILE and AWS_DEFAULT_REGION")
	rootCmd.Flags().BoolVarP(&currentFlag, "current", "c", false, "Print current AWS profile")
	rootCmd.Flags().StringVarP(&regionFlag, "region", "r", "", "Set AWS_DEFAULT_REGION")
	_ = rootCmd.RegisterFlagCompletionFunc("region", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return AWSRegions, cobra.ShellCompDirectiveNoFileComp
	})
}
