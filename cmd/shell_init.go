package cmd

import (
	"fmt"
	"io"

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
			_, err := io.WriteString(cmd.OutOrStdout(), ShellWrapperPosix)
			return err
		case "fish":
			_, err := io.WriteString(cmd.OutOrStdout(), ShellWrapperFish)
			return err
		default:
			return fmt.Errorf("unsupported shell %q — use zsh, bash, or fish", args[0])
		}
	},
}

const ShellWrapperPosix = `
# awsctx shell integration
awsctx() {
  local out line arg passthrough
  case "${1-}" in
    completion|shell-init|list-profile|ls|help|__complete|__completeNoDesc)
      command awsctx "$@"
      return $?
      ;;
  esac
  passthrough=0
  for arg in "$@"; do
    case "$arg" in
      --)
        break
        ;;
      # Keep pflag/strconv.ParseBool's false spellings in switching mode.
      -c=false|--current=false|-c=False|--current=False|-c=FALSE|--current=FALSE|-c=f|--current=f|-c=F|--current=F|-c=0|--current=0)
        passthrough=0
        ;;
      -c|--current|-c=*|--current=*)
        passthrough=1
        ;;
      -h|--help|-v|--version)
        command awsctx "$@"
        return $?
        ;;
    esac
  done
  if [ "$passthrough" -eq 1 ]; then
    command awsctx "$@"
    return $?
  fi
  out=$(command awsctx "$@") || return $?
  while IFS= read -r line; do
    case "$line" in
      'export AWS_PROFILE='*)
        export AWS_PROFILE="${line#export AWS_PROFILE=}"
        ;;
      'export AWS_DEFAULT_REGION='*)
        export AWS_DEFAULT_REGION="${line#export AWS_DEFAULT_REGION=}"
        ;;
      'unset AWS_PROFILE')
        unset AWS_PROFILE
        ;;
      'unset AWS_DEFAULT_REGION')
        unset AWS_DEFAULT_REGION
        ;;
      *)
        printf '%s\n' "$line"
        ;;
    esac
  done <<< "$out"
}
`

const ShellWrapperFish = `
# awsctx shell integration
function awsctx
  set -l passthrough 0
  switch $argv[1]
    case completion shell-init list-profile ls help __complete __completeNoDesc
      set passthrough 1
  end
  for arg in $argv
    switch $arg
      case --
        break
      # Keep pflag/strconv.ParseBool's false spellings in switching mode.
      case -c=false --current=false -c=False --current=False -c=FALSE --current=FALSE -c=f --current=f -c=F --current=F -c=0 --current=0
        set passthrough 0
      case -c --current '-c=*' '--current=*'
        set passthrough 1
      case -h --help -v --version
        command awsctx $argv
        return $status
    end
  end
  if test $passthrough -eq 1
    command awsctx $argv
    return $status
  end

  set -l out (command awsctx $argv)
  or return $status
  for line in $out
    switch $line
      case 'export AWS_PROFILE=*'
        set -gx AWS_PROFILE (string replace 'export AWS_PROFILE=' '' -- $line)
      case 'export AWS_DEFAULT_REGION=*'
        set -gx AWS_DEFAULT_REGION (string replace 'export AWS_DEFAULT_REGION=' '' -- $line)
      case 'unset AWS_PROFILE'
        set -e AWS_PROFILE
      case 'unset AWS_DEFAULT_REGION'
        set -e AWS_DEFAULT_REGION
      case '*'
        printf '%s\n' "$line"
    end
  end
end
`

func init() {
	rootCmd.AddCommand(shellInitCmd)
}
