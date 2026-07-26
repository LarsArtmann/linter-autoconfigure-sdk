# linter-autoconfigure-sdk — Project Context

Enduring context for AI sessions working in this repo. Domain-facing docs live in
the README; this file captures what is hard to discover from the code alone.

## What this is

Shared foundation (Go library) for linter auto-configuration tools. Owns the
plumbing that `golangci-lint-auto-configure` and `oxlint-auto-configure` duplicate:
config file round-trip, finding emission for config issues, and a BuildFlow
provider spec. YAML parsing is intentionally NOT here (each tool uses a
different YAML library). Early-stage; no active consumers yet.

Module: `github.com/larsartmann/linter-autoconfigure-sdk`. Requires Go 1.26+ and
`github.com/larsartmann/go-finding` (always latest).

**Always stay on the latest go-finding version.** go-finding v1.2+ imports
`encoding/json/v2` and `encoding/json/jsontext`, which are gated behind the
`//go:build goexperiment.jsonv2` build tag in Go 1.26.x. This is NOT a
permanent incompatibility — it just requires `GOEXPERIMENT=jsonv2` to be set in
the environment. The repo ships a `.envrc` (direnv) that sets this automatically;
`direnv allow` on first clone. See "GOEXPERIMENT=jsonv2" below for details.

## Build, test, lint

This project is automated with **BuildFlow** (no Makefile, no flake.nix):

- `buildflow` — full quality pipeline (detect mode). Exits 0 when healthy.
- `buildflow --fix` — detect + auto-fix. Exits 0 when healthy.
- `buildflow --fix --fail-on-findings` — strict; exits non-zero if ANY finding
  (including environmental warnings) remains. See "Known environmental warnings".
- `go test -race -count=1 ./...` — just the Go tests (covered by buildflow
  `test-race` and `test-coverage` steps).

Single step: `buildflow -s <step> -v`. Disable result cache during debugging:
`BUILDFLOW_NO_RESULT_CACHE=1 buildflow ...`.

## `reports/.gitkeep` — do not delete

BuildFlow's `test-coverage` step runs
`go test -coverprofile=reports/coverage.out ...`. `go test` will not create the
`reports/` directory, and BuildFlow does not create it either. The directory is
gitignored at the directory level (`reports/` in the buildflow-managed block),
so a normal tracked file cannot keep it alive.

`reports/.gitkeep` is force-added (`git add -f`) so the directory survives a
fresh clone and `test-coverage` can write into it. It is the one tracked file
inside the otherwise-ignored `reports/` tree. Leave it in place.

## GOEXPERIMENT=jsonv2 (required)

go-finding (latest) imports `encoding/json/v2`, which is gated behind
`//go:build goexperiment.jsonv2` in Go 1.26.x. The repo's `.envrc` sets
`export GOEXPERIMENT=jsonv2` and direnv loads it automatically on `cd`.

If direnv is unavailable (CI, containers), set the env var explicitly:
`GOEXPERIMENT=jsonv2 go build ./...` or `GOEXPERIMENT=jsonv2 buildflow`.

`go env -w GOEXPERIMENT=jsonv2` does NOT work on this NixOS host (read-only
`~/.config/go/env`). The `.envrc` is the durable mechanism. Do not rely on
`go env -w`.

## Design decisions

### `ConfigError` error-chain traversal

`ConfigError` wraps an underlying cause (`Err error`) and exposes it via custom
`Is`/`As` methods that delegate to the wrapped error. The full rationale
(`hierarchical-errors` analyzer constraint, pipeline-mode `//nolint` limitation)
lives in the godoc comment on `ConfigError` itself — that is the canonical source
of truth. This section points to it to avoid duplication.

## Conventions

- Return `*ConfigError` (a specific type), never bare `error`, from config I/O
  helpers. This is what the `hierarchical-errors` check enforces and why the
  type exists.
- No em dashes in code; prefer `errors.Is`/`errors.As` over string matching on
  error messages.
- Exported symbols carry godoc comments (matches existing style; also keeps
  buildflow lint quiet).
