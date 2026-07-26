package autoconfigure

import (
	"encoding/json"
	"errors"
	"io/fs"
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
		Rule:       finding.RuleName("missing-linter"),
		Message:    "errcheck is not enabled",
		Severity:   finding.SeverityWarning,
		File:       finding.FilePath(".golangci.yml"),
		Line:       5,
		Suggestion: "add errcheck to enabled linters",
	}

	f := FindingFromIssue(finding.ToolName("golangci-autoconfigure"), issue)

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
		Rule:     finding.RuleName("deprecated-linter"),
		Message:  "golint is deprecated",
		Severity: finding.SeverityInfo,
		File:     finding.FilePath(".golangci.yml"),
	}

	f := FindingFromIssue(finding.ToolName("golangci-autoconfigure"), issue)

	if f.FixStrategy == finding.FixStrategySuggest {
		t.Error("did not expect suggest fix strategy without suggestion")
	}
}

func TestFindingsFromIssues(t *testing.T) {
	issues := []ConfigIssue{
		{Rule: finding.RuleName("a"), Message: "a", Severity: finding.SeverityInfo, File: finding.FilePath("f")},
		{Rule: finding.RuleName("b"), Message: "b", Severity: finding.SeverityWarning, File: finding.FilePath("f")},
	}

	findings := FindingsFromIssues(finding.ToolName("tool"), issues)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
}

func TestReadConfig_MissingFile_ReturnsConfigErrorWrappingErrNotExist(t *testing.T) {
	_, err := ReadConfig(filepath.Join(t.TempDir(), "does-not-exist.yml"))

	var ce *ConfigError
	if !errors.As(err, &ce) {
		t.Fatalf("expected *ConfigError, got %T (%v)", err, err)
	}

	if ce.Op != OpRead {
		t.Errorf("expected Op=read, got %q", ce.Op)
	}

	if ce.Path == "" {
		t.Error("expected non-empty Path")
	}

	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("expected wrapped cause to match fs.ErrNotExist, got %v", err)
	}
}

func TestLoadJSON_MalformedJSON_ReturnsConfigErrorWrappingUnmarshalTypeError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")

	if err := os.WriteFile(path, []byte("{not valid json"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadJSON[map[string]any](path)

	var ce *ConfigError
	if !errors.As(err, &ce) {
		t.Fatalf("expected *ConfigError, got %T (%v)", err, err)
	}

	if ce.Op != OpUnmarshal {
		t.Errorf("expected Op=unmarshal, got %q", ce.Op)
	}

	// The underlying cause must still be reachable for callers that want it.
	var syntaxErr *json.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Errorf("expected underlying *json.SyntaxError to be reachable via errors.As, got %v", err)
	}
}

func TestSaveJSON_CreatesParentDirsAndRoundTrips(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "deep", "config.json")

	type cfg struct {
		Name string `json:"name"`
	}

	if err := SaveJSON(path, cfg{Name: "errcheck"}); err != nil {
		t.Fatalf("SaveJSON failed: %v", err)
	}

	got, err := LoadJSON[cfg](path)
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}

	if got.Name != "errcheck" {
		t.Errorf("unexpected name: %q", got.Name)
	}
}

func TestConfigError_SuccessReturnsNilTypedError(t *testing.T) {
	// Guards against the typed-nil interface gotcha: success must yield a true
	// nil error, not a nil *ConfigError boxed in a non-nil error interface.
	path := filepath.Join(t.TempDir(), "ok.yml")
	if err := os.WriteFile(path, []byte("linters: []"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ReadConfig(path)
	if err != nil {
		t.Fatalf("expected nil error on success, got %v (%T)", err, err)
	}
}

func TestConfigError_ErrorFormat(t *testing.T) {
	ce := &ConfigError{Op: OpRead, Path: "x.yml", Err: errors.New("boom")}
	want := "autoconfigure: read x.yml: boom"
	if got := ce.Error(); got != want {
		t.Errorf("unexpected Error(): %q, want %q", got, want)
	}
}
