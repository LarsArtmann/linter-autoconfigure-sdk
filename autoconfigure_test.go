package autoconfigure

import (
	"context"
	"encoding/json/jsontext"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	atomicwrite "github.com/larsartmann/go-atomic-write"
	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/toolsdk"
)

// Test-scoped error values. Static sentinels keep err113 quiet and give the
// assertions stable identities to match on.
var (
	errBoom        = errors.New("boom")
	errRootCause   = errors.New("root cause")
	errSentinel    = errors.New("sentinel cause")
	errUnrelated   = errors.New("unrelated")
	errAnalyzeFail = errors.New("analyze failed")
	errDiskFull    = errors.New("disk full")
)

func TestSaveAndLoadJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "config.json")

	type config struct {
		Linters []string `json:"linters"`
	}

	want := config{Linters: []string{"errcheck", "gofmt"}}

	if _, err := SaveJSON(path, want); err != nil {
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
	t.Parallel()

	_, err := LoadJSON[struct{}]("nonexistent.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestReadConfig(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	if _, ok := errors.AsType[*jsontext.SyntacticError](err); !ok {
		t.Errorf("expected underlying *jsontext.SyntacticError to be reachable via errors.AsType, got %v", err)
	}
}

func TestSaveJSON_CreatesParentDirsAndRoundTrips(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "deep", "config.json")

	type cfg struct {
		Name string `json:"name"`
	}

	if _, err := SaveJSON(path, cfg{Name: "errcheck"}); err != nil {
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
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	type cfg struct {
		Name string `json:"name"`
	}

	if _, err := SaveJSON(path, cfg{Name: "errcheck"}); err != nil {
		t.Fatalf("first SaveJSON: %v", err)
	}

	before, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat after first write: %v", err)
	}

	beforeContent, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after first write: %v", err)
	}

	if _, err := SaveJSON(path, cfg{Name: "errcheck"}); err != nil {
		t.Fatalf("second SaveJSON: %v", err)
	}

	after, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat after second write: %v", err)
	}

	// File identity (dev+inode), not mtime: an atomic rewrite goes through a
	// temp-file rename and always produces a new inode, while a skipped write
	// leaves the original file untouched. Unlike the previous mtime comparison,
	// this cannot false-pass on coarse-mtime filesystems.
	if !os.SameFile(before, after) {
		t.Error("file identity changed despite identical content (idempotency broken)")
	}

	afterContent, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after second write: %v", err)
	}

	if string(beforeContent) != string(afterContent) {
		t.Error("content changed despite identical input")
	}
}

func TestSaveJSON_ReportsChanged(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	type cfg struct {
		Name string `json:"name"`
	}

	changed, err := SaveJSON(path, cfg{Name: "errcheck"})
	if err != nil {
		t.Fatalf("first SaveJSON: %v", err)
	}

	if !changed {
		t.Error("expected changed=true on first write")
	}

	changed, err = SaveJSON(path, cfg{Name: "errcheck"})
	if err != nil {
		t.Fatalf("second SaveJSON: %v", err)
	}

	if changed {
		t.Error("expected changed=false for identical content")
	}

	changed, err = SaveJSON(path, cfg{Name: "gofmt"})
	if err != nil {
		t.Fatalf("third SaveJSON: %v", err)
	}

	if !changed {
		t.Error("expected changed=true for different content")
	}
}

