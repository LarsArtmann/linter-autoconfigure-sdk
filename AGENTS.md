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
`github.com/larsartmann/go-finding` v1.1+.

**Do NOT upgrade to go-finding v1.2+.** v1.2.0 switched to `encoding/json/v2` and
`encoding/json/jsontext`, which are excluded by build constraints in this NixOS
environment. The entire package fails to compile. Stick with v1.1.x.

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

## Known environmental warnings (not code issues)

### `go-auto-upgrade`: encoding/json v1 -> v2 migration skipped

BuildFlow suggests migrating `encoding/json` to `encoding/json/v2`. In this
NixOS environment the Go toolchain lives in a read-only Nix store and
`encoding/json/v2` is excluded by build constraints (`build constraints exclude
all Go files in .../encoding/json/v2`), so the migration would produce
non-compiling code. The tool detects this and skips the migration, but still
emits a warning finding.

This is expected and unactionable here. It surfaces as a "warning" finding; it
does not fail `buildflow` or `buildflow --fix` (both exit 0). It only affects
`--fail-on-findings`. If a CI gate needs to pass on warnings, use
`buildflow --fix --fail-on=error` instead. Do NOT migrate to `encoding/json/v2`
to silence this: it breaks compilation in this environment.

## Design decisions

### `ConfigError` uses `Is`/`As`, not `Unwrap`

`ConfigError` wraps an underlying cause (`Err error`) and exposes it to
`errors.Is` / `errors.As` via custom `Is`/`As` methods that delegate to the
wrapped error. It deliberately does **not** implement `Unwrap() error`.

Reason: the `hierarchical-errors` analyzer flags any function/method returning
the bare `error` interface. `Unwrap() error` is mandated by the `errors`
package contract and cannot be narrowed, so it is a permanent false positive.
Importantly, BuildFlow runs the analyzer in **pipeline mode**, which does not
apply `//nolint` source suppression (the standalone `hierarchical-errors` tool
does, but the buildflow integration does not). The `Is`/`As` methods return
`bool`, so they are not flagged, while still providing full chain traversal for
the standard `errors.Is(err, fs.ErrNotExist)` / `errors.As(err, &json.SyntaxError{})`
entry points.

If you ever change `ConfigError`'s shape, keep `Is`/`As` (or reintroduce a
suppression strategy) or the `hierarchical-errors` finding will return.

## Conventions

- Return `*ConfigError` (a specific type), never bare `error`, from config I/O
  helpers. This is what the `hierarchical-errors` check enforces and why the
  type exists.
- No em dashes in code; prefer `errors.Is`/`errors.As` over string matching on
  error messages.
- Exported symbols carry godoc comments (matches existing style; also keeps
  buildflow lint quiet).
