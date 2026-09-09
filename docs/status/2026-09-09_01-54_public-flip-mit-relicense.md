# Status Report: Public Flip + MIT Relicensing

**Date:** 2026-09-09 01:54 CEST
**Scope:** This session only — the go-public decision analysis, MIT relicensing,
and visibility flip of `LarsArtmann/linter-autoconfigure-sdk`. No other project
work was researched or included.

---

## Session summary

The repo was **PRIVATE** with a **PROPRIETARY (all rights reserved)** LICENSE
while its README had advertised an MIT badge since the initial commit. This
session: analyzed public-readiness (dependency visibility, secret screening,
license consistency, docs exposure), replaced LICENSE with MIT text identical
to sibling repos, pushed, flipped visibility to PUBLIC, and verified the
result. Post-flip verification caught the process gap: the full git history
(37 commits) had not been secret-scanned before exposure — remediated in this
report's preparation with gitleaks: **no leaks found**.

---

## a) FULLY DONE

| #  | Item                                                                                                                                                                                                                                                                                                   | Evidence                                                                                                    |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------- |
| a1 | Public-readiness diligence before recommending the flip: all three dependencies (`go-finding`, `go-atomic-write`, `go-error-family`) confirmed PUBLIC + MIT; both consumer tools (`golangci-lint-auto-configure`, `oxlint-auto-configure`) confirmed still PRIVATE; README↔LICENSE mismatch discovered | `gh repo view` on 5 repos; README.md:7,157-159 vs LICENSE (old)                                             |
| a2 | LICENSE replaced: PROPRIETARY → MIT, text identical to `go-finding`'s (Copyright (c) 2026 Lars Artmann)                                                                                                                                                                                                | Commit `23e74f1`, 1 file changed                                                                            |
| a3 | Pushed master and flipped visibility                                                                                                                                                                                                                                                                   | `git push` (`9d1373a..23e74f1`); `gh repo edit --visibility public --accept-visibility-change-consequences` |
| a4 | Verified end state: repo PUBLIC, GitHub license detection shows "MIT License"                                                                                                                                                                                                                          | `gh repo view --json visibility,licenseInfo` → `PUBLIC \| MIT License`                                      |
| a5 | Tracked-file inventory for public exposure reviewed; `.crush/` (private session data) confirmed gitignored, never tracked                                                                                                                                                                              | `git ls-files`, `git check-ignore .crush`                                                                   |
| a6 | Full-history secret scan completed (see d1 remediation): gitleaks over all 37 commits (320.8 KB), zero findings                                                                                                                                                                                        | `gitleaks git` → `no leaks found`, exit 0                                                                   |

## b) PARTIALLY DONE

| #  | Item                                 | What works                                                                  | What remains                                                                                                                                                                                                                   | Blocker                                | Effort                             |
| -- | ------------------------------------ | --------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------- | ---------------------------------- |
| b1 | Secret screening before the flip     | Current tree scanned via `git grep` (1 regex) during decision phase         | Full 37-commit history was NOT scanned before the flip; closed afterwards (a6) with gitleaks, but only post-exposure                                                                                                           | None                                   | S (done, ordering was the failure) |
| b2 | pkg.go.dev publication               | Repo public → module proxy can now serve it                                 | Propagation never triggered or verified; no `go get` from a clean external module attempted yet                                                                                                                                | None — needs one fetch + verification  | S                                  |
| b3 | docs/ exposure decision              | Flagged to user before flip; user instructed the flip (implicit acceptance) | No written decision: `docs/planning` (incl. internal `licenseforge` strategy analysis), `docs/reviews/*.html`, `docs/status/*` are now world-readable                                                                          | Owner risk-appetite call (question g1) | S                                  |
| b4 | Public-readiness of the repo surface | README exists and is decent (9 KB); MIT badge now truthful                  | README content never audited for external-consumer needs (~~GOEXPERIMENT=jsonv2 requirement lives only in AGENTS.md~~ _[claim was wrong: README.md:35-36 has documented it since `1818692`]_; install instructions unverified) | None                                   | S                                  |

## c) NOT STARTED

