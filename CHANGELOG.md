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
  and `ConfigError.Unwrap`

### Changed

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
- `SaveJSON` now writes atomically (temp file + `os.Rename`) with indented
  output (`json.MarshalIndent`), preventing config corruption on crash
- `FindingFromIssue` uses `finding.FilePos` for `Line==0` instead of
  fabricating line 1; normalizes fix strategy with `FixStrategyNone` when no
  suggestion is present
- `errors.As` calls migrated to `errors.AsType[E]` (Go 1.26 generic)
- AGENTS.md rewritten: stale "do not upgrade go-finding" note replaced with
  "stay on latest + GOEXPERIMENT=jsonv2" guidance

### Removed

- Stale `ConfigError` design-decision note claiming `Unwrap` was omitted due to
  a `hierarchical-errors` analyzer false positive — empirically verified the
  analyzer does NOT flag `Unwrap() error`; the method is now implemented

## [0.1.0] - 2026-01-01

### Added

- Initial release
