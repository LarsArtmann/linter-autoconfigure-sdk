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
