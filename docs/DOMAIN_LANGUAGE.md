# Domain Language

Ubiquitous vocabulary for the linter-autoconfigure-sdk. Terms map 1:1 to code;
definitions describe how the code actually uses them.

## Core concepts

- **Auto-configurer** — a tool (e.g. `golangci-lint-auto-configure`,
  `oxlint-auto-configure`) that inspects a project and rewrites a linter config
  to be optimal. This SDK is the shared plumbing between them; the
  domain-specific detection (Go project shape vs JS framework) stays in each
  tool.
- **Config round-trip** — reading a linter config, mutating it in memory, and
  writing it back: `ReadConfig` / `LoadJSON` / `SaveJSON`. YAML parsing is
  deliberately outside the SDK (tools use different YAML libraries).
- **Idempotent write** — a save that is a no-op when the marshalled content is
  byte-identical to what is on disk (no mtime bump, no spurious diff).
  Re-running an auto-configurer must not dirty the tree.
- **Atomic write** — fsync'd temp file + rename, so a crash cannot truncate a
  user's config. Provided by `go-atomic-write` (`WriteIfChanged`).

## Types (`autoconfigure.go`)

- **`ConfigError`** — typed failure for config I/O: `{Op, Path, Err}`,
  mirroring `os.PathError`. Wraps the underlying cause; `Unwrap`/`Is`/`As`
  delegate to it so standard error-chain traversal works.
- **`Op`** — the failed operation as a typed enum: `OpRead`, `OpUnmarshal`,
  `OpMarshal`, `OpMkdir`, `OpWrite`. Short lowercase verbs, `os.PathError`
  convention.
- **`ConfigIssue`** — one problem found in a config file:
  `{Rule, Message, Severity, File, Line, Suggestion}`. Auto-configurers
  produce these; the SDK converts them to findings.
- **`ProviderSpec`** — the minimal shape an auto-configurer supplies to wire
  into BuildFlow: `{Name, Description, ConfigFile, Analyze, Repair}`.
  Provisional (zero consumers so far).
- **`ErrNoRepair`** — sentinel signalling an auto-configurer is suggest-only
  (its `Repair` closure is nil); BuildFlow treats it as non-autofixable.
- **Analyze / Repair** — the closures in a `ProviderSpec`. Analyze inspects
  the config and returns issues; Repair rewrites it and returns a
  human-readable description of what changed.

## Borrowed from go-finding

- **Finding** — a structured, aggregatable problem report BuildFlow can gate
  repairs on. The SDK converts `ConfigIssue`s into these.
- **Branded types** — `finding.RuleName`, `finding.FilePath`,
  `finding.ToolName`: compile-time-safe strings. The SDK adopts them
  everywhere instead of bare `string`s ("full coupling beats half-coupling").
- **Fix strategy** — what a finding promises about remediation:
  `FixStrategySuggest` (a suggestion text exists) vs `FixStrategyNone`
  (explicitly set to avoid the zero-value split brain).
- **File-level position** — `finding.FilePos`: a finding whose line is
  unknown (`Line == 0`) rather than a fabricated line number.

## Environment

- **`GOEXPERIMENT=jsonv2`** — Go 1.26 build flag gating `encoding/json/v2`
  (imported by go-finding). Required for every build/test of this SDK until
  jsonv2 is GA in Go 1.27; the repo's `.envrc` handles it via direnv.
