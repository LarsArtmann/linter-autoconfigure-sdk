package autoconfigure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-finding"
)

func TestSaveAndLoadJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "config.json")

	type config struct {
		Linters []string `json:"linters"`
	}

	want := config{Linters: []string{"errcheck", "gofmt"}}

	if err := SaveJSON(path, want); err != nil {
		t.Fatalf("SaveJSON failed: %v", err)
	}

	got, err := LoadJSON[config](path)
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}

	if len(got.Linters) != 2 || got.Linters[0] != "errcheck" {
		t.Errorf("unexpected config: %+v", got)
	}
}

func TestLoadJSON_MissingFile(t *testing.T) {
	_, err := LoadJSON[struct{}]("nonexistent.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestReadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "linters.yml")

	if err := os.WriteFile(path, []byte("linters:\n  - errcheck\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	data, err := ReadConfig(path)
	if err != nil {
		t.Fatalf("ReadConfig failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty data")
	}
}

func TestFindingFromIssue_WithSuggestion(t *testing.T) {
	issue := ConfigIssue{
		Rule:       "missing-linter",
		Message:    "errcheck is not enabled",
		Severity:   finding.SeverityWarning,
		File:       ".golangci.yml",
		Line:       5,
		Suggestion: "add errcheck to enabled linters",
	}

	f := FindingFromIssue("golangci-autoconfigure", issue)

	if f.Rule != finding.RuleName("missing-linter") {
		t.Errorf("unexpected rule: %s", f.Rule)
	}

	if f.FixStrategy != finding.FixStrategySuggest {
		t.Error("expected suggest fix strategy")
	}

	if f.Suggestion != "add errcheck to enabled linters" {
		t.Errorf("unexpected suggestion: %s", f.Suggestion)
	}
}

func TestFindingFromIssue_NoSuggestion(t *testing.T) {
	issue := ConfigIssue{
		Rule:     "deprecated-linter",
		Message:  "golint is deprecated",
		Severity: finding.SeverityInfo,
		File:     ".golangci.yml",
	}

	f := FindingFromIssue("golangci-autoconfigure", issue)

	if f.FixStrategy == finding.FixStrategySuggest {
		t.Error("did not expect suggest fix strategy without suggestion")
	}
}

func TestFindingsFromIssues(t *testing.T) {
	issues := []ConfigIssue{
		{Rule: "a", Message: "a", Severity: finding.SeverityInfo, File: "f"},
		{Rule: "b", Message: "b", Severity: finding.SeverityWarning, File: "f"},
	}

	findings := FindingsFromIssues("tool", issues)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
}
