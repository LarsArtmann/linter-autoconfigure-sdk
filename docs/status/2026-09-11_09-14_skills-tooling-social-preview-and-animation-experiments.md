# Status Report — 2026-09-11 09:14 CEST

## Social-preview tooling, SKILLS repo upgrades, and the animation experiment

Session scope: continuation of the 07:45 report's session — SDK TODO sweep,
the social-preview redesign, its URL-fabrication incident, the SKILLS repo
upgrades (generator, platform matrix, `--animate`), and this report. Two
repos touched: `linter-autoconfigure-sdk` (docs) and `SKILLS` (skills +
scripts + docs). Format override: Markdown per explicit user instruction
(the status-report skill's canonical default is a styled HTML dashboard;
flagged per skill contract, one-off, not propagated).

---

## a) FULLY DONE

1. **Social preview card redesigned end-to-end.** Original critiqued (four
   competing text blocks, module path duplicated, no focal anchor); rebuilt
   as versioned SVG (`assets/branding/social-preview.svg`) + PNG: kicker
   chip, cyan-hyphen title, single tagline, terminal chip with traffic-light
   dots, Go-brand gradient bar. Verified at full size AND 320×160 card size,
   1280×640, 56 KB (< 1 MB). The redesign process itself is encoded in
   `website-launch/references/social-preview.md`.
2. **The URL-fabrication incident closed properly.** My guessed
   `/settings/social-preview` deep link 404'd; verified against
   docs.github.com that the section lives on General settings with NO deep
   URL; correct click path recorded in TODO_LIST evidence and the reference.
3. **SKILLS repo audited and upgraded.** Gate green (28/28 skills, 0 thin,
   no broken links); install locations are symlinks into the repo — zero
   drift, edits ship instantly. Four changes shipped: (a) new
   `website-launch/references/social-preview.md`; (b) `verify-external-claims`
   §0 now names constructed URLs as the top fabrication class (with today's
   incident); (c) `status-report` commit step defers to harness policy;
   (d) `website-launch` Phase 6 pointer. CHANGELOG wave entry written.
4. **`generate.sh` — parameterized card generator.** Flags
   `--title/--tagline/--install/--kicker/--output-dir`; computes title and
   install font sizes from the monospace 0.6em advance (auto-shrink, honest
   failure past the one-line floor); machine-checks GitHub's limits;
   renders the card-size thumbnail; prints the manual upload path.
   Validated: 0.06% pixel diff vs the hand-authored card, short-title cap,
   long-path auto-shrink.
