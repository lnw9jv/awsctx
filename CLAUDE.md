# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build            # build binary (injects version via ldflags)
make test             # run unit tests (go test ./...)
make test-integration # run unit + integration tests (-tags integration)
make install          # build and copy to /usr/local/bin

# Run a single test
go test ./internal/aws/ -run TestLoadProfiles
go test ./internal/state/ -run TestGetPrevious

# Run integration tests directly
go test -tags integration ./...
```

## Architecture

`awsctx` is a CLI tool built with [Cobra](https://github.com/spf13/cobra). The core design constraint is that **a child process cannot modify its parent shell's environment**. The tool works around this by printing `export AWS_PROFILE=<name>` to stdout; the shell wrapper function (installed via `awsctx shell-init zsh|bash|fish`) `eval`s that output so the variable propagates to the current shell session.

### Package layout

- **`cmd/`** — Cobra command definitions. `root.go` handles the main dispatch logic (flags, `-` for previous, interactive picker, or direct switch). `shell_init.go` contains the POSIX and Fish shell wrapper snippets as string constants. `completion.go` delegates to Cobra's built-in completion.

- **`internal/aws/`** — Reads `~/.aws/config` (or `$AWS_CONFIG_FILE`) using `gopkg.in/ini.v1`. Special care is taken because `ini.v1` merges `[default]` into a synthetic root section (index 0), so `hasDefaultSection` does a raw line scan to detect it before iterating named sections. Profile names are extracted by stripping the `profile ` prefix from `[profile foo]` sections.

- **`internal/state/`** — Manages the "previous profile" as a plain file at `~/.cache/awsctx/previous` (overridable via `AWSCTX_STATE_DIR`). Only two operations: `SetPrevious` and `GetPrevious`.

- **`internal/picker/`** — A from-scratch interactive TUI that opens `/dev/tty` directly, puts the terminal in raw mode, and renders a fuzzy-filtered list with arrow-key navigation. Does not use any TUI library — just ANSI escape codes and `golang.org/x/term`.

### Version injection

The version string is set at build time via `-ldflags "-X main.version=$(VERSION)"` where `VERSION` comes from `git describe`. The `version` var in `main.go` defaults to `"dev"` and is passed into `cmd.Execute(version)`.

### Integration tests

Integration tests live in `integration_test.go` at the root and use the `//go:build integration` tag. They are excluded from the default `go test ./...` run and require `-tags integration`.
