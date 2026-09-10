# TODO List

Short- and mid-term actionable work. Open items only — completed items live in
`CHANGELOG.md`. Long-term ideas and open questions live in `ROADMAP.md`.

2026-09-10 session closed T1, T2 (rendering verified; 24h re-check spun off as
T24), T3 (portable CI; buildflow-in-CI spun off as T21), T7, T8, T9, T11, T12,
T13, T14, T15 (asset created; manual upload spun off as T23), T16, T17, T18,
and T19 — see CHANGELOG `Repository` and `Added`/`Changed` sections.

| ID  | Task                                                                                                                                                  | Impact | Effort | Evidence / Source                                                        |
| --- | ---------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------------------------------------------------------------------ |
| T4  | Decide + cut first tag `v0.1.0` and create the GitHub Release with notes (needs owner go-ahead on API freeze; see ROADMAP Q2 — with T16 settled, the surface is a freeze candidate) | High   | S      | flip report f7/f8; `git tag` is empty                                    |
| T20 | Remove the deprecated `ErrNoRepair` alias at v1 (compile-compat shim; SDK never returns it; canonical signal is nil `toolsdk.Spec.Repair`)             | Low    | S      | `autoconfigure.go` ErrNoRepair; CHANGELOG Unreleased                     |
| T21 | Switch CI to `buildflow --fix --fail-on-findings` once BuildFlow is publicly installable (runners cannot fetch the private module today)               | High   | S      | TODO_LIST T3 residue; `.github/workflows/ci.yml` header comment          |
| T22 | Add required status checks to the `master` branch protection after CI's first green run (context names only become selectable then)                   | Medium | S      | branch protection currently has `required_status_checks: null`           |
| T23 | Upload `assets/branding/social-preview.png` as the GitHub social preview (Settings UI only; no API exists for it)                                    | Low    | S      | flip report f30-32; asset exists at 1280x640                             |
| T24 | Post-push re-checks: pkg.go.dev renders the post-T16 snapshot (`SaveJSON` two-value signature, updated `ExampleSaveJSON`); confirm CI goes green after the `mkdir -p reports` workflow fix (first run failed: coverprofile parent dir). Consumer compile vs the new pseudo-version is DONE (verified 2026-09-10 07:42 against `f742550`: fetch + build + run green) | Medium | S      | flip report f43; propagation lag is normal for new pseudo-versions       |
