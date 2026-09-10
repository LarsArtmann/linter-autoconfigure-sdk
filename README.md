# linter-autoconfigure-sdk

Shared foundation for linter auto-configuration tools — config round-trip, finding emission for config issues, and a provider spec for BuildFlow integration.

[![CI](https://github.com/LarsArtmann/linter-autoconfigure-sdk/actions/workflows/ci.yml/badge.svg)](https://github.com/LarsArtmann/linter-autoconfigure-sdk/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/larsartmann/linter-autoconfigure-sdk.svg)](https://pkg.go.dev/github.com/larsartmann/linter-autoconfigure-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/larsartmann/linter-autoconfigure-sdk)](https://goreportcard.com/report/github.com/larsartmann/linter-autoconfigure-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**[pkg.go.dev](https://pkg.go.dev/github.com/larsartmann/linter-autoconfigure-sdk)**

---

## Why?

Two existing auto-configurers — `golangci-lint-auto-configure` and `oxlint-auto-configure` — diverge on their domain-specific concepts (Go project shape vs JS framework detection, 4-tier priority enum vs profile presets). Those differences are legitimate and stay in each tool.

What they reinvent identically is the surrounding plumbing:

| Concern                                    | Before (per tool)                                             | After (this SDK)                                               |
| ------------------------------------------ | ------------------------------------------------------------- | -------------------------------------------------------------- |
| Read/write config files                    | Each tool hand-wraps `os.ReadFile` + parse + error handling   | `ReadConfig(path)` / `LoadJSON[T](path)` / `SaveJSON(path, v)` |
| Emit findings for config issues            | Each tool maps priority → Severity, fix → Suggestion, by hand | `FindingFromIssue(tool, ConfigIssue{...})`                     |
| Wire into BuildFlow as Detector + Repairer | Each tool writes its own adapter                              | `ProviderFromSpec` → canonical `toolsdk.Spec` (go-finding)     |

`linter-autoconfigure-sdk` owns that plumbing once. Adding a third auto-configurer (e.g. `biome-auto-configure`) becomes a config-schema exercise, not a from-scratch build.

---

## Installation

```bash
go get github.com/larsartmann/linter-autoconfigure-sdk@master
```

The `@master` pin is needed until the first tag (`v0.1.0`) is cut; after
that, plain `go get github.com/larsartmann/linter-autoconfigure-sdk` works.

Requires Go 1.26+ with `GOEXPERIMENT=jsonv2` set: go-finding imports
`encoding/json/v2`, which is experimental in Go 1.26 and standard in Go 1.27.
Either `export GOEXPERIMENT=jsonv2` or use a direnv-based `.envrc`.

Peer dependencies: the latest
[`go-finding`](https://github.com/larsartmann/go-finding) and
[`go-atomic-write`](https://github.com/larsartmann/go-atomic-write) modules.

---

## Usage

### Read a config file

```go
data, err := autoconfigure.ReadConfig(".golangci.yml")
// YAML parsing stays in each tool (different libraries: yaml.v3 vs go-yaml).
// This helper covers the shared read + existence check.
```

### Round-trip JSON configs

```go
type oxlintConfig struct {
    Plugins []string `json:"plugins"`
}

cfg, err := autoconfigure.LoadJSON[oxlintConfig](".oxlintrc.json")
// ... mutate cfg ...
changed, err := autoconfigure.SaveJSON(".oxlintrc.json", cfg) // creates parent dirs; changed=false when content already matched
```

### Emit findings for config issues

```go
issues := []autoconfigure.ConfigIssue{
    {
        Rule:       "missing-linter",
        Message:    "errcheck is not enabled",
        Severity:   finding.SeverityWarning,
        File:       ".golangci.yml",
        Line:       5,
        Suggestion: "add errcheck to enabled linters",
    },
}
findings := autoconfigure.FindingsFromIssues("golangci-autoconfigure", issues)
// → []finding.Finding with FixStrategySuggest attached where Suggestion != ""
```

### BuildFlow provider shape

```go
spec := autoconfigure.ProviderSpec{
    Name:        "golangci-autoconfigure",
    Description: "Optimize .golangci.yml",
    ConfigFile:  ".golangci.yml",
    Analyze: func(ctx context.Context) ([]autoconfigure.ConfigIssue, error) {
        // inspect config, return issues
    },
    Repair: func(ctx context.Context) (string, error) {
        // rewrite config, return description of changes
    },
}
```

`ProviderFromSpec(spec)` wraps this as the canonical BuildFlow provider contract — go-finding's `toolsdk.Spec` — ready for `toolsdk.Register` (adjust its `Trigger` / `DependsOn` on the returned value first if needed). The underlying analyze/repair closures also work standalone with no BuildFlow wiring.

```go
provider, err := autoconfigure.ProviderFromSpec(spec)
// provider.Detect implements finding.Detector (issues become findings)
// provider.Repair implements toolsdk.Repairer (nil when spec.Repair is nil:
// the canonical suggest-only signal)
```

---

## API

### Config I/O

| Function            | Signature                | Purpose                                                                                                                     |
| ------------------- | ------------------------ | --------------------------------------------------------------------------------------------------------------------------- |
| `ReadConfig(path)`  | `([]byte, *ConfigError)` | Read raw bytes; YAML parsing stays tool-specific                                                                            |
| `LoadJSON[T](path)` | `(*T, *ConfigError)`     | Read + unmarshal a JSON config                                                                                              |
| `SaveJSON(path, v)` | `(changed bool, *ConfigError)`         | Idempotent + crash-durable atomic write of indented JSON; creates parent dirs, skips the write when content is unchanged, and reports whether a write happened |

All three return a `*ConfigError` (implements `error`) whose `Op` (typed:
`OpRead`, `OpUnmarshal`, `OpMarshal`, `OpMkdir`, `OpWrite`), `Path`, and
`Err` fields describe the failure. The wrapped cause is reachable via
`errors.Is` / `errors.AsType` / `errors.Unwrap` against a `*ConfigError` (e.g.
`errors.Is(err, fs.ErrNotExist)`), so callers can distinguish missing files
from parse or I/O failures without parsing error strings.

### Finding emission

| Function                           | Signature                    | Purpose                                                                                                   |
| ---------------------------------- | ---------------------------- | --------------------------------------------------------------------------------------------------------- |
| `FindingFromIssue(tool, issue)`    | `(finding.Finding, error)`   | Convert one `ConfigIssue` to a `finding.Finding` (suggest-strategy auto-attached when `Suggestion != ""`) |
| `FindingsFromIssues(tool, issues)` | `([]finding.Finding, error)` | Slice version; propagates conversion errors                                                               |

### BuildFlow integration

| Function                      | Signature               | Purpose                                                                                                               |
| ----------------------------- | ----------------------- | --------------------------------------------------------------------------------------------------------------------- |
| `ProviderFromSpec(spec)`      | `(toolsdk.Spec, error)` | Convert a `ProviderSpec` into go-finding's canonical `toolsdk.Spec` for `toolsdk.Register`; validates required fields |
| `(*ProviderSpec).HasRepair()` | `bool`                  | Whether the spec supports auto-repair                                                                                 |

### Types

| Type           | Purpose                                                                                                                                                    |
| -------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ConfigError`  | `{Op, Path, Err}` — typed failure for config I/O; `Op` is a typed enum; supports `Unwrap`/`Is`/`As` for full error-chain traversal                         |
| `ConfigIssue`  | `{Rule, Message, Severity, File, Line, Suggestion}` — `Rule` is `finding.RuleName`, `File` is `finding.FilePath`                                           |
| `ProviderSpec` | `{Name, Description, ConfigFile, Analyze, Repair}` — auto-configurer declaration; `ConfigFile` is `finding.FilePath`; `HasRepair()` reports repair support |

---

## Design notes

- **YAML parsing is NOT in this SDK.** The two existing tools use different YAML libraries (yaml.v3 vs go-yaml) with different semantics. Forcing one would create friction. The SDK covers the shared byte-level read; YAML unmarshaling stays in each tool.
- **Branded types are adopted consistently.** `ConfigIssue.Rule` is `finding.RuleName`, `File` is `finding.FilePath`, `ProviderSpec.ConfigFile` is `finding.FilePath`, and `FindingFromIssue` takes `finding.ToolName`. The SDK is coupled to go-finding; taking the branded types buys compile-time safety at no extra cost.
- **`FindingFromIssue` auto-attaches `FixStrategySuggest` when `Suggestion != ""`; otherwise sets `FixStrategyNone` explicitly** to avoid the empty-string zero-value split brain. When `Line == 0`, a file-level position (`finding.FilePos`) is used instead of fabricating a line number.
- **No `ProjectType` enum.** The two existing tools have incompatible concepts (Go shape vs JS framework); sharing one enum would force false convergence. Each tool keeps its own detection layer.

---

## Consumers

Planned:

- [`golangci-lint-auto-configure`](https://github.com/LarsArtmann/golangci-lint-auto-configure)
- [`oxlint-auto-configure`](https://github.com/LarsArtmann/oxlint-auto-configure)
- Future: `biome-auto-configure`, etc.

No active consumers yet. The SDK provides atomic, crash-durable config writes (via go-atomic-write), branded finding types, and structured ConfigError wrapping — but value over stdlib remains modest until a consumer migrates.

## Status

Early (pre-v1). The config round-trip and finding-emission helpers have breaking signatures (typed `Op` enum, branded types, `(Finding, error)` and `(bool, *ConfigError)` returns) — no consumers exist yet, so breaking changes are acceptable. BuildFlow wiring is anchored to go-finding's canonical `toolsdk` contract (v1.10.0+). Requires `GOEXPERIMENT=jsonv2` on Go 1.26 (see [Installation](#installation)).

## License

MIT — see [LICENSE](LICENSE).
