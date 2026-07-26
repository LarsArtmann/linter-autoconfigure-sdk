// Package autoconfigure provides the shared foundation for linter
// auto-configuration tools (golangci-lint-auto-configure, oxlint-auto-configure,
// and future additions like biome-auto-configure).
//
// The two existing tools diverge on their domain-specific concepts — Go project
// shape (CLI/Library/Web) vs JS framework (React/Next/Vue) — and on their
// priority models (4-tier enum vs profile presets). Those differences are
// legitimate and stay in each tool.
//
// What they reinvent identically is the surrounding plumbing:
//   - Reading and writing YAML/JSON config files (round-trip)
//   - Emitting findings for config issues (priority → Severity, fix → suggestion)
//   - Wiring into BuildFlow as a Detector + Repairer
//
// This package owns that plumbing once. Adding a third auto-configurer becomes
// a config-schema exercise, not a from-scratch build.
package autoconfigure

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"encoding/json/jsontext"

	atomicwrite "github.com/larsartmann/go-atomic-write"
	"github.com/larsartmann/go-finding"
)

// --- Config round-trip (YAML / JSON) ---

// Op identifies the operation that failed during config I/O. It mirrors
// the convention from os.PathError: a short, lowercase verb.
type Op string

const (
	OpRead      Op = "read"
	OpUnmarshal Op = "unmarshal"
	OpMarshal   Op = "marshal"
	OpMkdir     Op = "mkdir"
	OpWrite     Op = "write"
)

// ConfigError describes a failure while reading, parsing, or writing a linter
// config file. It wraps the underlying cause with the operation attempted and
// the file path, so callers can produce precise diagnostics or branch with
// errors.Is / errors.As without parsing error strings.
//
// Op follows os.PathError's convention: "read", "unmarshal", "marshal",
// "mkdir", or "write". The underlying cause is reachable via the exported Err
// field and through the Is / As methods (see below), so both
// errors.Is(err, fs.ErrNotExist) and errors.As(err, &json.SyntaxError{}) still
// work against a *ConfigError.
type ConfigError struct {
	// Op is the operation that failed.
	Op Op
	// Path is the config file path involved.
	Path string
	// Err is the underlying cause, never nil for a returned ConfigError.
	Err error
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("autoconfigure: %s %s: %s", e.Op, e.Path, e.Err)
}

// Unwrap, Is, and As expose the wrapped cause for standard error-chain
// traversal. Callers can use errors.Is(err, fs.ErrNotExist),
// errors.As(err, &json.SyntaxError{}), or errors.Unwrap(err) interchangeably.
func (e *ConfigError) Unwrap() error        { return e.Err }
func (e *ConfigError) Is(target error) bool { return errors.Is(e.Err, target) }
func (e *ConfigError) As(target any) bool   { return errors.As(e.Err, target) }

// ReadConfig reads a config file's raw bytes. YAML parsing is deliberately
// left to each tool (different YAML libraries: golangci uses yaml.v3 / v4,
// oxlint may use go-yaml) — this helper covers the shared read + existence
// check so both tools stop hand-writing os.ReadFile with the same error wrapping.
func ReadConfig(path string) ([]byte, *ConfigError) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &ConfigError{Op: OpRead, Path: path, Err: err}
	}

	return data, nil
}

// LoadJSON reads and unmarshals a JSON config file into a new *T.
func LoadJSON[T any](path string) (*T, *ConfigError) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &ConfigError{Op: OpRead, Path: path, Err: err}
	}

	v := new(T)

	if err := json.Unmarshal(data, v); err != nil {
		return nil, &ConfigError{Op: OpUnmarshal, Path: path, Err: err}
	}

	return v, nil
}

// SaveJSON marshals v to indented JSON and writes it to path atomically,
// creating parent directories. The write is idempotent: if the marshalled
// content is byte-identical to the existing file, the write is skipped entirely
// (no mtime bump, no spurious diff). Otherwise the file is replaced via an
// fsync'd temp-file + atomic rename, so a crash cannot truncate the config.
// Race-safe: a concurrent modification between the content check and the
// rename surfaces as a non-nil *ConfigError wrapping
// atomicwrite.ErrConcurrentModification.
//
// Indented output is used because linter configs are typically human-edited.
func SaveJSON(path string, v any) *ConfigError {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return &ConfigError{Op: OpMkdir, Path: dir, Err: err}
	}

	data, err := json.Marshal(v, jsontext.WithIndentPrefix(""), jsontext.WithIndent("  "))
	if err != nil {
		return &ConfigError{Op: OpMarshal, Path: path, Err: err}
	}

	if _, err := atomicwrite.WriteIfChanged(path, data); err != nil {
		return &ConfigError{Op: OpWrite, Path: path, Err: err}
	}

	return nil
}

