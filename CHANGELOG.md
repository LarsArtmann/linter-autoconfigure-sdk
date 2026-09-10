# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- `Op` typed enum (`OpRead`, `OpUnmarshal`, `OpMarshal`, `OpMkdir`, `OpWrite`)
  replacing the bare `string` on `ConfigError.Op` — typos are now compile errors
- `ConfigError.Unwrap() error` method for idiomatic `errors.Unwrap` chain
  traversal (alongside the existing `Is`/`As` methods)
- `ProviderSpec.HasRepair() bool` helper to check repair support
- `example_test.go` with `ExampleLoadJSON`, `ExampleSaveJSON`,
  `ExampleFindingFromIssue` for pkg.go.dev discoverability
- `.envrc` (direnv) setting `GOEXPERIMENT=jsonv2` — required because go-finding
  (latest) imports `encoding/json/v2`, gated behind `goexperiment.jsonv2`
- Tests for `FindingFromIssue` error path, `Line==0` file-level position,
  `FindingsFromIssues` error propagation, `SaveJSON` marshal/mkdir errors,
  `SaveJSON` idempotency (no mtime bump on identical content), and
  `ConfigError.Unwrap`
- `ProviderFromSpec(spec ProviderSpec) (toolsdk.Spec, error)` — converts a
  provider spec into the canonical BuildFlow provider contract from
  go-finding's `toolsdk` sub-module: `Analyze` is wrapped as a
  `finding.Detector` (issues become findings stamped with the spec's tool
  name), a non-nil `Repair` is wrapped as a `toolsdk.Repairer`, and
  `ConfigFile` becomes `Inputs`. Validation errors name the offending field.
- `ExampleProviderFromSpec` godoc example plus tests for the field mapping,
  Detect/Repair adapters, suggest-only nil `Repairer`, validation errors, and
  `toolsdk.Register` acceptance of the converted spec
- Package doc now states the `GOEXPERIMENT=jsonv2` requirement for Go 1.26
  builds (standard in Go 1.27)
- Regression tests for `(*ConfigError).Is` / `.As` delegation to the wrapped
  cause (positive, negative, and `errors.Is` / `errors.AsType` round-trips)
- Filesystem-independent idempotency test: `os.SameFile` (dev+inode identity)
  replaces the 20ms-sleep mtime comparison that could false-pass on
  coarse-mtime filesystems (write-if-changed report d.3 / TODO_LIST T8)
- Concurrency test: parallel `SaveJSON` calls to the same path surface
  `*ConfigError` wrapping `atomicwrite.ErrConcurrentModification`
  (barrier-released writers, re-seeded rounds); write-error-path test
  (read-only dir wraps `fs.ErrPermission`); `changed`-flag test
- Minimal `.golangci.yml` (golangci-lint v2, standard linters) that also
  documents why the flat layout is intentional: no `internal/` (the whole
  module is public API) and no `examples/` dir (godoc examples live in
  `example_test.go` for pkg.go.dev rendering)
- `SECURITY.md` with a private-vulnerability-reporting contact and scope
- CI (`.github/workflows/ci.yml`): gofmt check, `go vet`, race-enabled tests,
  coverage artifact on every push/PR, with `GOEXPERIMENT=jsonv2` set. BuildFlow
  itself is private and not installable on runners, so CI runs the portable
  subset of the pipeline; swap for a buildflow job when BuildFlow goes public
- Issue and PR templates, `CODEOWNERS`, and `dependabot.yml` (gomod +
  github-actions ecosystems, weekly)
- Social preview image asset at `docs/branding/social-preview.png`
  (1280x640; upload via repo Settings is a manual step, no API exists)

### Changed

- **License changed from proprietary (all rights reserved) to MIT** and the
  repository was flipped from private to public (`23e74f1`); full git history
  secret-scanned with gitleaks after the flip — zero findings
- **BREAKING:** `ConfigIssue.Rule` is now `finding.RuleName` (was `string`)
- **BREAKING:** `ConfigIssue.File` is now `finding.FilePath` (was `string`)
- **BREAKING:** `FindingFromIssue` signature is now
  `(finding.ToolName, ConfigIssue) (finding.Finding, error)` — was
  `(string, ConfigIssue) finding.Finding`; errors from `builder.Build()` are
  propagated instead of silently swallowed
- **BREAKING:** `FindingsFromIssues` signature is now
  `(finding.ToolName, []ConfigIssue) ([]finding.Finding, error)` — was
  `(string, []ConfigIssue) []finding.Finding`
