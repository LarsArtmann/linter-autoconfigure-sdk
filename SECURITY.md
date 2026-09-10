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
