package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

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
		return writeOutput(cmd.OutOrStdout(), "No profiles found.\n")
	}
	return printProfileTable(cmd.OutOrStdout(), profiles, os.Getenv("AWS_PROFILE"))
}

// printProfileTable writes the profile list to w.
// Every data row is prefixed with "  " or "* " — this leading whitespace is load-bearing:
// shell wrappers interpret allowlisted operations at the start of stdout lines, so rows
// must never start with those operations even if a profile name resembles one.
func printProfileTable(w io.Writer, profiles []awscfg.Profile, current string) error {
	maxLen := len("NAME")
	for _, p := range profiles {
		if len(p.Name) > maxLen {
			maxLen = len(p.Name)
		}
	}

	var output strings.Builder
	fmt.Fprintf(&output, "  %-*s  %s\n", maxLen, "NAME", "ACCOUNT ID")
	for _, p := range profiles {
		marker := "  "
		if p.Name == current {
			marker = "* "
		}
		accountID := p.AccountID
		if accountID == "" {
			accountID = "-"
		}
		fmt.Fprintf(&output, "%s%-*s  %s\n", marker, maxLen, p.Name, accountID)
	}
	return writeOutput(w, output.String())
}

func init() {
	rootCmd.AddCommand(listCmd)
}