- **BREAKING:** `ConfigError.Op` is now typed `Op` (was `string`)
- `SaveJSON` now writes via `go-atomic-write.WriteIfChanged`: idempotent (skips
  the write entirely if the marshalled content is byte-identical to the
  existing file — no mtime bump, no spurious diff) and crash-durable
  (fsync'd temp file + atomic rename). Race-safe: a concurrent modification
  surfaces as `*ConfigError` wrapping `atomicwrite.ErrConcurrentModification`
- `FindingFromIssue` uses `finding.FilePos` for `Line==0` instead of
  fabricating line 1; normalizes fix strategy with `FixStrategyNone` when no
  suggestion is present
- `errors.As` calls migrated to `errors.AsType[E]` (Go 1.26 generic)
- AGENTS.md rewritten: stale "do not upgrade go-finding" note replaced with
  "stay on latest + GOEXPERIMENT=jsonv2" guidance
- `ErrNoRepair` deprecated — suggest-only is now expressed structurally:
  `ProviderFromSpec` leaves `toolsdk.Spec.Repair` nil, which is the canonical
  signal in the toolsdk contract (suggestions still flow through findings
  carrying `FixStrategySuggest`). The sentinel stays as a deprecated alias so
  pre-v1 code referencing it keeps compiling; this SDK never returns it, and
  returning it from a custom Repairer is a repair failure for BuildFlow, not a
  suggest-only signal. Removal planned at v1 (TODO_LIST T20)
- **BREAKING:** `ProviderSpec.ConfigFile` is now `finding.FilePath` (was
  `string`), matching `ConfigIssue.File`
- Unchecked `os.RemoveAll` returns in godoc examples are now explicitly
  discarded with `_ =` (errcheck clean)
- **BREAKING:** `SaveJSON` now returns `(changed bool, *ConfigError)` — was
  `*ConfigError`. Zero consumers exist, so the change is free now; repair
  flows can distinguish "config updated" from "config already correct"
  (write-if-changed report Q2 / TODO_LIST T16; the `SaveJSONIfChanged`
  variant was rejected because `SaveJSON` is already if-changed — the name
  would lie)
- `FindingsFromIssues` conversion errors now include the tool name
  (erraudit: context variable was lost on the error path)
- `ConfigError.As` carries a trailing `//nolint:legacyerrors` suppression
  with reason: `As` must delegate to an arbitrary caller-chosen target type,
  which `errors.AsType[E]` cannot express (erraudit false positive; the
  directive only works as a trailing comment, not on the line above)
- README audited as the public sales page: install snippet now pins
  `@master` (plain `go get` fails with zero tags until `v0.1.0`), the
  `GOEXPERIMENT=jsonv2` requirement is explained inline instead of deferring
  to AGENTS.md, CI badge added, and the consumer links verified (both
  consumer repos are public, resolving ROADMAP Q3)

### Removed

- Stale `ConfigError` design-decision note claiming `Unwrap` was omitted due to
  a `hierarchical-errors` analyzer false positive — empirically verified the
  analyzer does NOT flag `Unwrap() error`; the method is now implemented

### Dependencies

- Added `github.com/larsartmann/go-atomic-write` (direct) — provides the
  idempotent, crash-durable atomic write primitive used by `SaveJSON`.
  Transitive: `cespare/xxhash/v2` (fingerprinting), `gofrs/flock` (locking)
- `github.com/larsartmann/go-finding` kept on latest: v1.8.0 (`9d1373a`),
  then v1.9.2 (`1c7c5e5`), then v1.10.0 — build and tests verified green at
  each bump. v1.10.0 has no core-module API changes; it adds the `toolsdk`
  sub-module (the BuildFlow provider contract), consumed via
  `ProviderFromSpec`
- Added `github.com/larsartmann/go-finding/toolsdk` v1.10.0 (direct) — the
  canonical BuildFlow provider plugin contract (`Spec`, `Register`, `Trigger`)
- Note: the `go` directive sits at `1.26.7` because go-finding declares that
  floor; `go mod tidy` reinstates it after any normalization, so the fix
  belongs upstream

### Repository

- Branch protection enabled on `master`: force pushes and deletions blocked,
  admins not enforced (the auto-commit daemon keeps pushing as owner)
- Repo topics set (`go`, `linter`, `sdk`, `golangci-lint`, `oxlint`, `config`),
  description refreshed with the install command, homepage pointed at the
  pkg.go.dev page
- External fetchability verified end-to-end: `go get
  github.com/larsartmann/linter-autoconfigure-sdk@master` from a clean
  throwaway module resolves a pseudo-version via proxy.golang.org, and a
  consumer program using the SDK compiles and runs (flip report f3)
- pkg.go.dev listing triggered and rendering verified: MIT license detected,
  godoc for all exported symbols, all four `Example*` functions render
  (flip report f4/f43; re-check after 24h for the post-T16 snapshot)
- TODO_LIST T17 sweep decided: internal project names in tracked `docs/`
  are ACCEPTED, not redacted — they are irreversibly in public git history,
  gitleaks ran clean at the visibility flip, and redacting working-tree
  copies would be cosmetic. ROADMAP Q1 still owns the docs-fate decision

> No version has been tagged yet. Everything above `Unreleased`-grade until
> the first tag (`v0.1.0`, see TODO_LIST/ROADMAP); pkg.go.dev serves only
> pseudo-versions in the meantime.
