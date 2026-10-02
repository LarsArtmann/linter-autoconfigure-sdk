# toolsdk Deep-Dive Session — Status & Self-Review

**Date:** 2026-10-03 01:01 · **Scope:** this session only (release-worthiness
check → toolsdk utilization audit → panic fix → docs)
**Series:** follows `2026-09-23_15-47_release-v0.7.0-status-and-self-review.md`
**Format note:** written as `.md` at the user's explicit path demand; the
status-report skill's HTML default is overridden for this one report only.

---

## a) FULLY DONE

1. **Release-worthiness verdict for post-v0.7.0 commits**: zero Go source
   changes since the tag — only toolsdk v1.13.0→v1.14.0 + go-error-family
   floor bumps, AGENTS.md, a status doc, daemon metadata churn. Verdict:
   nothing release-worthy yet; consumers get toolsdk v1.14.0 via MVS anyway.
2. **toolsdk utilization audit (library-deep-dive skill, all 7 phases)** —
   `docs/research/2026-10-03_go-finding-toolsdk-deep-dive.html`:
   - Phase 1: cataloged the touched surface (5 of 10 `Spec` fields,
     `RepairerFunc`, `RepairResult`, `DryRunFromContext`) across
     autoconfigure.go / bootstrap.go + 2 test files.
   - Phase 2: full v1.14.0 surface via `go doc -all` of the pinned module
     cache; toolsdk/CHANGELOG 1.12→1.14; BuildFlow consumption semantics
     verified in source (HealthCheck is warn-only at
     execution/pipeline.go:88-90, surfaced in the summary; `ToolFromSpec`
     converts every Spec field); both consumer repos read.
   - Phase 3/4: gap analysis + scoring → adoption 80/100 post-fix.
   - Discovery: a **third consumer** (go-version-auto-configure, vendored by
     BuildFlow) using `ProviderFromSpec` — was missing from AGENTS.md's
     consumer inventory. Now recorded.
3. **Confirmed + fixed a real bug**: `Repair(nil)` on bootstrap specs
   panicked in `toolsdk.DryRunFromContext` (bootstrap.go:219, pre-fix) —
   empirically reproduced against **released v0.7.0** in a clean module
   (panic stack: toolsdk dryrun.go:17 ← bootstrap.go:219). The v0.4.1
   `WorkingDir(nil)` hotfix bug class, missed at that sweep. Fix: all five
   derived entry points (both Detect adapters, both repair closures, the
   drift HealthCheck) now normalize via `toolsdk.EnsureContext`; consumer
   closures never observe a nil ctx.
4. **Regression tests**: `TestBootstrapNilContext` (Detect+Repair+HealthCheck
   under nil ctx, t.Chdir to temp) and
   `TestProviderFromSpec_NilContext_NormalizedBeforeClosures` (closures
   assert non-nil ctx).
5. **Gates green**: `go test -race -count=1 ./...` ok; buildflow golangci-lint
   step clean after two rewrites (see d); plain-mode pipeline exits 0
   (40 success / 0 failed).
6. **Docs updated in the right files**:
   - TODO_LIST: **T22** filed (bridge the declarative surface:
     ExtraInputs / HealthCheck on ProviderSpec / Trigger / DependsOn /
     ModuleFanOut) with consumer evidence.
   - AGENTS.md: audit summary, third consumer, nil-ctx contract.
   - CHANGELOG `[Unreleased]/Fixed`: the panic fix entry.
   - AGENTS.md strict baseline re-verified and rewritten: **9 → 17 known
     advisories**; verified via a temp worktree on commit 68cc3d7 that all
     8 art-dupl + 2 stdlib2lo findings pre-date this session — the jump is
     the buildflow binary upgrade (art-dupl newly reporting; go-auto-upgrade
     renamed stdlib2lo), NOT my code. Worktree removed cleanly.

## b) PARTIALLY DONE

1. **T22** — proposed with full evidence and target API sketch (report
   §Proposal), deliberately NOT implemented: additive public-API growth on
   both spec types is an owner-level design call before the v1 freeze. Needs
   your go/no-go.