func TestSaveJSON_ConcurrentWritesToSamePath_SurfaceConcurrentModification(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	type cfg struct {
		Writer string `json:"writer"`
	}

	const (
		rounds  = 200
		writers = 8
	)

	for round := range rounds {
		// Re-seed known content so every writer's payload differs from disk.
		if _, err := SaveJSON(path, cfg{Writer: "seed"}); err != nil {
			t.Fatalf("seed save in round %d: %v", round, err)
		}

		start := make(chan struct{})
		errs := make([]*ConfigError, writers) //nolint:makezero // distinct-index writes are race-free

		var wg sync.WaitGroup
		for w := range writers {
			wg.Add(1)
			go func(w int) {
				defer wg.Done()

				<-start

				_, errs[w] = SaveJSON(path, cfg{Writer: fmt.Sprintf("round-%d-writer-%d", round, w)})
			}(w)
		}

		close(start)
		wg.Wait()

		for _, err := range errs {
			if err == nil {
				continue
			}

			if err.Op != OpWrite {
				t.Errorf("expected Op=write, got %q", err.Op)
			}

			if !errors.Is(err, atomicwrite.ErrConcurrentModification) {
				t.Errorf("expected wrapped atomicwrite.ErrConcurrentModification, got %v", err)
			}

			return
		}
	}

	t.Fatalf("ErrConcurrentModification never surfaced in %d rounds of %d concurrent SaveJSON calls", rounds, writers)
}

func TestSaveJSON_WriteErrorInReadOnlyDir_ReturnsConfigError(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("running as root: read-only directories do not block root writes")
	}

	dir := t.TempDir()

	roDir := filepath.Join(dir, "readonly")
	if err := os.Mkdir(roDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.Chmod(roDir, 0o555); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(roDir, 0o755) })

	path := filepath.Join(roDir, "config.json")

	_, err := SaveJSON(path, map[string]string{"k": "v"})
	if err == nil {
		t.Fatal("expected write error, got nil")
	}

	if err.Op != OpWrite {
		t.Errorf("expected Op=write, got %q", err.Op)
	}

	if !errors.Is(err, fs.ErrPermission) {
		t.Errorf("expected wrapped fs.ErrPermission, got %v", err)
	}
}

func TestSaveJSON_Idempotent_RewritesOnDifferentContent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	type cfg struct {
		Name string `json:"name"`
	}

	if _, err := SaveJSON(path, cfg{Name: "first"}); err != nil {
		t.Fatalf("first SaveJSON: %v", err)
	}

	if _, err := SaveJSON(path, cfg{Name: "second"}); err != nil {
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
	t.Parallel()

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
	t.Parallel()

	ce := &ConfigError{Op: OpRead, Path: "x.yml", Err: errBoom}

	want := "autoconfigure: read x.yml: boom"
	if got := ce.Error(); got != want {
		t.Errorf("unexpected Error(): %q, want %q", got, want)
	}
}

func TestConfigError_Unwrap(t *testing.T) {
	t.Parallel()

	cause := errRootCause
	ce := &ConfigError{Op: OpRead, Path: "x.yml", Err: cause}

	if unwrapped := errors.Unwrap(ce); !errors.Is(unwrapped, cause) {
		t.Errorf("expected Unwrap to return the cause, got %v", unwrapped)
	}
}

type stubAsCauseError struct{ msg string }

func (s *stubAsCauseError) Error() string { return s.msg }

func TestConfigError_IsDelegatesToWrappedCause(t *testing.T) {
	t.Parallel()

	sentinel := errSentinel
	ce := &ConfigError{Op: OpWrite, Path: "config.json", Err: sentinel}

	if !ce.Is(sentinel) {
		t.Error("expected Is to match the wrapped sentinel")
	}

	if !errors.Is(ce, sentinel) {
		t.Error("expected errors.Is to reach the wrapped sentinel through ConfigError")
	}

	if ce.Is(errUnrelated) {
		t.Error("expected Is to be false for an unrelated target")
	}

	if errors.Is(ce, errUnrelated) {
		t.Error("expected errors.Is to be false for an unrelated target")
	}
}

