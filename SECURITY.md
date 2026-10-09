# Security Policy

## Supported versions

This SDK is pre-1.0: only the latest commit on `master` receives security fixes.

## Reporting a vulnerability

Please report suspected vulnerabilities privately via
[GitHub security advisories](https://github.com/LarsArtmann/linter-autoconfigure-sdk/security/advisories/new).

Do not open a public issue for a suspected vulnerability. You should receive a
response within 72 hours. Confirmed issues are fixed in a patch release and
credited in the advisory unless you prefer to remain anonymous.

## Scope

This library writes config files atomically and reports linter findings. It does
not execute linters, parse untrusted YAML, or accept network input. Vulnerabilities
in file-handling logic (e.g. symlink or permission handling in the atomic write
path) are in scope.

## Security Practices

- **Atomic, crash-durable writes.** `SaveJSON` writes through
  `go-atomic-write` (`WriteIfChanged`): temp file, rename, and an idempotent
  skip when content is unchanged, so a crash never leaves a truncated or
  half-written config behind.
- **Typed errors instead of string matching.** All config I/O returns
  `*ConfigError` with a traversable error chain (`Unwrap`/`Is`/`As`); callers
  match typed causes (e.g. `fs.ErrNotExist`, `*jsontext.SyntacticError`) rather
  than parsing messages.
- **Deterministic serialization.** Every production `encoding/json/v2` marshal
  passes `json.Deterministic(true)`; the `cmd/jsondeterminism` analyzer (run in
  CI) rejects bare Marshal calls so map iteration order can never silently
  defeat write-if-changed idempotency.
- **Validated inputs at the boundary.** `ProviderFromSpec` rejects malformed
  provider declarations with exported sentinel errors
  (`ErrNameRequired`, `ErrDescriptionRequired`, `ErrAnalyzeRequired`); JSON
  decoding is schema-driven into strong types.
- **Minimal blast radius.** The SDK performs no network I/O, spawns no
  processes, and executes no embedded code; its dependency set is kept to the
  LarsArtmann family plus `golang.org/x/tools`.
- **Continuous vulnerability scanning.** The BuildFlow pipeline runs
  `govulncheck` (Go symbols) and `vulnix` (closure) on every full run; CI runs
  `go vet` and race tests on every push.
