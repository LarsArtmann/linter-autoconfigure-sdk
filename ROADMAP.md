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

_Graduated 2026-09-23: v0.4.0 shipped the first consumer migration, the
`ProviderFromSpec` BuildFlow adapter, and `ProviderSpec` validation; v0.5.0
shipped the bootstrap provider lifecycle (`BootstrapSpec[T]` /
`BootstrapProviderFromSpec`, TODO_LIST T40) and v0.6.0 the `jsondeterminism`
analyzer + vettool cmd (T41); T43's format disposition is decided and drafted
(docs/planning/2026-09-23_pkg-format-disposition.md — upstream proposal
pending). The candidates below are still raw._

- **Write variants.** `SaveYAML` counterpart (per-tool YAML libs make this
  tricky), `SaveJSONCompact` for machine-only configs,
  `ReadConfigWithFingerprint` for read-modify-write transactions,
  `WithIndent` option for `SaveJSON`, `LoadJSONWith[T]` accepting
  `jsontext.Options` (e.g. `RejectUnknownMembers`).
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
- **Q2 — first tag timing.** _Resolved 2026-09-10: `v0.1.0` cut; the surface
  froze then and evolves via SemVer 0.x minor bumps (v0.4.0 adds the I/O
  matrix, diff engine, and ConfigFiles)._
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
