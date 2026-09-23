# SDK Absorption — Full Masterplan Executed (T31–T43 + P21–P23 Closed)

**Date:** 2026-09-23 (session ~03:30–05:30 CEST)
**Plan:** `docs/planning/2026-09-22_23-48_SDK-ABSORPTION-MASTERPLAN.md` (all three trains)
**Predecessor:** `2026-09-23_03-14_execution-trains-1-2-shipped.md` (trains 1–2 + the 3 parked owner gates)

---

## a) The parked gates g1–g3, resolved by the owner's blanket "GET SHIT DONE"

- **g1 (push oxlint's 5 commits):** moot — the pma daemon had already pushed them by session start (verified `origin/master..HEAD` empty). The gate re-ran green before the v0.8.0 release anyway.
- **g2 (oxlint v0.7.2 vs v0.8.0):** v0.8.0 (minor), matching the prior session's lean — user-visible deltas (Summary wording, provider Inputs 2→4, override-drift rendering).
- **g3 (BuildFlow):** waited for the parallel session's wave-5 close (verified green + pushed), then took the SDK indirect bump (v0.4.1, later v0.6.0) without touching their consumer-input pairing.

## b) Releases shipped this session (7 tags, all verified)

| Repo | Tag | What | Verification |
|---|---|---|---|
| SDK | **v0.5.0** | Bootstrap provider lifecycle (`BootstrapSpec[T]` + `BootstrapProviderFromSpec`) | pre-tag adversarial smoke (clean-dir, local replace — the v0.4.0 lesson applied), CI green incl. new API, proxy, `scripts/verify-release.sh`-equivalent clean-dir run, GitHub release Latest, pkg.go.dev renders |
| SDK | **v0.6.0** | `jsondeterminism` analyzer + `cmd/jsondeterminism` vettool | bite-check analysistest (reverted-gap fixture MUST flag), offender exit-1 / clean exit-0 e2e, CI green with the analyzer as a step, proxy + clean-dir compile-all-APIs, release Latest, pkg.go.dev renders fully |
| oxlint | **v0.8.0** | SDK v0.4.1 migration (I/O, diff engine, ConfigFiles; flake input v0.4.1 + vendorHash fixed on the way) | full pre-release gate green, CI green, GoReleaser release SUCCESS with complete artifact set, Latest |
| oxlint | **v0.9.0** | provider onto `BootstrapProviderFromSpec` (~110 lines deleted, `provider_test.go` semantically unchanged) + `validate` drift advisory (`--fail-on-drift`) | gate green, CI green, Release SUCCESS, Latest |
| oxlint | **v0.9.1** | jsondeterminism gate wiring + SDK v0.6.0 (2 real test-fixture findings fixed) | gate green (now includes the analyzer), CI green, Release SUCCESS, Latest |
| golangci | **v0.9.0** | diff-engine alias (`ChangeType = autoconfigure.Kind`), stranded v0.8.2 changelog entries repaired into their section | gate 9/9, CI green (after a stale `vendorHash.nix` fix), Release SUCCESS, post-release verify 9/9 |
| golangci | **v0.10.0** | atomic config/backup writes (audit table in changelog) + fleet-wide `json.Deterministic(true)` (36 sites) + analyzer gate + SDK v0.6.0 | gate 9/9, CI green, Release SUCCESS, post-release verify 9/9 |
| BuildFlow | (no tag) | SDK indirect v0.3.1 → v0.4.1 → v0.6.0 (3× go.mod + flake + lock) | `nix build .` green each time, pushed after their wave landed |

## c) T40–T43 outcomes

- **T40 (bootstrap lifecycle):** design decision — a separate `BootstrapSpec[T]` type, NOT a `Generate` field on `ProviderSpec` (two lifecycles; one struct offering both invites ambiguous specs). 18 contract tests ported from oxlint's `provider_test.go` semantics (no-overwrite, dry-run, shadow-config, drift advisory, malformed, normalize-suppresses-drift, sentinel match). oxlint migrated: 304 → 194 lines in provider.go.
- **T41 (determinism enforcement):** analyzer-vs-helper decision → analyzer + `cmd` vettool (no golangci plugin fragility; runs via each repo's existing gate/CI). Conservative by design: `opts...` spreads unflagged; `Deterministic(false)` = explicit opt-out. First consumer runs caught 2 (oxlint) + 36 (golangci) real findings.
- **T42 (atomic writes):** audit table + disposition in golangci CHANGELOG. `config.NewOSFS()` exposes the atomic default FS; report writers / git hook / dev codegen documented-keep with reasons. Crash-safety test: no temp residue, full replacement, 0600 preserved.
- **T43 (format disposition):** matrix + go-finding issue draft in `docs/planning/2026-09-23_pkg-format-disposition.md` (claims verified against the module cache: opts unexported at go-finding json.go:12-20; no view types upstream). Filing is one `gh issue create` away — left for the owner to file in their voice.

## d) P21 carried TODOs

- **T27:** `scripts/verify-release.sh vX.Y.Z` — proxy check + clean-dir `go get` + compile-every-exported-API smoke (verified against v0.6.0).
- **T28:** `readme_snippets_test.go` compile guard — **caught real drift on its first run** (`FindingsFromIssues` two-value return + non-compilable illustrative closures); README snippets now compile-true.
- **T29:** `scripts/check-social-preview.sh` (1280x640 + <1MB via PNG IHDR parse — endianness bug caught and fixed on the way; SVG parity check when rsvg-convert present). Wired into CI.
- **T21:** buildflow CI job was implemented and live-tested — **it failed because `LarsArtmann/BuildFlow` is PRIVATE again** (API-verified 2026-09-23; the 2026-09-11 "public" verification is stale). Job reverted; T21 reinstated in TODO_LIST with the corrected evidence. The local pipeline itself is now FULLY green: `license-check` (go-licenses can't parse go 1.27 stdlib) and `go-structure-linter` (intentional flat layout) are skipped via `.buildflow.yml` `skip_steps` with documented reasons — first 100%-green buildflow in repo history.

## e) Honest fuckups this session

1. **The analyzer's spread case didn't work on first write** — the Ellipsis check was missing and the bite-test caught it immediately (testdata `goodSpread` flagged wrongly). Bite-checking works.
2. **My smoke harness had the bug, not the SDK, twice** (NormalizeExpected copying a missing key as ""; a `must`/`mustErr` type mismatch) — each caught by the smoke before anything shipped.
3. **`golangci-lint` direct runs surfaced 3 findings the earlier buildflow runs had masked** (the "9 tools unavailable" health-check state) — ll/varnamelen/err113 in new test code, plus a pre-existing SA1012 on the v0.4.1 nil-ctx regression test (nolint'ed with reason).
4. **`gh release create` for v0.6.0 silently skipped** when chained after a failing `go run` with `&&` — noticed via the `releases/latest` check and re-run. Release-verify discipline (check Latest pointer) caught it.
5. **README Status section said "v0.5.x" while shipping v0.6.0** — fixed in the closing sweep; pkg.go.dev v0.6.0 served the stale text for the tag (immutable; acceptable for a status line).

## f) End state

- TODO_LIST: only T20 (ErrNoRepair removal at v1, owner-gated) and T30 (GIF empirical validation, needs owner's hands) remain.
- All four repos: clean trees, zero unpushed commits, CI green at HEAD.
- SDK buildflow: exit 0. oxlint gate: green including the new analyzer step. golangci gate: 9/9 twice.
- pkg.go.dev: v0.6.0 fully rendered (bootstrap + analyzer docs, examples); the module's version ladder is proxy-complete through v0.6.0.