2. **Consumer-side findings** (go-version-auto-configure: hand-rolled
   nil-unsafe `WorkingDir` copy at provider.go:135-144; Inputs
   under-declaration: reads go.work/.github/nix, declares `go.mod` only →
   weak BuildFlow result-cache invalidation; oxlint: post-hoc
   `spec.Inputs` prepend) — identified and documented, fixes not applied
   (other repos).
3. **Nil-ctx contract documentation** — godoc updated on the two bridge
   functions, but **FEATURES.md and README were NOT updated** with the new
   standalone-caller guarantee. Drift risk: FEATURES says nothing either
   way; README shows provider examples that now have a stronger contract.

## c) NOT STARTED

1. **Consumer compile-verification against the modified SDK working tree**
   (replace-directive build of oxlint + go-version). Changes are additive
   with no signature changes, so risk is ~0, but I asserted that mentally
   instead of compiling. Cheap, should be done before the next tag.
2. **Options bridging** (toolsdk v1.14.0 `Spec.Options`) — deferred by
   design, no consumer demand (report finding 5).
3. **Export of `errConfigDrift`** — export-on-demand decision re-affirmed;
   no consumer matches on it today (searched both repos).
4. **README/FEATURES entries** for the nil-ctx guarantee (see b3).

## d) TOTALLY FUCKED UP

1. **The buildflow --fix autofight (real, self-inflicted, recovered)**: I
   wrote the regression tests with an outer-variable capture pattern
   (`analyzeCtx = ctx`) and long trailing `//nolint` comments. The fatcontext
   autofix rewrote the captures into shadowing `:=` declarations (**breaking
   compilation**), and golines reflowing dropped two nolint comments. I then
   ran the fixer AGAIN before re-reading its output, compounding it. Net:
   two extra rewrite cycles that a moment of forethought would have avoided.
   The final form (typed-nil `var nilCtx context.Context` + closure-internal
   `t.Error` assertions) was autofix-stable on the first try.
   **Lesson:** in buildflow-covered repos, write lint-autofix-stable tests
   from the start; after ANY `--fix` run, re-read the diff before trusting
   green.
2. **First test attempt had a missing `context` import** — background run
   failed to compile. Sloppy first draft, fixed in one pass.

## e) WHAT WE SHOULD IMPROVE (session lessons)

1. **Load the buildflow skill BEFORE the first buildflow invocation**, not
   when first reaching for `--fix`. I invoked single-step golangci runs
   before loading the skill (loaded it late, got lucky).
2. **Anticipate autofixers** (see d1): prefer patterns linters cannot
   "improve" into bugs; typed-nil variables over literal-nil + nolint.
3. **Proactive findings-diff discipline**: I only chased the 9→17 advisory
   jump because strict mode failed. Better: capture a findings snapshot
   before starting work in a session that will touch code, so "did I add
   findings?" is a diff, not an investigation (the worktree trick worked but
   was reactive).
4. **Report rigor**: the deep-dive's "7/10 capabilities" denominator is not
   itemized anywhere in the HTML; the adoption score (80/100) has no stated
   weight formula. Both are defensible but challengeable. Next report:
   enumerate the denominator and the weighting.
5. **Stale LSP diagnostics** showed pre-edit positions all session; I
   distrusted them correctly but never ran `lsp_restart`. Minor hygiene.
6. **FEATURES.md should have been updated in the same breath as CHANGELOG**
   — the nil-ctx contract is a user-visible behavior guarantee; FEATURES is
   the honest inventory. Two-minute fix, missed.

## f) Up to 50 things to get done next

Already tracked (TODO_LIST): **T22** (declarative-surface bridge fields,
High/M), **T21** (CI buildflow job, blocked on private BuildFlow), **T20**
(remove ErrNoRepair at v1), **T30** (social GIF validation, needs your
hands), **T44** (dprint formatter-of-record reconciliation).

