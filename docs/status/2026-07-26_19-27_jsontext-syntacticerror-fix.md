# Status Update — `jsontext.SyntacticError` fix in `linter-autoconfigure-sdk`

**Date:** 2026-07-26 19:27 UTC
**Session:** Single-task bug-fix session
**Author:** Crush (MiniMax-M3)

---

## TL;DR

Buildflow pipeline failed on 3 steps (`go-fix`, `govalid-generate`, `test-race`) with
`undefined: json.SyntaxError` at `autoconfigure_test.go:167:34`. Root cause: the test
referenced a v1 `encoding/json` type while the SDK uses `encoding/json/v2` (gated by
`GOEXPERIMENT=jsonv2`). jsonv2 emits `*jsontext.SyntacticError` instead. Replaced the
assertion, the import, and the godoc comments. All 3 originally-failing steps now pass.

---

## a) FULLY DONE

- Replaced `*json.SyntaxError` assertion with `*jsontext.SyntacticError` in
  `autoconfigure_test.go:167-168`.
- Replaced `encoding/json/v2` import with `encoding/json/jsontext` in
  `autoconfigure_test.go:5` (the v2 import was unused once the assertion moved to
  `jsontext`).
- Updated godoc for `ConfigError` type in `autoconfigure.go:52-58` to reference
  `errors.AsType[*jsontext.SyntacticError](err)` and explain the v1→v2 mapping.
- Updated godoc for `Unwrap/Is/As` methods in `autoconfigure.go:70-74` likewise.
- Updated `AGENTS.md:73-74` design-decision note to cite `*jsontext.SyntacticError`.
- Verified: `go build ./...` clean, `go test -race -count=1 ./...` passes,
  `buildflow -s govalid-generate` exit 0, `buildflow -s test-race` exit 0,
  `buildflow --fix` passes the 3 originally-failing steps.

## b) PARTIALLY DONE

Nothing — task was scoped tight and shipped end-to-end.

## c) NOT STARTED

- ~~The remaining 7 strict-mode (`--fail-on-findings`) findings from `buildflow`. They are
  **pre-existing environmental warnings**, not introduced by this session. Left alone
  per "don't fix unrelated bugs" rule. Listed in section (e).~~ done (state 2026-09-09: pipeline exit 0; strict lane still red on pre-existing errcheck/erraudit/MD013/lychee findings, tracked in TODO_LIST)
- ~~Stale LSP diagnostics in the tool output showing `jsontext.SyntacticError requires
go1.27` warnings — these are LSP cache artifacts. The actual `go test` / `buildflow`
  runs were clean.~~ done (still cosmetic as of 2026-09-09: the stdversion warnings appear in LSP output only; builds and tests are clean)

## d) TOTALLY FUCKED UP

Nothing broken. Diff is minimal and surgical (11 lines added, 7 removed across 3 files).

## e) WHAT WE SHOULD IMPROVE

### Leftover strict-mode findings (pre-existing, NOT introduced this session)

These all exist on master before this fix and were not touched. Flagging for the next
session:

1. **`erraudit` — 1 finding.** `autoconfigure.go:197` `FindingsFromIssues` returns generic
   `error` instead of a specific error type. Pre-existing pattern; the function
   legitimately wraps conversion failures via `fmt.Errorf("convert issue %q: %w", ...)`.
   The strict-typed-error convention enforced elsewhere in the file (`*ConfigError`)
   would suggest either a typed wrapper here or a documented exception. Decision needed:
   does the SDK treat batch-conversion failures as `*ConfigError`, or are they
   semantically different enough to warrant a new type (e.g. `ConversionError`)? I
   would have addressed this if it were not explicitly out of scope.
