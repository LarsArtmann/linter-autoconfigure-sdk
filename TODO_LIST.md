# TODO List

Short- and mid-term actionable work. Open items only — completed items live in
`CHANGELOG.md`. Long-term ideas and open questions live in `ROADMAP.md`.

2026-09-10 (second session) cut `v0.1.0`: CHANGELOG section cut, README/FEATURES
synced, tag pushed, proxy indexed, clean-dir `go get@v0.1.0` + consumer compile
verified, GitHub Release created (marked pre-release per v0.x convention), and
required status checks added to `master` (T4, T22, most of T24 done). The
2026-09-10 first session closed T1, T2, T3, T7, T8, T9, T11, T12, T13, T14,
T15, T16, T17, T18, T19 — see CHANGELOG `[0.1.0]`. 2026-09-11 closed T24: the
pkg.go.dev `v0.1.0` page renders (two-value `SaveJSON`, `ExampleSaveJSON`).
2026-09-11 (second session) closed T23: the redesigned social preview is
LIVE — the repo page's og:image serves the uploaded card byte-identical
(57,801 bytes, repository-images.githubusercontent.com), verified via
generate.sh `--verify`.

HARVEST ruling (2026-09-11, defaulted per the 09:14 status report g3 after
the question went unanswered twice): bounded f-list items were harvested into
this table (T27–T30); unbounded ideas stay in ROADMAP.md. User-blocking
g-questions defaulted as: social channels unknown → GitHub/README are the
surfaces that matter until answered; empirical GIF validation = T30 (needs
your hands, not mine).

2026-09-22 (masterplan session): T26 SUPERSEDED — v0.2.0 and v0.3.0 are cut
(see CHANGELOG); oxlint already pins tags without a replace, and the remaining
replace-drop/bump work moved to T31/T32. T31–T43 are harvested from
`docs/planning/2026-09-22_23-48_SDK-ABSORPTION-MASTERPLAN.md`, which owns the
full Pareto breakdown, execution graph, verification strategy, and the
consumer-repo tasks (oxlint/golangci/BuildFlow migration steps).

