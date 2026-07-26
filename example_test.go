package autoconfigure

import (
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
	defer os.RemoveAll(dir)
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
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "nested", ".oxlintrc.json")

	cfg := exampleLintConfig{Linters: []string{"errcheck", "gofmt"}}
	if err := SaveJSON(path, cfg); err != nil {
		fmt.Println("error:", err)
		return
	}

	data, _ := os.ReadFile(path)
	fmt.Println(string(data))
	// Output:
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
