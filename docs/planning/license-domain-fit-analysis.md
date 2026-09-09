# License-Domain Fit Analysis

**Date:** 2026-07-26
**Question investigated:** Should `licenseforge` depend on `linter-autoconfigure-sdk` (or vice versa), and is the SDK too hardcoded to extend to the license domain?

---

## TL;DR

1. **Direction 1 — `licenseforge` → `linter-autoconfigure-sdk`**: NO. Domains don't match; SDK is linter-config-shaped; licenseforge already has richer equivalents of everything the SDK offers.
2. **Direction 2 — `linter-autoconfigure-sdk` → `licenseforge` (consumes findings to fix them)**: NO under the current SDK, **YES if the SDK is renamed/repuposed** and grows `SaveText` + `ApplyTemplate` + `UpdateCopyrightYear` helpers.
3. **Are we too hardcoded?** YES — only in naming and godoc. The Go types (`Op`, `ConfigError`, `ConfigIssue`, `ProviderSpec`, `ReadConfig`, `LoadJSON`, `SaveJSON`, `FindingFromIssue`) are already domain-agnostic. The lies live in package name, godoc, and one or two prose places.

The right move is **broaden the SDK's scope (rename + add helpers)** rather than reject the license domain as out-of-scope.

---

## Direction 1: `licenseforge` depending on `linter-autoconfigure-sdk` — NO

### Why not

- **Domain mismatch.** `linter-autoconfigure-sdk` solves three things: linter config round-trip, linter-config-issue finding emission, BuildFlow provider spec. `licenseforge` solves: project-context detection, LICENSE generation, LICENSE validation, git hooks, workflow orchestration. They share "read/write a file" as a concept but nothing else.
- **Architecture mismatch.** SDK is flat package, plain `error`, branded go-finding types. `licenseforge` is full Clean Architecture with `samber/do/v2` DI, `samber/mo` Result types everywhere (per `licenseforge/AGENTS.md`).
- **Findings emission is the wrong shape.** Licenseforge's `pkg/validation/finding_adapter.go` already does more than the SDK's `FindingFromIssue`: it derives Confidence per issue type, maps Categories, attaches domain Tags, and assembles SARIF Reports. Adopting the SDK would be a downgrade.
- **Wrong layer.** Licenseforge's `pkg/fileops` (transactions + path sanitization + rollback per `PARTS.md`) already covers what `SaveJSON` offers — at a richer level.
- **SDK is pre-v1 / unstable.** `linter-autoconfigure-sdk/README.md:155` explicitly says breaking changes are acceptable with zero consumers.
- **Dependency bloat.** Licenseforge would gain a third config-I/O path (alongside `pkg/config`/`koanf` and `pkg/fileops`) for no value.

### Verdict

Do not add `linter-autoconfigure-sdk` as a dependency of `licenseforge`.

---

## Direction 2: `linter-autoconfigure-sdk` adding license-fix capability — YES, with conditions

### The seam

licenseforge detects license issues → emits findings → `linter-autoconfigure-sdk` consumes them and writes the fix. This is the architecturally clean division of labor.

### What's blocking it (and the fix)

| Blocker                                                                         | Severity       | Fix                                                                                                                   |
| ------------------------------------------------------------------------------- | -------------- | --------------------------------------------------------------------------------------------------------------------- |
| Package doc hardcodes "linter" (`autoconfigure.go:1`)                           | Naming — fatal | Broaden to "project file" / "project artifact"                                                                        |
| `ConfigIssue` godoc hardcodes "linter config file" (`autoconfigure.go:139`)     | Naming — fatal | Broaden to "project file"                                                                                             |
| `SaveJSON` is JSON-only — license files are plain text                          | Functional gap | Add `SaveText(path, content) *ConfigError` with same atomic + idempotent semantics                                    |
| No template substitution (license templates use `{{AUTHOR}}`, `{{YEAR}}`, etc.) | Functional gap | Add `ApplyTemplate(dst, tmpl string, vars map[string]string) (*ConfigError, []string)` returning substituted warnings |
| No copyright-year update helper                                                 | Functional gap | Add `UpdateCopyrightYear(path string, year int) *ConfigError`                                                         |
| Module name `linter-autoconfigure-sdk` lies if scope broadens                   | Naming — fatal | Rename or rebrand (see "Rename decision" below)                                                                       |
| `ConfigFile` field godoc hardcodes ".golangci.yml" (`autoconfigure.go:223`)     | Naming — minor | Broaden example to "LICENSE", "README.md", etc.                                                                       |

### What does NOT need to change

The Go types are already domain-agnostic. `ConfigIssue{Rule, Message, Severity, File, Line, Suggestion}` works for linter configs, license files, README headers, anything file-shaped. `FindingFromIssue`, `FindingsFromIssues`, `ReadConfig`, `LoadJSON`, `ProviderSpec`, `ErrNoRepair` all work as-is.

### Rename decision (your call)

| Option                                              | Pros                                          | Cons                                                     |
| --------------------------------------------------- | --------------------------------------------- | -------------------------------------------------------- |
| **Keep `linter-autoconfigure-sdk`** + broaden godoc | Zero churn, no module-redirect needed         | Name lies; pkg.go.dev badge will mislead                 |
| **Rename to `project-autofix-sdk`**                 | Honest scope; communicates generality         | Forces go.mod redirect; breaks any planned consumer URLs |
| **Rename to `autoconfigure-sdk`** (drop "linter")   | Short, honest, lets linter + license both fit | Less discoverable for the original use case              |

