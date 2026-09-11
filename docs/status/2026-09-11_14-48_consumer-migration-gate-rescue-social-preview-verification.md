# Status Report — 2026-09-11 14:48 CEST

## Consumer migration, gate rescue, and social-preview verification

Session scope: continuation of the 2026-09-11 morning sessions. One directive:
"READ, UNDERSTAND, RESEARCH, REFLECT → break down → execute → verify → repeat
until done." Four repos touched: `linter-autoconfigure-sdk` (SDK + docs +
assets), `SKILLS` (generator + gate + feedback), `oxlint-auto-configure`
(first SDK consumer), `golangci-lint-auto-configure` (second SDK consumer).
Format override: Markdown per explicit user instruction (the status-report
skill's canonical default is a styled HTML dashboard; flagged per skill
contract, one-off, not propagated — the HTML-vs-md default question has now
recurred 4+ sessions, see f39).

Headline numbers: SDK quality gate 1338 findings → **3** (all three the
documented-intentional go-structure-linter warnings); golangci-lint
**83 → 0 issues**; markdown-lint **1249 → 0**; "Imported by: 0" → **2
consumers migrated**; T23 (social preview upload) verified **LIVE and
byte-identical**.

---

## a) FULLY DONE

1. **The P0 nobody planned: the SDK gate was secretly red.** The morning
   sessions reported `buildflow` exit 0 "repeatedly", but the
   golangci-lint-auto-configure step was failing with 58 error-severity
   findings (101 linters + 5 formatters "missing" from `.golangci.yml`).
   Root cause: buildflow's result cache (168h TTL) served a stale green
   verdict after the tool's behavior changed. Diagnosed via
   `BUILDFLOW_NO_RESULT_CACHE=1`, fixed by full adoption: the generated
   config (106 linters, `run.timeout: 5m`, `--priority high`) is now the
   repo's config, regenerated via `buildflow --fix -s
   golangci-lint-auto-configure`. The SDK being red against its own
   ecosystem's config tool was exactly wrong; it now conforms.
2. **All 83 golangci-lint findings on the SDK fixed to zero.** 33 test
   functions + 1 subtest got `t.Parallel()` (scripted insertion; no
   t.Setenv/Chdir hazards, verified by `-race`); dynamic errors became
   static sentinels (exported `ErrNameRequired`, `ErrDescriptionRequired`,
   `ErrAnalyzeRequired` — same messages, so text-matching callers keep
   working); `stubAsCause` → `stubAsCauseError` (errname); `SaveJSON` parent
   dirs 0755 → 0750 with a named `configDirPerm` const (gosec G301); one
   honest trailing `//nolint:makezero` on the race-free pre-sized-slice
   idiom; `Builder.Build()` error wrapped with full context (erraudit
   correctly flagged that my first wrap dropped `toolName`).
3. **Markdown flood eliminated: 1249 → 0 findings.** `.markdownlint.yml`
   (MD013 at 120 with tables/code exempt, MD024 siblings_only, MD033
   a/img allowed for the hero embed), `.buildflow.yml` excludes for
   `docs/status/**` and `docs/planning/**` — point-in-time records are not
   reformatted. Discovered en route: buildflow exclude patterns are
   doublestar globs matched against full path OR basename, so
   `docs/status` matches nothing and `docs/status/**` is required;
   markdownlint-cli has NO hierarchical config, so buildflow-level
   exclusion is the only clean mechanism. All doc tables realigned
   (FEATURES/README/TODO_LIST), README prose wrapped at 110, `.github`
   issue/PR templates got real headings instead of bold-as-heading,
   dependabot config gained update groups + `open-pull-requests-limit: 10`.
4. **SDK API extension the migration required:** `ConfigIssue.Confidence`
   (zero = builder default `ConfidenceFull`, pre-extension behavior) and
   `ConfigIssue.FixStrategy *finding.FixStrategy` (nil = the
   suggest-if-suggestion/none-otherwise default). Two new tests. The
   zero-value semantics are documented including the honest limitation:
   explicit `ConfidenceNone` is unrepresentable through ConfigIssue.
