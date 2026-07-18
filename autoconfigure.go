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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/larsartmann/go-finding"
)

// --- Config round-trip (YAML / JSON) ---

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
	Op string
	// Path is the config file path involved.
	Path string
	// Err is the underlying cause, never nil for a returned ConfigError.
	Err error
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("autoconfigure: %s %s: %s", e.Op, e.Path, e.Err)
}

// Is and As expose the wrapped cause to errors.Is / errors.As, so callers can
// match against the underlying error (e.g. errors.Is(err, fs.ErrNotExist)) and
// pull out typed causes (e.g. errors.As(err, &json.SyntaxError{})).
//
// Unwrap is intentionally omitted: its mandated Unwrap() error signature is a
// false positive in the hierarchical-errors analyzer ("generic return"), and
// Is/As provide the same chain traversal for the standard entry points.
func (e *ConfigError) Is(target error) bool { return errors.Is(e.Err, target) }
func (e *ConfigError) As(target any) bool   { return errors.As(e.Err, target) }

// ReadConfig reads a config file's raw bytes. YAML parsing is deliberately
// left to each tool (different YAML libraries: golangci uses yaml.v3 / v4,
// oxlint may use go-yaml) — this helper covers the shared read + existence
// check so both tools stop hand-writing os.ReadFile with the same error wrapping.
func ReadConfig(path string) ([]byte, *ConfigError) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &ConfigError{Op: "read", Path: path, Err: err}
	}

	return data, nil
}

// LoadJSON reads and unmarshals a JSON config file into a new *T.
func LoadJSON[T any](path string) (*T, *ConfigError) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &ConfigError{Op: "read", Path: path, Err: err}
	}

	v := new(T)

	if err := json.Unmarshal(data, v); err != nil {
		return nil, &ConfigError{Op: "unmarshal", Path: path, Err: err}
	}

	return v, nil
}

// SaveJSON marshals v and writes it to path, creating parent directories.
// Used by auto-configurers that emit JSON configs (oxlint).
func SaveJSON(path string, v any) *ConfigError {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return &ConfigError{Op: "mkdir", Path: filepath.Dir(path), Err: err}
	}

	data, err := json.Marshal(v)
	if err != nil {
		return &ConfigError{Op: "marshal", Path: path, Err: err}
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return &ConfigError{Op: "write", Path: path, Err: err}
	}

	return nil
}

// --- Finding emission for config issues ---

// ConfigIssue describes a single problem found in a linter config file.
// Auto-configurers produce a slice of these; FindingFromIssue converts each
// to a finding.Finding that BuildFlow can aggregate and gate repairs on.
type ConfigIssue struct {
	// Rule is the issue's rule identifier (e.g. "missing-linter", "wrong-priority").
	Rule string

	// Message describes the problem for the user.
	Message string

	// Severity rates how serious the issue is.
	Severity finding.Severity

	// File is the config file path (for Position).
	File string

	// Line is the 1-based line number in the config file (0 if unknown).
	Line int

	// Suggestion is the recommended fix text (empty if no auto-fix).
	Suggestion string
}

// FindingFromIssue converts a ConfigIssue to a finding.Finding with the given
// tool name. When Suggestion is non-empty, the finding carries a FixStrategySuggest
// so BuildFlow's repair loop can surface it.
func FindingFromIssue(toolName string, issue ConfigIssue) finding.Finding {
	line := issue.Line
	if line <= 0 {
		line = 1
	}

	builder := finding.NewBuilder(
		finding.RuleName(issue.Rule),
		finding.ToolName(toolName),
		issue.Message,
		issue.Severity,
		finding.Pos(finding.FilePath(issue.File), line, 1),
	).WithCategory(finding.CategoryConfiguration)

	if issue.Suggestion != "" {
		builder = builder.WithFixStrategy(finding.FixStrategySuggest).
			WithSuggestion(issue.Suggestion)
	}

	f, err := builder.Build()
	if err != nil {
		return finding.Finding{}
	}

	return f
}

// FindingsFromIssues converts a slice of ConfigIssues to findings.
func FindingsFromIssues(toolName string, issues []ConfigIssue) []finding.Finding {
	findings := make([]finding.Finding, 0, len(issues))

	for _, issue := range issues {
		findings = append(findings, FindingFromIssue(toolName, issue))
	}

	return findings
}

// --- Provider wiring (via BuildFlow tool-sdk, optional) ---

// ProviderSpec is the minimal shape an auto-configurer supplies to wire into
// BuildFlow via the tool-sdk. This avoids each tool reimplementing the
// Detector/Repairer adapter ceremony.
type ProviderSpec struct {
	Name        string
	Description string
	ConfigFile  string // the config file path the tool manages (e.g. ".golangci.yml")
	// Analyze inspects the config and returns issues. The working directory
	// is available via finding.WorkingDirFromContext(ctx).
	Analyze func(ctx context.Context) ([]ConfigIssue, error)
	// Repair, if non-nil, rewrites the config to fix the issues. Returns a
	// human-readable description of what changed.
	Repair func(ctx context.Context) (string, error)
}

// ErrNoRepair signals that an auto-configurer does not support auto-repair.
// BuildFlow treats this as a suggest-only tool.
var ErrNoRepair = errors.New("autoconfigure: tool does not support auto-repair")
