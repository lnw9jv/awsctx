# Repository Guidelines

## Project Structure & Module Organization

`main.go` is the executable entry point. Cobra commands and shell integration live in `cmd/`. Core packages are under `internal/`: `aws/` parses AWS configuration, `state/` stores the previous profile, and `picker/` wraps the embedded fzf UI. Unit tests sit beside their packages as `*_test.go`. End-to-end CLI coverage lives in the root `integration_test.go` behind the `integration` build tag. Shell snippets are Go string constants in `cmd/shell_init.go`.

## Build, Test, and Development Commands

- `make build` builds `./awsctx` and injects a version from `git describe`.
- `make test` runs all default unit tests with `go test ./...`.
- `make test-integration` builds and exercises the CLI with `-tags integration`.
- `go test ./cmd -run TestRegion` runs one focused test.
- `go vet ./...` performs the manual static-analysis check; there is no configured lint target.
- `make install` moves the built binary to `/usr/local/bin`; use it only when a system install is intended.

Go 1.26.2 or newer is required by `go.mod`.

## Coding Style & Naming Conventions

Run `gofmt` on every changed Go file; use Go's tab-based indentation and standard import grouping. Follow idiomatic names: exported identifiers use `PascalCase`, internal functions and variables use `camelCase`, and files use short lowercase names such as `shell_init.go`. Keep Cobra command definitions in `cmd/`, wrap errors with context using `%w`, and avoid adding dependencies for behavior the standard library handles clearly.

## Testing Guidelines

Use Go's `testing` package and name tests `TestBehavior`. Add package-level unit tests next to changed code. Put full-process behavior in `integration_test.go` with `//go:build integration`. Integration tests must isolate filesystem state by setting both `AWS_CONFIG_FILE` and `AWSCTX_STATE_DIR` to temporary locations. No coverage threshold is configured, but every behavior change should include regression coverage.

## Shell Output & Configuration Safety

Shell wrappers evaluate stdout. On switching paths, stdout must contain only `export KEY=VALUE` or `unset KEY` lines; send errors and informational messages to stderr. Never commit AWS credentials, real account configuration, or generated state files.

## Commit & Pull Request Guidelines

Recent commits use concise, imperative subjects such as `Fix fish shell wrapper...`, `Add awsctx list...`, and `Replace INI library...`. Keep commits focused and explain non-obvious tradeoffs in the body. Pull requests should summarize user-visible behavior, list tests run, link relevant issues, and include before/after terminal output when shell commands or formatting change. Screenshots are generally unnecessary for this CLI.
