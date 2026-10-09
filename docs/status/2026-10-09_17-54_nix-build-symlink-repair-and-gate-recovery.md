# Status Report — Nix-Build Symlink Repair & BuildFlow Gate Recovery

| | |
|---|---|
| **Date** | 2026-10-09 17:54 CEST |
| **Repo** | linter-autoconfigure-sdk @ master (daemon commits c82b7a8, 0513273, c6f2f11) |
| **Trigger** | `buildflow --fix --build-mode=full --budget 5m` exited **69** (nix-build failure + findings gate) |
| **Outcome** | `buildflow` exits **0**, "passed with warnings" (50/57 steps; 43 advisory findings, all documented baseline) |
| **Format note** | `.md` written per explicit user instruction; the status-report skill's canonical format is HTML (one-off override, not a new default) |

## Self-critique (asked directly)

**What did I forget?**
- I did not check the architecture of the nix-store Go I grabbed (`head -1` picked an **arm64** go-1.27.1), producing a bogus
  `gcc_arm64.S` test failure and one wasted test cycle. `nix develop -c` was the correct move from the start; AGENTS.md even
  warned the PATH recipe churns.
- I did not run `nix fmt` immediately after editing flake.nix, so my own unformatted edit blocked the treefmt check, which
  blocked nix-hash-fix's verify (a self-inflicted extra round-trip).
- I did not examine WHY the daemon committed `result` despite the `/result` ignore rule (tracked-before-rule? force-add?) —
  I made the repo immune but left the daemon behavior uninvestigated.
- I did not re-create `./result` (`nix build .# -o result`) after trashing it, so vulnix's scan-target state is unverified.
- I did not review the daemon's commits diff-by-diff after repair steps ran (go-mod-tidy/update had license to change go.mod).

**What could I have done better?**
- Diagnosis order: read `nix log <drv>` FIRST, before interpreting nix-hash-fix loops. I eventually did, but the "no fixable
  findings" loop message initially pointed me at hashes, not symlinks — trusting the summarizer over the raw log cost minutes.
- Two sloppy tool calls (a broken `agentic_fetch`-in-bash pipeline; an `rg -rn` accidental-replace). Both recovered, neither
  should have shipped.
- Could have batched the verification ladder (nix build → flake check → tests → buildflow) into fewer, bigger background runs.

**What could I still improve?**
- The `result`/`.direnv` symlink class is now structurally impossible in the flake source, but the daemon can still *commit*
  junk; the true root cause (pma force-add behavior) is upstream and untouched.
- Remaining preflight warns (BuildFlow binary freshness — a concurrent session moved BuildFlow HEAD twice mid-run; lychee
  private-link policy) are outside this repo's scope and were documented, not fixed.
- govulncheck reports 25 stdlib/closure CVE advisories (GO-2026-6617 HTTP/2 HPACK race et al.); I documented them but did not
  attempt a toolchain bump (go 1.27.2 is available in the nix store).

## a) FULLY DONE

| # | Item | Receipt |
|---|------|---------|
| 1 | Diagnosed nix-build failure: daemon-committed `result` symlink (dangling /nix/store target) killed the prepared-source FOD under `noBrokenSymlinks` | `nix log 4dl6mpr….drv` names `<src>/result` explicitly |
| 2 | Untracked + removed the dangling `result` symlink (`git rm --cached` + trash; daemon committed in c82b7a8) | `git ls-files result` → empty |
| 3 | Structural fix: `flake.nix` sets `src` to `lib.fileset.toSource { root = ./.; fileset = lib.fileset.gitTracked ./.; }` (fleet precedent: go-finding) so untracked/ignored junk (`result`, `.direnv/`, `.envrc`) can never enter the source in any tree state | flake.nix:21-27; prepared-source drv now builds |
| 4 | vendorHash repaired via `buildflow -s nix-hash-fix --fix` (never hand-pasted): `sha256-QoV9Q1iJ+RR/Uksm2xE6zT9vBJ47DnMaF6CQSI+B3fA=` | `nix-hash-fix` re-run: 6/6 targets, 0 findings |
| 5 | Cleared the treefmt blocker (my own flake.nix formatting) via `nix fmt` before the hash fix's verify build | treefmt check green in subsequent runs |
| 6 | SECURITY.md: added honest 'Security Practices' section (atomic writes, typed errors, determinism analyzer, sentinel validation, vuln scanning) → `securitymd` error cleared | absent from subsequent findings lists |
| 7 | AGENTS.md condensed 307 → 207 lines (≤120 cols, 0 overlong), every fact preserved, plus new gotchas: store-symlink class, vendorHash-cascade diagnosis, arm64-Go trap, securitymd requirement | `awk length>120` → 0; preflight `agents-md-size` warn gone (35 ok / 2 warn) |
| 8 | Full pipeline green: `buildflow --fix --build-mode=full` → exit 0 "passed with warnings"; plain `buildflow` gate → exit 0 | 50/57 steps; 43 findings all advisory baseline |
| 9 | Tests verified: `go test -race -count=1 ./...` green in devShell (go1.27.1 linux/amd64); `nix flake check` green; `nix build` green | devShell run + flake check output |
| 10 | Daemon picked up all work (no manual commits, per harness contract) | c82b7a8, 0513273, c6f2f11 |

