package autoconfigure

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/larsartmann/go-finding"
)

type exampleLintConfig struct {
	Linters []string `json:"linters"`
}

func ExampleLoadJSON() {
	dir, _ := os.MkdirTemp("", "example")
	defer func() { _ = os.RemoveAll(dir) }()

	path := filepath.Join(dir, ".oxlintrc.json")
	_ = os.WriteFile(path, []byte(`{"linters":["errcheck","gofmt"]}`), 0o644)

	cfg, err := LoadJSON[exampleLintConfig](path)
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println(cfg.Linters)
	// Output: [errcheck gofmt]
}

func ExampleSaveJSON() {
	dir, _ := os.MkdirTemp("", "example")
	defer func() { _ = os.RemoveAll(dir) }()

	path := filepath.Join(dir, "nested", ".oxlintrc.json")

	cfg := exampleLintConfig{Linters: []string{"errcheck", "gofmt"}}

	changed, err := SaveJSON(path, cfg)
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println("changed:", changed)

	data, _ := os.ReadFile(path)
	fmt.Println(string(data))
	// Output:
	// changed: true
	// {
	//   "linters": [
	//     "errcheck",
	//     "gofmt"
	//   ]
	// }
}

func ExampleFindingFromIssue() {
	issue := ConfigIssue{
		Rule:       finding.RuleName("missing-linter"),
		Message:    "errcheck is not enabled",
		Severity:   finding.SeverityWarning,
		File:       finding.FilePath(".golangci.yml"),
		Line:       5,
		Suggestion: "add errcheck to enabled linters",
	}

	f, err := FindingFromIssue(finding.ToolName("golangci-autoconfigure"), issue)
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println(f.Rule, f.Severity, f.FixStrategy)
	// Output: missing-linter warning suggest
}

func ExampleProviderFromSpec() {
	spec := ProviderSpec{
		Name:        "golangci-autoconfigure",
		Description: "keeps .golangci.yml aligned with the project shape",
		ConfigFile:  ".golangci.yml",
		Analyze: func(ctx context.Context) ([]ConfigIssue, error) {
			return []ConfigIssue{{Rule: "missing-linter", Message: "errcheck is not enabled"}}, nil
		},
		Repair: func(ctx context.Context) (string, error) {
			return "enabled errcheck", nil
		},
	}

	provider, err := ProviderFromSpec(spec)
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println(provider.Name, provider.Inputs, provider.Repair != nil)
	fmt.Println(provider.Detect.Name())
	// Output:
	// golangci-autoconfigure [.golangci.yml] true
	// golangci-autoconfigure
}

func ExampleMarshalJSONIndented() {
	data, err := MarshalJSONIndented(map[string][]string{"categories": {"correctness"}})
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Printf("%s", data)
	// Output:
	// {
	//   "categories": [
	//     "correctness"
	//   ]
	// }
}

func ExampleParseJSON() {
	cfg, err := ParseJSON[exampleLintConfig]([]byte(`{"linters":["errcheck"]}`))
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println(cfg.Linters)
	// Output: [errcheck]
}

func ExampleSaveJSONBytes() {
	dir, _ := os.MkdirTemp("", "example")
	defer func() { _ = os.RemoveAll(dir) }()

	path := filepath.Join(dir, ".oxlintrc.json")

	data, err := MarshalJSONIndented(exampleLintConfig{Linters: []string{"errcheck"}})
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	// Trailing newline is the caller's contract: append it to match the
	// tool's existing file format.
	changed, err := SaveJSONBytes(path, append(data, '\n'))
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println("changed:", changed)
	// Output: changed: true
}

func ExampleWorkingDir() {
	ctx := finding.WithWorkingDir(context.Background(), "/repo")

	fmt.Println(WorkingDir(ctx), WorkingDir(context.Background()))
	// Output: /repo .
}

func ExampleFirstExisting() {
	dir, _ := os.MkdirTemp("", "example")
	defer func() { _ = os.RemoveAll(dir) }()

	_ = os.WriteFile(filepath.Join(dir, ".oxlintrc.json"), []byte("{}"), 0o644)

	path, found := FirstExisting(dir, ".oxlintrc.json", ".oxlintrc.jsonc")
	fmt.Println(filepath.Base(path), found)

	missing, found := FirstExisting(dir, ".eslintrc.json", ".eslintrc.jsonc")
	fmt.Println(filepath.Base(missing), found)
	// Output:
	// .oxlintrc.json true
	// .eslintrc.json false
}

func ExampleDiffMaps() {
	before := map[string]string{"no-console": "off", "no-debugger": "off"}
	after := map[string]string{"no-console": "warn"}

	changes := DiffMaps(before, after, "rules.")

	fmt.Println(Summary(changes))
	fmt.Print(FormatDiff(changes))
	// Output:
	// Added: 0, Modified: 1, Removed: 1
	// ~ rules.no-console: off → warn
	// - rules.no-debugger: off
}

func ExampleBootstrapProviderFromSpec() {
	spec := BootstrapSpec[map[string]string]{
		Name:        "fake-auto-configure",
		Description: "generates .fakerc.json for recognizable projects",
		ConfigFile:  ".fakerc.json",
		ConfigFiles: []finding.FilePath{".fakerc.json", ".fakerc.jsonc"},
		MissingRule: "FAKE_CONFIG_MISSING",
		FixCommand:  "fake-auto-configure configure",
		CountLabel:  "rules",
		Generate: func(ctx context.Context) (map[string]string, int, error) {
			return map[string]string{"plugins": "react"}, 3, nil
		},
		Compare: func(existing, expected map[string]string) []Change {
			return DiffMaps(existing, expected, "")
		},
	}

	provider, err := BootstrapProviderFromSpec(spec)
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println(provider.Name, provider.Inputs)
	fmt.Println(provider.Detect != nil, provider.Repair != nil, provider.HealthCheck != nil)
	// Output:
	// fake-auto-configure [.fakerc.json .fakerc.jsonc]
	// true true true
}
