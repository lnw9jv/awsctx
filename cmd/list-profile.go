package cmd

import (
	"fmt"
	"io"
	"os"

	awscfg "github.com/lnw9jv/awsctx/internal/aws"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list-profile",
	Aliases: []string{"ls"},
	Short:   "List all AWS profiles with account IDs",
	Args:    cobra.NoArgs,
	RunE:    runList,
}

func runList(cmd *cobra.Command, args []string) error {
	profiles, err := awscfg.LoadProfileDetails(awscfg.ConfigPath())
	if err != nil {
		return err
	}
	if len(profiles) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No profiles found.")
		return nil
	}
	printProfileTable(cmd.OutOrStdout(), profiles, os.Getenv("AWS_PROFILE"))
	return nil
}

// printProfileTable writes the profile list to w.
// Every data row is prefixed with "  " or "* " — this leading whitespace is load-bearing:
// the shell wrapper evals any stdout line starting with "export"/"unset", so rows must
// never start with those words even if a profile name contains them.
func printProfileTable(w io.Writer, profiles []awscfg.Profile, current string) {
	maxLen := len("NAME")
	for _, p := range profiles {
		if len(p.Name) > maxLen {
			maxLen = len(p.Name)
		}
	}

	fmt.Fprintf(w, "  %-*s  %s\n", maxLen, "NAME", "ACCOUNT ID")
	for _, p := range profiles {
		marker := "  "
		if p.Name == current {
			marker = "* "
		}
		accountID := p.AccountID
		if accountID == "" {
			accountID = "-"
		}
		fmt.Fprintf(w, "%s%-*s  %s\n", marker, maxLen, p.Name, accountID)
	}
}

func init() {
	rootCmd.AddCommand(listCmd)
}