## b) PARTIALLY DONE

| # | Item | State |
|---|------|-------|
| 1 | Preflight warnings | Reduced 5 → 2. Remaining: BuildFlow binary freshness (concurrent session in BuildFlow repo; advisory) and lychee private-links (fleet policy undecided). Both documented, neither fixable from this repo. |
| 2 | govulncheck 25 advisories | Documented as baseline; no reachability triage (which of the 25 are actually called?), no toolchain bump attempted. |
| 3 | "9 tools unavailable" health warns | Noticed and documented (interrogate missing; shellcheck/vulnix not in project devShell, running via `nix run nixpkgs#…` fallback). Not fixed. |
| 4 | Daemon root cause | Repo made structurally immune (gitTracked filter), but the daemon's commit behavior that caused both incidents (2026-10-04 `.direnv/`, today `result`) is unexamined upstream. |
| 5 | vulnix scan target | Trashed the stale `./result`; did not rebuild with `-o result` afterward, so vulnix's target freshness claim is unverified. |

## c) NOT STARTED (open items surfaced this session)

| # | Item |
|---|------|
| 1 | TODO T20 — remove deprecated `ErrNoRepair` at v1 |
| 2 | TODO T21 — BuildFlow CI job (blocked: BuildFlow repo private) |
| 3 | TODO T22 — bridge declarative toolsdk surface (Trigger/DependsOn/ModuleFanOut/Options/extra Inputs/HealthCheck) |
| 4 | HARVEST this report's section (f) into TODO_LIST.md (docs-health HARVEST) |
| 5 | file-size-check refactors: autoconfigure_test.go (1356), autoconfigure.go (533), bootstrap_test.go (551) — all over the 350-line max |
| 6 | Tracked-but-ignored file audit (`git ls-files -i -c --exclude-standard`) — is `.envrc` tracked-ignored too? |
| 7 | gitleaks + codespell on-demand scans — never run this session |
| 8 | nix-hash-fix failure history now reads 16/20 (80%) — stale stats may misfire loop-detection heuristics on future runs |

## d) TOTALLY FUCKED UP (then fixed; honest residuals)

| # | Item | Status |
|---|------|--------|
| 1 | Morning state: pipeline exit 69 with a **failure loop** (nix-hash-fix 3x identical "no fixable findings" while the real cause was the dangling symlink) — cascading red across nix-build, nix-build-verify, nix-hash-fix, nix-flake-check | FIXED (items a1-a5) |
| 2 | The daemon has now twice committed /nix/store symlinks that break `nix build` (2026-10-04 `.direnv/`, 2026-10-09 `result`) — systemic, will recur wherever the fleet lacks a source filter | MITIGATED here (gitTracked filter); upstream daemon behavior unfixed |
| 3 | My arm64-Go PATH pick produced a fake "[build failed]" test run | Process error, corrected; documented as a gotcha in AGENTS.md |
| 4 | My unformatted flake.nix edit briefly turned one broken pipeline into two failures (treefmt + hash) | Fixed same session; lesson: `nix fmt` after every flake edit |

Nothing is currently fucked up: working tree clean, all commits landed, gate exits 0.

## e) WHAT WE SHOULD IMPROVE

1. **Diagnosis discipline:** raw evidence first (`nix log <drv>`), summarizer output second. The loop message was a red herring.
2. **Shell hygiene in AI sessions:** always `nix develop -c` for toolchain-sensitive commands; never hand-assemble store paths
   (arch traps, path churn).
