# Contributing

Thanks for your interest in contributing!

## How to Contribute

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Development Setup

The project requires `GOEXPERIMENT=jsonv2` (Go 1.26 gates `encoding/json/v2`).
On `cd` into the repo, the `.envrc` sets it automatically via direnv — run
`direnv allow` once after cloning. Without direnv, prefix every command with
`GOEXPERIMENT=jsonv2`.

Quality pipeline (BuildFlow — no Makefile):

    buildflow              # full pipeline, detect mode
    buildflow --fix        # detect + auto-fix
    go test -race -count=1 ./...   # just the Go tests

## Reporting Issues

Please use GitHub Issues to report bugs or request features.
