# Features

Honest inventory of what this SDK does, by status. Statuses:
`FULLY_FUNCTIONAL`, `PARTIALLY_FUNCTIONAL`, `BROKEN`, `PLANNED`.
Evidence cites code; claims were verified against the test suite (green,
97.6% statement coverage, 2026-09-09).

## Config I/O

| Feature                                         | Status                     | Notes / Evidence                                                                                                                                                          |
| ----------------------------------------------- | -------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Read raw config bytes (`ReadConfig`)            | FULLY_FUNCTIONAL           | `autoconfigure.go:84`; missing-file wraps `fs.ErrNotExist` (test `autoconfigure_test.go:126`)                                                                             |
| JSON load (`LoadJSON[T]`)                       | FULLY_FUNCTIONAL           | `autoconfigure.go:94`; malformed input wraps `*jsontext.SyntacticError` (test `autoconfigure_test.go:147`)                                                                |
| Atomic + idempotent JSON save (`SaveJSON`)      | FULLY_FUNCTIONAL           | `autoconfigure.go:119`; creates parent dirs, skips write when content unchanged (no mtime bump, tests `autoconfigure_test.go:194-253`), crash-durable via go-atomic-write |
| Typed config errors (`ConfigError` + `Op` enum) | FULLY_FUNCTIONAL           | `autoconfigure.go:37-78`; `errors.Is`/`AsType`/`Unwrap` chain traversal tested                                                                                            |
| YAML parsing                                    | — (out of scope by design) | Each tool keeps its own YAML library; SDK stops at byte-level reads                                                                                                       |

## Finding emission

| Feature                                                | Status           | Notes / Evidence                                                                                                                                                                                                    |
| ------------------------------------------------------ | ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Convert config issues to findings (`FindingFromIssue`) | FULLY_FUNCTIONAL | `autoconfigure.go:169`; attaches `FixStrategySuggest` when a suggestion exists, `FixStrategyNone` otherwise; file-level position when `Line == 0`; errors propagated (tests `autoconfigure_test.go:65-124,286-316`) |
| Batch conversion (`FindingsFromIssues`)                | FULLY_FUNCTIONAL | `autoconfigure.go:197`; whole batch fails on any invalid issue (test `autoconfigure_test.go:318`)                                                                                                                   |
| Runnable godoc examples                                | FULLY_FUNCTIONAL | `example_test.go` — `ExampleLoadJSON`, `ExampleSaveJSON`, `ExampleFindingFromIssue`                                                                                                                                 |

## BuildFlow integration

| Feature                                                            | Status           | Notes / Evidence                                                                                                                                                                      |
| ------------------------------------------------------------------ | ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Provider shape (`ProviderSpec` + `HasRepair`)                      | FULLY_FUNCTIONAL | `autoconfigure.go:224-239`; typed `ConfigFile` (`finding.FilePath`); closures usable standalone without BuildFlow                                                                     |
| BuildFlow adapter (`ProviderFromSpec` → go-finding `toolsdk.Spec`) | FULLY_FUNCTIONAL | `autoconfigure.go:241`; Detect adapter emits findings via `FindingFromIssue`, nil Repair stays nil (canonical suggest-only signal), validated by a `toolsdk.Register` acceptance test |