3. **Format-after-edit:** run `nix fmt` immediately after touching flake.nix; treefmt gates half the pipeline's verify paths.
4. **Defense in depth for the daemon:** the gitTracked filter protects the build; a fast preflight/preflake guard (fail on
   tracked symlinks into /nix/store) would turn a 30-minute investigation into a 5-second error message.
5. **Baseline-carry vs fix-at-source:** art-dupl 8, file-size 3, branching-flow 1 are documented-accepted; the quality bar
   says kill them properly (test helpers, file splits, nil-guard) rather than carry the baseline forever.
6. **Report harvesting:** section (f) below is TODO_LIST/ROADMAP fuel; entombing it in a timestamped file loses it.

## f) UP TO 50 THINGS TO GET DONE NEXT (brainstorm — HARVEST-routed, not commitments)

| # | Task | Impact | Notes |
|---|------|--------|-------|
| 1 | HARVEST this list into TODO_LIST.md / ROADMAP.md | High | docs-health HARVEST mode |
| 2 | Bump Go to 1.27.2 (present in store) and re-run govulncheck; triage GO-2026-6617 et al. for reachability | High | security hygiene |
| 3 | Add shellcheck + vulnix to devShells.default | Med | clears 2 "unavailable" warns at the root |
| 4 | Install/exclude `interrogate` | Low | clears another unavailable-tool warn |
| 5 | Decide fleet lychee policy: GITHUB_TOKEN vs exclude-regex | High | needs Lars (Q1) |
| 6 | Rebuild + reinstall BuildFlow binary after the concurrent session settles | Med | `nix build . && nix run .#reinstall` in BuildFlow repo |
| 7 | Upstream BuildFlow: preflight guard failing fast on tracked /nix/store symlinks | High | would have caught today's incident in seconds |
| 8 | Investigate pma daemon: why it committed `result` despite `/result` in .gitignore; harden | High | true root cause of both incidents |
| 9 | Repo-local pre-commit guard: reject staging /nix/store symlinks (stopgap if #7/#8 are slow) | Med | |
| 10 | Regression-test the flake source filter (eval assertion that result/.direnv are absent from src) | Med | belt-and-suspenders beyond gitTracked |
| 11 | Split autoconfigure_test.go (1356 lines) into focused test files | Med | kills file-size advisory |
| 12 | Split bootstrap_test.go (551) / review autoconfigure.go (533) | Med | same class |
| 13 | Extract shared table-test helpers to kill the 8 art-dupl clones properly | Med | fix-at-source vs baseline-carry |
| 14 | Add explicit nil-guard or restructure `return *parsed` at bootstrap.go:323 to legitimately silence branching-flow | Low | cheap fix beats documented false positive |
| 15 | Prune/reset nix-hash-fix failure history (16/20) so loop-detection heuristics stay honest | Low | check `buildflow history` capabilities |
| 16 | T20: remove `ErrNoRepair` at v1 | Med | needs v1 plan |
| 17 | T21: BuildFlow CI job when the repo goes public | Med | blocked |
| 18 | T22: bridge declarative toolsdk surface into ProviderFromSpec | High | removes post-hoc Spec mutation in both live consumers |
| 19 | Fix go-version-auto-configure under-declared Inputs (go.work/.github/nix) | Med | upstream consumer task (toolsdk audit) |
| 20 | Remove go-version's hand-rolled WorkingDir copy; use the SDK export | Low | upstream consumer task |
| 21 | Consumer sweep: confirm both consumers + go-version build against latest SDK tag | Med | go-ecosystem-upgrade pattern |
| 22 | Tracked-but-ignored audit: `git ls-files -i -c --exclude-standard`; untrack strays (.envrc?) | Med | same class as today's bug |
| 23 | Run `buildflow -s gitleaks` and `buildflow -s codespell` (on-demand, never ran this session) | Med | |
| 24 | Re-run `nix build .# -o result` and vulnix so the scan target is fresh | Low | |
| 25 | Evaluate `nix flake check --all-systems` (aarch64 currently omitted → nix-flake-check advisory) | Low | |
| 26 | Verify the 3 "no-op" steps in the full run — wire inputs or document | Low | `buildflow --verbose` |
| 27 | Sanity-check the 55 "not applicable" steps once — anything that should apply but doesn't? | Low | |
| 28 | Verify markdown-lint full-mode skip rationale (config-skipped) is still valid | Low | |
| 29 | Check FEATURES.md freshness re: flake.nix + SECURITY.md (docs-health VERIFY) | Low | |
| 30 | Check docs/DOMAIN_LANGUAGE.md existence/applicability | Low | |
| 31 | CHANGELOG entry for repo-level changes (SECURITY.md, flake filter) per repo policy | Low | |
| 32 | README: short "Nix" note documenting the flake + src-filter rationale for external users | Low | |
| 33 | vulnix closure CVE noise (binutils/bison/coreutils/gcc): whitelist or slim the build closure | Med | |
| 34 | gofumpt/oxfmt drift check: review daemon commits 140da6e..c6f2f11 diff-by-diff | Low | repairs had write access |
| 35 | Decide: does the flake filter warrant a v0.8.1 consumer-facing note? | Low | probably no (repo-local infra) |
| 36 | dependabot-auto-configure: verify config matches current modules | Low | |
| 37 | Verify pre-commit coverage of CI-only guards (social-preview, README snippets) | Low | |
| 38 | Freshness-check flake.nix comment about go_1_26 default vs goPkgAttr auto-select (doc rot candidate) | Low | |
| 39 | stdlib2lo: if the step supports excludes, encode the "larsartmann-family deps only" policy in config instead of prose | Low | |
| 40 | CI: evaluate a nix build job (install-nix-action) — roadmap | Low | runners/infra dependent |
| 41 | Add the symlink incident to crush-config references/lessons.md (cross-project lesson, by commit) | Med | per global memory policy |
| 42 | File verified upstream issue: nix-hash-fix should diagnose drv failure cause before hash repairs (verify-before-filing first) | Med | |
| 43 | Consider `nix settings`/sandbox note documenting WHY store symlinks are fatal in-sandbox (teardown for future sessions) | Low | partially in AGENTS.md now |
| 44 | Audit remaining docs/status reports for stale claims this session invalidated (docs-health ANNOTATE) | Low | |
| 45 | Confirm .buildflow.yml `skip_steps` list is still minimal post-2026-10-07 auto-defer change | Low | |
| 46 | Keep an eye on BuildFlow HEAD churn (2 moves this session) — concurrency between sessions is real | Info | |
| 47 | Evaluate whether `file-size-check` max=350 is right for `_test.go` files fleet-wide | Low | policy call |
| 48 | Propose go-version-auto-configure as third flake-consumer for the gitTracked src pattern (it reads go.work) | Low | fleet spread of today's fix |
| 49 | Check whether nixpkgs go_1_27 has moved past 1.27.2 and whether goPkgAttr auto-select picks it up correctly | Low | |
| 50 | Re-verify pkg.go.dev/social-preview guards still pass after SECURITY.md grew (link scan ran: 48 files, no failures) | Info | already green this run |

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Lychee fleet policy:** authenticate private-repo link checks with GITHUB_TOKEN, or exclude the `larsartmann` namespace in
   lychee.toml? The warn says the policy is undecided fleet-wide; I can implement either but shouldn't pick for the fleet.
2. **Advisory baselines: carry or kill?** art-dupl 8 / file-size 3 / branching-flow 1 are documented-accepted. Do you want them
   actually eliminated (test-helper extraction, file splits, a nil-guard), or is carrying the documented baseline the intended
   posture? I lean kill-at-source, but it churns ~2000 lines of tests for zero behavior change.
3. **Where does the daemon fix land?** Is pma (the auto-commit daemon) fixable upstream so it never tracks/commits /nix/store
   symlinks, or should the permanent defense be repo-side (pre-commit guard + BuildFlow preflight guard)? This decides tasks
   #7/#8/#9.

## Verification receipts

```text
nix log 4dl6mpr….drv      → "symlink …/result points to a missing target" (root cause)
git ls-files result       → empty after git rm --cached (c82b7a8)
buildflow -s nix-hash-fix --fix → vendorHash=sha256-QoV9Q1iJ…; detect 6/6 targets, 0 findings
nix build .# && nix flake check → green (all checks passed)
nix develop -c go test -race -count=1 ./... → ok (all packages, go1.27.1 linux/amd64)
buildflow --fix --build-mode=full → exit 0, "passed with warnings 50/57"
buildflow (plain)         → exit 0; preflight 35 ok / 2 warn / 0 fail
awk length>120 AGENTS.md  → 0 (207 lines total, under the 220 cap)
```

**Status: report written. Waiting for instructions.** (Manual commit skipped per harness contract; the auto-commit daemon
picks this file up — `docs/status/**` is excluded from all BuildFlow checks by `.buildflow.yml`.)
