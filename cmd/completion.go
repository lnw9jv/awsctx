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

var AWSRegions = []string{
	"af-south-1\tAfrica (Cape Town)",
	"ap-east-1\tAsia Pacific (Hong Kong)",
	"ap-northeast-1\tAsia Pacific (Tokyo)",
	"ap-northeast-2\tAsia Pacific (Seoul)",
	"ap-northeast-3\tAsia Pacific (Osaka)",
	"ap-south-1\tAsia Pacific (Mumbai)",
	"ap-south-2\tAsia Pacific (Hyderabad)",
	"ap-southeast-1\tAsia Pacific (Singapore)",
	"ap-southeast-2\tAsia Pacific (Sydney)",
	"ap-southeast-3\tAsia Pacific (Jakarta)",
	"ap-southeast-4\tAsia Pacific (Melbourne)",
	"ap-southeast-5\tAsia Pacific (Malaysia)",
	"ap-southeast-7\tAsia Pacific (Thailand)",
	"ca-central-1\tCanada (Central)",
	"ca-west-1\tCanada (Calgary)",
	"cn-north-1\tChina (Beijing)",
	"cn-northwest-1\tChina (Ningxia)",
	"eu-central-1\tEurope (Frankfurt)",
	"eu-central-2\tEurope (Zurich)",
	"eu-north-1\tEurope (Stockholm)",
	"eu-south-1\tEurope (Milan)",
	"eu-south-2\tEurope (Spain)",
	"eu-west-1\tEurope (Ireland)",
	"eu-west-2\tEurope (London)",
	"eu-west-3\tEurope (Paris)",
	"il-central-1\tIsrael (Tel Aviv)",
	"me-central-1\tMiddle East (UAE)",
	"me-south-1\tMiddle East (Bahrain)",
	"mx-central-1\tMexico (Central)",
	"sa-east-1\tSouth America (São Paulo)",
	"us-east-1\tUS East (N. Virginia)",
	"us-east-2\tUS East (Ohio)",
	"us-gov-east-1\tAWS GovCloud (US-East)",
	"us-gov-west-1\tAWS GovCloud (US-West)",
	"us-west-1\tUS West (N. California)",
	"us-west-2\tUS West (Oregon)",
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