// --- Finding emission for config issues ---

// ConfigIssue describes a single problem found in a linter config file.
// Auto-configurers produce a slice of these; FindingFromIssue converts each
// to a finding.Finding that BuildFlow can aggregate and gate repairs on.
type ConfigIssue struct {
	// Rule is the issue's rule identifier (e.g. "missing-linter", "wrong-priority").
	Rule finding.RuleName

	// Message describes the problem for the user.
	Message string

	// Severity rates how serious the issue is.
	Severity finding.Severity

	// File is the config file path (for Position).
	File finding.FilePath

	// Line is the 1-based line number in the config file (0 if unknown).
	Line int

	// Suggestion is the recommended fix text (empty if no auto-fix).
	Suggestion string
}

// FindingFromIssue converts a ConfigIssue to a finding.Finding with the given
// tool name. When Suggestion is non-empty, the finding carries a FixStrategySuggest
// so BuildFlow's repair loop can surface it; otherwise FixStrategyNone is set
// explicitly to avoid the empty-string zero-value split brain.
//
// When issue.Line is 0 (unknown), the finding receives a file-level Position via
// finding.FilePos rather than a fabricated line number.
func FindingFromIssue(toolName finding.ToolName, issue ConfigIssue) (finding.Finding, error) {
	var pos finding.Position
	if issue.Line > 0 {
		pos = finding.Pos(issue.File, issue.Line, 1)
	} else {
		pos = finding.FilePos(issue.File)
	}

	builder := finding.NewBuilder(
		issue.Rule,
		toolName,
		issue.Message,
		issue.Severity,
		pos,
	).WithCategory(finding.CategoryConfiguration)

	if issue.Suggestion != "" {
		builder = builder.WithFixStrategy(finding.FixStrategySuggest).
			WithSuggestion(issue.Suggestion)
	} else {
		builder = builder.WithFixStrategy(finding.FixStrategyNone)
	}

	return builder.Build()
}

// FindingsFromIssues converts a slice of ConfigIssues to findings. If any
// issue fails to convert, the entire batch fails and the cause is wrapped.
func FindingsFromIssues(toolName finding.ToolName, issues []ConfigIssue) ([]finding.Finding, error) {
	findings := make([]finding.Finding, 0, len(issues))

	for _, issue := range issues {
		f, err := FindingFromIssue(toolName, issue)
		if err != nil {
			return nil, fmt.Errorf("convert issue %q: %w", issue.Rule, err)
		}
		findings = append(findings, f)
	}

	return findings, nil
}

// --- Provider wiring (via BuildFlow tool-sdk, optional) ---

// ProviderSpec is the minimal shape an auto-configurer supplies to wire into
// BuildFlow via the tool-sdk. This avoids each tool reimplementing the
// Detector/Repairer adapter ceremony.
//
// Provisional: no consumer has migrated to this shape yet. The fields may
// evolve when the first auto-configurer adopts the SDK. The Analyze and Repair
// closures are usable standalone today.
type ProviderSpec struct {
	Name        string
	Description string
	ConfigFile  string // the config file path the tool manages (e.g. ".golangci.yml")
	// Analyze inspects the config and returns issues. The working directory
	// is available via finding.WorkingDirFromContext(ctx).
	Analyze func(ctx context.Context) ([]ConfigIssue, error)
	// Repair, if non-nil, rewrites the config to fix the issues. Returns a
	// human-readable description of what changed. When nil, the spec is
	// suggest-only and callers should return ErrNoRepair from repair attempts.
	Repair func(ctx context.Context) (string, error)
}

// HasRepair reports whether this spec supports auto-repair.
func (s ProviderSpec) HasRepair() bool { return s.Repair != nil }

// ErrNoRepair is returned by a ProviderSpec whose Repair function is nil,
// signalling that the auto-configurer does not support auto-repair. BuildFlow
// treats this as a suggest-only tool.
var ErrNoRepair = errors.New("autoconfigure: tool does not support auto-repair")
