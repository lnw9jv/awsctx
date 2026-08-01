package cmd

import (
	"fmt"
	"os"
	"strings"

	awscfg "github.com/lnw9jv/awsctx/internal/aws"
	"github.com/lnw9jv/awsctx/internal/picker"
	"github.com/lnw9jv/awsctx/internal/state"
	"github.com/spf13/cobra"
)

var (
	unsetFlag   bool
	currentFlag bool
	regionFlag  string
	profileFlag string
)

var rootCmd = &cobra.Command{
	Use:   "awsctx [profile]",
	Short: "Switch AWS profiles",
	Long:  "awsctx — switch AWS_PROFILE like kubectx.\nRun 'awsctx shell-init zsh|bash|fish' to set up shell integration.",
	Args:  validateRootArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		stdout := cmd.OutOrStdout()
		if unsetFlag {
			if err := writeOutput(stdout, "unset AWS_PROFILE\nunset AWS_DEFAULT_REGION\n"); err != nil {
				return err
			}
			return writeOutput(cmd.ErrOrStderr(), statusMessage("", "", true))
		}
		if currentFlag {
			profile := os.Getenv("AWS_PROFILE")
			if profile == "" {
				profile = "none"
			}
			return writeOutput(stdout, profile+"\n")
		}
		target := profileFlag
		if len(args) == 1 {
			target = args[0]
		}
		if target == "-" {
			prev, err := state.GetPrevious(state.DefaultDir())
			if err != nil {
				return err
			}
			target = prev // falls through to switchProfile, which validates it still exists
		}
		if target == "" && regionFlag != "" {
			if err := writeOutput(stdout, fmt.Sprintf("export AWS_DEFAULT_REGION=%s\n", regionFlag)); err != nil {
				return err
			}
			return writeOutput(cmd.ErrOrStderr(), statusMessage("", regionFlag, false))
		}
		if target != "" {
			profileOutput, err := profileExport(target, nil)
			if err != nil {
				return err
			}
			var output strings.Builder
			output.WriteString(profileOutput)
			if regionFlag != "" {
				fmt.Fprintf(&output, "export AWS_DEFAULT_REGION=%s\n", regionFlag)
			}
			if err := writeOutput(stdout, output.String()); err != nil {
				return err
			}
			savePreviousProfile(target, cmd.ErrOrStderr())
			return writeOutput(cmd.ErrOrStderr(), statusMessage(target, regionFlag, false))
		}
		profiles, err := awscfg.LoadProfiles(awscfg.ConfigPath())
		if err != nil {
			return err
		}
		selected, err := picker.Pick(profiles, os.Getenv("AWS_PROFILE"))
		if err != nil {
			return err
		}
		profileOutput, err := profileExport(selected, profiles)
		if err != nil {
			return err
		}
		var output strings.Builder
		output.WriteString(profileOutput)
		if regionFlag != "" {
			fmt.Fprintf(&output, "export AWS_DEFAULT_REGION=%s\n", regionFlag)
		}
		if err := writeOutput(stdout, output.String()); err != nil {
			return err
		}
		savePreviousProfile(selected, cmd.ErrOrStderr())
		return writeOutput(cmd.ErrOrStderr(), statusMessage(selected, regionFlag, false))
	},
}

func validateRootArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
		return err
	}
	if len(args) == 1 && args[0] == "" {
		return fmt.Errorf("profile cannot be empty")
	}

	hasProfile := cmd.Flags().Changed("profile")
	hasCurrent := currentFlag
	hasUnset := unsetFlag
	hasRegion := cmd.Flags().Changed("region")
	hasPositional := len(args) == 1

	if hasProfile && profileFlag == "" {
		return fmt.Errorf("--profile cannot be empty")
	}
	if hasRegion && regionFlag == "" {
		return fmt.Errorf("--region cannot be empty")
	}
	if hasRegion {
		if err := validateExportValue("region", regionFlag); err != nil {
			return err
		}
	}
	if hasPositional && hasProfile {
		return fmt.Errorf("profile cannot be set by both argument and --profile")
	}
	if hasCurrent && (hasPositional || hasProfile || hasUnset || hasRegion) {
		return fmt.Errorf("--current cannot be combined with profile, --profile, --unset, or --region")
	}
	if hasUnset && (hasPositional || hasProfile || hasRegion) {
		return fmt.Errorf("--unset cannot be combined with profile, --profile, or --region")
	}
	return nil
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
	rootCmd.Flags().StringVarP(&profileFlag, "profile", "p", "", "AWS profile to switch to")
	_ = rootCmd.RegisterFlagCompletionFunc("region", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return AWSRegions, cobra.ShellCompDirectiveNoFileComp
	})
	_ = rootCmd.RegisterFlagCompletionFunc("profile", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		profiles, err := awscfg.LoadProfiles(awscfg.ConfigPath())
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		completions := []string{"-\tSwitch to previous profile"}
		for _, p := range profiles {
			completions = append(completions, p+"\tAWS Profile")
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	})
}
