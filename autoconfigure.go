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
	"os"
	"path/filepath"

	"github.com/larsartmann/go-finding"
)

// --- Config round-trip (YAML / JSON) ---

// ReadConfig reads a config file's raw bytes. YAML parsing is deliberately
// left to each tool (different YAML libraries: golangci uses yaml.v3 / v4,
// oxlint may use go-yaml) — this helper covers the shared read + existence
// check so both tools stop hand-writing os.ReadFile with the same error wrapping.
func ReadConfig(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// LoadJSON reads and unmarshals a JSON config file into a new *T.
func LoadJSON[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	v := new(T)

	if err := json.Unmarshal(data, v); err != nil {
		return nil, err
	}

	return v, nil
}

// SaveJSON marshals v and writes it to path, creating parent directories.
// Used by auto-configurers that emit JSON configs (oxlint).
func SaveJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
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
