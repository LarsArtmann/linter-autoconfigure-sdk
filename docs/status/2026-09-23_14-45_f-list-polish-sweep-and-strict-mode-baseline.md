# Polishing Sweep — f-List Execution + Strict-Mode Truth-Setting

**Date:** 2026-09-23 ~14:45 CEST · **Repo:** `linter-autoconfigure-sdk` (master)
**Predecessor report:** `2026-09-23_13-26_masterplan-wrapup-self-review.md` (its f-list drove this session)

---

## Closed from the predecessor's f-list

| Item | What was done | Verification |
| --- | --- | --- |
| f#1 README Consumers stale | oxlint row now says `BootstrapProviderFromSpec` (since its v0.9.0) + SDK I/O/discovery/diff; golangci row adds `Change`/`Kind` vocabulary + jsondeterminism gate. Claims grep-verified in both local consumer repos at use-time (lesson d/#5 applied) | local greps hit `pkg/provider/provider.go` (oxlint) and `internal/cli/cmd_validate.go` + `pkg/diff/differ.go` (golangci) |
| f#2 `ExampleFirstExisting` | Added, covering both the found path and the none-exists fallback (`first candidate + false`) | `go test -run ExampleFirstExisting` green; shows on next pkg.go.dev render |
| f#3 CHANGELOG Unreleased | Four Added entries: verify-release.sh, README snippet guard, social-preview guard, license-check skip_steps | readme_snippets guard still green |
| f#7 strict proof | Strict (`--fail-on-findings`) is NOT green: **9 advisory findings, all steps 38/38 pass**. All 9 dispositioned (see below) and documented in AGENTS.md as the known baseline | full strict run captured; findings enumerated |
| f#10 pkg.go.dev render checks | v0.3.1, v0.4.0, v0.5.0 all fetched directly: full index + examples + docs render. (v0.5.0 page carries the known stale Consumers text — unrecoverable at tag, fixed on master) | direct fetches 2026-09-23 |
| f#18 package doc | Plumbing bullet list now includes diff engine, discovery, and bootstrap lifecycle | vet/gofmt clean |
| f#19 Go Reference badge | Already present on master — verified, no change needed | README head |
| f#42 masterplan traceability | Header now links all three closing status reports | - |
| f#43/44 GOEXPERIMENT retirement | ci.yml env block + comment removed. Discovery: the repo's `.envrc` NEVER exported it directly; the export came from `~/.config/direnv/lib/zz-smart-nix.sh` `use_go_env` (user-global, not repo-owned — nothing to retire in-repo). AGENTS.md GOEXPERIMENT section rewritten to match reality | grep: no GOEXPERIMENT in repo outside docs/history |
| f#45 ProviderFromSpec Inputs rationale | Field-mapping bullet now explains WHY every recognized filename is an Input (re-run on any appearance/change, curated formats never stomped) | - |
| f#47 verify-release.sh surface | Already covers `FirstExisting` + `WorkingDir` ("Discovery + working dir" smoke section) — verified, no change needed | script read |

## f#7 disposition of the 9 strict-mode advisories (all accepted, documented)

1. branching-flow 5: nil-deref at `bootstrap.go:318` is a **proven false positive**
   (`ParseJSON` returns `new(T)`-backed non-nil on success; the analyzer cannot see
   across functions). The other 4 are phantom-type style suggestions for
   `Change.Old/New` and `FixCommand`/`CountLabel` — plain strings by design.
2. cqrs-lint 2: false positives — go.mod contains no go-cqrs-lite.
3. go-auto-upgrade 2: samber/lo adoption nudges; SDK keeps third-party deps to the
   larsartmann family.

AGENTS.md "Build, test, lint" now states this baseline (replacing the stale
"strict is green up to 2 warnings" text).

## New environment knowledge (in AGENTS.md)

- Non-direnv shells: `GOTOOLCHAIN=local` beats `.buildflow.yml` env; working
  recipe is `env -u GOTOOLCHAIN PATH=<nix go-1.27>/bin:... buildflow ...`.
- The system govulncheck is built with go1.26 and fails against 1.27 stdlib
  with ~24 bogus findings; install a matching one
  (`GOBIN=... go install golang.org/x/vuln/cmd/govulncheck@latest` under 1.27).
- BuildFlow `--format finding` currently errors on its own output
  (`finding.Tags[1] "original-severity:error" is invalid: tags must be
  non-empty`) — a BuildFlow-side report bug, worth upstreaming when T21's
  visibility question resolves.

## Also verified

- T21 precondition re-checked at use-time: `gh api repos/LarsArtmann/BuildFlow`
  → still private (2026-09-23). Stays blocked.
- f#6: golangci daemon commit `96b0afc` inspected — a 2-line go.sum prune;
  `go build ./...` green after it.
- Full `go test ./...` green (incl. README snippet guard), gofmt clean, vet
  clean, jsondeterminism analyzer clean.

## Still open (owner-gated)

- g1 (file T43 go-finding issue), g2 (curate golangci release notes), g3 (T21
  unblock path), T30 (GIF empirical validation — owner hands), T20 (v1 gate).
- f#4/f#5 (consumer AGENTS.md syncs) and the remaining fleet items are
  consumer-repo sessions, not done here.

*Assisted-by: Crush <crush@charm.land>*
