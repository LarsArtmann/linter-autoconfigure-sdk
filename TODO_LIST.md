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

| ID  | Task                                                                                                                                                                                                                                                                     | Impact | Effort | Evidence / Source                                                                            |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------ | ------ | -------------------------------------------------------------------------------------------- |
| T20 | Remove the deprecated `ErrNoRepair` alias at v1 (compile-compat shim; SDK never returns it; canonical signal is nil `toolsdk.Spec.Repair`)                                                                                                                               | Low    | S      | `autoconfigure.go` ErrNoRepair; CHANGELOG `[0.1.0]`                                          |
| T21 | Switch CI to `buildflow --fix --fail-on-findings` now that BuildFlow is PUBLIC (`go mod download github.com/larsartmann/buildflow@v0.6.0` verified 2026-09-11 — the privacy blocker is LIFTED; remaining work is the ci.yml workflow swap, which needs a watched CI run) | High   | S      | `.github/workflows/ci.yml`; proxy check 2026-09-11                                           |
| T26 | Cut `v0.2.0` (ConfigIssue.Confidence/FixStrategy, validation sentinels, lint-clean codebase, first consumers), then drop both consumers' local `replace` directives and `go get` the tag (AGENTS.md release runbook)                                                     | High   | S      | CHANGELOG `[Unreleased]`; replaces in golangci/oxlint-auto-configure go.mod (2026-09-11)     |
| T27 | Script the release verification (clean-dir `go get@vX.Y.Z` + consumer compile) so it is one command, not AGENTS.md prose                                                                                                                                                 | Medium | S      | AGENTS.md Releases section; 0.1.0 runbook                                                    |
| T28 | Guard README code snippets against drift (compile the README's Go snippets in a test)                                                                                                                                                                                    | Medium | S      | README install/usage blocks; report f18                                                      |
| T29 | CI guard for social-preview assets (render from svg, assert 1280x640 + < 1 MB before push)                                                                                                                                                                               | Low    | S      | assets/branding/; SKILLS generator machine-checks are the local half (report f19)            |
| T30 | Empirical GIF validation: one scratch-repo upload (GitHub card slot) + one Discord/Slack post — needs YOUR hands                                                                                                                                                         | Low    | S      | assets/branding/social-preview-animated.gif; platform matrix claims are doc-grade until then |
