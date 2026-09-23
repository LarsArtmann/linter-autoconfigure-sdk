# Features

Honest inventory of what this SDK does, by status. Statuses:
`FULLY_FUNCTIONAL`, `PARTIALLY_FUNCTIONAL`, `BROKEN`, `PLANNED`.
Evidence cites symbols (not line numbers, which rot); claims were verified
against the test suite (green, 100.0% statement coverage, 2026-09-10).

## Config I/O

| Feature                                         | Status                     | Notes / Evidence                                                                                                                                                                                                                                                                         |
| ----------------------------------------------- | -------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Read raw config bytes (`ReadConfig`)            | FULLY_FUNCTIONAL           | `ReadConfig` in `autoconfigure.go`; missing-file wraps `fs.ErrNotExist` (test `TestReadConfig_*`)                                                                                                                                                                                        |
| JSON load (`LoadJSON[T]`)                       | FULLY_FUNCTIONAL           | `LoadJSON` in `autoconfigure.go`; malformed input wraps `*jsontext.SyntacticError` (test `TestLoadJSON_*`)                                                                                                                                                                               |
| Byte-level parse (`ParseJSON[T]`)               | FULLY_FUNCTIONAL           | `ParseJSON` in `autoconfigure.go`; the counterpart of `LoadJSON` for bytes you already hold; errors carry no `Path` (tests `TestParseJSON_*`)                                                                                                                                            |
| Canonical marshal (`MarshalJSONIndented`)       | FULLY_FUNCTIONAL           | Deterministic map keys + 2-space indent, no trailing newline; `SaveJSON` builds on it byte-identically (tests `TestMarshalJSONIndented_*`, `TestSaveJSON_OutputsMarshalJSONIndentedBytes`)                                                                                               |
| Atomic + idempotent JSON save (`SaveJSON`)      | FULLY_FUNCTIONAL           | `SaveJSON` in `autoconfigure.go`; creates parent dirs, skips write when content unchanged (no inode change, race-tested), crash-durable via go-atomic-write; returns `changed bool`; concurrent modification surfaces as `*ConfigError` wrapping `atomicwrite.ErrConcurrentModification` |
| Byte-faithful atomic save (`SaveJSONBytes`)     | FULLY_FUNCTIONAL           | Raw-bytes `SaveJSON`; trailing newline is the caller's contract (tests `TestSaveJSONBytes_*`)                                                                                                                                                                                            |
| Working directory resolution (`WorkingDir`)     | FULLY_FUNCTIONAL           | `finding.WorkingDirFromContext` + `"."` fallback (tests `TestWorkingDir_*`)                                                                                                                                                                                                              |
| Typed config errors (`ConfigError` + `Op` enum) | FULLY_FUNCTIONAL           | `ConfigError` in `autoconfigure.go`; `errors.Is`/`AsType`/`Unwrap` chain traversal tested; renders cleanly without a `Path`                                                                                                                                                              |
| YAML parsing                                    | — (out of scope by design) | Each tool keeps its own YAML library; SDK stops at byte-level reads                                                                                                                                                                                                                      |

## Finding emission

| Feature                                                | Status           | Notes / Evidence                                                                                                                                                                                                        |
| ------------------------------------------------------ | ---------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Convert config issues to findings (`FindingFromIssue`) | FULLY_FUNCTIONAL | `FindingFromIssue` in `autoconfigure.go`; attaches `FixStrategySuggest` when a suggestion exists, `FixStrategyNone` otherwise; file-level position when `Line == 0`; errors propagated (tests `TestFindingFromIssue_*`) |
| Batch conversion (`FindingsFromIssues`)                | FULLY_FUNCTIONAL | `FindingsFromIssues` in `autoconfigure.go`; whole batch fails on any invalid issue (test `TestFindingsFromIssues_*`)                                                                                                    |
| Runnable godoc examples                                | FULLY_FUNCTIONAL | `example_test.go` — `ExampleLoadJSON`, `ExampleSaveJSON`, `ExampleMarshalJSONIndented`, `ExampleParseJSON`, `ExampleSaveJSONBytes`, `ExampleWorkingDir`, `ExampleDiffMaps`, `ExampleFindingFromIssue`, `ExampleBootstrapProviderFromSpec`                   |

## Config diff

| Feature                                       | Status           | Notes / Evidence                                                                                                                               |
| --------------------------------------------- | ---------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| Change model (`Change{Kind, Path, Old, New}`) | FULLY_FUNCTIONAL | `Change` in `diff.go`; `KindAdded`/`KindRemoved`/`KindModified`, deliberately no `unchanged` dead state                                        |
| Map/set/blob comparators                      | FULLY_FUNCTIONAL | `DiffMaps` / `DiffSets` / `DiffBlobs` in `diff.go`; deterministic path-sorted output; reorder-only input yields no changes (tests `TestDiff*`) |
| Value rendering (`StringValue`)               | FULLY_FUNCTIONAL | Bare strings; structured values as deterministic compact JSON; `%v` fallback (test `TestStringValue`)                                          |
| Rendering (`Summary`, `FormatDiff`)           | FULLY_FUNCTIONAL | `"Added: N, Modified: N, Removed: N"`; `+`/`-`/`~` lines sorted by path, `"No changes."` when empty (tests `TestSummary`, `TestFormatDiff_*`)  |

## Config discovery

| Feature                                   | Status           | Notes / Evidence                                                                                                                                 |
| ----------------------------------------- | ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| Multi-candidate discovery (`ConfigFiles`) | FULLY_FUNCTIONAL | `ProviderSpec.ConfigFiles` derives `Inputs` from all candidates; `ConfigFile` stays the write target (tests `TestProviderFromSpec_ConfigFiles*`) |
| First-existing lookup (`FirstExisting`)   | FULLY_FUNCTIONAL | Returns the first existing candidate, or the first path + `false` when none exists (test `TestFirstExisting`)                                    |

## BuildFlow integration

| Feature                                                            | Status                        | Notes / Evidence                                                                                                                                                                                        |
| ------------------------------------------------------------------ | ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Provider shape (`ProviderSpec` + `HasRepair`)                      | FULLY_FUNCTIONAL              | `ProviderSpec` / `HasRepair` in `autoconfigure.go`; typed `ConfigFile` (`finding.FilePath`); closures usable standalone without BuildFlow                                                               |
| BuildFlow adapter (`ProviderFromSpec` → go-finding `toolsdk.Spec`) | FULLY_FUNCTIONAL              | `ProviderFromSpec` in `autoconfigure.go`; Detect adapter emits findings via `FindingFromIssue`, nil Repair stays nil (canonical suggest-only signal), validated by a `toolsdk.Register` acceptance test |
| Bootstrap lifecycle (`BootstrapSpec[T]` + `BootstrapProviderFromSpec`) | FULLY_FUNCTIONAL          | `bootstrap.go`; derives missing-only Detect, never-overwrite dry-run-aware Repair, and advisory drift HealthCheck on the diff engine; contract tests ported from oxlint's `provider_test.go` semantics (tests `TestBootstrap*`) |
| Deprecated sentinel (`ErrNoRepair`)                                | FULLY_FUNCTIONAL (deprecated) | alias restored 2026-09-10 for pre-v1 compile compatibility; the SDK never returns it and returning it from a custom Repairer is a repair failure; removal at v1 tracked as TODO_LIST T20                |
