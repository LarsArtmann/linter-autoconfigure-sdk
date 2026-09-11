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
| Atomic + idempotent JSON save (`SaveJSON`)      | FULLY_FUNCTIONAL           | `SaveJSON` in `autoconfigure.go`; creates parent dirs, skips write when content unchanged (no inode change, race-tested), crash-durable via go-atomic-write; returns `changed bool`; concurrent modification surfaces as `*ConfigError` wrapping `atomicwrite.ErrConcurrentModification` |
| Typed config errors (`ConfigError` + `Op` enum) | FULLY_FUNCTIONAL           | `ConfigError` in `autoconfigure.go`; `errors.Is`/`AsType`/`Unwrap` chain traversal tested                                                                                                                                                                                                |
| YAML parsing                                    | — (out of scope by design) | Each tool keeps its own YAML library; SDK stops at byte-level reads                                                                                                                                                                                                                      |

## Finding emission

| Feature                                                | Status           | Notes / Evidence                                                                                                                                                                                                        |
| ------------------------------------------------------ | ---------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Convert config issues to findings (`FindingFromIssue`) | FULLY_FUNCTIONAL | `FindingFromIssue` in `autoconfigure.go`; attaches `FixStrategySuggest` when a suggestion exists, `FixStrategyNone` otherwise; file-level position when `Line == 0`; errors propagated (tests `TestFindingFromIssue_*`) |
| Batch conversion (`FindingsFromIssues`)                | FULLY_FUNCTIONAL | `FindingsFromIssues` in `autoconfigure.go`; whole batch fails on any invalid issue (test `TestFindingsFromIssues_*`)                                                                                                    |
| Runnable godoc examples                                | FULLY_FUNCTIONAL | `example_test.go` — `ExampleLoadJSON`, `ExampleSaveJSON`, `ExampleFindingFromIssue`                                                                                                                                     |

## BuildFlow integration

| Feature                                                            | Status                        | Notes / Evidence                                                                                                                                                                                        |
| ------------------------------------------------------------------ | ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Provider shape (`ProviderSpec` + `HasRepair`)                      | FULLY_FUNCTIONAL              | `ProviderSpec` / `HasRepair` in `autoconfigure.go`; typed `ConfigFile` (`finding.FilePath`); closures usable standalone without BuildFlow                                                               |
| BuildFlow adapter (`ProviderFromSpec` → go-finding `toolsdk.Spec`) | FULLY_FUNCTIONAL              | `ProviderFromSpec` in `autoconfigure.go`; Detect adapter emits findings via `FindingFromIssue`, nil Repair stays nil (canonical suggest-only signal), validated by a `toolsdk.Register` acceptance test |
| Deprecated sentinel (`ErrNoRepair`)                                | FULLY_FUNCTIONAL (deprecated) | alias restored 2026-09-10 for pre-v1 compile compatibility; the SDK never returns it and returning it from a custom Repairer is a repair failure; removal at v1 tracked as TODO_LIST T20                |
