# Summary

What this PR changes and why. Reference the issue it closes, if any (`Closes #N`).

## Checklist

- [ ] `go test -race -count=1 ./...` passes (with `GOEXPERIMENT=jsonv2` set)
- [ ] Public API changes are reflected in README, FEATURES.md, and CHANGELOG.md
- [ ] New exported symbols carry godoc comments
- [ ] No new dependency on a YAML library (out of scope by design)