5. **Platform matrix researched and encoded** (five parallel agents against
   primary sources, 2026-09-11): animated GIF plays only on Discord/Slack/
   Telegram; X/LinkedIn freeze first frame (X's docs verbatim: "Only the
   first frame of an animated GIF will be used"); Mastodon/Reddit static;
   WebP/APNG/animated-AVIF browser support is irrelevant because the
   crawler is the gatekeeper; LinkedIn lacks WebP; animated SVG works
   nowhere. Encoded into the reference with per-row verification status.
6. **`--animate typing` shipped and verified.** Native delta-optimized GIF:
   32 frames, 37 KB measured, loop ≈ 4 s. The architectural win: the static
   card IS the animation's final frame (one SVG builder emits both), so the
   artifacts cannot drift. Tested: animate mode, static regression, error
   paths, frame-count report. Four bugs were caught by frame-by-frame
   verification BEFORE shipping (see e.13–16).
7. **ROADMAP theme 7 added** (SKILLS): dedicated social-preview repo is a
   post-2-real-launches consideration, with the day-one rejection recorded.
8. **Quality gates green throughout:** SDK `buildflow` exit 0 repeatedly;
   SKILLS `check-skills.sh` 28/28 after every change (143 markdown files,
   links resolve). Daemon committed both repos continuously; SDK log
   `1fff6c5`, SKILLS log `0605ba6` (also `b6163c2` — a shfmt reformat of
   generate.sh, which explains the two "modified since read" tool trips).

## b) PARTIALLY DONE

1. **T23 — social preview upload.** Still manual (no API; no deep URL). The
   asset is ready at `assets/branding/social-preview.png`; post-upload
   og:image verification not yet possible (needs the upload first).
2. **Animation empirical validation.** The GIF exists and is
   machine-checked, but "plays on Discord/Slack/Telegram" rests on platform
   documentation, not on OUR artifact being posted. GitHub's own card-slot
   animation behavior remains unverified. Not fabricable from a terminal —
   needs one scratch-repo upload + one chat post.
3. **README embed for the GIF.** The animated asset's guaranteed home is
   the SDK README, but no embed snippet was added to README.md yet (didn't
   touch README without your word — it's user-facing copy).
4. **Status-report HARVEST obligation.** Both this and the 07:45 report's
   section-f lists are still brainstorm-only; docs-health HARVEST into
   TODO_LIST/ROADMAP has not run (you instructed wait-for-instructions;
   the skill flags this as an open loop).
5. **SDK TODO_LIST/CHANGELOG** from phase 1 are consistent, but the
   three questions from the 07:45 report (BuildFlow ETA, consumer priority,
   CHANGELOG policy) remain unanswered, so T21 evidence stays
   single-source and the docs-only-CHANGELOG policy remains my unilateral
   call.

## c) NOT STARTED

1. **T20** — v1-gated by design (correct).
2. **T21 execution** — blocked: BuildFlow still private (404 re-check
   today); the automated proxy-detector job (prior report f5) unbuilt.
3. **First consumer migration** — golangci-lint-auto-configure /
   oxlint-auto-configure onto the SDK. pkg.go.dev still shows
   "Imported by: 0". Biggest leverage item in the whole portfolio; not a
   TODO_LIST item yet, only README "Planned".
4. **SKILLS TODO_LIST items** — T33 (site videos), T34 (linter-building
   decisions, blocked on you), T35 (annotate pre-fix status reports), T30
   (blocked on first real jj PR run).
5. **markdownlint MD013 fix** in the SDK (would make
   `buildflow --fix --fail-on-findings` green) — recommended twice, unbuilt.
6. **Website-launch 802-line phase split** — flagged three times now.

## d) TOTALLY FUCKED UP

1. **Nothing.** No data loss, no broken gates, no shipped bugs, no
   unintended diffs. The four animation bugs (self-deleting prototype, `$$`
   double prompt, truncated final frame, regressing hold frames) were all
   caught in the /tmp prototype by frame-by-frame verification and never
   reached the repo — that is the process working, and they are recorded as
   engineering notes in the reference, not hidden.
2. **Closest miss, stated honestly:** I again reached for memory before
   primary sources on VHS tape syntax — the unquoted `Output` path failed
   to parse, and only after the failure did I notice `vhs new` (the tool's
   own example/docs generator) sitting in `--help`. The fabricated-URL
   lesson from the 07:45 report, same shape, smaller blast radius.

## e) WHAT WE SHOULD IMPROVE (brutal self-review)

1. **Golden-frame-first ordering.** All four animation bugs shared one root
   cause: I animated before rendering the final frame and comparing it to
   the static card. Rendering frame-final FIRST would have caught the `$$`,
   truncation, and regression bugs in one round trip. The integrated script
   now has that property by construction (static = final frame), but the
   prototype phase paid tuition for it.
2. **`md5sum` was the wrong comparator** for PNGs (magick re-encodes; bytes
   differ, pixels identical). Pixel AE is the tool. Caught it, but only
   after confusing myself for one round.
3. **Integration test ran without its key flag.** First `--animate` test
   omitted `--animate` — I verified the wrong thing and burned a round
   trip. Test checklists should come from the flag list, not memory.
4. **Formatter guard trips.** The SKILLS repo formats shell via shfmt on
   commit (`b6163c2`); I hit "modified since read" twice. I read
   AGENTS.md's first 40 lines only — the tooling conventions live deeper.
   Read repo meta fully before writing files, not after the first collision.
5. **`magick identify -format '%n'`** prints per-frame counts (the
   `3232…32` report bug). For GIFs: `identify | wc -l`. Trivial, but it
   shipped in the first integrated run — the machine-check section deserves
   the same frame-review skepticism as the pixels.
6. **Empirical gap is the honest weak spot:** every platform-animation claim
   in the reference is documentation-grade, not artifact-grade. The
   reference hedges correctly, but the hedge should be temporary.
7. **Stupid things we do anyway (noted, unfixed):** TODO_LIST tables are
   space-padded (whitespace-exact-match risk + MD013 flood); release
   runbook is AGENTS.md prose, not a checklist; SKILLS AGENTS.md tooling
   conventions (shfmt/dprint) are buried mid-file where newcomers trip on
   them; the daemon's commits go unpushed (SKILLS remote state never
   checked — unknown whether work is safely off-machine).
8. **Question hygiene:** three questions from the 07:45 report are still
   open. Questions that go unanswered twice should be converted into
   defaults with an expiry, not re-asked forever.

## f) Up to 50 things to get done next (brainstorm — most are ROADMAP fuel)

**NOW (unblocked, this session's momentum):**

1. You: upload the social preview (Settings → General → Social preview →
   Edit) — closes T23; then I verify og:image live.
2. Empirical animation test: post the GIF once in Discord/Slack and embed in
   the SDK README — converts documentation-grade claims into artifact-grade.
3. Scratch-repo test for GitHub's own card-slot GIF animation (the last
   unverified matrix cell).
4. Add the GIF embed snippet to the SDK README (needs your wording OK).
5. `--verify <owner/repo>` on generate.sh: post-upload og:image check.
6. `--audit` sweep: list LarsArtmann repos missing social previews.
7. `--check-env` doctor: fc-list guards for JetBrains Mono/Noto Sans (the
   0.6em math silently breaks under font substitution).
8. Shell gate in SKILLS: `bash -n` + shellcheck over all `scripts/*.sh` in
   check-skills.sh (the scripts surface keeps growing).
9. File today's failure modes as `docs/feedback/new/` entries (SKILLS
   feedback loop; constructed-URL + VHS time-box + the four animation bugs).
