# Roadmap

Long-term direction and raw ideas not yet refined into bounded tasks. Actionable
work lives in `TODO_LIST.md`. When an idea here becomes bounded, graduate it.

## Vision

Become the canonical shared foundation for linter auto-configuration tools
(golangci-lint, oxlint, biome, eslint, ruff, ...). The SDK earns or loses its
right to exist the day the **first real consumer migrates** to it — until then
everything about the abstraction is provisional. The package doc in
`autoconfigure.go:1-17` states this intent.

## Candidate directions (raw ideas)

- **First consumer migration.** Wire `golangci-lint-auto-configure` or
  `oxlint-auto-configure` onto `ConfigError`, `ConfigIssue`,
  `FindingFromIssue`, and `ProviderSpec` end-to-end. This is the SDK's real
  validation milestone; it will settle which exports are stable.
- **`ProviderFromSpec(spec)` BuildFlow adapter.** The README promises it as a
  future helper. Build it when a consumer exists, not before (avoiding
  speculative API design against BuildFlow's tool-sdk).
- **Write variants.** `SaveYAML` counterpart (per-tool YAML libs make this
  tricky), `SaveJSONCompact` for machine-only configs,
  `ReadConfigWithFingerprint` for read-modify-write transactions,
  `WithIndent` option for `SaveJSON`, `LoadJSONWith[T]` accepting
  `jsontext.Options` (e.g. `RejectUnknownMembers`).
- **`ProviderSpec` validation helper** catching obvious misconfiguration
  (empty Name, missing Analyze, Repair contract violations).
- **`MustLoadJSON` / `MustSaveJSON`** panic-on-error variants for fixtures.
- **Extract `configerr` sub-package** (`ConfigError`, `Op`, `Unwrap/Is/As`)
  if non-autoconfigure packages want the error machinery without the domain
  types.
- **Testing depth.** BDD suite for the auto-configure flow (onsi/ginkgo),
  fuzz tests for `LoadJSON`/`SaveJSON` (no panics on arbitrary input),
  property-based Save→Load round-trip invariants.
- **Pipeline hardening.** Add `gitleaks` and `gosec` to the buildflow
  pipeline; add a license-badge ↔ LICENSE consistency check; add a markdown
  link/badge checker (the dead template-LICENSE link was the first of its
  class).
- **Reviews worth running at this size:** `library-deep-dive` on go-finding
  and go-atomic-write (confirm the SDK uses both to their full potential),
  `data-model-review` on the exported types, `full-code-review` (~240 LOC +
  ~370 LOC tests, cheap to do whole).
- **Go-public checklist as a reusable skill** — the ordered checklist
  (history secret scan → tracked-file review → internal-docs review → license
  consistency → tag/CI plan → flip → verify → pkg.go.dev check) that session
  2026-09-09 executed in the wrong order.
- **Rename/repurpose proposal** (from
  `docs/planning/license-domain-fit-analysis.md`): broaden the SDK to a
  general "project file autofix" SDK (`SaveText`, `ApplyTemplate`,
  `UpdateCopyrightYear`) with a rename (`project-autofix-sdk` or
  `autoconfigure-sdk`). Evaluate against the first-consumer milestone, or
  formally reject and keep the linter scope.

## Open questions (owner decisions)

- **Q1 — fate of the public `docs/` tree.** `docs/planning` (incl. the
  internal `licenseforge` strategy analysis), `docs/reviews/*.html`, and
  `docs/status/*` became world-readable in the 2026-09-09 visibility flip.
  Keep (radical transparency), forward-delete (history retains them), or
  rewrite history? Risk-appetite call; see flip report section g.
  _T17 sweep (2026-09-10) resolved to ACCEPT all mentions: the names are
  irreversibly in public git history, gitleaks ran clean at the flip, and
  redacting working-tree copies would be cosmetic. Deletion remains covered
  by this open question._
- **Q2 — first tag timing.** Zero tags exist; pkg.go.dev serves only
  pseudo-versions until `v0.1.0` is cut. Is the current API surface ready to
  freeze, or should planned breaking changes land first?
- **Q3 — will the consumer tools ever go public?** The README justifies the
  SDK by referencing two repos. _Resolved 2026-09-10: both
  `golangci-lint-auto-configure` and `oxlint-auto-configure` are public
  (verified via GitHub API), so the README's consumer links resolve and the
  why-it-exists story stands on its own._

## Non-goals

- **YAML parsing in the SDK.** The tools use different YAML libraries by
  design; the SDK stops at shared byte-level reads.
- **A shared `ProjectType` enum.** Go project shape and JS framework
  detection are incompatible concepts; forcing one would be false convergence.
- **A second build system** (flake.nix, Makefile). BuildFlow owns the
  pipeline.
