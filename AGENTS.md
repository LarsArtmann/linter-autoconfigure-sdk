# linter-autoconfigure-sdk — Project Context

Enduring context for AI sessions working in this repo. Domain-facing docs live in the README; this file captures what
is hard to discover from the code alone.

## What this is

Shared foundation (Go library) for linter auto-configuration tools. Owns the plumbing that
`golangci-lint-auto-configure` and `oxlint-auto-configure` duplicate: config file round-trip, finding emission for
config issues, the config diff engine, and the BuildFlow provider bridge (`ProviderSpec` → `ProviderFromSpec` →
go-finding's canonical `toolsdk.Spec`). YAML parsing is intentionally NOT here (each tool uses a different YAML
library). Public repo, MIT-licensed. Boundary decision: linter-recommendation findings with per-linter
categories/tags (golangci's `missing-linter`) stay app-side; ConfigIssue models config-health issues only (no
Category/Tags fields by design). Ecosystem: `go-finding` (finding model + toolsdk contract) and `go-atomic-write`
(write primitive) are dependencies; `go-linter-sdk` is a sibling, NOT a dependency or consumer — it scaffolds linters
that find code issues, this SDK owns config-file plumbing; orthogonal layers (decided 2026-09-10, re-confirmed
2026-09-23). Session orientation: `TODO_LIST.md` is the open-work queue, `FEATURES.md` the honest feature inventory.