| #      | Item                                                                                                                               | Why not started                                                                                              | Priority   |
| ------ | ---------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------ | ---------- |
| ~~c1~~ | ~~CHANGELOG entry documenting relicensing + public flip~~ done at `e46c225`                                                        | ~~Not part of the explicit instruction; commit message carries the rationale but CHANGELOG.md is untouched~~ | ~~High~~   |
| ~~c2~~ | ~~README §License fix: still links `LarsArtmann/template-LICENSE` repo instead of the now-existing `./LICENSE`~~ done at `e46c225` | ~~Discovered late (a1); not instructed~~                                                                     | ~~Medium~~ |
| c3     | First tag/release — repo has **zero tags**; pkg.go.dev will show only `v0.0.0` pseudo-versions                                     | Release readiness is an owner decision (question g2)                                                         | High       |
| c4     | CI entirely absent: no `.github/` directory, no workflows, nothing runs buildflow/test/lint on push or PR                          | Pre-existing gap, discovered this session while inventorying tracked files                                   | High       |
| ~~c5~~ | ~~AGENTS.md not updated with new enduring facts: repo is public, MIT, no tags, no CI~~ done at `e46c225`                           | ~~Memory-maintenance rule applies; deferred to post-report~~                                                 | ~~Medium~~ |
| c6     | GitHub repo polish: topics, branch protection, SECURITY.md, issue/PR templates, social preview                                     | Not instructed; now expected for a public repo                                                               | Medium     |
| ~~c7~~ | ~~docs-health HARVEST of section (f) into TODO_LIST.md / ROADMAP.md~~ done at `e46c225`                                            | ~~Report first (this file); harvest is the next step by design~~                                             | ~~High~~   |

## d) TOTALLY FUCKED UP

| #  | What is broken                                                                                                                                                                                                                                                                                     | Severity                                    | Root cause                                                                                                                                                                                                                                                                              | Status / Mitigation                                                                                                                                             |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| d1 | **Visibility flip executed BEFORE a full-history secret scan.** 37 commits (~321 KB of history) became world-readable having been screened only via a single-regex working-tree grep. History exposure is effectively irreversible (forks, caches, GitHub event streams).                          | High consequence / low realized probability | I treated the user's explicit "public it" as authorization to move fast and substituted a shallow tree grep for real diligence; irreversible actions demand deeper verification than reversible ones (the exact lesson `verify-before-filing` encodes, applied to the wrong direction). | **Retired post-hoc:** gitleaks scanned all 37 commits → `no leaks found`, exit 0. No rewrite needed. The process failure stands; the outcome is verified clean. |
| d2 | **README advertised MIT while LICENSE said PROPRIETARY — a license split-brain that existed from the initial commit (2026-07-18) until today.** Anyone auditing the private repo saw contradictory license signals; had the repo gone public in that state, the badge would have been a legal lie. | Medium (was)                                | LICENSE and README were created in different commits and never cross-checked; no automated consistency check exists.                                                                                                                                                                    | Fixed this session (a2). Prevention → e3.                                                                                                                       |

Nothing else in this session qualifies: no tests were broken (no code changed),
no data lost, no rollback needed.

## e) WHAT WE SHOULD IMPROVE

| #  | Suboptimal practice                                                                                              | Impact                                                                                                    | Concrete fix                                                                                                                                                                                                                     |
| -- | ---------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| e1 | Irreversible actions (visibility flip, deletion, force ops) get the same shallow verification as reversible ones | One gitleaks run stood between "lucky" and "leaked"; the ordering failure is invisible when it goes right | A fixed **go-public checklist** (as a skill or skill extension): history secret scan → tracked-file review → internal-docs review → license consistency → tag/CI plan → flip → verify. Load it before any `--visibility` change. |
| e2 | Material repo-level changes skip the CHANGELOG                                                                   | Relicensing is exactly what future audits look for; it is currently only in a commit message              | Rule: license/scope/ownership changes get a CHANGELOG entry in the same commit                                                                                                                                                   |
| e3 | README badges/links can drift from reality (MIT badge vs proprietary file for 7 weeks)                           | Silent legal contradiction; embarrassed public launch                                                     | buildflow step (or lint script): verify license badge text matches LICENSE file's detected license                                                                                                                               |
| e4 | pkg.go.dev state treated as "automatic, someone will see it"                                                     | A public Go module nobody has verified can sit broken on pkg.go.dev (bad godoc, pseudo-versions only)     | Add "trigger proxy fetch + verify listing + check pkg.go.dev rendering" to the go-public checklist                                                                                                                               |
| e5 | No CI on a repo that has a full local quality pipeline (buildflow)                                               | Quality gates exist only on Lars' machine; public contributors and the auto-commit daemon push unverified | CI workflow that runs `buildflow --fix --fail-on-findings` (see c4)                                                                                                                                                              |
| e6 | Enduring project facts (public status, license) updated nowhere durable                                          | Next session starts from stale AGENTS.md context                                                          | Update AGENTS.md in the same session as the change (c5)                                                                                                                                                                          |

