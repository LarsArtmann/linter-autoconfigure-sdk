# linter-autoconfigure-sdk — Project Context

Enduring context for AI sessions working in this repo. Domain-facing docs live in
the README; this file captures what is hard to discover from the code alone.

## What this is

Shared foundation (Go library) for linter auto-configuration tools. Owns the
plumbing that `golangci-lint-auto-configure` and `oxlint-auto-configure` duplicate:
config file round-trip, finding emission for config issues, the config diff
engine, and the BuildFlow provider bridge (`ProviderSpec` → `ProviderFromSpec`
→ go-finding's canonical `toolsdk.Spec`). YAML parsing is intentionally NOT
here (each tool uses a different YAML library). Public repo
(github.com/LarsArtmann/linter-autoconfigure-sdk), MIT-licensed.
Boundary decision: linter-recommendation findings with per-linter
categories/tags (golangci's `missing-linter`) stay app-side; ConfigIssue
models config-health issues only (no Category/Tags fields by design).
Ecosystem: `go-finding` (finding model + toolsdk contract) and
`go-atomic-write` (write primitive) are dependencies; `go-linter-sdk` is a
sibling, NOT a dependency or consumer — it scaffolds authoring linters that
find code issues, this SDK owns config-file plumbing; orthogonal layers
(decided 2026-09-10, re-confirmed 2026-09-23). Session orientation:
`TODO_LIST.md` is the open-work queue, `FEATURES.md` the honest feature
inventory.

**Exported API inventory (v0.6.0):** I/O — `ReadConfig`, `LoadJSON[T]`,
`ParseJSON[T]` (path-less), `MarshalJSONIndented` (deterministic + 2-space),
`SaveJSON` (if-changed, atomic), `SaveJSONBytes` (byte-faithful; trailing
newline is CALLER's contract), `WorkingDir(ctx)` (nil-ctx safe, "." fallback).
Diff — `Change{Kind,Path,Old,New}`, `Kind` string enum (KindAdded/
KindRemoved/KindModified — deliberately NO unchanged kind), `DiffMaps`,
`DiffSets`, `DiffBlobs` (set semantics), `StringValue`, `Summary` ("Modified:",
not "Changed:"), `FormatDiff` (+/-/~ lines, "No changes." when empty).
Discovery — `ProviderSpec.ConfigFiles []finding.FilePath` (Inputs derive from
it; `ConfigFile` remains the single-file write target), `FirstExisting(root,
candidates...)` (returns first candidate + false when none exist).
Provider — `ProviderFromSpec` (+ `HasRepair()`), validation sentinels
`ErrNameRequired`/`ErrDescriptionRequired`/`ErrAnalyzeRequired`, deprecated
`ErrNoRepair` (removal at v1 = TODO T20).

**Determinism enforcement (v0.6.0):** the `determinism` subpackage +
`cmd/jsondeterminism` vettool flag bare `encoding/json/v2` Marshal calls
(the pre-v0.3.1 `SaveJSON` bug class). CI runs
`go run ./cmd/jsondeterminism ./...` after vet. `json.Deterministic(false)`
is the explicit opt-out; opaque `opts...` spreads are not flagged
(conservative). Bite-check: `determinism/testdata` analysistest.

**Consumer migration state (2026-09-23, late):** both consumers build against
TAGS (no replace directives remain in the fleet). oxlint-auto-configure
(v0.9.1) is FULLY on the SDK: I/O helpers, diff engine aliases
`Change`/`Kind`, `ConfigFiles`+`FirstExisting` discovery, AND the bootstrap
provider lifecycle via `BootstrapProviderFromSpec` (its `validate` also runs a
drift advisory); its gate + CI run `cmd/jsondeterminism`.
golangci-lint-auto-configure (v0.10.0) uses `FindingFromIssue`, the diff
engine (`ChangeType = autoconfigure.Kind`), atomic config/backup writes via
go-atomic-write behind `config.NewOSFS()`, and `json.Deterministic(true)` on
every marshal (enforced by the same analyzer in its gate + CI). BuildFlow
pins the SDK v0.6.0 as indirect only.

Module: `github.com/larsartmann/linter-autoconfigure-sdk`. Requires Go 1.27+
(floor re-settled from `1.27.1` to minor-form `1.27` on 2026-09-23, see gotcha),
`github.com/larsartmann/go-finding` (always latest),
`github.com/larsartmann/go-finding/toolsdk` (always latest — the canonical
BuildFlow provider contract, consumed via `ProviderFromSpec`), and
`github.com/larsartmann/go-atomic-write` (always latest — used by `SaveJSON`
for idempotent, crash-durable writes).

**go.mod gotcha (recurring):** patch-form floors are dependency-imposed and
move with the dep graph — never hand-normalize the `go` directive, run
`go mod tidy` and accept the result. History: go-finding declared `go 1.26.7`
(pre-v0.3.0); v0.3.0 dropped to minor-form `go 1.27`; v0.3.1 through v0.6.0
sat at `go 1.27.1` (an upstream dep's patch floor); since 2026-09-23
(`golang.org/x/tools` v0.50.0 joined as a direct dep for the analyzer) tidy
settles this module at `go 1.27` — verified stable under a fresh tidy
2026-09-23. Consumers' floors ride along.

**Always stay on the latest go-finding version.** Since Go 1.27,
`encoding/json/v2` and `encoding/json/jsontext` are standard (no build tag);
the old `GOEXPERIMENT=jsonv2` requirement is gone. The repo carries no
GOEXPERIMENT anywhere as of 2026-09-23 (retired from CI; the `.envrc` never
exported it directly — the direnv `use_go_env` helper auto-detects jsonv2
imports outside the repo, which is moot here because the go 1.27 floor
makes 1.26 builds impossible for this module).

## Build, test, lint

This project is automated with **BuildFlow** (no Makefile, no flake.nix):

- `buildflow` — full quality pipeline (detect mode). Exits 0 when healthy.
- `buildflow --fix` — detect + auto-fix. Exits 0 when healthy.
- `buildflow --fix --fail-on-findings` — strict; exits non-zero if ANY finding
  remains. As of 2026-09-23 the strict baseline is 9 known advisories (all
  steps green, 38/38; plain mode exits 0):
  branching-flow 5 (the nil-deref warning at `bootstrap.go` `return *parsed`
  is a proven false positive — `ParseJSON` returns `new(T)`-backed non-nil on
  success; the rest are phantom-type style suggestions for the plain-string
  `Change.Old/New` and `FixCommand`/`CountLabel` fields, which are plain by
  design), cqrs-lint 2 (false positives — this repo imports no go-cqrs-lite),
  and go-auto-upgrade 2 (samber/lo adoption nudges; the SDK keeps third-party
  deps to the larsartmann family). Revisit only if the advisories start
  flagging real code.
- `go test -race -count=1 ./...` — just the Go tests (covered by buildflow
  `test-race` and `test-coverage` steps).
- Known tool bug (2026-09-22; SKIPPED via config since 2026-09-23):
  `license-check` (go-licenses) fails on the go 1.27 floor — it cannot load
  go 1.27 toolchain stdlib packages ("Package log/slog does not have module
  info. Non go modules projects are no longer supported"). Unrelated to code
  state; excluded via `.buildflow.yml` `skip_steps: [license-check]` (with
  reason), so the pipeline is FULLY green. Re-enable after a go-licenses
  release supporting 1.27 toolchains.

Single step: `buildflow -s <step> -v`. Release verification is scripted:
`./scripts/verify-release.sh vX.Y.Z` (proxy check + clean-dir go get +
compile-every-API smoke — run it after every tag). CI also runs the
jsondeterminism analyzer, the README snippet compile guard
(`readme_snippets_test.go` — caught real snippet drift on first run), and the
social-preview guard (`scripts/check-social-preview.sh`). A BuildFlow CI job
is NOT possible yet: the BuildFlow repo is PRIVATE again (2026-09-23; the
2026-09-11 public verification is stale) — T21. Disable result cache during debugging:
`BUILDFLOW_NO_RESULT_CACHE=1 buildflow ...` — the result cache has a 168h TTL
and can serve stale green results after a tool upgrade changed the verdict
(this masked a real golangci-lint-auto-configure gate failure on 2026-09-11).

**buildflow in non-direnv shells** (AI tool shells, cron): the shell's
`GOTOOLCHAIN=local` beats `.buildflow.yml` env, and the system go 1.26.7
fails the go.mod floor. Working recipe (verified 2026-09-23):
`env -u GOTOOLCHAIN BUILDFLOW_NO_RESULT_CACHE=1 PATH=<go127-bin>:/tmp/gv127:$PATH buildflow --fail-on-findings`
where `<go127-bin>` is a nix-store `*-go-1.27*/bin` and `/tmp/gv127` holds a
govulncheck BUILT WITH 1.27
(`GOBIN=/tmp/gv127 go install golang.org/x/vuln/cmd/govulncheck@latest`) —
the system govulncheck is built with 1.26 and fails against 1.27 stdlib
with ~24 bogus "package requires newer Go version" findings.

## Lint configuration (2026-09-11 wave)

- `.golangci.yml` is GENERATED by `golangci-lint-auto-configure` (via
  `buildflow --fix -s golangci-lint-auto-configure`); it enables ~106 linters
  plus a 5m run timeout. Hand-edits may be overwritten by the next fix run.
  The header comment (flat-layout rationale) survives regeneration.
- `.markdownlint.yml`: MD013 at 120 with tables/code_blocks exempt; MD024
  siblings_only (changelog sections repeat heading names by convention).
- `.buildflow.yml` excludes `docs/status/**` and `docs/planning/**` from all
  checks: they are point-in-time historical records and must not be
  reformatted or link-checked against rotted references. markdownlint-cli has
  NO hierarchical config (a `.markdownlint.yml` in a subdirectory is
  ignored), so buildflow-level exclusion is the only clean mechanism; the
  `docs/status/.markdownlint.yml` exists only for direct markdownlint runs
  inside that directory.
- BuildFlow exclude patterns are doublestar globs matched against the full
  path OR the basename: `docs/status` matches nothing, `docs/status/**`
  matches everything under it.
- erraudit enforces that error wraps on error paths include every in-scope
  context variable in the message (e.g. a wrap in
  `FindingFromIssue` must carry `toolName`, not just the rule name).
- ProviderFromSpec validation returns exported sentinels
  (`ErrNameRequired`, `ErrDescriptionRequired`, `ErrAnalyzeRequired`) with
  the same messages the old dynamic errors had; match with `errors.Is`.

## CI (GitHub Actions)

`.github/workflows/ci.yml` runs the portable subset on push/PR: gofmt check,
`go vet`, race tests with the coverage artifact uploaded. BuildFlow itself is
a private module and
cannot be installed on runners, so it is NOT in CI; swap the workflow for a
buildflow job if/when BuildFlow goes public (TODO_LIST T21). Branch protection
on `master` blocks force pushes and deletions but does not enforce on admins,
so the auto-commit daemon keeps working; the `Build, vet, test` check is a
required status check (admins bypass it, PRs must pass it).

## Releases

`v0.1.0` tagged 2026-09-10 (annotated tag, first release). Procedure that
worked: cut CHANGELOG, sync README/FEATURES into the release commit, wait for
CI green on that exact commit, `git tag -a` + push tag + `gh release create
--latest` (a PLAIN release marked Latest — GitHub's API rejects
`--prerelease` combined with Latest with HTTP 422, verified live at v0.3.1;
drop the `--prerelease` flag entirely even for v0.x), then verify proxy
(`go list -m -versions`) and a clean-dir `go get@vX.Y.Z` + consumer compile.
ALWAYS run the adversarial clean-dir smoke BEFORE tagging: the v0.4.0 smoke
caught a `WorkingDir(nil)` panic after the tag was already immutable, which
forced the v0.4.1 hotfix. Gotchas seen live: the daemon does not
always push promptly (manual `git push origin master` of its commit was
needed once); pkg.go.dev 404s for a fresh tag even after the proxy serves it
(minutes-to-longer lag; the proxy is the source of truth). Tags are immutable
once the proxy caches them — never re-tag, always cut a new version.
Release history: v0.1.0, v0.2.0, v0.3.0 (go 1.27 floor), v0.3.1 (determinism
fix), v0.4.0 (I/O matrix + diff engine + ConfigFiles), v0.4.1 (nil-ctx fix), v0.5.0 (bootstrap provider), v0.6.0 (determinism analyzer + vettool cmd).

## `reports/` is buildflow-owned (nothing tracked inside)

BuildFlow's `test-coverage` step writes `reports/coverage.out`. The whole
`reports/` tree is gitignored, and buildflow creates the directory itself when
it is missing (verified empirically 2026-09-09: the step passes green on a tree
with no `reports/` dir). No tracked placeholder file is needed — do not
force-add files into `reports/`.

## GOEXPERIMENT=jsonv2 (historical; fully retired 2026-09-23)

`encoding/json/v2` was gated behind `//go:build goexperiment.jsonv2` in Go
1.26.x; for that era CI set `GOEXPERIMENT=jsonv2` and the direnv
`use_go_env` helper auto-exported it (not the `.envrc` itself). Since the
go 1.27 floor (v0.3.0) jsonv2 is standard, and since the go 1.27.1
dependency-imposed floor a 1.26 build of this module is impossible anyway,
so the repo carries no GOEXPERIMENT anywhere.

`go env -w` does NOT work on this NixOS host (read-only
`~/.config/go/env`). Do not rely on `go env -w`.

**Non-interactive shell gotcha:** bash shells that skip direnv resolve `go`
to the system 1.26.7, which fails the go.mod floor with "go.mod requires
go >= 1.27 (running go 1.26.7; GOTOOLCHAIN=local)". Prefix commands with a
nix-store Go 1.27, e.g.
`PATH=/nix/store/dvp2lzfa22nnpqfl5dbi680a93chdsjv-go-1.27.1/bin:$PATH go test ./...`
(store path may churn after nixos rebuilds; re-derive with
`ls -d /nix/store/*-go-1.27*/bin`).

## Design decisions

### `SaveJSON` returns `(changed bool, *ConfigError)`

Decided 2026-09-10 (TODO_LIST T16 / write-if-changed report Q2): the `changed`
bool from `atomicwrite.WriteIfChanged` is surfaced instead of discarded, so
repair flows can distinguish "config updated" from "config already correct".
A `SaveJSONIfChanged` variant was rejected on naming grounds: `SaveJSON` is
ALREADY if-changed (idempotent skip), so the variant name would lie. Zero
consumers existed when the signature changed, so the break was free.

### Flat layout is intentional

No `internal/` (the whole module is public API; there is nothing to hide) and
no `examples/` dir (runnable examples live in `example_test.go` so pkg.go.dev
renders them inline). Documented in `.golangci.yml`; the go-structure-linter
warnings about both are accepted, not fixed.

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
- No em dashes in code; prefer `errors.Is`/`errors.AsType` over string matching
  on error messages.
- `(*ConfigError).As` must keep `errors.As` (it delegates to an arbitrary
  caller-chosen target type; `AsType[E]` cannot express that). erraudit accepts
  this only via a TRAILING `//nolint:legacyerrors` comment on the same line —
  a standalone directive on the line above does NOT suppress.
- Exported symbols carry godoc comments (matches existing style; also keeps
  buildflow lint quiet).
- All production `encoding/json/v2` marshals pass `json.Deterministic(true)`
  (policy mirrors go-finding's `json.go` marshalOpts). Non-deterministic map
  keys silently defeat `WriteIfChanged` idempotency — the 2026-09-22 audit
  caught exactly that in `SaveJSON`. Any new jsonv2 output path gets the
  option on day one.