Session-derived (unrouted — HARVEST candidates):

7. Cut **v0.8.0** (or v0.7.1) carrying the nil-ctx fix; run
   `./scripts/verify-release.sh` + adversarial clean-dir smoke BEFORE
   tagging (v0.4.0 lesson).
8. Update **FEATURES.md** with the nil-ctx standalone-caller guarantee.
9. Update **README** provider-bridge section with the same contract.
10. **Consumer compile-check**: build oxlint-auto-configure and
    go-version-auto-configure against this working tree (replace directive)
    before tagging.
11. go-version-auto-configure: delete hand-rolled `workingDir`, call
    `autoconfigure.WorkingDir` (also fixes its own nil-ctx panic risk).
12. go-version-auto-configure: declare full Inputs read surface
    (go.work, .github/workflows/*, flake.nix/other nix pins).
13. oxlint-auto-configure: move `spec.Inputs` prepend into T22's
    `ExtraInputs` once it exists.
14. Decide T22 scope with this session's evidence (question g1).
15. Add a `lsp_restart` habit when diagnostics contradict the gate (or
    verify the golangci_lint_ls staleness root cause once).
16. Document the "autofix-stable test patterns" lesson in AGENTS.md
    conventions (typed-nil var; closure-internal assertions; short nolint).
17. Re-check the 17-advisory baseline after the next buildflow upgrade
    (art-dupl/stdlib2lo churn will move again).
18. Evaluate the 8 art-dupl test clones once (10-14 tokens in *_test.go
    setup): accept as table-test noise in AGENTS.md, or extract a tiny
    helper — currently they are unadjudicated.
19. Consider a findings-snapshot helper (buildflow --format finding to a
    file) as a pre-session habit in this repo.
20. CHANGELOG: decide whether the toolsdk v1.14.0 floor bump deserves an
    Unreleased "Changed" line at release time (currently only the fix is
    logged).
21. In the deep-dive report: itemize the 10-capability denominator + weight
    formula (appendix note, point-in-time doc — annotate, don't rewrite).
22. Roadmap fuel: per-tool `Options` bridging design sketch (profile knob
    for oxlint) so T22 has a sibling when demand lands.
23. Roadmap fuel: biome-auto-configure as the third bootstrap consumer —
    the SDK's stated purpose; nothing blocks it but T22 would make it a
    pure config-schema exercise.
24. When T22 lands: add `TestProviderFromSpec_TriggerPassThrough` &
    friends (field-forwarding table tests).
25. When T22 lands: update both consumers' provider.go to drop post-hoc
    mutations (one commit each, their repos).
26. Verify BuildFlow's vendored copies of both consumer providers get
    re-vendored after consumer releases (BuildFlow pins SDK v0.6.0 indirect
    — bump rides the next BuildFlow train).
27. Check whether `docs/reviews/` brutal-self-review series (last
    2026-07-26) should get this session's episode appended as HTML next
    time the skill runs (this session folded it into the status report per
    your instruction).

(27 items; the remaining headroom of 50 would be padding — declined. All
items above are session-observable; nothing was researched beyond this
session's scope per your instruction.)

## g) Questions I cannot answer myself

1. **T22 green light?** Grow `ProviderSpec`/`BootstrapSpec` with the five
   pass-through fields (ExtraInputs, HealthCheck, Trigger, DependsOn,
   ModuleFanOut) before v1 freezes the API — or keep the bridges minimal
   and keep blessing post-hoc mutation? Both defensible; it is your API.
2. **Release timing:** cut a patch/minor release for the nil-ctx fix now,
   or hold and ship it together with T22 so consumers take one bump?
   (The fix is in-tree and green either way.)
3. **Cross-repo routing:** apply the two go-version-auto-configure fixes
   (WorkingDir copy, Inputs under-declaration) from here in that repo now,
   or route them into its own TODO_LIST for its next train?

---

*Point-in-time snapshot. Written from session memory + the session's own
artifacts; no new research performed. Waiting for instructions.*