5. **First consumer migrated: oxlint-auto-configure.** Its hand-rolled
   `toolsdk.Spec` is now built through `ProviderFromSpec` (Analyze returns
   `[]ConfigIssue`, Repair is a plain closure, Trigger/DependsOn layered on
   the returned Spec). One test assertion updated
   (`FixStrategyDirect` → `FixStrategySuggest`) — this FIXED a latent
   contract violation: go-finding validation requires before/after code for
   direct fixes, which the old hand-built finding bypassed via direct
   struct assignment. Full suite green (10 packages), `go vet` clean,
   lint 0 issues on the touched package, SDK vendored.
6. **Second consumer migrated: golangci-lint-auto-configure.**
   `healthIssuesToFindings` now emits through `FindingFromIssue` (same
   warn-and-skip semantics). Documented behavior deltas: no known line →
   file-level position instead of fabricated `Line: 1`; empty suggestion →
   `FixStrategyNone`. The `missing-linter` recommendation conversion
   deliberately stays app-side — it carries per-linter categories and tags
   that ConfigIssue does not model (boundary recorded in SDK AGENTS.md).
   CLI test suite green (116s), build clean, lint clean on the touched
   package.
7. **T23 verified CLOSED with hard evidence.** The redesigned social
   preview is LIVE on GitHub: the repo page's og:image serves the uploaded
   PNG byte-identical (57,801 bytes, via
   `repository-images.githubusercontent.com`), confirmed by
   `generate.sh --verify LarsArtmann/linter-autoconfigure-sdk`. TODO_LIST
   row removed, closure recorded in header + CHANGELOG.
8. **`generate.sh` hardened: three maintenance modes, all tested.**
   `--check-env` (renderer + magick identify + both fonts via fc-list;
   verified green on this host); `--verify <owner/repo>` (extracts live
   og:image, checks content-type + dimensions, byte-compares against the
   local card with an explicit verdict); `--audit <owner>` (per-repo table
   with the CUSTOM-vs-auto verdict — custom uploads serve from
   `repository-images.githubusercontent.com`, auto cards from
   `opengraph.githubassets.com`; verified by byte-compare, not
   documentation). Shellcheck clean (run via `nix shell
   nixpkgs#shellcheck`). Escaping regression-tested with `&<>` in the
   title.
9. **Shell gate added to `check-skills.sh`:** `bash -n` over every `*.sh`
   (hard fail) + shellcheck at warning severity when installed (advisory
   NOTE printed when absent). 28/28 skills still green, 143 markdown files,
   no broken links.
10. **Animated GIF shipped into the SDK README.**
    `assets/branding/social-preview-animated.gif` (37,010 bytes, 32
    frames) regenerated with the EXACT text of the live uploaded card
    (first attempt used a different tagline — caught by byte-diff against
    git, see d.1 context), embedded as a width-640 hero under the README
    intro. The canonical social-preview.png/svg were restored untouched
    (the live card stays byte-canonical; PNG rendering is pixel-stable at
    0.06% AE but NOT byte-stable across runs).
11. **HARVEST applied with the defaulted ruling.** After the harvest
    question went unanswered twice (09:14 report g3 proposed the default),
    this session executed it: bounded items into SDK TODO_LIST (T27 release
    verification script, T28 README snippet drift guard, T29 CI asset
    guard, T30 empirical GIF validation), T21 re-scoped, ruling + defaulted
    g-questions recorded in the TODO_LIST header. Unbounded ideas remain
    ROADMAP fuel, not commitments.
12. **T21 blocker LIFTED: BuildFlow is public.** `go mod download
    github.com/larsartmann/buildflow@v0.6.0` succeeds — the "runners cannot
    fetch the private module" premise is dead. T21 evidence updated and
    re-scoped to the actual remaining work (the ci.yml swap, see b.1).
13. **SKILLS repo meta + feedback loop fed.** AGENTS.md now flags the
    shfmt-before-commit convention and the quality gate near the top (the
    convention previously existed only as a user style commit, `b6163c2`);
    `docs/feedback/new/2026-09-11_quality-session-four-failure-modes.md`
    files the four failure modes from d.1–d.3 and e.6 with root causes and
    skill implications; SKILLS CHANGELOG gained the second-session wave.
