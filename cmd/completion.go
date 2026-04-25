package cmd

import (
	awscfg "github.com/lnw9jv/awsctx/internal/aws"
	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:       "completion [zsh|bash|fish]",
	Short:     "Generate shell completion script",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"zsh", "bash", "fish"},
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "zsh":
			return rootCmd.GenZshCompletion(cmd.OutOrStdout())
		case "bash":
			return rootCmd.GenBashCompletion(cmd.OutOrStdout())
		case "fish":
			return rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)

	rootCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		profiles, err := awscfg.LoadProfiles(awscfg.ConfigPath())
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		completions := []string{"-\tSwitch to previous profile"}
		for _, p := range profiles {
			completions = append(completions, p+"\tAWS Profile")
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}