**Exported API inventory (unchanged since v0.6.0):** I/O — `ReadConfig`, `LoadJSON[T]`, `ParseJSON[T]` (path-less),
`MarshalJSONIndented` (deterministic + 2-space), `SaveJSON` (if-changed, atomic), `SaveJSONBytes` (byte-faithful;
trailing newline is CALLER's contract), `WorkingDir(ctx)` (nil-ctx safe, "." fallback). Diff —
`Change{Kind,Path,Old,New}`, `Kind` string enum (KindAdded/KindRemoved/KindModified — deliberately NO unchanged
kind), `DiffMaps`, `DiffSets`, `DiffBlobs` (set semantics), `StringValue`, `Summary` ("Modified:", not "Changed:"),
`FormatDiff` (+/-/~ lines, "No changes." when empty). Discovery — `ProviderSpec.ConfigFiles []finding.FilePath`
(Inputs derive from it; `ConfigFile` remains the single-file write target), `FirstExisting(root, candidates...)`
(first candidate + false when none exist). Provider — `ProviderFromSpec` (+ `HasRepair()`), validation sentinels
`ErrNameRequired`/`ErrDescriptionRequired`/`ErrAnalyzeRequired`, deprecated `ErrNoRepair` (removal at v1 = TODO T20).

**Determinism enforcement (v0.6.0):** the `determinism` subpackage + `cmd/jsondeterminism` vettool flag bare
`encoding/json/v2` Marshal calls (the pre-v0.3.1 `SaveJSON` bug class). CI runs `go run ./cmd/jsondeterminism ./...`
after vet. `json.Deterministic(false)` is the explicit opt-out; opaque `opts...` spreads are not flagged
(conservative). Bite-check: `determinism/testdata` analysistest.

**Consumer migration state (2026-09-23):** both consumers build against TAGS (no replace directives remain in the
fleet). oxlint-auto-configure (v0.9.1) is FULLY on the SDK: I/O helpers, diff engine aliases `Change`/`Kind`,
`ConfigFiles`+`FirstExisting` discovery, AND the bootstrap provider lifecycle via `BootstrapProviderFromSpec` (its
`validate` also runs a drift advisory); its gate + CI run `cmd/jsondeterminism`. golangci-lint-auto-configure
(v0.10.0) uses `FindingFromIssue`, the diff engine (`ChangeType = autoconfigure.Kind`), atomic config/backup writes
via go-atomic-write behind `config.NewOSFS()`, and `json.Deterministic(true)` on every marshal (same analyzer in its
gate + CI). BuildFlow pins the SDK v0.6.0 as indirect only. Third consumer: go-version-auto-configure (vendored by
BuildFlow) is the second live `ProviderFromSpec` user — surfaced by the 2026-10-03 toolsdk audit.

**toolsdk utilization audit (2026-10-03, `docs/research/2026-10-03_go-finding-toolsdk-deep-dive.html`):** adoption
80/100 post-fix. Core conversion, dry-run channel, anti-lie RepairResult, and the registry boundary are fully
leveraged. (1) FIXED in-tree — nil-ctx panic in bootstrap repair (`toolsdk.DryRunFromContext` on a nil ctx, the
v0.4.1 bug class): every derived capability normalizes via `toolsdk.EnsureContext`; regression tests
`TestBootstrapNilContext` and `TestProviderFromSpec_NilContext_NormalizedBeforeClosures`. (2) TODO T22 — the
declarative surface is unbridged (Trigger / DependsOn / ModuleFanOut / Options / extra Inputs, HealthCheck on
ProviderSpec): both live provider consumers mutate the returned Spec post-hoc; go-version also under-declares Inputs
(reads go.work/.github/nix but declares go.mod only, weakening BuildFlow result-cache invalidation) and hand-rolls a
nil-unsafe copy of the SDK's exported `WorkingDir`. Drift-as-HealthCheck verified sound against BuildFlow source
(failures are warn-only, surfaced in the workflow summary).

Module: `github.com/larsartmann/linter-autoconfigure-sdk`. Requires Go 1.27+ (minor-form floor, see gotcha),
`github.com/larsartmann/go-finding` + `go-finding/toolsdk` (always latest — the canonical BuildFlow provider
contract, consumed via `ProviderFromSpec`), and `github.com/larsartmann/go-atomic-write` (always latest —
`SaveJSON`'s idempotent, crash-durable write primitive).

**go.mod gotcha (recurring):** patch-form floors are dependency-imposed and move with the dep graph — never
hand-normalize the `go` directive, run `go mod tidy` and accept the result. History: go-finding declared `go 1.26.7`
(pre-v0.3.0), then `go 1.27` (v0.3.0), then `go 1.27.1` (v0.3.1–v0.6.0, an upstream dep's patch floor); since
2026-09-23 (`golang.org/x/tools` v0.50.0 joined as a direct dep for the analyzer) tidy settles this module at `go 
1.27` — verified stable under a fresh tidy. Consumers' floors ride along.

**Always stay on the latest go-finding version.** Since Go 1.27, `encoding/json/v2`/`jsontext` are standard; the
GOEXPERIMENT=jsonv2 era is fully retired — the repo carries no GOEXPERIMENT anywhere (a 1.26 build is impossible
under the 1.27 floor anyway).

## Build, test, lint

Automated with **BuildFlow** (no Makefile) + **flake.nix** (2026-10-04, go-standard template from go-nix-helpers):
`use flake` provides the devShell (Go 1.27.1, GOTOOLCHAIN=local), `nix build` the package, `nix flake check` passes.
All larsartmann deps are public-on-proxy but are wired via `deps` + local flake inputs anyway: go-standard always
marks `larsartmann/*` private, so a deps-less FOD bypasses the proxy into sandbox-blocked direct VCS. Gotchas:

- **Store symlinks must never enter the flake source.** The daemon twice committed /nix/store symlinks (`.direnv/` GC
  roots 2026-10-04, the `result` build-output link 2026-10-09); a foreign store target is invisible in the sandbox, so
  the prepared-source FOD dies under `noBrokenSymlinks` and nix-build, nix-hash-fix, nix-build-verify, nix-flake-check
  all cascade-fail with misleading hash/fix loops. Structural fix (2026-10-09): flake.nix sets `src` to
  `lib.fileset.toSource { root = ./.; fileset = lib.fileset.gitTracked ./.; }` (fleet precedent: go-finding), so
  untracked/ignored junk (`result`, `.direnv/`, `.envrc`) never enters the source regardless of tree state. Keep them
  untracked anyway.
- **vendorHash staleness is a cascade symptom, not a root cause.** Repair steps (go-mod-tidy/go-mod-update) can
  change go.sum; after the daemon commits it, `nix build` fails with a hash mismatch while nix-hash-fix loops with "no
  fixable findings" — the REAL failure is a different drv (read `nix log <drv>`). Fix the real error first, then
  `buildflow -s nix-hash-fix --fix` settles the hash (never paste hashes by hand). Its verify builds all flake checks,
  so an unformatted flake.nix (treefmt) blocks the hash write — run `nix fmt` first.
- `enableCheck = false` in flake.nix: `TestREADMESnippetsCompile` builds README snippets in a temp module that cannot
  resolve deps under the sandbox's GOPROXY=off. Tests run in the devShell (`go test ./...`), not the hermetic build —
  same class as project-discovery-daemon's check disable.
- `buildflow` — full quality pipeline (detect); `buildflow --fix` — detect + auto-fix; both exit 0 when healthy, and
  the gate fails only on remaining error-severity findings. Known advisory baseline (2026-10-09, plain mode exits 0;
  all detect-only warnings): art-dupl 8 (10-14 token clones in `*_test.go` table-test setup, verified pre-existing on
  68cc3d7), branching-flow 1 (nil-deref at `bootstrap.go` `return *parsed` is a proven false positive — `ParseJSON`
  returns `new(T)`-backed non-nil on success), file-size-check 3 (files over 350 lines), govulncheck/vulnix (Go stdlib

* nixpkgs closure CVE advisories), nix-flake-check arch-omission note. Earlier baselines (branching-flow phantom-type
  suggestions for plain-string `Change.Old/New` and `FixCommand`/`CountLabel`, cqrs-lint false positives on a repo that
  imports no go-cqrs-lite, stdlib2lo samber/lo nudges under the go-auto-upgrade name) resolve or rename with tool
  updates; the SDK deliberately keeps third-party deps to the larsartmann family. Revisit only if advisories start
  flagging real code.

- `go test -race -count=1 ./...` — just the Go tests (covered by buildflow `test-race`/`test-coverage`).
- Known tool bug (2026-09-22; auto-deferred since 2026-10-07): `license-check` (go-licenses) cannot load go 1.27
  toolchain stdlib packages ("Package log/slog does not have module info"). Unrelated to code state; the
  `.buildflow.yml` skip was REMOVED — BuildFlow auto-defers the failure class (visible no-op + INFO finding), so the
  pipeline is green without config.
- `securitymd` requires a 'Security Practices' section in SECURITY.md (added 2026-10-09 after it gated an error).

Single step: `buildflow -s <step> -v`. Release verification is scripted: `./scripts/verify-release.sh vX.Y.Z` (proxy
check + clean-dir go get + compile-every-API smoke — run after every tag). CI also runs the jsondeterminism analyzer,
the README snippet compile guard (`readme_snippets_test.go` — caught real snippet drift on first run), and the
social-preview guard (`scripts/check-social-preview.sh`). A BuildFlow CI job is NOT possible yet: the BuildFlow repo
is PRIVATE again (2026-09-23; the 2026-09-11 public verification is stale) — T21. Debugging:
`BUILDFLOW_NO_RESULT_CACHE=1 buildflow ...` — the result cache has a 168h TTL and can serve stale green results after
a tool upgrade changed the verdict (this masked a real golangci-lint-auto-configure gate failure on 2026-09-11).

**Non-interactive shells** (AI tool shells, cron — direnv skipped):

- Run Go inside the devShell: `nix develop -c bash -c 'go test ./...'`. Do NOT hand-build PATH from `ls -d 
/nix/store/*-go-1.27*/bin` — the first match can be an arm64 Go that fails x86 builds with bogus `gcc_arm64.S`
  assembly errors (observed 2026-10-09).
- The system go 1.26.7 fails the go.mod floor under GOTOOLCHAIN=local, and the system govulncheck (built with 1.26)
  fails against 1.27 stdlib with ~24 bogus findings. Verified buildflow recipe (2026-09-23): `env -u GOTOOLCHAIN 
BUILDFLOW_NO_RESULT_CACHE=1 PATH=<go127-bin>:/tmp/gv127:$PATH buildflow --fail-on-findings` where `/tmp/gv127` holds
  govulncheck BUILT WITH 1.27 (`GOBIN=/tmp/gv127 go install golang.org/x/vuln/cmd/govulncheck@latest`).
- `go env -w` does NOT work on this NixOS host (read-only `~/.config/go/env`).

## Lint configuration

- `.golangci.yml` is GENERATED by `golangci-lint-auto-configure` (`buildflow --fix -s golangci-lint-auto-configure`):
  ~106 linters plus a 5m run timeout. Hand-edits may be overwritten by the next fix run; the flat-layout header comment
  survives regeneration.
- `.markdownlint.yml`: MD013 at 120 with tables/code_blocks exempt; MD024 siblings_only (changelog sections repeat
  heading names by convention).
- `.buildflow.yml` excludes `docs/status/**` and `docs/planning/**` from all checks: point-in-time historical records
  must not be reformatted or link-checked against rotted references. markdownlint-cli has NO hierarchical config (a
  subdirectory `.markdownlint.yml` is ignored), so buildflow-level exclusion is the only clean mechanism;
  `docs/status/.markdownlint.yml` exists only for direct markdownlint runs inside that dir. Exclude patterns are
  doublestar globs matched against the full path OR the basename: `docs/status` matches nothing, `docs/status/**`
  matches everything under it.
- erraudit enforces that error wraps on error paths include every in-scope context variable (e.g. a wrap in
  `FindingFromIssue` must carry `toolName`, not just the rule name).
- `ProviderFromSpec` validation sentinels carry the same messages the old dynamic errors had; match with `errors.Is`.

## CI (GitHub Actions)

`.github/workflows/ci.yml` runs the portable subset on push/PR: gofmt check, `go vet`, race tests with the coverage
artifact uploaded. BuildFlow is a private module — NOT in CI; swap the workflow for a buildflow job if/when it goes
public (T21). Branch protection on `master` blocks force pushes and deletions but does not enforce on admins, so the
auto-commit daemon keeps working; the `Build, vet, test` check is a required status check (admins bypass it, PRs must
pass it).

## Releases

Procedure: cut CHANGELOG, sync README/FEATURES into the release commit, wait for CI green on that exact commit, `git 
tag -a` + push tag + `gh release create --latest` (a PLAIN release marked Latest — GitHub's API rejects
`--prerelease` combined with Latest with HTTP 422, verified live at v0.3.1; drop the flag entirely even for v0.x),
then verify proxy (`go list -m -versions`) and clean-dir `go get@vX.Y.Z` + consumer compile. ALWAYS run the
adversarial clean-dir smoke BEFORE tagging: the v0.4.0 smoke caught a `WorkingDir(nil)` panic after the tag was
already immutable, forcing the v0.4.1 hotfix. Gotchas seen live: the daemon does not always push promptly (manual
`git push origin master` was needed once); pkg.go.dev 404s for a fresh tag even after the proxy serves it
(minutes-to-longer lag; the proxy is the source of truth). Tags are immutable once the proxy caches them — never
re-tag, always cut a new version.
History: v0.1.0 (first, 2026-09-10), v0.2.0, v0.3.0 (go 1.27 floor), v0.3.1 (determinism fix), v0.4.0 (I/O matrix +
diff engine + ConfigFiles), v0.4.1 (nil-ctx fix), v0.5.0 (bootstrap provider), v0.6.0 (determinism analyzer + vettool
cmd), v0.7.0 (release guards + floor settled to minor-form 1.27; API unchanged), v0.8.0 (nil-ctx normalized in ALL
derived provider capabilities via toolsdk.EnsureContext — Repair/Detect/HealthCheck from both ProviderFromSpec and
BootstrapProviderFromSpec; toolsdk floor v1.14.0; API unchanged).

## Misc

- `reports/` is buildflow-owned: `test-coverage` writes `reports/coverage.out`; the whole tree is gitignored and
  buildflow creates the dir when missing (verified 2026-09-09). No tracked placeholder needed — do not force-add files
  into `reports/`.

## Design decisions

### `SaveJSON` returns `(changed bool, *ConfigError)`

Decided 2026-09-10 (TODO_LIST T16): the `changed` bool from `atomicwrite.WriteIfChanged` is surfaced instead of
discarded, so repair flows can distinguish "config updated" from "config already correct". A `SaveJSONIfChanged`
variant was rejected on naming grounds: `SaveJSON` is ALREADY if-changed (idempotent skip), so the variant name would
lie. Zero consumers existed when the signature changed, so the break was free.

### Flat layout is intentional

No `internal/` (the whole module is public API; there is nothing to hide) and no `examples/` dir (runnable examples
live in `example_test.go` so pkg.go.dev renders them inline). Documented in `.golangci.yml`; the go-structure-linter
warnings about both are accepted, not fixed.

### `ConfigError` error-chain traversal

`ConfigError` wraps an underlying cause (`Err error`) and exposes it through standard `Unwrap() error`, `Is(error) 
bool`, and `As(any) bool` methods (all delegating), so `errors.Is`, `errors.AsType`, and `errors.Unwrap` all traverse
the chain. Tests verify sentinel matching (`fs.ErrNotExist`) and typed-cause extraction (`*jsontext.SyntacticError`,
the jsonv2 equivalent of v1's `json.SyntaxError`).

## Conventions

- Return `*ConfigError` (a specific type), never bare `error`, from config I/O helpers. This is what the
  `hierarchical-errors` check enforces and why the type exists.
- No em dashes in code; prefer `errors.Is`/`errors.AsType` over string matching on error messages.
- `(*ConfigError).As` must keep `errors.As` (it delegates to an arbitrary caller-chosen target type; `AsType[E]`
  cannot express that). erraudit accepts this only via a TRAILING `//nolint:legacyerrors` comment on the same line — a
  standalone directive on the line above does NOT suppress.
- Exported symbols carry godoc comments (matches existing style; also keeps buildflow lint quiet).
- All production `encoding/json/v2` marshals pass `json.Deterministic(true)` (policy mirrors go-finding's `json.go`
  marshalOpts). Non-deterministic map keys silently defeat `WriteIfChanged` idempotency — the 2026-09-22 audit caught
  exactly that in `SaveJSON`. Any new jsonv2 output path gets the option on day one.
