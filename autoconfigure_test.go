package autoconfigure

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

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

	f, err := FindingFromIssue(finding.ToolName("golangci-autoconfigure"), issue)
	if err != nil {
		t.Fatalf("FindingFromIssue failed: %v", err)
	}

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

	f, err := FindingFromIssue(finding.ToolName("golangci-autoconfigure"), issue)
	if err != nil {
		t.Fatalf("FindingFromIssue failed: %v", err)
	}

	if f.FixStrategy == finding.FixStrategySuggest {
		t.Error("did not expect suggest fix strategy without suggestion")
	}
}

func TestFindingsFromIssues(t *testing.T) {
	issues := []ConfigIssue{
		{Rule: finding.RuleName("a"), Message: "a", Severity: finding.SeverityInfo, File: finding.FilePath("f")},
		{Rule: finding.RuleName("b"), Message: "b", Severity: finding.SeverityWarning, File: finding.FilePath("f")},
	}

	findings, err := FindingsFromIssues(finding.ToolName("tool"), issues)
	if err != nil {
		t.Fatalf("FindingsFromIssues failed: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
}

func TestReadConfig_MissingFile_ReturnsConfigErrorWrappingErrNotExist(t *testing.T) {
	_, err := ReadConfig(filepath.Join(t.TempDir(), "does-not-exist.yml"))

	ce, ok := errors.AsType[*ConfigError](err)
	if !ok {
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

	ce, ok := errors.AsType[*ConfigError](err)
	if !ok {
		t.Fatalf("expected *ConfigError, got %T (%v)", err, err)
	}

	if ce.Op != OpUnmarshal {
		t.Errorf("expected Op=unmarshal, got %q", ce.Op)
	}

	// The underlying cause must still be reachable for callers that want it.
	if _, ok := errors.AsType[*json.SyntaxError](err); !ok {
		t.Errorf("expected underlying *json.SyntaxError to be reachable via errors.AsType, got %v", err)
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

func TestSaveJSON_Idempotent_NoRewriteOnSameContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	type cfg struct {
		Name string `json:"name"`
	}

	if err := SaveJSON(path, cfg{Name: "errcheck"}); err != nil {
		t.Fatalf("first SaveJSON: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat after first write: %v", err)
	}
	firstMtime := info.ModTime()

	// Force the clock forward so an actual rewrite would be detectable.
	time.Sleep(20 * time.Millisecond)

	if err := SaveJSON(path, cfg{Name: "errcheck"}); err != nil {
		t.Fatalf("second SaveJSON: %v", err)
	}

	info, err = os.Stat(path)
	if err != nil {
		t.Fatalf("stat after second write: %v", err)
	}

	if !info.ModTime().Equal(firstMtime) {
		t.Errorf("mtime changed despite identical content (idempotency broken)")
	}
}

func TestSaveJSON_Idempotent_RewritesOnDifferentContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	type cfg struct {
		Name string `json:"name"`
	}

	if err := SaveJSON(path, cfg{Name: "first"}); err != nil {
		t.Fatalf("first SaveJSON: %v", err)
	}

	if err := SaveJSON(path, cfg{Name: "second"}); err != nil {
		t.Fatalf("second SaveJSON: %v", err)
	}

	got, err := LoadJSON[cfg](path)
	if err != nil {
		t.Fatalf("LoadJSON: %v", err)
	}

	if got.Name != "second" {
		t.Errorf("expected rewritten content 'second', got %q", got.Name)
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

func TestConfigError_Unwrap(t *testing.T) {
	cause := errors.New("root cause")
	ce := &ConfigError{Op: OpRead, Path: "x.yml", Err: cause}

	if unwrapped := errors.Unwrap(ce); unwrapped != cause {
		t.Errorf("expected Unwrap to return the cause, got %v", unwrapped)
	}
}

func TestFindingFromIssue_LineZero_ProducesFileLevelPosition(t *testing.T) {
	issue := ConfigIssue{
		Rule:     finding.RuleName("deprecated-linter"),
		Message:  "golint is deprecated",
		Severity: finding.SeverityInfo,
		File:     finding.FilePath(".golangci.yml"),
		Line:     0,
	}

	f, err := FindingFromIssue(finding.ToolName("tool"), issue)
	if err != nil {
		t.Fatalf("FindingFromIssue failed: %v", err)
	}

	if f.Position.Line != 0 {
		t.Errorf("expected file-level finding (Line=0), got Line=%d", f.Position.Line)
	}
}

func TestFindingFromIssue_EmptyRule_ReturnsError(t *testing.T) {
	issue := ConfigIssue{
		Rule:     finding.RuleName(""),
		Message:  "missing rule",
		Severity: finding.SeverityWarning,
		File:     finding.FilePath(".golangci.yml"),
	}

	if _, err := FindingFromIssue(finding.ToolName("tool"), issue); err == nil {
		t.Error("expected error for empty rule, got nil")
	}
}

func TestFindingsFromIssues_InvalidIssue_ReturnsError(t *testing.T) {
	issues := []ConfigIssue{
		{Rule: finding.RuleName(""), Message: "bad", Severity: finding.SeverityInfo, File: finding.FilePath("f")},
	}

	if _, err := FindingsFromIssues(finding.ToolName("tool"), issues); err == nil {
		t.Error("expected error for invalid issue, got nil")
	}
}

func TestSaveJSON_MarshalError_ReturnsConfigError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")

	err := SaveJSON(path, func() {})
	if err == nil {
		t.Fatal("expected marshal error, got nil")
	}

	if err.Op != OpMarshal {
		t.Errorf("expected Op=marshal, got %q", err.Op)
	}
}

func TestSaveJSON_MkdirFails_WhenParentIsAFile(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(blocker, "sub", "config.json")
	err := SaveJSON(path, map[string]string{"k": "v"})
	if err == nil {
		t.Fatal("expected mkdir error, got nil")
	}

	if err.Op != OpMkdir {
		t.Errorf("expected Op=mkdir, got %q", err.Op)
	}
}

func TestProviderSpec_HasRepair(t *testing.T) {
	withRepair := ProviderSpec{Repair: func(ctx context.Context) (string, error) { return "", nil }}
	if !withRepair.HasRepair() {
		t.Error("expected HasRepair=true when Repair is set")
	}

	withoutRepair := ProviderSpec{}
	if withoutRepair.HasRepair() {
		t.Error("expected HasRepair=false when Repair is nil")
	}
}