func TestConfigError_AsDelegatesToWrappedCause(t *testing.T) {
	t.Parallel()

	cause := &stubAsCauseError{msg: "typed cause"}
	ce := &ConfigError{Op: OpUnmarshal, Path: "config.json", Err: cause}

	var got *stubAsCauseError
	if !ce.As(&got) {
		t.Fatal("expected As to extract the wrapped cause's type")
	}

	if got != cause {
		t.Errorf("expected As to set the target to the wrapped cause, got %v", got)
	}

	var wrong *os.LinkError
	if ce.As(&wrong) {
		t.Error("expected As to be false for a type the cause does not implement")
	}

	if viaErrorsAs, ok := errors.AsType[*stubAsCauseError](ce); !ok || viaErrorsAs != cause {
		t.Error("expected errors.AsType to reach the wrapped cause through ConfigError")
	}
}

func TestFindingFromIssue_LineZero_ProducesFileLevelPosition(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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

func TestFindingFromIssue_Confidence_PassedThroughWhenSet(t *testing.T) {
	t.Parallel()

	issue := ConfigIssue{
		Rule:       finding.RuleName("config-missing"),
		Message:    "no config",
		Severity:   finding.SeverityWarning,
		File:       finding.FilePath(".oxlintrc.json"),
		Confidence: finding.ConfidenceHigh,
	}

	f, err := FindingFromIssue(finding.ToolName("tool"), issue)
	if err != nil {
		t.Fatalf("FindingFromIssue failed: %v", err)
	}

	if f.Confidence != finding.ConfidenceHigh {
		t.Errorf("expected ConfidenceHigh, got %q", f.Confidence)
	}
}

func TestFindingFromIssue_FixStrategyOverride_KeepsSuggestion(t *testing.T) {
	t.Parallel()

	direct := finding.FixStrategyDirect
	issue := ConfigIssue{
		Rule:        finding.RuleName("config-missing"),
		Message:     "no config",
		Severity:    finding.SeverityWarning,
		File:        finding.FilePath(".oxlintrc.json"),
		Suggestion:  "run the auto-configurer",
		FixStrategy: &direct,
	}
	f, err := FindingFromIssue(finding.ToolName("tool"), issue)
	if err != nil {
		t.Fatalf("FindingFromIssue failed: %v", err)
	}

	if f.FixStrategy != finding.FixStrategyDirect {
		t.Errorf("expected overridden FixStrategyDirect, got %q", f.FixStrategy)
	}

	if f.Suggestion != "run the auto-configurer" {
		t.Errorf("expected suggestion to survive the strategy override, got %q", f.Suggestion)
	}
}

func TestFindingsFromIssues_InvalidIssue_ReturnsError(t *testing.T) {
	t.Parallel()

	issues := []ConfigIssue{
		{Rule: finding.RuleName(""), Message: "bad", Severity: finding.SeverityInfo, File: finding.FilePath("f")},
	}

	if _, err := FindingsFromIssues(finding.ToolName("tool"), issues); err == nil {
		t.Error("expected error for invalid issue, got nil")
	}
}

func TestSaveJSON_MarshalError_ReturnsConfigError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")

	_, err := SaveJSON(path, func() {})
	if err == nil {
		t.Fatal("expected marshal error, got nil")
	}

	if err.Op != OpMarshal {
		t.Errorf("expected Op=marshal, got %q", err.Op)
	}
}

func TestSaveJSON_MkdirFails_WhenParentIsAFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(blocker, "sub", "config.json")

	_, err := SaveJSON(path, map[string]string{"k": "v"})
	if err == nil {
		t.Fatal("expected mkdir error, got nil")
	}

	if err.Op != OpMkdir {
		t.Errorf("expected Op=mkdir, got %q", err.Op)
	}
}

func TestProviderSpec_HasRepair(t *testing.T) {
	t.Parallel()

	withRepair := ProviderSpec{Repair: func(ctx context.Context) (string, error) { return "", nil }}
	if !withRepair.HasRepair() {
		t.Error("expected HasRepair=true when Repair is set")
	}

	withoutRepair := ProviderSpec{}
	if withoutRepair.HasRepair() {
		t.Error("expected HasRepair=false when Repair is nil")
	}
}

