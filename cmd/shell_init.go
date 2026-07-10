package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var shellInitCmd = &cobra.Command{
	Use:       "shell-init [zsh|bash|fish]",
	Short:     "Print shell integration snippet",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"zsh", "bash", "fish"},
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "zsh", "bash":
			fmt.Print(ShellWrapperPosix)
		case "fish":
			fmt.Print(ShellWrapperFish)
		default:
			return fmt.Errorf("unsupported shell %q — use zsh, bash, or fish", args[0])
		}
		return nil
	},
}

const ShellWrapperPosix = `
# awsctx shell integration
awsctx() {
  local out
  out=$(command awsctx "$@") || return $?
  if [[ "$out" == export* || "$out" == unset* ]]; then
    eval "$out"
  else
    echo "$out"
  fi
}
`

const ShellWrapperFish = `
# awsctx shell integration
function awsctx
  set -l out (command awsctx $argv)
  or return $status
  for line in $out
    switch $line
      case 'export *'
        set -l kv (string split -m1 '=' -- (string replace -r '^export ' '' -- $line))
        set -gx $kv[1] $kv[2]
      case 'unset *'
        set -e (string replace -r '^unset ' '' -- $line)
      case '*'
        echo $line
    end
  end
end
`

func init() {
	rootCmd.AddCommand(shellInitCmd)
}