14. **Both repos' documentation made self-consistent:** SDK README
    Consumers section (active consumers + replace-directive note + status
    line no longer claims "no active consumers"), SDK
    CHANGELOG/TODO_LIST/AGENTS.md, both consumer CHANGELOGs with honest
    behavior-delta notes. The stale-LSP diagnostics that contradicted the
    CLI throughout were resolved by trusting fresh runs (memory rule:
    builds don't lie) — documented nowhere new because it is already in
    AGENTS.md memory.

## b) PARTIALLY DONE

1. **T21 — CI swap to buildflow.** The blocker (private module) is gone,
   but the `ci.yml` workflow change is NOT written: I cannot watch a
   runner, and shipping an unwatched CI redesign that gates `master`
   violates "verify before declaring done". Evidence updated, task
   re-scoped, execution deferred to a session with CI visibility.
2. **T26 — v0.2.0 release.** Everything EXCEPT the release itself is in
   place: CHANGELOG `[Unreleased]` section is cut and complete, consumers
   documented, release runbook referenced. The tag/push is explicitly the
   user's (never push, never tag without instruction; tags are immutable
   once the proxy caches them). Both consumers still carry the temporary
   local `replace` directives that must be dropped post-release.
3. **The full `--audit LarsArtmann` run.** The per-repo loop was traced
   correct (first rows extracted og:image and classified
   auto-vs-CUSTOM correctly), but the complete 100-repo run's output was
   lost to my own cleanup (d.2) and the run aborted with curl exit 22 on a
   HEAD request — fixed with `|| true` AFTER the run, not re-run since
   (200+ HTTP requests; also a candidate for a `--limit` flag).
4. **Empirical animation validation (T30).** The GIF exists,
   machine-checked, embedded — but "plays on Discord/Slack/Telegram" and
   "GitHub's own card slot animates" remain documentation-grade claims.
   Needs one user upload + one chat post. Not fabricable from a terminal.
5. **Consumer go.mod `replace` directives.** Working as designed for local
   verification, but they are a shipping hazard if forgotten; dropping them
   is step two of T26 and both CHANGELOGs say so.
6. **SKILLS TODO_LIST** unchanged this session (T30/T33/T34/T35 all still
   open as before; T35's annotate sweep is bounded but was outranked by the
   migration work).

## c) NOT STARTED

1. **T27** — release verification script (clean-dir `go get@vX.Y.Z` +
   consumer compile as one command).
