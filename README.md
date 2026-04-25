# awsctx

Switch AWS profiles like [kubectx](https://github.com/ahmetb/kubectx) — fast, fuzzy, no dependencies.

```
awsctx          # interactive picker
awsctx dev      # switch to profile "dev"
awsctx -        # switch to previous profile
awsctx -c       # show current profile
awsctx -u       # unset AWS_PROFILE
```

## Install

### Build from source

```bash
git clone https://github.com/lnw9jv/awsctx
cd awsctx
make build
sudo make install   # copies binary to /usr/local/bin
```

### Shell integration (required)

Without the shell wrapper, `awsctx` prints the export statement but cannot set the variable in your shell.

**zsh / bash** — add to `~/.zshrc` or `~/.bashrc`:

```bash
eval "$(awsctx shell-init zsh)"   # or bash
```

**fish** — add to `~/.config/fish/config.fish`:

```fish
awsctx shell-init fish | source
```

## Usage

| Command | Description |
|---|---|
| `awsctx` | Open interactive fuzzy picker |
| `awsctx <profile>` | Switch to named profile |
| `awsctx -` | Switch to previous profile |
| `awsctx -c` | Print current profile |
| `awsctx -u` | Unset `AWS_PROFILE` |
| `awsctx shell-init zsh\|bash\|fish` | Print shell integration snippet |
| `awsctx completion zsh\|bash\|fish` | Print completion script |
| `awsctx --version` | Print version |

### Shell completion

```bash
# zsh
awsctx completion zsh > "${fpath[1]}/_awsctx"

# bash
awsctx completion bash > /etc/bash_completion.d/awsctx

# fish
awsctx completion fish > ~/.config/fish/completions/awsctx.fish
```

## How it works

`awsctx` reads profiles from `~/.aws/config` (or `$AWS_CONFIG_FILE`). When you switch, it prints `export AWS_PROFILE=<name>` to stdout — the shell wrapper `eval`s that output so the variable propagates to your current shell session. The previous profile is saved to `~/.cache/awsctx/previous`.

## Requirements

- Go 1.22+ (to build)
- `~/.aws/config` with `[profile <name>]` sections