## f) Up to 50 things we should get done next

Ranked by impact. Impact: Critical/High/Medium/Low. Effort: S <30min, M 30min-2h, L >2h.
**This section is HARVEST input → `TODO_LIST.md` / `ROADMAP.md`** (c7).

| #      | Task                                                                                                                                                                                                                                                  | Impact     | Effort | Category          |
| ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- | ------ | ----------------- |
| ~~1~~  | ~~docs-health HARVEST: pull this table into TODO_LIST.md (1-15) and ROADMAP.md (16+)~~ done at `e46c225`                                                                                                                                              | ~~High~~   | ~~S~~  | ~~Documentation~~ |
| 2      | Record the docs/ decision (g1) once made; if removing, forward-delete and note history retains them                                                                                                                                                   | High       | S      | Cleanup           |
| 3      | Verify module fetchable from a clean external module: `go get github.com/larsartmann/linter-autoconfigure-sdk`                                                                                                                                        | High       | S      | Quality           |
| 4      | Trigger pkg.go.dev listing (proxy fetch) and verify the rendered package page (godoc, examples, license)                                                                                                                                              | High       | S      | Quality           |
| ~~5~~  | ~~Add CHANGELOG entry: MIT relicensing + public flip (c1)~~ done at `e46c225`                                                                                                                                                                         | ~~High~~   | ~~S~~  | ~~Documentation~~ |
| 6      | Add CI: `.github/workflows/ci.yml` running `buildflow --fix --fail-on-findings` on push/PR (c4)                                                                                                                                                       | High       | M      | Quality           |
| 7      | Decide + cut first tag `v0.1.0` so pkg.go.dev shows a real version (c3, g2)                                                                                                                                                                           | High       | S      | Release           |
| 8      | Create GitHub Release with notes for the tag                                                                                                                                                                                                          | Medium     | S      | Release           |
| ~~9~~  | ~~Run `buildflow --fix --fail-on-findings` once to confirm the public tree is pipeline-green (not run since LICENSE change)~~ done — run 2026-09-09 — pipeline exit 0; strict lane still red on pre-existing findings (TODO_LIST T5/T11)              | ~~Medium~~ | ~~S~~  | ~~Quality~~       |
| ~~10~~ | ~~Fix README §License to link `./LICENSE` instead of the template repo (c2)~~ done at `e46c225`                                                                                                                                                       | ~~Medium~~ | ~~S~~  | ~~Documentation~~ |
| ~~11~~ | ~~Verify README documents the `GOEXPERIMENT=jsonv2` / direnv requirement for building from source (currently AGENTS.md-only; external consumers will hit it)~~ done — already documented — README.md:35-36 covers GOEXPERIMENT since 1818692          | ~~High~~   | ~~S~~  | ~~Documentation~~ |
| 12     | Audit README as a public sales page: install snippet, API example, comparison to hand-writing config                                                                                                                                                  | Medium     | M      | Documentation     |
| ~~13~~ | ~~Update AGENTS.md: public status, MIT, tag/CI state (c5)~~ done at `e46c225`                                                                                                                                                                         | ~~Medium~~ | ~~S~~  | ~~Documentation~~ |
| 14     | Add SECURITY.md with a security contact                                                                                                                                                                                                               | Medium     | S      | Documentation     |
| 15     | Enable branch protection on master (no direct pushes to a public default branch)                                                                                                                                                                      | Medium     | S      | Quality           |
| 16     | Review `docs/planning/license-domain-fit-analysis.md` for public exposure: it proposes renaming/repurposing the SDK (strategy). Move insight to ROADMAP, relocate file                                                                                | Medium     | S      | Cleanup           |
| 17     | Decide fate of `docs/reviews/2026-07-26_architecture-and-data-model.html` (styled internal dashboard, now public)                                                                                                                                     | Medium     | S      | Cleanup           |
| 18     | Sweep all tracked `docs/status/*.md` + `docs/planning/*.md` for internal project names (licenseforge, BuildFlow internals) and redact or accept                                                                                                       | Medium     | S      | Cleanup           |
| 19     | Add GitHub topics: `go`, `linter`, `golangci-lint`, `oxlint`, `sdk`, `config`                                                                                                                                                                         | Low        | S      | Quality           |
| 20     | Review repo description; consider adding install command                                                                                                                                                                                              | Low        | S      | Quality           |
| 21     | dependabot.yml (or renovate) for the public repo                                                                                                                                                                                                      | Medium     | S      | Quality           |
| ~~22~~ | ~~Review CONTRIBUTING.md (403 bytes) for external-contributor readiness: build steps, GOEXPERIMENT note, dprint, buildflow~~ done at `e46c225`                                                                                                        | ~~Medium~~ | ~~M~~  | ~~Documentation~~ |
| 23     | Verify package-level godoc on `autoconfigure.go` reads well as the pkg.go.dev landing text                                                                                                                                                            | Medium     | S      | Documentation     |
| 24     | Verify `example_test.go` Examples render on pkg.go.dev (they are the public API demo)                                                                                                                                                                 | Medium     | S      | Documentation     |
| ~~25~~ | ~~Verify no build artifacts ever entered history (e.g. `reports/coverage.out`) — `git log --all -- reports/`~~ done — verified — git log --all -- reports/ is empty; no build artifacts ever tracked                                                  | ~~Medium~~ | ~~S~~  | ~~Security~~      |
| ~~26~~ | ~~Confirm dprint/lint pipeline does not want to reformat LICENSE (plain text) — covered by item 9~~ done — covered — pipeline runs green, LICENSE untouched by dprint                                                                                 | ~~Low~~    | ~~S~~  | ~~Quality~~       |
| 27     | Quick review of `.envrc`, `git-town.toml`, `.gitattributes`, `.editorconfig` for anything not meant for public eyes                                                                                                                                   | Low        | S      | Security          |
| 28     | Review `.gitignore` for personal/machine-specific entries worth generalizing                                                                                                                                                                          | Low        | S      | Cleanup           |
| 29     | Decide Issues vs Discussions for the public repo                                                                                                                                                                                                      | Low        | S      | Quality           |
| 30     | Add issue + PR templates                                                                                                                                                                                                                              | Low        | S      | Documentation     |
| 31     | Add CODEOWNERS                                                                                                                                                                                                                                        | Low        | S      | Quality           |
| 32     | Set GitHub social preview image                                                                                                                                                                                                                       | Low        | S      | Quality           |
| 33     | Document in README which tools consume this SDK (currently both private — the "why does this exist" story is invisible)                                                                                                                               | Medium     | S      | Documentation     |
| 34     | Decide whether consumer tools eventually go public (g3); if yes, plan their own go-public checklists (e1)                                                                                                                                             | Medium     | M      | Feature           |
| ~~35~~ | ~~State the v0 semver policy (breaking changes allowed until v1) in README or CONTRIBUTING~~ done — already stated — README Status section documents breaking-changes-acceptable pre-v1                                                               | ~~Medium~~ | ~~S~~  | ~~Documentation~~ |
| 36     | Automate license-badge consistency check (e3) as a buildflow step or script                                                                                                                                                                           | Medium     | M      | Quality           |
| 37     | Encode the go-public checklist (e1) as a reusable skill or skill extension                                                                                                                                                                            | High       | M      | Quality           |
| ~~38~~ | ~~Add markdown link/badge checker for README (dead template-LICENSE link is the first example of the class)~~ done — exists — buildflow's lychee link checker flagged the pkg.go.dev 404s on 2026-09-09                                               | ~~Low~~    | ~~S~~  | ~~Quality~~       |
| ~~39~~ | ~~Annotate this report once its items complete (docs-health ANNOTATE mode, non-destructive)~~ done (docs-health pass 2026-09-09)                                                                                                                      | ~~Low~~    | ~~S~~  | ~~Documentation~~ |
| 40     | ROADMAP: evaluate the rename/repurpose proposal from license-domain-fit analysis (broaden SDK scope, add SaveText/ApplyTemplate helpers) or formally reject it                                                                                        | Low        | M      | Roadmap           |
| 41     | Link the SDK from larsartmann profile README and sibling public libs' related-projects sections                                                                                                                                                       | Low        | S      | Documentation     |
| ~~42~~ | ~~Consider `goreleaser`/release automation only if this ever ships binaries — for a library, tags + GitHub Releases suffice (note to prevent over-tooling)~~ **Won't implement — library — tags + GitHub Releases suffice, per the item's own note.** | ~~Low~~    | ~~S~~  | ~~Roadmap~~       |
| 43     | Re-check proxy/pkg.go.dev propagation 24h after items 3-4 (propagation can lag)                                                                                                                                                                       | Low        | S      | Quality           |
| 44     | Add coverage reporting (buildflow already writes `reports/coverage.out`) as a CI artifact or badge                                                                                                                                                    | Low        | M      | Quality           |
| ~~45~~ | ~~Sweep CHANGELOG.md formatting against Keep-a-Changelog since it now becomes a public-facing artifact~~ done at `e46c225`                                                                                                                            | ~~Low~~    | ~~S~~  | ~~Documentation~~ |

