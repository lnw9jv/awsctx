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
  set out (command awsctx $argv)
  or return $status
  if string match -q 'export *' $out; or string match -q 'unset *' $out
    eval (echo $out | sed 's/export /set -gx /' | sed 's/unset /set -e /')
  else
    echo $out
  end
end
`

func init() {
	rootCmd.AddCommand(shellInitCmd)
}
