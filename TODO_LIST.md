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

2026-09-23 (second execution session): T40–T43 CLOSED — v0.5.0 shipped the
bootstrap provider lifecycle (BootstrapSpec + BootstrapProviderFromSpec, 18
contract tests ported from oxlint's provider_test.go); oxlint's provider
migrated onto it (~110 lines deleted, tests semantically unchanged; v0.9.0)
and validate gained the drift advisory (--fail-on-drift). v0.6.0 shipped the
jsondeterminism analyzer + cmd vettool (T41), bite-checked against the
original SaveJSON gap; both consumers wire it into their gates (oxlint v0.9.1,
golangci next release) and golangci's config/backup writes went atomic with a
crash-safety test (T42). T27/T28/T29/T21 landed: scripts/verify-release.sh,
the README snippet compile guard (which caught real snippet drift on its
first run), the social-preview asset guard, and a BuildFlow CI job (license-
check now skipped via .buildflow.yml — the local pipeline is FULLY green for
the first time). T43's disposition matrix + go-finding issue draft live in
docs/planning/2026-09-23_pkg-format-disposition.md.

2026-09-23 (execution session): T31–T39 CLOSED — v0.3.1 shipped the determinism
fix to all three consumers; v0.4.0 shipped the I/O matrix (T33), WorkingDir
(T34), the generic diff engine (T36), and ConfigFiles/FirstExisting (T38), with
the docs wave (T39). oxlint's golden-byte pin landed before its I/O migration
(F30 pattern). The consumer-side migrations T35/T37/T38-oxlint landed on the
v0.4.1 tag (both consumers' `pkg/diff` now alias the SDK `Change`/`Kind`
vocabulary; oxlint's I/O and discovery run fully on SDK helpers; verified
2026-09-23: trees green, pushed). Remaining: the T40–T43 train.

| ID  | Task                                                                                                                                       | Impact | Effort | Evidence / Source                                                                            |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------ | ------ | ------ | -------------------------------------------------------------------------------------------- |
| T20 | Remove the deprecated `ErrNoRepair` alias at v1 (compile-compat shim; SDK never returns it; canonical signal is nil `toolsdk.Spec.Repair`) | Low    | S      | `autoconfigure.go` ErrNoRepair; CHANGELOG `[0.1.0]`                                          |
| T30 | Empirical GIF validation: one scratch-repo upload (GitHub card slot) + one Discord/Slack post — needs YOUR hands                           | Low    | S      | assets/branding/social-preview-animated.gif; platform matrix claims are doc-grade until then |
