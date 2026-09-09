# linter-autoconfigure-sdk — Project Context

Enduring context for AI sessions working in this repo. Domain-facing docs live in
the README; this file captures what is hard to discover from the code alone.

## What this is

Shared foundation (Go library) for linter auto-configuration tools. Owns the
plumbing that `golangci-lint-auto-configure` and `oxlint-auto-configure` duplicate:
config file round-trip, finding emission for config issues, and a BuildFlow
provider spec. YAML parsing is intentionally NOT here (each tool uses a
different YAML library). Early-stage; no active consumers yet. Public repo
(github.com/LarsArtmann/linter-autoconfigure-sdk), MIT-licensed.

Module: `github.com/larsartmann/linter-autoconfigure-sdk`. Requires Go 1.26+,
`github.com/larsartmann/go-finding` (always latest), and
`github.com/larsartmann/go-atomic-write` (always latest — used by `SaveJSON`
for idempotent, crash-durable writes).

**Always stay on the latest go-finding version.** go-finding v1.2+ imports
`encoding/json/v2` and `encoding/json/jsontext`, which are gated behind the
`//go:build goexperiment.jsonv2` build tag in Go 1.26.x. This is NOT a
permanent incompatibility — it just requires `GOEXPERIMENT=jsonv2` to be set in
the environment. The repo's `.envrc` handles this automatically via direnv
(`direnv allow` on first clone). See "GOEXPERIMENT=jsonv2" below for details.

## Build, test, lint

This project is automated with **BuildFlow** (no Makefile, no flake.nix):

- `buildflow` — full quality pipeline (detect mode). Exits 0 when healthy.
- `buildflow --fix` — detect + auto-fix. Exits 0 when healthy.
- `buildflow --fix --fail-on-findings` — strict; exits non-zero if ANY finding
  remains, including pre-existing environmental warnings (currently the
  errcheck + go-structure-linter findings tracked in TODO_LIST).
- `go test -race -count=1 ./...` — just the Go tests (covered by buildflow
  `test-race` and `test-coverage` steps).

Single step: `buildflow -s <step> -v`. Disable result cache during debugging:
`BUILDFLOW_NO_RESULT_CACHE=1 buildflow ...`.

## `reports/` is buildflow-owned (nothing tracked inside)

BuildFlow's `test-coverage` step writes `reports/coverage.out`. The whole
`reports/` tree is gitignored, and buildflow creates the directory itself when
it is missing (verified empirically 2026-09-09: the step passes green on a tree
with no `reports/` dir). No tracked placeholder file is needed — do not
force-add files into `reports/`.

## GOEXPERIMENT=jsonv2 (required)

go-finding (latest) imports `encoding/json/v2`, which is gated behind
`//go:build goexperiment.jsonv2` in Go 1.26.x. The repo's `.envrc` contains
`use_go_env` — a direnv helper from the LarsArtmann home-managed direnv lib
that auto-detects `encoding/json/v2` imports in `.go` files and exports
`GOEXPERIMENT=jsonv2` on `cd`.

If direnv is unavailable (CI, containers), set the env var explicitly:
`GOEXPERIMENT=jsonv2 go build ./...` or `GOEXPERIMENT=jsonv2 buildflow`.

`go env -w GOEXPERIMENT=jsonv2` does NOT work on this NixOS host (read-only
`~/.config/go/env`). The `.envrc` is the durable mechanism. Do not rely on
`go env -w`.

## Design decisions

### `ConfigError` error-chain traversal

`ConfigError` wraps an underlying cause (`Err error`) and exposes it through the
standard `Unwrap() error`, `Is(error) bool`, and `As(any) bool` methods. All
three delegate to the wrapped error, so `errors.Is`, `errors.AsType`, and
`errors.Unwrap` all traverse the chain. Tests verify sentinel matching
(`fs.ErrNotExist`) and typed-cause extraction (`*jsontext.SyntacticError`, the
jsonv2 equivalent of v1's `json.SyntaxError`).

## Conventions

- Return `*ConfigError` (a specific type), never bare `error`, from config I/O
  helpers. This is what the `hierarchical-errors` check enforces and why the
  type exists.
- No em dashes in code; prefer `errors.Is`/`errors.As` over string matching on
  error messages.
- Exported symbols carry godoc comments (matches existing style; also keeps
  buildflow lint quiet).