## g) Questions I can NOT figure out myself

**Q1 — docs/ fate.** `docs/planning`, `docs/reviews`, and `docs/status` are now
public, including internal strategy content (`licenseforge` analysis, brutal
reviews, rename/repurpose proposals). Options: (a) keep public as radical
transparency, (b) forward-delete (git history still contains them),
(c) delete + history rewrite (force-push, coordinate with any forks/clones).
This is a risk-appetite call only you can make. I tried: reviewed every tracked
doc's content — the sensitive-ness judgment (embarrassing vs merely internal)
is yours, not mine.

**Q2 — first tag now or later?** The repo has zero tags; until one exists,
pkg.go.dev shows only `v0.0.0` pseudo-versions and there is nothing stable to
point consumers at. Is the API surface ready to freeze as `v0.1.0` today, or do
you have known breaking changes planned that should land first? I cannot infer
your release intent from the code.

**Q3 — will the consumer tools (`golangci-lint-auto-configure`,
`oxlint-auto-configure`) ever go public?** The SDK's public README currently
justifies its existence by referencing two repos nobody can see. If they will
become public, I write the SDK story accordingly; if they stay private, I need
a public-facing motivation that stands alone. Your roadmap for those repos is
not discoverable from here.

---

_Point-in-time snapshot — will go stale. Harvest (f) into TODO_LIST.md/ROADMAP.md,
then annotate this file non-destructively as items complete._

_[Both executed 2026-09-09: section (f) harvested into TODO_LIST.md (T1-T19)
and ROADMAP.md; 18 rows resolved inline above.]_

---

## Resolution (2026-09-09, docs-health pass)

Harvest complete: TODO_LIST.md carries the bounded items (T1-T19), ROADMAP.md
carries the strategic ones plus the open questions Q1-Q3 from section (g).
Resolved this pass (18 rows): c1, c2, c5, c7, f1, f5, f9, f10, f11, f13, f22,
f25, f26, f35, f38, f39, f42, f45 — the CHANGELOG now documents the
relicensing, README links ./LICENSE and go-atomic-write, AGENTS.md reflects
the public/MIT state (ghost `reports/.gitkeep` section removed after empirical
verification), CONTRIBUTING.md explains buildflow + GOEXPERIMENT, and the
docs set (TODO_LIST/FEATURES/ROADMAP/DOMAIN_LANGUAGE) was built at `e46c225`.
Still open: first tag, CI, pkg.go.dev verification, GitHub polish, docs/
exposure decision (ROADMAP Q1).
