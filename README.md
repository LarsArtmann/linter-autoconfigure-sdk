# linter-autoconfigure-sdk

Shared foundation for linter auto-configuration tools — config round-trip, finding emission for config issues, and a provider spec for BuildFlow integration.

[![Go Reference](https://pkg.go.dev/badge/github.com/larsartmann/linter-autoconfigure-sdk.svg)](https://pkg.go.dev/github.com/larsartmann/linter-autoconfigure-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/larsartmann/linter-autoconfigure-sdk)](https://goreportcard.com/report/github.com/larsartmann/linter-autoconfigure-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**[pkg.go.dev](https://pkg.go.dev/github.com/larsartmann/linter-autoconfigure-sdk)**

---

## Why?

Two existing auto-configurers — `golangci-lint-auto-configure` and `oxlint-auto-configure` — diverge on their domain-specific concepts (Go project shape vs JS framework detection, 4-tier priority enum vs profile presets). Those differences are legitimate and stay in each tool.

What they reinvent identically is the surrounding plumbing:

| Concern | Before (per tool) | After (this SDK) |
|---|---|---|
| Read/write config files | Each tool hand-wraps `os.ReadFile` + parse + error handling | `ReadConfig(path)` / `LoadJSON[T](path)` / `SaveJSON(path, v)` |
| Emit findings for config issues | Each tool maps priority → Severity, fix → Suggestion, by hand | `FindingFromIssue(tool, ConfigIssue{...})` |
| Wire into BuildFlow as Detector + Repairer | Each tool writes its own adapter | `ProviderSpec` shape (analyze + repair closures) |

`linter-autoconfigure-sdk` owns that plumbing once. Adding a third auto-configurer (e.g. `biome-auto-configure`) becomes a config-schema exercise, not a from-scratch build.

---

## Installation

```bash
go get github.com/larsartmann/linter-autoconfigure-sdk
```

Requires Go 1.26+ and [`go-finding`](https://github.com/larsartmann/go-finding) v1.2+.

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
err = autoconfigure.SaveJSON(".oxlintrc.json", cfg) // creates parent dirs
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

A future `ProviderFromSpec(spec)` helper will wrap this as a `toolsdk.Spec` (BuildFlow tool-sdk), but the underlying analyze/repair closures are usable standalone today.

---

## API

### Config I/O

| Function | Signature | Purpose |
|---|---|---|
| `ReadConfig(path)` | `([]byte, error)` | Read raw bytes; YAML parsing stays tool-specific |
| `LoadJSON[T](path)` | `(*T, error)` | Read + unmarshal a JSON config |
| `SaveJSON(path, v)` | `error` | Marshal + write JSON, creating parent dirs |

### Finding emission

| Function | Signature | Purpose |
|---|---|---|
| `FindingFromIssue(tool, issue)` | `finding.Finding` | Convert one `ConfigIssue` to a `finding.Finding` (suggest-strategy auto-attached when `Suggestion != ""`) |
| `FindingsFromIssues(tool, issues)` | `[]finding.Finding` | Slice version |

### Types

| Type | Purpose |
|---|---|
| `ConfigIssue` | `{Rule, Message, Severity, File, Line, Suggestion}` — describes one config problem |
| `ProviderSpec` | `{Name, Description, ConfigFile, Analyze, Repair}` — BuildFlow provider shape |
| `ErrNoRepair` | Sentinel: tool does not support auto-repair (suggest-only) |

---

## Design notes

- **YAML parsing is NOT in this SDK.** The two existing tools use different YAML libraries (yaml.v3 vs go-yaml) with different semantics. Forcing one would create friction. The SDK covers the shared byte-level read; YAML unmarshaling stays in each tool.
- **`ConfigIssue.File` is a plain string, not a branded `FilePath`.** Keeps the SDK decoupled from finding's branded types; consumers wrap at the boundary.
- **`FindingFromIssue` auto-attaches `FixStrategySuggest` when `Suggestion != ""`.** Matches the ecosystem convention that suggestions surface as fixable findings.
- **No `ProjectType` enum.** The two existing tools have incompatible concepts (Go shape vs JS framework); sharing one enum would force false convergence. Each tool keeps its own detection layer.

---

## Consumers

Planned:
- [`golangci-lint-auto-configure`](https://github.com/LarsArtmann/golangci-lint-auto-configure)
- [`oxlint-auto-configure`](https://github.com/LarsArtmann/oxlint-auto-configure)
- Future: `biome-auto-configure`, etc.

No active consumers yet. The weakest of the 5 SDKs — value over stdlib is modest until a second auto-configurer lands.

## Status

Early. The config round-trip and finding-emission helpers are stable. The `ProviderSpec` shape is provisional and may evolve when the first consumer migrates.

## License

MIT — see [LarsArtmann/template-LICENSE](https://github.com/LarsArtmann/template-LICENSE).
