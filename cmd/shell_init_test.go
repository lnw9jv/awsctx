package cmd_test

import (
	"strings"
	"testing"

	"github.com/lnw9jv/awsctx/cmd"
)

func TestShellInitZsh(t *testing.T) {
	if !strings.Contains(cmd.ShellWrapperPosix, "command awsctx") {
		t.Error("zsh wrapper missing 'command awsctx'")
	}
	if !strings.Contains(cmd.ShellWrapperPosix, "eval") {
		t.Error("zsh wrapper missing eval")
	}
}

func TestShellInitFish(t *testing.T) {
	if !strings.Contains(cmd.ShellWrapperFish, "command awsctx") {
		t.Error("fish wrapper missing 'command awsctx'")
	}
	if !strings.Contains(cmd.ShellWrapperFish, "set -gx") {
		t.Error("fish wrapper missing 'set -gx' for export translation")
	}
}