func TestErrNoRepair_DeprecatedAliasStable(t *testing.T) {
	t.Parallel()

	if ErrNoRepair == nil {
		t.Fatal("expected ErrNoRepair to be non-nil")
	}

	const want = "autoconfigure: tool does not support auto-repair"
	if ErrNoRepair.Error() != want {
		t.Errorf("expected message %q, got %q", want, ErrNoRepair.Error())
	}

	if !errors.Is(ErrNoRepair, ErrNoRepair) {
		t.Error("expected ErrNoRepair to match itself as a sentinel target")
	}
}

func TestProviderFromSpec_MapsFieldsToToolsDKSpec(t *testing.T) {
	t.Parallel()

	spec := ProviderSpec{
		Name:        "golangci-autoconfigure",
		Description: "keeps .golangci.yml aligned with the project shape",
		ConfigFile:  ".golangci.yml",
		Analyze:     func(ctx context.Context) ([]ConfigIssue, error) { return nil, nil },
		Repair:      func(ctx context.Context) (string, error) { return "", nil },
	}

	converted, err := ProviderFromSpec(spec)
	if err != nil {
		t.Fatalf("ProviderFromSpec failed: %v", err)
	}

	if converted.Name != spec.Name || converted.Description != spec.Description {
		t.Errorf("Name/Description not passed through: %+v", converted)
	}

	if len(converted.Inputs) != 1 || converted.Inputs[0] != ".golangci.yml" {
		t.Errorf("expected Inputs=[.golangci.yml], got %v", converted.Inputs)
	}

	if converted.Detect == nil {
		t.Fatal("expected Detect to be set")
	}

	if converted.Detect.Name() != spec.Name {
		t.Errorf("expected detector name %q, got %q", spec.Name, converted.Detect.Name())
	}

	if converted.Repair == nil {
		t.Error("expected Repair to be set when spec.Repair is non-nil")
	}
}

func TestProviderFromSpec_Detect_ConvertsIssuesToFindings(t *testing.T) {
	t.Parallel()

	issues := []ConfigIssue{
		{
			Rule:     "missing-linter",
			Message:  "errcheck is not enabled",
			Severity: finding.SeverityWarning,
			File:     ".golangci.yml",
			Line:     5,
		},
		{
			Rule:     "wrong-priority",
			Message:  "gofmt has the wrong priority",
			Severity: finding.SeverityError,
			File:     ".golangci.yml",
		},
	}
	spec := ProviderSpec{
		Name:        "golangci-autoconfigure",
		Description: "desc",
		Analyze:     func(ctx context.Context) ([]ConfigIssue, error) { return issues, nil },
	}

	converted, err := ProviderFromSpec(spec)
	if err != nil {
		t.Fatalf("ProviderFromSpec failed: %v", err)
	}

	findings, err := converted.Detect.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(findings) != len(issues) {
		t.Fatalf("expected %d findings, got %d", len(issues), len(findings))
	}

	first := findings[0]
	if first.ToolName != finding.ToolName(spec.Name) {
		t.Errorf("expected ToolName %q, got %q", spec.Name, first.ToolName)
	}

	if first.Rule != "missing-linter" || first.Severity != finding.SeverityWarning {
		t.Errorf("unexpected first finding: %+v", first)
	}

	if findings[1].Severity != finding.SeverityError {
		t.Errorf("expected second finding severity error, got %q", findings[1].Severity)
	}
}

func TestProviderFromSpec_Detect_PropagatesAnalyzeError(t *testing.T) {
	t.Parallel()

	analyzeFailed := errAnalyzeFail
	spec := ProviderSpec{
		Name:        "oxlint-autoconfigure",
		Description: "desc",
		Analyze:     func(ctx context.Context) ([]ConfigIssue, error) { return nil, analyzeFailed },
	}

	converted, err := ProviderFromSpec(spec)
	if err != nil {
		t.Fatalf("ProviderFromSpec failed: %v", err)
	}

	if _, err := converted.Detect.Detect(context.Background()); !errors.Is(err, analyzeFailed) {
		t.Errorf("expected Detect error to wrap the Analyze error, got %v", err)
	}
}

