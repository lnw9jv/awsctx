# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build            # build binary (injects version via ldflags)
make test             # go test ./...
make test-integration # go test -tags integration ./...
make install          # build and copy to /usr/local/bin

# Run a single test
go test ./internal/aws/ -run TestLoadProfiles
go test ./internal/state/ -run TestGetPrevious
go test ./cmd/ -run TestRegion

# Run integration tests directly
go test -tags integration ./...
```

There is no lint target; run `go vet ./...` manually.

## Architecture

`awsctx` is a CLI tool built with [Cobra](https://github.com/spf13/cobra). The core design constraint is that **a child process cannot modify its parent shell's environment**. The tool works around this by printing `export AWS_PROFILE=<name>` to stdout; the shell wrapper function (installed via `awsctx shell-init zsh|bash|fish`) `eval`s that output so the variable propagates to the current shell session.

### Package layout

- **`cmd/`** — Cobra command definitions. `root.go` handles the main dispatch logic (flags, `-` for previous, interactive picker, or direct switch). `switch.go` contains `switchProfile()` as a helper function — it is **not** a subcommand. `list-profile.go` is the `list-profile` subcommand (alias `ls`; profiles + account IDs, no env mutation). `shell_init.go` contains the POSIX and Fish shell wrapper snippets as string constants. `completion.go` delegates to Cobra's built-in completion.

- **`internal/aws/`** — Reads `~/.aws/config` (or `$AWS_CONFIG_FILE`) using `gopkg.in/ini.v1`. Exposes two functions: `LoadProfiles` (returns `[]string` names) and `LoadProfileDetails` (returns `[]Profile` with `Name` and `AccountID`). `AccountID` resolves via `sectionAccountID` with three fallbacks in order: `sso_account_id` key → `account_id` key → account segment parsed from `role_arn` (e.g. `arn:aws:iam::123456789012:role/Name`). Special care is taken because `ini.v1` merges `[default]` into a synthetic root section (index 0), so `hasDefaultSection` does a raw line scan to detect it before iterating named sections. Profile names are extracted by stripping the `profile ` prefix from `[profile foo]` sections.

- **`internal/state/`** — Manages the "previous profile" as a plain file at `~/.cache/awsctx/previous` (overridable via `AWSCTX_STATE_DIR`). Only two operations: `SetPrevious` and `GetPrevious`.

- **`internal/picker/`** — `Pick()` embeds fzf in-process via the `github.com/junegunn/fzf/src` library (compiled into the binary, so no external `fzf` install is required). It parses options with `fzf.ParseOptions(true, ["--ansi", "--no-preview"])`, feeds profile names through `opts.Input` (a goroutine, current profile colored green via ANSI) and reads the selection back from `opts.Output`, then calls `fzf.Run(opts)`. Exit codes are mapped: `ExitInterrupt` (130) → "cancelled", `ExitNoMatch` (1) → "no profile selected". **stdout guard**: `guardStdout()` repoints `os.Stdout` to `/dev/tty` (falling back to stderr) for the duration of `Run()` and restores it after — this protects the stdout invariant below, since fzf must never write to the stdout the shell wrapper `eval`s.

### stdout/stderr contract

**Critical invariant**: only `export KEY=VALUE` and `unset KEY` lines go to stdout — the shell wrapper `eval`s everything on stdout verbatim. Errors and informational output must go to stderr (Cobra handles this for returned errors). Adding any non-export `fmt.Println` to a code path that the shell wrapper invokes silently corrupts the shell environment.

### Flag interactions

- `-p <profile>` and a positional arg are interchangeable; `-p` is preferred when tab-completion is needed because Cobra wires profile completion to the flag.
- `-c` reads `$AWS_PROFILE` from the environment and prints it; it does not touch state and writes to stdout (no `export` prefix — not `eval`'d by the wrapper).
- `-r <region>` can be used standalone (no profile arg) to set only `AWS_DEFAULT_REGION` without switching the profile.
- `-u` unsets both `AWS_PROFILE` and `AWS_DEFAULT_REGION`; combining with `-r` re-exports the region after unsetting.
- All code paths that switch or print a profile also check `regionFlag` at the end and emit `export AWS_DEFAULT_REGION=<region>` if set.
- `switchProfile()` validates the profile exists in `~/.aws/config` before emitting the export; an unknown profile returns an error (goes to stderr via Cobra).

### Version injection

The version string is set at build time via `-ldflags "-X main.version=$(VERSION)"` where `VERSION` comes from `git describe`. The `version` var in `main.go` defaults to `"dev"` and is passed into `cmd.Execute(version)`.

### Integration tests

Integration tests live in `integration_test.go` at the root and use the `//go:build integration` tag. They are excluded from the default `go test ./...` run and require `-tags integration`. Each test builds the binary into a temp dir and controls config/state via env vars: `AWS_CONFIG_FILE` points to a temp config file, `AWSCTX_STATE_DIR` points to a temp dir — both must be set together to fully isolate a test.

### `list-profile` output format

`awsctx list-profile` writes to stdout (not eval'd by the shell wrapper). Current profile is marked with `*`; profiles without a resolvable account ID show `-` in that column.

## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

Rules:
- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).
