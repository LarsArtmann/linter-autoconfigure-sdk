#!/usr/bin/env bash
# One-command release verification for a tagged SDK version: proxy check +
# clean-dir `go get` + compile-every-exported-API smoke. Codifies the release
# runbook that used to live only in AGENTS.md prose.
#
# Usage: ./scripts/verify-release.sh vX.Y.Z
set -euo pipefail

version="${1:?usage: verify-release.sh vX.Y.Z (e.g. v0.6.0)}"
module="github.com/larsartmann/linter-autoconfigure-sdk"

step() { printf '==> %s\n' "$1"; }

step "proxy serves ${version}"
served="$(go list -m -versions "${module}" | tr ' ' '\n' | grep -Fx "${version}" || true)"
if [ -z "${served}" ]; then
	echo "ERROR: proxy does not list ${version} yet (minutes-to-longer lag; the proxy is the source of truth, pkg.go.dev lags further)" >&2
	exit 1
fi

workdir="$(mktemp -d)"
trap 'rm -rf "${workdir}"' EXIT

step "clean-dir go get ${module}@${version}"
(
	cd "${workdir}"
	go mod init verify >/dev/null
	go get "${module}@${version}" >/dev/null
)

step "compile + run the full API surface"
cat >"${workdir}/main.go" <<'EOF'
package main

import (
	"context"
	"fmt"

	"github.com/larsartmann/go-finding"
	autoconfigure "github.com/larsartmann/linter-autoconfigure-sdk"
	"github.com/larsartmann/linter-autoconfigure-sdk/determinism"
)

type smokeConfig struct {
	Plugins []string `json:"plugins,omitempty"`
}

func main() {
	// Config I/O matrix.
	data, cerr := autoconfigure.MarshalJSONIndented(map[string]int{"a": 1})
	if cerr != nil {
		panic(cerr)
	}
	parsed, cerr := autoconfigure.ParseJSON[smokeConfig]([]byte(`{"plugins":["react"]}`))
	if cerr != nil {
		panic(cerr)
	}
	changed, cerr := autoconfigure.SaveJSONBytes("smoke.json", append(data, '\n'))
	if cerr != nil {
		panic(cerr)
	}

	// Diff engine.
	before := map[string]string{"no-console": "off", "no-debugger": "off"}
	after := map[string]string{"no-console": "warn"}
	changes := autoconfigure.DiffMaps(before, after, "rules.")
	setChanges := autoconfigure.DiffSets([]string{"a", "b"}, []string{"b", "c"}, "plugins.")
	blobChanges := autoconfigure.DiffBlobs([]string{`{"x":1}`}, []string{`{"x":1}`}, "overrides.")
	summary := autoconfigure.Summary(changes)
	formatted := autoconfigure.FormatDiff(changes)
	stringified := autoconfigure.StringValue(map[string]int{"b": 2, "a": 1})

	// Discovery + working dir.
	path, found := autoconfigure.FirstExisting(".", ".golangci.yml")
	workingDir := autoconfigure.WorkingDir(context.Background())

	// Finding emission.
	finding1, err := autoconfigure.FindingFromIssue("smoke", autoconfigure.ConfigIssue{
		Rule:      "missing-linter",
		Message:   "errcheck is not enabled",
		Severity:  finding.SeverityWarning,
		File:      ".golangci.yml",
		Line:      5,
		Suggestion: "add errcheck to linters.enable",
	})
	batch, err := autoconfigure.FindingsFromIssues("smoke", nil)

	// Provider bridges.
	spec := autoconfigure.ProviderSpec{
		Name:        "smoke",
		Description: "smoke",
		ConfigFile:  ".smoke.yml",
		Analyze:     func(context.Context) ([]autoconfigure.ConfigIssue, error) { return nil, nil },
		Repair:      func(context.Context) (string, error) { return "fixed", nil },
	}
	provider, perr := autoconfigure.ProviderFromSpec(spec)
	hasRepair := spec.HasRepair()

	bootstrap := autoconfigure.BootstrapSpec[map[string]string]{
		Name:        "smoke-bootstrap",
		Description: "smoke",
		ConfigFile:  ".smokeb.json",
		Generate:    func(context.Context) (map[string]string, int, error) { return nil, 0, nil },
		Compare:     func(existing, expected map[string]string) []autoconfigure.Change { return nil },
	}
	bootstrapProvider, berr := autoconfigure.BootstrapProviderFromSpec(bootstrap)

	// Determinism analyzer.
	analyzer := determinism.NewAnalyzer()

	fmt.Println(changed, parsed.Plugins, len(changes), len(setChanges), len(blobChanges))
	fmt.Println(summary, formatted, stringified)
	fmt.Println(path, found, workingDir)
	fmt.Println(finding1.Rule, err, len(batch))
	fmt.Println(provider.Name, hasRepair, perr)
	fmt.Println(bootstrapProvider.Name, berr, analyzer.Name)
	fmt.Println(
		autoconfigure.ErrNameRequired != nil,
		autoconfigure.ErrDescriptionRequired != nil,
		autoconfigure.ErrAnalyzeRequired != nil,
		autoconfigure.ErrConfigFileRequired != nil,
		autoconfigure.ErrGenerateRequired != nil,
		autoconfigure.ErrCompareRequired != nil,
	)
}
EOF
(
	cd "${workdir}"
	go mod tidy >/dev/null
	go run . >/dev/null
)

step "OK: ${version} proxy-verified, resolves clean, full API surface compiles and runs"