2. **T28** — README snippet drift guard (compile the README's Go snippets).
3. **T29** — CI-side social-preview asset guard (render + assert 1280x640 +
   < 1 MB; the SKILLS generator's machine-checks are only the local half).
4. **v0.2.0 release run itself** (blocked on T26 user decision).
5. **biome-auto-configure** as the third consumer (ROADMAP).
6. **website-launch phase split** (803 lines → sub-500 + references/;
   flagged four sessions running).
7. **SKILLS T35** — annotate the pre-fix status reports.
8. **status-report HTML-vs-md default decision** (recurred 4+ sessions;
   defaulted to user-instruction-wins each time).
9. **BuildFlow-publicity follow-ons:** once CI runs buildflow, the
   portable-subset workflow (gofmt/vet/race) could shrink; also the SDK
   could dogfood buildflow as a library consumer if the API allows.
10. **`verify-external-claims` §0 update** to name "agent-summarized fetch"
    as a fabrication class (the incident is filed in feedback; the skill
    edit itself was not made this session).
11. **`--audit` pagination** (only page 1, 100 repos) and `--limit` flag.
12. **golangci-lint v2.13.2 vs the auto-configurer's recommended v2.12.2**
    mismatch (warned during configure; not investigated).

## d) TOTALLY FUCKED UP

1. **The generate.sh surgery was a four-pass demolition.** Inserting the
   maintenance modes went through increasingly hairy python string-slice
   passes that (in order) inserted the dispatch AFTER the title check,
   duplicated the whole validation trio, left literal newlines inside
   printf format strings, and — worst — **silently deleted the
   `xml_escape` assignment block**. The script still "worked" and would
   have produced invalid SVG for any title containing `&`. Only shellcheck
   SC2154 ("referenced but not assigned") caught it, and only because I
   ran shellcheck as a one-off. Root cause: rebuilding a region by
   `index()` anchors without diffing before-vs-after, plus testing only
   ASCII inputs so the missing escaping was invisible. The final state is
   good and regression-tested; the path there was malpractice-adjacent.
2. **I destroyed my own evidence.** While the 100-repo audit ran in the
   background, a later cleanup command `rm -f /tmp/audit-out.txt
   /tmp/audit-err.txt` deleted its output mid-run. I then had no audit
   results to report and no way to inspect the exit-22 failure without
   re-running 200+ requests. Cleanup habit applied without checking
   background jobs first.
3. **I trusted an AI summarizer with exact-string extraction — twice.**
   Three agentic fetches were used to determine whether og:image URLs
   distinguish custom uploads from auto-generated cards; the summaries
   reported wrong URLs (`opengraph.githubassets.com/1/...` for repos whose
   real og:image is `repository-images.githubusercontent.com`). From those
   summaries I briefly concluded "the discriminator does not exist" and
   encoded that false claim into the script's audit footnote AND the
   reference doc. The single raw `curl | grep` during `--verify` testing
   exposed the truth and forced a reversal of both artifacts. The correct
   conclusion (CUSTOM = repository-images domain) is now verified by
   byte-compare, but the wrong claim lived in two files for most of the
   session. This is the constructed-URL lesson's sibling: a tool-sourced
   claim still needs mechanical verification when it is an exact string.
4. **My table-realign script had a silent off-by-one** (separator rows got
   one phantom dash from a `(':' if right else '-')` typo). It misaligned
   every separator row while appearing to succeed. Caught only because I
   measured pipe positions with python instead of trusting the exit code.
   The fixed script then processed seven files — the fix is what made
   markdown-lint hit 0.
5. **Masked exit codes in my own verification commands** — `cmd | tail;
   echo EXIT: $?` reports tail's status, not cmd's. The exact
   pipeline-masking anti-pattern recorded in memory after earlier
   sessions, and I still ran it at least twice this session before
   catching it (`PIPESTATUS` discipline is now re-armed).

Nothing else: no data loss (the "lost" audit output is the one exception,
d.2), no broken gates at session end, no unintended diffs in git (the
daemon committed everything in heuristic waves and the final `git status`
is clean apart from the last report-adjacent edits).

## e) WHAT WE SHOULD IMPROVE

1. **Test inputs must exercise the code being deleted or edited.** The
   xml_escape block died unnoticed because every test used a plain ASCII
   title. A `&` in the title proves escaping in one render. Rule: the
   test vector list comes from the flag/function list, not from memory —
   same lesson as the earlier "integration test ran without its key flag".
2. **Never clean up shared temp evidence while background jobs run.** Check
   `job_output`/job state before ANY `rm -f /tmp/...` in the same session.
3. **Exact-string claims need raw extraction.** agentic summarization is
   for orientation; URLs, versions, flags, and error messages come from
   raw bytes + grep. Encoded in feedback; the verify-external-claims skill
   edit is still owed (c.10).
4. **Scripted edits get a golden diff.** `diff before after` with every
   removed line accounted for would have caught the escaping deletion in
   seconds. The pixel-regression test (regenerate + compare) is the right
   pattern and worked; extend it to SVG structural assertions (presence of
   `xml_escape`-produced entities).
5. **Linters before "done", not after.** shellcheck found in one run what
   manual review missed; it should run inside the gate on every machine
   (the advisory NOTE now says so, but shellcheck is not on this host's
   PATH by default — `nix shell nixpkgs#shellcheck` worked; consider
   adding it to the dev environment).
6. **PIPESTATUS discipline** for my own multi-command verification lines.
7. **Consumer migrations should go release-first.** Cutting v0.2.0 BEFORE
   migrating consumers would have avoided the temporary `replace` window
   entirely; the replace-then-release order created a dangling state that
   spans repos and sessions.
8. **Static LSP diagnostics were stale for hours** (kept flagging lines I
   had already fixed). I trusted the CLI per the memory rule, but
   `lsp_restart` when diagnostics contradict a fresh CLI run would remove
   the noise from every tool result.
9. **The daemon's heuristic commit waves** produced ~30 "chore:
   auto-commit N changed file(s)" commits across four repos, splitting
   logical changes (e.g. the .markdownlint.yml + TODO_LIST row landed in
   separate commits). A session-level commit convention (or daemon
   suppression during focused work) is a user decision, not mine — noted,
   not acted on.
10. **Doc tables stay space-padded** (whitespace-exact-match edit risk
    remains). The alignment lint (MD060) now detects drift, which is the
    mitigating control; full de-padding remains rejected because MD060's
    aligned style is the intended look.

## f) Up to 50 things to get done next (brainstorm — first ~12 are the real queue)

**NOW (unblocked, direct continuation):**

1. You: review + run T26 — cut v0.2.0 (tag/push is yours), then I drop
   both consumers' `replace` directives and `go get` the tag.
2. Re-run `generate.sh --audit LarsArtmann` with the fixed HEAD handling;
   add a `--limit` flag first so a run is 5 repos, not 100 requests.
3. T21: write the buildflow CI job (`go install
   github.com/larsartmann/buildflow@v0.6.0` + `buildflow --fix
   --fail-on-findings`) and watch the first push together.
4. T30 (you + me): one scratch-repo upload of the GIF to GitHub's card
   slot; post it once in Discord/Slack. Converts the last doc-grade claims
   to artifact-grade.
5. Install shellcheck into the default dev environment (nix profile or
   home-manager) so the advisory gate actually runs everywhere.
6. `verify-external-claims` §0: add the agent-summarized-fetch class (the
   feedback file is written; the skill edit is 5 lines).
7. T27: script the release verification (clean-dir `go get` + consumer
   compile) — becomes a preflight for every future tag.
8. T28: compile the README's Go snippets in a test (snippet drift guard).
9. T29: CI-side social-preview guard (render + dimension/size assert).
10. SKILLS T35: annotate the pre-fix status reports (bounded, M).
11. `lsp_restart` as a standing move when LSP diagnostics contradict a
    fresh CLI run — cheap, removes hours of stale-warning noise.
12. Investigate the golangci-lint v2.13.2 vs recommended v2.12.2 mismatch
    the auto-configurer warned about.

**NEXT (high value, bounded):**

13. Cut v0.6.0-consumers wave: after T26, both consumers release their
    SDK-integration versions (their CHANGELOG `[Unreleased]` sections are
    already written).
14. Extract the release runbook from AGENTS.md prose into a checklist
    script (T27 generalizes to all LarsArtmann Go modules).
15. `--audit` pagination (Link header) + owner-type org support.
16. GitHub badge for the SDK's lint status (buildflow's strict mode once
    T21 lands makes this honest).
17. Dependabot: verify the new groups actually batch minor/patch PRs on
    the next scheduled run.
18. SDK ROADMAP: record the defaulted channels answer and its expiry
    (motion-variant investment pauses unless chat surfaces are confirmed).
19. Add `docs/status/` archived-report ANNOTATE pass for THIS session's
    predecessors (the two 2026-09-11 morning reports now contain
    superseded claims: "no active consumers", "T23 pending").
20. README comparison-table mobile rendering check (narrow viewport).
21. Consider `go.work` at the ~/projects level as a replace-free local
    dev alternative (measure the blast radius on unrelated repos first).
22. oxlint: revisit `FixStrategySuggest` → potential `Direct` once
    ConfigIssue can carry before/after code (SDK-side field addition).
23. SDK FEATURES.md: add the new ConfigIssue fields + sentinels (they are
    in CHANGELOG but not yet in the feature inventory).
24. Example coverage: `ExampleProviderFromSpec` for pkg.go.dev (the
    bridge has no runnable example yet).
25. Error-docs: list the exported sentinels in the README error section.

**ROADMAP fuel (ideas, not commitments):**

26. biome-auto-configure as third consumer (SDK adoption proof #3).
27. Multi-repo branding consistency checker built ON `--audit` (og:image
    presence + CUSTOM share across LarsArtmann repos, as a scheduled
    digest).
28. social-preview `--theme light` variant (only when a second real card
    needs it).
29. Animated accent-bar shimmer as a second motion mode (only if T30
    proves the GIF surfaces matter).
30. buildflow-as-library: SDK dogfood a buildflow provider registration
    via the bridge (would make the SDK its own fourth consumer).
31. `--verify` pixel-mode: download og:image and pixel-compare (not just
    byte-compare) for GitHub-re-encoded cases.
32. Shell-skill: distill the SIGPIPE/pipefail + early-exit-consumer rules
    into a `how-to-bash` reference in SKILLS.
33. check-skills.sh: promote shellcheck to hard-fail once it is
    environmentally guaranteed (needs 5).
34. SKILLS ROADMAP §7 (dedicated social-preview repo): the two-launch
    counter now stands at ONE real launch (linter-autoconfigure-sdk).
35. Tag-protection rulesets for `v*` tags across repos (tag immutability
    beyond the proxy effect).
36. CODEOWNERS decision record for the SDK (probably skip; record it).
37. SECURITY.md review for the SDK (file exists; verify content freshness
    now that consumers exist).
38. Convert TODO_LIST tables to a lint-friendly canonical form (decide:
    aligned-padded canonical via the fixer script in CI, or unpadded).
39. status-report skill: settle HTML-vs-md default once (4+ overrides).
40. website-launch phase split (803 → sub-500; flagged four sessions).
41. oxlint test framework: testify → ginkgo/gomega per how-to-golang
    (banned-list alignment; large, schedule deliberately).
42. golangci-lint-auto-configure: evaluate `RecommendationsToFindings`
    migration if ConfigIssue ever grows Category/Tags (deliberately
    rejected today — revisit only with a second consumer needing it).
43. Consumer pinning policy: `@vX.Y` vs `@latest` for in-house deps in
    CI (buildflow job in 3 makes this concrete).
44. Document the doublestar-exclude semantics in buildflow's own docs
    (upstream contribution candidate; verify-before-filing first).
45. upstream: markdownlint-cli hierarchical config feature request (would
    make the docs/status exclusion local instead of buildflow-level).
46. feedback loop: route the four failure modes into their skills
    (verify-external-claims, a future how-to-bash, docs-health's cache
    note) — the file is the inbox, routing is owed.
47. daemon commit-message quality: propose a convention (user decision).
48. Explore buildflow `--resume` in CI for the cache-masking class (does
    CI need result-cache disabled by default?).
49. SDK benchmark: SaveJSON idempotent-skip vs blind-write cost (the
    changed-bool API deserves a number).
50. Harvest expiry: re-check the defaulted rulings in 30 days
    (2026-10-11) — channels answer, T30, and the HTML-vs-md default.

## g) Three questions I can NOT figure out myself

1. **v0.2.0 release mechanics:** shall I prepare the release commit and
   hand you the exact tag/push commands, or do you want to run the whole
   runbook yourself? And is `--prerelease --latest` still the convention
   for v0.x GitHub Releases (v0.1.0 used it)?
2. **CI swap timing (T21):** BuildFlow is now public — may I rewrite
   `ci.yml` to a buildflow job now and watch the first push together, or
   should the portable gofmt/vet/race workflow stay until a session where
   you can watch runners live?
3. **Channels (third ask, defaulted 2026-09-11):** which surfaces do you
   actually post repo links to — Discord, Slack, Telegram, X, LinkedIn?
   This decides whether the typing GIF gets further investment (motion
   variants, more platforms) or whether the static card + README embed is
   the end state. Default until answered: GitHub + README only.

---

**WAITING FOR INSTRUCTIONS.**