| ID  | Task                                                                                                                                                                                                                                                                     | Impact   | Effort | Evidence / Source                                                                            |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------- | ------ | -------------------------------------------------------------------------------------------- |
| T20 | Remove the deprecated `ErrNoRepair` alias at v1 (compile-compat shim; SDK never returns it; canonical signal is nil `toolsdk.Spec.Repair`)                                                                                                                               | Low      | S      | `autoconfigure.go` ErrNoRepair; CHANGELOG `[0.1.0]`                                          |
| T21 | Switch CI to `buildflow --fix --fail-on-findings` now that BuildFlow is PUBLIC (`go mod download github.com/larsartmann/buildflow@v0.6.0` verified 2026-09-11 — the privacy blocker is LIFTED; remaining work is the ci.yml workflow swap, which needs a watched CI run) | High     | S      | `.github/workflows/ci.yml`; proxy check 2026-09-11                                           |
| T27 | Script the release verification (clean-dir `go get@vX.Y.Z` + consumer compile) so it is one command, not AGENTS.md prose                                                                                                                                                 | Medium   | S      | AGENTS.md Releases section; 0.1.0 runbook                                                    |
| T28 | Guard README code snippets against drift (compile the README's Go snippets in a test)                                                                                                                                                                                    | Medium   | S      | README install/usage blocks; report f18                                                      |
| T29 | CI guard for social-preview assets (render from svg, assert 1280x640 + < 1 MB before push)                                                                                                                                                                               | Low      | S      | assets/branding/; SKILLS generator machine-checks are the local half (report f19)            |
| T30 | Empirical GIF validation: one scratch-repo upload (GitHub card slot) + one Discord/Slack post — needs YOUR hands                                                                                                                                                         | Low      | S      | assets/branding/social-preview-animated.gif; platform matrix claims are doc-grade until then |
| T31 | Ship SDK `v0.3.1` (determinism fix, commits `fc9aacf`+`d872d1e`): CHANGELOG + AGENTS updates done, push master, tag, proxy + clean-dir verify — TAG TIMING IS AN OWNER GATE (plan Q1)                                                                                    | Critical | S      | plan §3 P01; CHANGELOG `[Unreleased]`                                                        |
| T32 | Consumer bumps to `v0.3.1`: oxlint (go get + flake input + vendorHash + vendor), golangci (drop local `replace`, go.mod:68), BuildFlow (indirect, tools + execution go.mod)                                                                                              | Critical | M      | plan §3 P02–P04; all three go.mods pin v0.2.0 (2026-09-22)                                   |
| T33 | Add `MarshalJSONIndented` + `ParseJSON[T]` + `SaveJSONBytes` (bytes/paths matrix completion); refactor `SaveJSON` onto the shared marshal helper; tests + godoc examples                                                                                                 | Critical | M      | plan §3 P05; 4 duplicated marshal-opts copies across consumers (plan §1.2 D1–D3)             |
| T34 | Add `WorkingDir(ctx)` helper (`finding.WorkingDirFromContext` + `"."` fallback; absent in go-finding/toolsdk too)                                                                                                                                                        | High     | S      | plan §3 P06; oxlint provider.go:68 hand-rolls it                                             |
| T35 | oxlint: migrate `ToJSON`/`FromJSON`/`writeConfig`×2 onto T33/T34 APIs, golden-byte pinned; drops its direct go-atomic-write dep                                                                                                                                          | High     | M      | plan §3 P07; cmd_configure.go:293 + provider.go:300 bypass `SaveJSON`                        |
| T36 | Add generic config-diff engine: `Change{Kind,Path,Old,New}` (no dead `Unchanged` state), `DiffMaps`/`DiffSets`/`DiffBlobs`, `Summary`/`FormatDiff` — kills the oxlint↔golangci diff split brain                                                                          | Critical | L      | plan §3 P08; both repos' `pkg/diff` (plan §1.2 D6)                                           |
| T37 | Migrate both consumers' `pkg/diff` onto T36 (oxlint `Rule`→`Path` rename ripple; golangci int `ChangeType`→`Kind`, `Description`→derived)                                                                                                                                | High     | M      | plan §3 P09–P10                                                                              |
| T38 | `ProviderSpec.ConfigFiles []FilePath` + `FirstExisting` helper + `Inputs` derivation (multi-candidate config discovery; decide single-`ConfigFile` v1 deprecation)                                                                                                       | High     | M      | plan §3 P11; oxlint `hasConfig` + forced `Inputs` override                                   |
| T39 | Release `v0.4.0` (T33–T38) + docs wave: README API tables, FEATURES, example_test, ROADMAP graduation                                                                                                                                                                    | Critical | S      | plan §3 P13–P14                                                                              |
| T40 | `ProviderSpec.Generate` hook: SDK-owned Detect-missing / never-overwrite dry-run-aware Repair / advisory drift HealthCheck on T36; oxlint provider migrates (~150 lines deleted); release `v0.5.0`                                                                       | High     | L      | plan §3 P15–P16, P19; oxlint provider.go:138–262                                             |
| T41 | `Deterministic(true)` enforcement analyzer (`json.Marshal` without the option) wired into SDK + both consumers; bite-checked against the original `SaveJSON` gap                                                                                                         | Medium   | L      | plan §3 P18; oxlint report f24                                                               |
| T42 | golangci atomic-write migration: audit `os.WriteFile` sites, migrate config writes to `SaveJSONBytes`/`WriteWithPerm` preserving 0600 perms                                                                                                                              | Medium   | L      | plan §3 P17; loader.go:460 writes non-atomically                                             |
| T43 | `pkg/format` disposition: propose `FindingView`/`PrintSummary`/`PrintFindingsTable` + exported marshal-opts helper upstream to go-finding                                                                                                                                | Low      | M      | plan §3 P20; go-finding json.go opts are unexported                                          |
