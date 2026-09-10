# Status Update — 2026-09-10 04:38 CEST — Status-Report Commit, Lint Fallout, HARVEST Deferral

| Field            | Value                                                                                                    |
| ---------------- | -------------------------------------------------------------------------------------------------------- |
| Session scope    | Everything since report `2026-09-10_04-25` (the go-linter-sdk relationship report): writing that report, |
|                  | committing it, and post-commit verification of my own artifact                                            |
| Code changes     | **Zero product code.** One new file: the 04:25 status report, committed as `4576a11`                      |
| Verification run | `buildflow` full pipeline (exit 0), `buildflow --verbose` (tool health), `buildflow -s markdown-lint`     |
|                  | (finding attribution for the new file) — all read-only except the report commit itself                    |
| Report format    | `.md` per owner's explicit request — same flagged override as last time                                   |

## a) FULLY DONE

| # | What                                                                                                                                     | Evidence                                            |
| - | ---------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------- |
| a1 | Status report #1 written to spec: sections a–g, 50 ranked next-tasks in 6 tiers, 3 owner questions, HARVEST handoff section              | `4576a11`, 193 insertions                           |
| a2 | Format override honored AND flagged per the skill's escape-hatch rule (.md instead of the skill's HTML default), in both report and chat | report header + closing message of previous turn    |
| a3 | Commit followed the mandated format: detailed body explaining why the report exists, what it flags, and what is queued                    | `git log` — `4576a11` message                       |
| a4 | **Repo health claims re-verified today** (closing item b3 of report #1): `buildflow` exits 0 — tests green, exactly the 2 known errcheck findings (T5: `example_test.go:17/:33`), no new product-code findings | buildflow run 04:33, exit 0 |
| a5 | Lint-fallout attribution completed: the new report file's markdown-lint contribution was counted and rule-classified rather than guessed  | ~348 of 1145 total findings (696/2 JSON entries), predominantly MD013 line-length |

## b) PARTIALLY DONE

| # | What works now | What remains open | Blocker | Effort |
| - | -------------- | ----------------- | ------- | ------ |
| b1 | Report #1's section (f) is written and committed as HARVEST input | **HARVEST itself has not run** — TODO_LIST/ROADMAP still lack the 10 new Tier-1 items | The status-report skill says "if the session continues and TODO_LIST.md was not updated, run HARVEST now"; the owner's standing instruction says "report, then wait." Explicit owner instruction wins; deferral is deliberate and documented, but the loop stays open until the owner says "harvest" | S |
| b2 | The 3 owner questions from report #1 are asked (in report + closing message) | Unanswered; this second report request arrived instead | Owner decision — re-asked in section (g) | — |
| b3 | Repo health is now freshly verified (04:33) | Verification is manual-only; no CI exists (T3), so the green claim rots again tomorrow | T3 unstarted | M |

## c) NOT STARTED

| # | What | Why |
| - | ---- | --- |
| c1 | HARVEST of report #1's (f) into TODO_LIST/ROADMAP | Deferred on explicit "wait for instructions" (see b1) |
| c2 | Report #1 Tier 1, items 1–10 (ecosystem map, `ProviderSpec.Detector()`, ConfigIssue ADR, source-level re-validation of go-linter-sdk) | Proposed, committed as prose, zero bytes executed |
| c3 | All of TODO_LIST T1–T19 and ROADMAP candidates | Untouched this session; T5's errcheck findings re-confirmed live today |
| c4 | Formatter-of-record reconciliation (dprint, see d3) | Discovered this delta; not started |
| c5 | markdown-lint policy decision (see d1/e4) | Discovered and quantified this delta; not started |

## d) TOTALLY FUCKED UP

Nothing user-facing is broken; `buildflow` exits 0 and no product code changed. What follows
is what this delta actually got wrong:

| # | What's wrong | Severity | Root cause | Mitigation |
| - | ------------ | -------- | ---------- | ---------- |
| d1 | **I committed the single noisiest markdown file in the repo without running any check on it first.** The 04:25 report adds ~348 markdown-lint findings — roughly 30% of the repo's entire 1145 — in one commit. Mostly MD013 (80-col limit vs my wide tables), which the repo's existing docs also violate (~800 pre-existing), so it is *consistent with the status quo* — but "every change raises the bar" did not happen; I actively deepened the noise floor. Nothing gated it: buildflow exits 0 on warnings, and I never looked until the owner asked "what did you forget." | Medium (docs-only, zero runtime impact, but self-inflicted and now committed) | Treated a docs-only artifact as exempt from the test-after-changes discipline. Docs are artifacts too; I verified nothing before `git commit` | e1/e2: `buildflow --staged-only` before every commit — buildflow itself prints this exact tip, which I ignored; plus the f-list policy fix |
| d2 | **The status-report skill's HARVEST mandate and the owner's instruction collided, and I resolved it silently.** Skill: "if the session continues and TODO_LIST.md was not updated, run HARVEST now." Owner: "report, then wait." I chose the owner (correct), but only surfaced the conflict in a subsection rather than treating an unresolved skill-vs-owner precedence as a first-class report item | Low | Skill defaults vs explicit instruction precedence is decided ad hoc each time | e5: record the precedence rule once |
| d3 | **Formatter-of-record drift, newly discovered.** `dprint.json` is tracked and commit `7c75323` ("docs: let dprint realign annotated table cells") proves dprint formatted this repo's docs. Today: `dprint` is not on PATH, it is absent from buildflow's step list (the 9 unavailable tools are all irrelevant JS/TS/Python steps: jest, vitest, svelte-check, knip, …), and nobody noticed until now. The mechanism that produced `7c75323` is currently unenforceable — meaning report `4576a11`'s tables were committed to whatever alignment I typed by hand | Medium (latent; no enforcement gap fires until someone edits tables again) | Environment drift: the tool existed in some earlier session's shell and was never wired into buildflow where it would be health-checked | f2 |

## e) WHAT WE SHOULD IMPROVE

| # | Pattern that's suboptimal | Impact | Concrete fix |
| - | ------------------------- | ------ | ------------ |
| e1 | Docs-only commits skip verification entirely | Lint noise compounds silently until a report like this one counts it | Make `buildflow --staged-only` the universal pre-commit step for every artifact, docs included — it exists, it's one command, and buildflow suggests it in its own output |
| e2 | Verification happens only when challenged ("what did you forget?") | Quality depends on the owner asking the right question | Invert the order: run the staged check before composing the closing message of any task, every time |
| e3 | Repo-wide MD013 at 80 columns vs table-heavy docs is structurally unwinnable — ~1145 standing warnings make the linter wallpaper | Real findings drown in line-length noise; nobody will ever read 1145 warnings | One explicit policy decision: either configure markdown-lint for wide tables (MD013 off or per-syntax) repo-wide, or accept and record that the warnings are by-design wallpaper — but decide it once instead of letting each new doc re-grow the pile |
| e4 | Skill-vs-owner precedence resolved ad hoc per collision | Each resolution is invisible unless someone writes a report about it | One line in the global AGENTS.md: explicit owner instructions override skill defaults; when they conflict, note the deferral in the report and move on — no silent choices |
| e5 | Report #1's Tier-1 items exist only as prose in a timestamped file | The ecosystem map, Detector adapter, and ConfigIssue ADR rot in `docs/status/` while every future session re-derives them | HARVEST on owner go-ahead (b1) — this is the same lesson as report #1's d1, now with a second instance: findings must land in living docs, not snapshots |

## f) Things to get done next

Honest scoping: this delta produced 5 genuinely new items (#1–5); inventing 49 more would be
padding. The remaining 20 are the highest-impact still-open carry-overs from report #1's
50-item list, re-ranked with today's evidence (all still 0% started). Full 50-item detail
lives in report `2026-09-10_04-25`, section (f).

| #  | Task                                                                                                                                                    | Impact | Effort | Category      | Source |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- | ------ |
| 1  | Decide markdown-lint policy repo-wide: configure MD013/wide-table handling once, or formally accept the wallpaper (then stop counting it)                  | High   | S      | Decision      | NEW — d1 |
| 2  | Reconcile the formatter of record: wire dprint into buildflow (or into `--staged-only` pre-commit flow) or delete `dprint.json` — no third orphaned state | High   | S      | Quality       | NEW — d3 |
| 3  | Adopt `buildflow --staged-only` as the mandatory pre-commit gate for ALL artifacts, including docs                                                        | High   | S      | Process       | NEW — e1 |
| 4  | Run HARVEST: route report #1's 10 new Tier-1 items + this report's #1–3 into TODO_LIST/ROADMAP                                                           | High   | S      | Documentation | NEW — b1 |
| 5  | Slim buildflow config for a pure-Go repo: 9 permanently-unavailable JS/TS/Python tool steps and 73 not-applicable steps are pure noise in every run       | Low    | S      | Cleanup       | NEW — d3 |
| 6  | Ecosystem map into both AGENTS.mds (go-finding / go-atomic-write / go-linter-sdk / this SDK / BuildFlow)                                                 | High   | S      | Documentation | #1 rep. |
| 7  | Add `ProviderSpec.Detector()` adapter (Analyze + FindingsFromIssues ≡ `finding.Detector`)                                                                | High   | M      | Feature       | #3 rep. |
| 8  | ConfigIssue-vs-direct-findings ADR (converter philosophy vs Repair contract)                                                                             | High   | S      | Decision      | #6 rep. |
| 9  | Verify external fetch: `go get` from a clean throwaway module (repo is public; fetch never proven)                                                       | Critical | S    | Quality       | T1      |
| 10 | Trigger + verify pkg.go.dev listing (godoc, examples, license)                                                                                           | High   | S      | Quality       | T2      |
| 11 | Add CI workflow (`buildflow --fix --fail-on-findings`, `GOEXPERIMENT=jsonv2`) — ends manual-only health claims like a4/b3                                 | High   | M      | Quality       | T3      |
| 12 | Cut `v0.1.0` + GitHub Release (gated on question g2)                                                                                                     | High   | S      | Release       | T4      |
| 13 | Fix the 2 live errcheck findings (`os.RemoveAll` in `example_test.go:17/:33`) — re-confirmed today                                                       | Medium | S      | Bug           | T5      |
| 14 | Source-level re-validation of go-linter-sdk claims (read `rule.go`/`registry.go`/`errors.go`; verify proxy fetchability)                                  | Medium | S      | Quality       | #4/#5 rep. |
| 15 | First consumer migration: wire `golangci-lint-auto-configure` onto the SDK end-to-end — the milestone that decides the SDK's right to exist              | Critical | L    | Feature       | R       |
| 16 | `ErrNoRepair` sentinel test + doc example (exported, untested)                                                                                           | Medium | S      | Quality       | T6      |
| 17 | `(*ConfigError).As`/`.Is` delegation regression tests (only `AsType` asserted)                                                                           | Medium | S      | Quality       | T7      |
| 18 | Filesystem-independent idempotency test (replace 20ms-sleep mtime comparison)                                                                            | Medium | S      | Quality       | T8      |
| 19 | `SaveJSON` `changed`-bool decision (zero consumers = free now)                                                                                           | Medium | S      | Decision      | T16     |
| 20 | Rename/repurpose decision: `project-autofix-sdk` broadening — formally accept or reject                                                                  | High   | S      | Decision      | R       |
| 21 | SECURITY.md + branch protection on public `master`                                                                                                       | Medium | S      | Quality       | T12/T13 |
| 22 | README public sales-page audit from a clean machine (gated on question g3)                                                                               | Medium | M      | Documentation | T18     |
| 23 | GOEXPERIMENT=jsonv2 note in package doc (`autoconfigure.go:1-17`)                                                                                        | Medium | S      | Documentation | T10     |
| 24 | Fuzz + property-based Save→Load round-trip tests                                                                                                         | Medium | M      | Quality       | R       |
| 25 | `ProviderFromSpec` → `toolsdk.Spec` BuildFlow adapter (only after consumer exists, per ROADMAP)                                                          | High   | M      | Feature       | R       |

## g) Questions for the owner (unchanged from report #1 — still open, still not self-answerable)

1. **Sibling strategy:** Is `go-linter-sdk` the long-term home for shared linter-ecosystem
   scaffolding this SDK should align with (possibly hosting shared BuildFlow adapters), or
   fully independent siblings with only a documented ecosystem map? Decides #6–8, #14.
2. **API freeze (ROADMAP Q2):** Freeze for `v0.1.0` now, or land the likely breaking moves
   first (#7 Detector adapter, #19 `changed`-bool, #8 ConfigIssue ADR)? Gates #12.
3. **Consumers public (ROADMAP Q3):** Will `golangci-lint-auto-configure` /
   `oxlint-auto-configure` ever go public? Decides the README story (#22) and whether #15 is
   public validation or private plumbing.

## Handoff

Sections (d)/(e) contain two meta-findings worth harvesting beyond the task lists:
`buildflow --staged-only` as universal pre-commit gate (#3), and the skill-vs-owner
precedence note (#e4). Section (f) is HARVEST-ready on owner go-ahead.

Nothing else was touched. Waiting for instructions.
