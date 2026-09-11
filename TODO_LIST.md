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

| ID  | Task                                                                                                                                                                           | Impact | Effort | Evidence / Source                                                                                     |
| ---- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | ------- | ------------------------------------------------------------------------------------------------------ |
| T20 | Remove the deprecated `ErrNoRepair` alias at v1 (compile-compat shim; SDK never returns it; canonical signal is nil `toolsdk.Spec.Repair`)                                     | Low    | S      | `autoconfigure.go` ErrNoRepair; CHANGELOG `[0.1.0]`                                                   |
| T21 | Switch CI to `buildflow --fix --fail-on-findings` once BuildFlow is publicly installable (runners cannot fetch the private module today; re-checked 2026-09-11, still private) | High   | S      | TODO_LIST T3 residue; `.github/workflows/ci.yml` header comment                                       |
| T23 | Upload `assets/branding/social-preview.png` as the GitHub social preview (Settings → General → Social preview → Edit → Upload an image; no deep URL exists; no API)            | Low    | S      | flip report f30-32; redesigned 2026-09-11; source is social-preview.svg; 1280x640, 56KB (< 1MB limit) |