func TestProviderFromSpec_RepairAdapter_WrapsRepairClosure(t *testing.T) {
	t.Parallel()

	repairFailed := errDiskFull
	spec := ProviderSpec{
		Name:        "golangci-autoconfigure",
		Description: "desc",
		Analyze:     func(ctx context.Context) ([]ConfigIssue, error) { return nil, nil },
		Repair: func(ctx context.Context) (string, error) {
			return "enabled errcheck", nil
		},
	}

	converted, err := ProviderFromSpec(spec)
	if err != nil {
		t.Fatalf("ProviderFromSpec failed: %v", err)
	}

	result, err := converted.Repair.Repair(context.Background())
	if err != nil {
		t.Fatalf("Repair failed: %v", err)
	}

	if result.Description != "enabled errcheck" {
		t.Errorf("expected description %q, got %q", "enabled errcheck", result.Description)
	}

	spec.Repair = func(ctx context.Context) (string, error) { return "", repairFailed }

	failing, err := ProviderFromSpec(spec)
	if err != nil {
		t.Fatalf("ProviderFromSpec failed: %v", err)
	}

	if _, err := failing.Repair.Repair(context.Background()); !errors.Is(err, repairFailed) {
		t.Errorf("expected Repair error to wrap the closure error, got %v", err)
	}
}

func TestProviderFromSpec_SuggestOnly_LeavesRepairNil(t *testing.T) {
	t.Parallel()

	spec := ProviderSpec{
		Name:        "biome-autoconfigure",
		Description: "desc",
		Analyze:     func(ctx context.Context) ([]ConfigIssue, error) { return nil, nil },
	}

	converted, err := ProviderFromSpec(spec)
	if err != nil {
		t.Fatalf("ProviderFromSpec failed: %v", err)
	}

	if converted.Repair != nil {
		t.Errorf("expected nil Repair for suggest-only spec, got %T", converted.Repair)
	}

	if converted.Detect == nil {
		t.Error("expected Detect to be set for suggest-only spec")
	}
}

func TestProviderFromSpec_Validation(t *testing.T) {
	t.Parallel()

	analyze := func(ctx context.Context) ([]ConfigIssue, error) { return nil, nil }

	cases := []struct {
		name    string
		spec    ProviderSpec
		wantSub string
	}{
		{"empty name", ProviderSpec{Description: "d", Analyze: analyze}, "Name"},
		{"empty description", ProviderSpec{Name: "n", Analyze: analyze}, "Description"},
		{"nil analyze", ProviderSpec{Name: "n", Description: "d"}, "Analyze"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := ProviderFromSpec(tc.spec)
			if err == nil {
				t.Fatal("expected validation error")
			}

			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Errorf("expected error to mention %q, got %v", tc.wantSub, err)
			}
		})
	}
}

func TestProviderFromSpec_ConvertedSpecPassesToolsDKRegisterValidation(t *testing.T) {
	t.Parallel()

	spec := ProviderSpec{
		Name:        "golangci-autoconfigure",
		Description: "desc",
		Analyze:     func(ctx context.Context) ([]ConfigIssue, error) { return nil, nil },
	}

	converted, err := ProviderFromSpec(spec)
	if err != nil {
		t.Fatalf("ProviderFromSpec failed: %v", err)
	}

	before := len(toolsdk.All())

	registered := toolsdk.Register(converted)
	if registered.Name != converted.Name {
		t.Errorf("expected registered spec %q, got %q", converted.Name, registered.Name)
	}

	if got := len(toolsdk.All()); got != before+1 {
		t.Errorf("expected registry to grow from %d to %d, got %d", before, before+1, got)
	}
}