10. Read SKILLS AGENTS.md §tooling conventions (shfmt/dprint) before next
    file write; add a one-line pointer near the top for future sessions.
11. SDK: fix markdownlint MD013 config → strict buildflow green.
12. SDK: reformat TODO_LIST/CHANGELOG tables unpadded → kills the
    whitespace-edit risk at the root.
13. Convert the 07:45 report's three unanswered questions into defaults
    with an expiry date (see g).

**NEXT (high value, bounded):**

14. SDK: migrate golangci-lint-auto-configure onto the SDK (kills
    "Imported by: 0" — the single biggest leverage item).
15. SDK: migrate oxlint-auto-configure.
16. Cut SDK v0.2.0 after the first consumer merge.
17. Script the SDK release verification (clean-dir `go get` + consumer
    compile).
18. Guard README code snippets against drift (doc-snippet compile test).
19. Dimension/size guard CI check for social-preview assets.
20. Extract website-launch's release/launch prose into checkbox runbooks.
21. Protect `v*` tags via GitHub rulesets (tag immutability).
22. Update README "Consumers" after each migration.
23. CI + pkg.go.dev badges in the SDK README.
24. Scheduled proxy-check job for BuildFlow publicity (automates T21).
25. Cross-check T21 evidence via `go list -m github.com/larsartmann/buildflow@latest`.
26. Post-release T+24h verification checklist as a runnable script.
27. GOEXPERIMENT=jsonv2 teardown plan for Go 1.27.
28. Upstream go-finding `go 1.26.7` patch-floor fix (verify-before-filing
    first).
29. SKILLS: docs-health pass over docs/status/ (ANNOTATE/ARCHIVE the ten
    resolved reports; T35).
30. Renovate config for SHA-pinned GitHub Actions.
31. Coverage badge/threshold from the SDK's CI artifact.
32. `ExampleFindingsFromIssues` (missing from pkg.go.dev example set).
33. Check the README comparison table's narrow/mobile rendering.
34. GitHub repo topics/description refresh across LarsArtmann repos.
35. SECURITY.md for the SDK (cheap, pre-consumers).
36. CODEOWNERS decision (probably skip; record it).
37. HARVEST both reports' f-lists into TODO_LIST/ROADMAP properly (g3).
38. Website-launch phase split (802 → sub-500 with references/).
39. Status-report HTML-vs-md default: decide (recurred 3+ times).
40. `how-to-write-skills.md` location decision (ROADMAP open question).

**ROADMAP fuel (ideas, not commitments):**

41. biome-auto-configure as the third consumer.
42. Social-preview theme variants (light theme, non-Go kicker) — only after
    a second real card needs one.
43. Animated accent-bar shimmer variant (second motion mode) — only if #2
    proves the GIF surfaces matter to you.
44. Multi-repo branding consistency checker (og:image presence + size).
45. Tag→release automation once tooling allows.
46. evaluate skill-creator's eval machinery for top-5 skills' triggering.
47. ROADMAP §7 graduation: dedicated social-preview repo after 2–3 real
    launches (criteria already written).
48. jj-fork-pr-workflow: first real PR run → flip T30 / README 🆕→🟢.
49. SKILLS backup-retention cleanup decision (ROADMAP open question, aged).
50. Record whether SKILLS has a remote and whether the daemon pushes to it
    — off-machine safety is currently unverified.

## g) Up to 3 questions I can NOT figure out myself

1. **Channels:** which do you actually use to share repo links — Discord,
   Slack, Telegram, X, LinkedIn? This decides whether `--animate` was worth
   building (and whether motion variants get more investment) or whether
   the static card plus README embeds are all that matter.
2. **Empirical validation:** may I spend one scratch-repo cycle on the
   unverified cells (GitHub card-slot GIF animation + real unfurl tests)?
   It needs your GitHub hands for the upload and one chat post from you.
3. **Harvest ruling:** should I run docs-health HARVEST now on both
   reports' f-lists into the two repos' TODO_LIST/ROADMAP, or keep them as
   brainstorm until you pick items? (Second time asking — if unanswered
   again, I propose the default: harvest bounded items only, leave the rest
   in ROADMAP, and note the ruling as defaulted.)

---

**WAITING FOR INSTRUCTIONS.**
