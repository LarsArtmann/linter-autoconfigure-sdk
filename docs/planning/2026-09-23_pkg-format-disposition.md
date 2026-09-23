# pkg/format Disposition — Decision Matrix + go-finding Proposal Draft

_Date: 2026-09-23 · Task: TODO_LIST T43 (masterplan P20)_
_Input: oxlint-auto-configure `pkg/format/format.go` (199 lines), go-finding v1.13.0 API surface (verified against the module cache)._

## Decision matrix

| Candidate | Upstream to go-finding? | Absorb into SDK? | Keep local? | Decision + why |
|---|---|---|---|---|
| `FindingView` / `SummaryView` projection types | YES — generic over ANY tool's findings; go-finding owns the `Finding` model, so its rendering projection belongs beside it | NO — the SDK is config-plumbing, not finding rendering; absorbing it would grow the SDK past its README contract | interim | **Propose upstream.** Every auto-configurer that renders findings (oxlint today, golangci's report writers conceptually) re-derives the same projection |
| `PrintSummary` / `PrintFindingsTable` / `PrintFindingsJSON` | YES — human rendering of the shared model; the top-N-by-rule/file rollups are tool-agnostic | NO — same boundary argument | interim keep | **Propose upstream.** Keep local until go-finding lands an equivalent, then delete |
| Exported marshal-opts helper (`marshalOpts`/`prettyMarshalOpts` in go-finding `json.go:12-20`, verified unexported) | YES — trivial export; today every LarsArtmann consumer re-declares these options (the SDK's `MarshalJSONIndented`, oxlint's format/report writers) | MOOT — the SDK keeps its own canonical opts regardless (its 2-space+Deterministic form is its public byte contract since v0.1.0) | — | **Propose upstream.** Reduces copy-paste drift across the fleet; the SDK does not switch to it (byte-contract ownership) |
| `Map`/`formatMap` string helpers | NO — trivial local glue | NO | YES | **Keep local** — 10 lines, zero reuse value |
| `topN`/`sortedKeys`/`namedCount` internals | ride along with PrintSummary | NO | — | **Ride along** with the upstream proposal |

## Wiring plan after upstream adoption

1. go-finding ships `findingview` (or extends `report`): view types + printers.
2. oxlint deletes `pkg/format` (or reduces it to oxlint-specific column tweaks).
3. golangci's report writers evaluate adoption independently (its report JSON schema is its own).

## go-finding issue draft (filed by owner or next session; text ready)

> **Title:** Proposal: exported finding-rendering helpers (view types + printers) and exported marshal options
>
> **What:** `pkg/format`-style rendering is re-implemented per consumer. oxlint-auto-configure carries `FindingView`/`SummaryView` + `PrintSummary`/`PrintFindingsTable`/`PrintFindingsJSON` (199 lines) — projections of `finding.Finding` that any findings-producing tool needs. Separately, `json.go`'s `marshalOpts`/`prettyMarshalOpts` (deterministic map keys, 2-space indent) are unexported, so every downstream repo re-declares the same options.
>
> **Why:** the projections and printers are model-adjacent (they encode what a "finding summary" IS), so they drift per consumer otherwise. The marshal options are already the fleet's de-facto standard bytes; exporting them removes N copies.
>
> **Evidence:** oxlint `pkg/format/format.go` (types at :23/:37, printers at :49/:147/:165); go-finding `json.go:12-20` (unexported opts). linter-autoconfigure-sdk deliberately keeps its own `MarshalJSONIndented` (public byte contract), so this is additive, not a migration.
>
> **Proposed shape:** `finding/view.go` or subpackage: `type View struct{...}`, `func PrintSummary(w io.Writer, ...)`, `func PrintTable(...)`, plus `func MarshalOpts() []json.Options` / `PrettyMarshalOpts() []json.Options` exports.
>
> **Non-goals:** SARIF (already owned by `report.ToSARIF`), pipeline, tool-specific columns.

## Verification of claims (done before this doc)

- go-finding v1.13.0 root package has NO `FindingView`/`PrintSummary` (grep over module cache).
- `marshalOpts`/`prettyMarshalOpts` unexported at `json.go:12-20` (read).
- oxlint `pkg/format` surface and line counts (read; 199 lines).