2. **`go-structure-linter` — 3 findings.**
   - Missing `.golangci.yml` (the SDK does not currently configure golangci-lint; the
     linter's own pipeline handles its config).
   - Missing `internal/` directory (project uses flat layout — `autoconfigure.go` and
     `autoconfigure_test.go` at the package root).
   - Missing `examples/` directory (despite `example_test.go` existing at repo root).
3. **`golangci-lint` — 2 findings.** `example_test.go:17` and `:33` have unchecked
   `os.RemoveAll` return values. Trivial to fix (`_ = os.RemoveAll(...)` or
   `t.Cleanup(...)` wrapper) but pre-existing.
4. ~~**`gomod-check` — 1 finding.** `go.mod:11` mixes direct and indirect requires in a~~ done (fixed — go.mod has separate blocks now)
   ~~single block (Go 1.17+ wants separate blocks). Buildflow's `go-mod-tidy` /~~
   ~~`go-mod-normalize` ran but did not split. Manual fix: re-run `go mod tidy` after~~
   ~~removing one dep from direct usage, or move the offending entry into the indirect~~
   ~~block.~~

### Things I did NOT do but probably should have

5. ~~Did not verify the fix against a sample run with `GOEXPERIMENT=jsonv2` set via a~~ **Won't implement — moot — the use_go_env direnv helper handles the environment.**
   ~~Nix shell — relied on the `.envrc` from the repo. If direnv were not active, the~~
   ~~test would fail differently (likely a clean build error from `encoding/json/v2`~~
   ~~itself). Worth documenting a one-liner "without direnv" test command.~~
6. ~~Did not add a CHANGELOG entry. `CHANGELOG.md` exists (per the git history) and this~~ done at `1c72b36`
   ~~is a behavior-affecting fix for downstream consumers (when they exist). Convention~~
   ~~uncertain — the recent `changelog` commit (1c72b36) suggests the file is curated.~~
7. ~~Did not check whether the planning doc~~ done (docs-health pass 2026-09-09)
   ~~`docs/planning/2026-07-26_06-05_make-architecture-and-data-model-superb.md:137`~~
   ~~references `json.SyntaxError` in any actionable way. The grep found it in the~~
   ~~F35 row ("Replace `errors.As(err, &syntaxErr)` with~~
   ~~`errors.AsType[*json.SyntaxError]` in tests"). That task is now already done by my~~
   ~~fix — should be marked complete or removed.~~
8. ~~Did not check the HTML review file `docs/reviews/2026-07-26_architecture-and-data-model.html`~~ **Won't implement — frozen point-in-time artifact, left alone by policy; its fate rides the docs/ decision (ROADMAP Q1).**
   ~~line 265 — same `json.SyntaxError` reference is in a frozen point-in-time artifact;~~
   ~~per `update-old-docs` skill policy these should not be rewritten in place but the~~
   ~~fact should be noted in a follow-up.~~
9. Did not add a regression test specifically for `*jsontext.SyntacticError` round-trip
   via the `errors.As`/`errors.Is` methods on `*ConfigError`. The existing test covers
   `errors.AsType`, but not the `(*ConfigError).As(any)` method which delegates to
   `errors.As(e.Err, target)`. These are equivalent in practice but the godoc promises
   both, so the assertion could be richer.
10. ~~Did not run `govalid-generate` output diff after the fix. The govalid tool generates~~ done (verified green through 2026-09-09)
    ~~code; if any generated file referenced `json.SyntaxError`, it would re-introduce~~
    ~~the compile error. `buildflow -s govalid-generate` succeeded, so this is fine, but~~
    ~~I did not inspect the diff.~~

### Process / meta observations

11. The bug surfaced because v1→v2 migration was done incompletely — the test was
    not updated when `LoadJSON` switched to `json.Unmarshal` from `encoding/json/v2`.
    This is the kind of split brain that a `go fix` modernization pass would catch. A
    project-level lint rule banning references to v1 `encoding/json` types when
    `encoding/json/v2` is imported would have caught this at edit time.
12. The fix is correct but the commit history will not show a clear "v2 migration
    regression" intent — the diff reads like a type rename. If a v2 migration is
    planned, this fix should ride that PR, not stand alone.
13. AGENTS.md was the right place to capture the design fact (jsonv2 → SyntacticError).
    Did not add a corresponding godoc note in `autoconfigure.go` about the jsonv2
    build-tag dependency (it is mentioned only in the package-level comment via
    `import "encoding/json/v2"`). Worth a sentence at the top: "Requires
    GOEXPERIMENT=jsonv2 in Go 1.26; GA in Go 1.27."
14. Did not consider whether a v1 fallback wrapper would be more useful for downstream
    consumers. Some consumers may be on Go 1.25 or have GOEXPERIMENT unset. Out of
    scope for this fix, but a strategic question.

## f) Up to 50 things we should get done next (session recommendations, sorted by impact)

These are project-level improvements visible from this session. Not all are mine to
pick up — flagging for triage.

### Immediate (this would have been in-scope if asked)

1. Fix the unchecked `os.RemoveAll` in `example_test.go:17` and `:33` — trivial, 2-line.
2. ~~Mark F35 in `docs/planning/2026-07-26_06-05_make-architecture-and-data-model-superb.md`~~ done (docs-health pass 2026-09-09)
   ~~as done (or remove) — done by this session's fix.~~
3. ~~Move the offending entry in `go.mod:11` into the indirect block (or do a fresh~~ done (go.mod now has separate direct and indirect require blocks)
   ~~`go mod tidy`) to clear the `gomod-check` warning.~~
4. Add `internal/` directory or document why the flat layout is intentional, to
   satisfy `go-structure-linter`.
5. Move `example_test.go` into an `examples/` directory (or add a stub), to satisfy
   the same linter.
6. Add a `.golangci.yml` at the repo root, even a minimal one, to silence the
   missing-config finding.

### Short-term (next session or two)

7. Decide on the error-type strategy for `FindingsFromIssues`: introduce a typed
   error (e.g. `ConversionError`), or document the exception in the function godoc
   and silence the `erraudit` finding per-line.
8. Add a regression test covering `(*ConfigError).As(any) bool` delegation to
   `errors.As(e.Err, target)` — currently only `errors.AsType` is asserted.
9. Add a one-line "Requires `GOEXPERIMENT=jsonv2`" note to the package-level doc
   comment in `autoconfigure.go:1-17`.
10. ~~Add a CHANGELOG entry for this fix.~~ done at `1c72b36`
11. Add a project-level lint rule (revive or custom analyzer) that flags references to
    `json.SyntaxError` when `encoding/json/v2` is imported in the same file.
12. ~~Audit the rest of the test file (`autoconfigure_test.go`) and main file for other~~ done (verified — no v1 encoding/json type references remain in *.go)
    ~~v1→v2 migration leftovers (e.g. `json.Marshaler`, `json.Unmarshaler`, `json.RawMessage`).~~
13. ~~Decide whether to add a v1 fallback wrapper for consumers on Go < 1.26 (likely~~ done (decided no — the SDK pins Go 1.26+ (README, AGENTS.md))
    ~~no — the SDK already pins to 1.26+ per AGENTS.md).~~
14. ~~Verify the `govalid-generate` output did not change after the fix (cosmetic, but~~ done (verified — buildflow govalid-generate green through 2026-09-09)
    ~~confirms no generated-code regression).~~
15. Add an integration test that confirms the full Save→Load→Malformed-Load cycle
    produces the expected `*ConfigError` chain end-to-end.

### Mid-term (next quarter)

16. Build out a first real consumer (`golangci-lint-auto-configure` or
    `oxlint-auto-configure`) that exercises `ConfigError`, `ConfigIssue`,
    `FindingFromIssue`, and `ProviderSpec.HasRepair` together.
17. Add BDD tests covering the auto-configure flow using onsi/ginkgo, per the
    `bdd-testing` skill convention.
18. Add a `pareto-planning` pass over the SDK to identify the next 20% of work that
    delivers 80% of consumer value.
19. Publish the SDK on pkg.go.dev once a tagged release exists (currently no tags on
    master).
20. Set up a CI workflow that runs `buildflow` on PRs (currently the project uses
    `buildflow` locally; CI integration is undocumented).
21. ~~Add `nix flake check` support for hermetic CI builds, mirroring the pattern from~~ **Won't implement — rejected — flake.nix explicitly excluded; buildflow owns the pipeline.**
    ~~sibling projects (per the `nix-review` / `nix-flake-migration` skill conventions).~~
22. Replace the remaining generic-`error` returns (`FindingsFromIssue`,
    `FindingsFromIssues`) with a typed batch-conversion error, mirroring
    `*ConfigError`.
23. Add a `go-finding` integration test that exercises the full finding-emission
    contract with BuildFlow's repair loop (no in-tree consumer exercises this yet).
24. Add `examples/` directory with a runnable example that wires the SDK into a
    BuildFlow detector + repairer pair, end-to-end.
25. ~~Write a `docs/DOMAIN_LANGUAGE.md` for the SDK (the AGENTS.md describes it but the~~ done at `e46c225`
    ~~domain glossary does not exist yet — per global AGENTS.md docs policy).~~
26. Run a `library-deep-dive` on `go-atomic-write` to confirm `SaveJSON` is using the
    library to its full potential (write-if-changed, atomic rename, fsync, concurrent
    modification detection are all there, but `Validate` / `Verify` / hash options
    are not used).
27. Run a `library-deep-dive` on `go-finding` to confirm `FindingFromIssue` uses the
    library's full builder API (currently uses `NewBuilder` + `WithFixStrategy` +
    `WithSuggestion` + `WithCategory` — verify nothing is missing).
28. Add a `name-review` pass over the SDK exports: `ConfigError`, `ConfigIssue`,
    `Op`, `ProviderSpec`, `FindingFromIssue`, `FindingsFromIssues`. Names are
    reasonable but `ProviderSpec` is generic; `RepairSpec` or `BuildFlowSpec` might
    be more honest.
29. Run `code-quality-scan` and `deduplicate-code` to surface duplication between the
    SDK and the planned consumers.
30. Run `full-code-review` on the SDK — only ~370 LOC + ~370 LOC tests, fast to
    review end-to-end.
31. Run `architecture-review` on the SDK to confirm the package boundary is correct
    for the planned consumer split.
32. Run `data-model-review` on `ConfigIssue`, `ConfigError`, `Op`, `ProviderSpec` —
    types are small but warrant a first-principles pass.
33. ~~Migrate from `.envrc` to `flake.nix` for Go toolchain + `GOEXPERIMENT=jsonv2`~~ **Won't implement — rejected — .envrc via the use_go_env direnv helper handles the env.**
    ~~consistency (per AGENTS.md, this project is automated with BuildFlow — no~~
    ~~`flake.nix` exists).~~
34. Add a `Makefile`-style convenience target via `buildflow -s <step>` aliases in
    README, so contributors know which step does what.
35. Add `gitleaks` integration (currently skipped via config per build output).
36. Add `gosec` or equivalent to the buildflow pipeline.
37. Add a benchmark for `SaveJSON` write-if-changed path (the mtime-preservation test
    exists but no `BenchmarkSaveJSON` does).
38. Add a fuzz test for `LoadJSON` against arbitrary malformed JSON to confirm
    `*ConfigError` wrapping never panics.
39. Add property-based testing (rapid or gopter) for the Save→Load round-trip
    invariants.
40. Update README.md with a "Quick start" section showing the typical 5-step consumer
    setup (per the `website-launch` skill pattern).

### Long-term / strategic

41. Make the SDK the canonical foundation for ALL linter auto-configurers
    (golangci, oxlint, biome, eslint, ruff, etc.). The package doc comments state
    this intent; needs at least one downstream consumer to validate the abstraction.
42. Document the SDK's compatibility policy: Go version, jsonv2 build tag, dependency
    floor on `go-finding` / `go-atomic-write` (currently "always latest", which is a
    stability risk for downstream).
43. Establish a semantic-versioning policy and tag the first 1.0.0.
44. ~~Add a public roadmap (the SDK has no `ROADMAP.md` yet — only planning/ docs).~~ done at `e46c225`
45. ~~Add a `CONTRIBUTING.md` for downstream consumers who want to add new auto-configurers.~~ done (exists at repo root; updated at e46c225)
46. Consider extracting the error-type machinery (`ConfigError`, `Op`, `Unwrap/Is/As`)
    into a separate sub-package (`configerr`) so other packages can reuse it without
    pulling in the auto-configure domain types.
47. Consider adding `MustLoadJSON` / `MustSaveJSON` panic-on-error variants for use
    in `init()` (not yet needed, but the SDK will likely grow test fixtures that need
    it).
48. Add a `WithIndent` option to `SaveJSON` so consumers can match the project's
    style (currently hardcoded to `"  "`).
49. Add a `LoadJSONWith[T any]` variant that accepts a `jsontext.Options` so consumers
    can opt into `RejectUnknownMembers` etc.
50. Add a `ProviderSpec` validation function that catches obvious misconfiguration
    (e.g. `Repair` returns a non-nil string error with no clear contract).

## g) Questions I CANNOT figure out myself

These require human judgment or external context I do not have:

**Q1.** Should `FindingsFromIssues`'s generic `error` return be retrofitted to a typed
error (introducing a new `ConversionError` or similar), or should the SDK document this
single function as an explicit exception to the typed-error convention? The function
legitimately wraps a `fmt.Errorf` for batch conversion, but the project's strict
`erraudit` finding wants a specific type.

**Q2.** Should this fix ride on a future v2-migration PR (with CHANGELOG entry framed
as "jsonv2 alignment"), or stand alone as a regression fix? The diff is small and the
intent is a v2 type rename — bundling or splitting affects how downstream consumers
will read the commit log when the SDK gains its first tag.

**Q3.** Are there planned consumers for this SDK that I should be aware of when
deciding the API surface? The package doc mentions `golangci-lint-auto-configure`,
`oxlint-auto-configure`, and "future additions like `biome-auto-configure`" — if any of
these are already in development elsewhere on this machine, knowing their shape would
let me predict which SDK exports are stable vs still malleable.

---

## Resolution (2026-09-09, docs-health pass)

19 items resolved inline (f: 11, e: 6, c: 2). Open items are tracked in
`TODO_LIST.md` (T5, T7, T10, T11) and `ROADMAP.md` (reviews, fuzz/property
tests, consumer milestone, write variants). Items e11-e14 are process/meta
observations, deliberately left unannotated. The g-questions map to: Q1 →
open (erraudit still reports 2 findings), Q2 → moot (no tag cut yet, commit
stands alone), Q3 → ROADMAP "first consumer migration" milestone.
