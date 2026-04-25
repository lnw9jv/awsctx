# awsctx Design — 2026-04-24

## Scope

AWS profile switcher CLI in Go, inspired by kubectx.

### In scope
- List and fuzzy-pick AWS profiles from `~/.aws/config`
- Switch profile via `export AWS_PROFILE=<name>` (shell function wrapper)
- Switch to previous profile (`awsctx -`)
- Unset current profile (`awsctx -u`)
- Show current profile (`awsctx -c`)
- Shell integration init (`awsctx shell-init zsh|bash|fish`)
- Shell autocomplete (`awsctx completion zsh|bash|fish`)
- Version flag (`awsctx --version`)
- Help (cobra built-in)

### Non-goals
- Account ID / alias display (no STS calls)
- fzf binary dependency (built-in pure Go fuzzy picker)
- Modifying `~/.aws/credentials` or `~/.aws/config`
- Multiple simultaneous profiles

## Architecture

```
awsctx/
├── main.go
├── cmd/
│   ├── root.go        # cobra root; fzf picker when no args
│   ├── current.go     # -c flag
│   ├── unset.go       # -u flag
│   ├── previous.go    # - argument
│   ├── shell_init.go  # shell-init subcommand
│   └── completion.go  # completion subcommand
├── internal/
│   ├── aws/
│   │   └── config.go  # parse ~/.aws/config profiles, respect AWS_CONFIG_FILE
│   └── state/
│       └── state.go   # read/write ~/.cache/awsctx/previous
└── go.mod
```

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/ktr0731/go-fzf` | Built-in fuzzy picker (pure Go) |
| `gopkg.in/ini.v1` | Parse AWS INI config |

## Command Table

| Command | Stdout | Shell effect |
|---|---|---|
| `awsctx` | (fzf UI) → `export AWS_PROFILE=<x>` | Sets env var |
| `awsctx <profile>` | `export AWS_PROFILE=<x>` | Sets env var |
| `awsctx -` | `export AWS_PROFILE=<prev>` | Sets to previous |
| `awsctx -u` | `unset AWS_PROFILE` | Unsets env var |
| `awsctx -c` | current profile name | Info only |
| `awsctx shell-init zsh\|bash\|fish` | shell wrapper snippet | Setup only |
| `awsctx completion zsh\|bash\|fish` | completion script | Setup only |
| `awsctx --version` | version string | Info only |
| `awsctx --help` | usage (cobra) | Info only |

## Shell Wrapper (zsh example)

```zsh
awsctx() {
  local out
  out=$(command awsctx "$@") || return $?
  if [[ "$out" == export* || "$out" == unset* ]]; then
    eval "$out"
  else
    echo "$out"
  fi
}
```

Same pattern for bash. Fish uses `eval (command awsctx $argv)` with fish-specific guards.

## State

- `~/.cache/awsctx/previous` — plain text, one line, previous profile name
- Written on every successful switch
- Read by `awsctx -`

## Error Handling

| Scenario | Behaviour |
|---|---|
| `~/.aws/config` missing | Error: hint to create config |
| Profile not found | Error before any state change |
| `awsctx -` with no history | Error: "no previous profile" |
| `AWS_CONFIG_FILE` set | Use that path instead |

## Build & Version

```bash
go build -ldflags "-X main.version=$(git describe --tags)" .
```

## Testing Strategy

- Unit test: AWS config parser (various profile name formats)
- Unit test: state read/write
- Unit test: shell-init output per shell
- Integration test: full switch flow with temp config file