My recommendation: rename. A module named after one domain shouldn't host another.

---

## Hardcoded-ness audit (where the lies live)

| Location                   | Current text                                                                                               | Issue                                    |
| -------------------------- | ---------------------------------------------------------------------------------------------------------- | ---------------------------------------- |
| `autoconfigure.go:1`       | `// Package autoconfigure provides the shared foundation for linter auto-configuration tools`              | Hardcoded to linters                     |
| `autoconfigure.go:2-3`     | `// (golangci-lint-auto-configure, oxlint-auto-configure, and future additions like biome-auto-configure)` | Examples — fine                          |
| `autoconfigure.go:47`      | `// ConfigError describes a failure while reading, parsing, or writing a linter config file`               | Hardcoded                                |
| `autoconfigure.go:81-83`   | `// left to each tool (different YAML libraries: golangci uses yaml.v3 / v4, oxlint may use go-yaml)`      | Tool-specific prose — fine for context   |
| `autoconfigure.go:118`     | `// Indented output is used because linter configs are typically human-edited`                             | Hardcoded + lying for license files      |
| `autoconfigure.go:139`     | `// ConfigIssue describes a single problem found in a linter config file`                                  | Hardcoded                                |
| `autoconfigure.go:143`     | `// Rule is the issue's rule identifier (e.g. "missing-linter", "wrong-priority")`                         | Examples — fine                          |
| `autoconfigure.go:223`     | `ConfigFile string // the config file path the tool manages (e.g. ".golangci.yml")`                        | Example too narrow — fine but improvable |
| `README.md` throughout     | `linter auto-configuration`, linter examples only                                                          | Domain-locked marketing copy             |
| `go.mod:1`                 | `module github.com/larsartmann/linter-autoconfigure-sdk`                                                   | Hardcoded name                           |
| ~~`LICENSE` vs `README.md:7`~~ | ~~PROPRIETARY file, MIT badge~~ | ~~Unrelated split-brain — separate fix~~ fixed at `23e74f1` (LICENSE is MIT; README §License links ./LICENSE since 2026-09-09) |

---

## Action plan (when you greenlight)

### Phase 1 — Honest naming (zero Go-type changes)

1. Rename module: `linter-autoconfigure-sdk` → `project-autofix-sdk` (or your pick).
2. Update `autoconfigure.go` package doc to reference "project files / project artifacts".
3. Broaden godocs on `ConfigError`, `ConfigIssue`, `SaveJSON`'s indented-output comment, `ConfigFile`.
4. Update `README.md` to include license + README as first-class examples alongside linter configs.
5. ~~Fix `LICENSE`/`README.md:7` license split-brain (unrelated to this work but noted).~~ done at `23e74f1`

### Phase 2 — Add license-shaped helpers

1. **`SaveText(path, content) *ConfigError`** — atomic, idempotent, mkdir-parent. Same semantics as `SaveJSON` minus the JSON marshalling. Use case: writing `LICENSE`, `README.md`, copyright headers.
2. **`ApplyTemplate(dst, tmpl string, vars map[string]string) (*ConfigError, []string)`** — substitute `{{KEY}}` tokens from embedded tmpl string, write atomically. Returns warnings (e.g. "unresolved variables: X, Y"). Use case: licenseforge `Proprietary` license with `--pricing-structure "..."` flags.
3. **`UpdateCopyrightYear(path string, year int) *ConfigError`** — scan existing file for `Copyright (c) YYYY[-YYYY] Author`, rewrite to `Copyright (c) YYYY Author` (or extend the range if current year > existing year). Idempotent. Use case: licenseforge `UpdateCopyright` workflow.
4. Add tests for all three (table-driven: missing file, malformed template, year present, year range extension, concurrent modification race against `atomicwrite.ErrConcurrentModification`).

### Phase 3 — Wire licenseforge

This is **not** in this SDK's scope. licenseforge (or a new `licenseforge-fix` tool) consumes the new SDK helpers. The SDK stays a focused library; the integration lives downstream.

---

## Risks

| Risk                                                                | Mitigation                                                                                                          |
| ------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| Rename breaks planned consumers (none exist yet per README:151)     | Greenfield — no migration cost                                                                                      |
| Template engine re-invents text/template                            | Use `text/template` from stdlib; document the choice                                                                |
| Copyright-year regex is locale-fragile (e.g. `©` vs `(c)` vs `(C)`) | Document supported forms; refuse + report otherwise                                                                 |
| `SaveText` could shadow licenseforge's `pkg/fileops` capabilities   | Document that this SDK is plain-text-only; no transactions, no rollback. Multi-step repairs belong in licenseforge. |

---

## Out of scope for this analysis

- ~~Fixing the existing LICENSE vs README.md license split-brain (separate ticket).~~ done at `23e74f1`
- Building `licenseforge-fix` itself (lives in licenseforge's repo).
- Migrating any current linter-auto-configurer to the new helpers (no consumers exist).

---

## Status (2026-09-09, docs-health pass)

The license split-brain noted throughout is fixed (`23e74f1`). The proposal
itself (rename/repurpose + SaveText/ApplyTemplate/UpdateCopyrightYear) remains
**open and unexecuted** — routed to ROADMAP.md ("Rename/repurpose proposal")
for evaluation against the first-consumer milestone. Note this analysis now
lives in a public repo and names internal projects (`licenseforge`); its
exposure is part of the docs/ fate decision (ROADMAP Q1, TODO_LIST T17).
