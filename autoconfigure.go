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
//
// Build with Go 1.26 requires GOEXPERIMENT=jsonv2 (encoding/json/v2 becomes
// standard in Go 1.27).
package autoconfigure

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	atomicwrite "github.com/larsartmann/go-atomic-write"
	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/toolsdk"
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
// errors.Is(err, fs.ErrNotExist) and errors.AsType[*jsontext.SyntacticError](err)
// still work against a *ConfigError. (The jsonv2 unmarshaler emits
// *jsontext.SyntacticError for malformed input; v1's json.SyntaxError is the
// legacy equivalent.)
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
// errors.AsType[*jsontext.SyntacticError](err), or errors.Unwrap(err)
// interchangeably.
func (e *ConfigError) Unwrap() error        { return e.Err }
func (e *ConfigError) Is(target error) bool { return errors.Is(e.Err, target) }

func (e *ConfigError) As(target any) bool {
	return errors.As(
		e.Err,
		target,
	) //nolint:legacyerrors // As delegates to an arbitrary caller-chosen target type; errors.AsType[E] cannot express that
}

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
// The changed return reports whether the file was actually written: false
// means the on-disk content already matched and the write was skipped. Repair
// flows use it to distinguish "config updated" from "config already correct".
//
// Indented output is used because linter configs are typically human-edited.
func SaveJSON(path string, v any) (bool, *ConfigError) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false, &ConfigError{Op: OpMkdir, Path: dir, Err: err}
	}

	data, err := json.Marshal(v, jsontext.WithIndentPrefix(""), jsontext.WithIndent("  "))
	if err != nil {
		return false, &ConfigError{Op: OpMarshal, Path: path, Err: err}
	}

	changed, err := atomicwrite.WriteIfChanged(path, data)
	if err != nil {
		return false, &ConfigError{Op: OpWrite, Path: path, Err: err}
	}

	return changed, nil
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
			return nil, fmt.Errorf("convert issue %q for tool %q: %w", issue.Rule, toolName, err)
		}

		findings = append(findings, f)
	}

	return findings, nil
}

// --- Provider wiring (via go-finding toolsdk) ---

// ProviderSpec is the shape an auto-configurer supplies to wire into BuildFlow.
// It speaks the auto-configurer's domain language: Analyze reports ConfigIssues
// against the managed config file, Repair describes what was rewritten.
//
// ProviderFromSpec converts a ProviderSpec into the canonical BuildFlow provider
// contract, go-finding's toolsdk.Spec, ready for toolsdk.Register. The Analyze
// and Repair closures are also usable standalone without any BuildFlow wiring.
type ProviderSpec struct {
	Name        string
	Description string
	// ConfigFile is the config file path the tool manages (e.g. ".golangci.yml").
	ConfigFile finding.FilePath
	// Analyze inspects the config and returns issues. The working directory
	// is available via finding.WorkingDirFromContext(ctx).
	Analyze func(ctx context.Context) ([]ConfigIssue, error)
	// Repair, if non-nil, rewrites the config to fix the issues. Returns a
	// human-readable description of what changed. When nil, the spec is
	// suggest-only: ProviderFromSpec leaves toolsdk.Spec.Repair nil, the
	// canonical signal that BuildFlow should not attempt repairs.
	Repair func(ctx context.Context) (string, error)
}

// HasRepair reports whether this spec supports auto-repair.
func (s ProviderSpec) HasRepair() bool { return s.Repair != nil }

// ErrNoRepair is the pre-bridge sentinel for "this tool does not support
// auto-repair", kept so pre-v1 code referencing it keeps compiling.
//
// Deprecated: the canonical suggest-only signal is structural. ProviderFromSpec
// leaves toolsdk.Spec.Repair nil when ProviderSpec.Repair is nil, and BuildFlow
// reads that directly. This SDK never returns ErrNoRepair, and returning it
// from a custom Repairer is a repair failure for BuildFlow, not a suggest-only
// signal. Removed at v1.
var ErrNoRepair = errors.New("autoconfigure: tool does not support auto-repair")

// ProviderFromSpec converts a ProviderSpec into the canonical BuildFlow provider
// contract: go-finding's toolsdk.Spec (module github.com/larsartmann/go-finding/toolsdk).
// The result can be handed to toolsdk.Register for BuildFlow discovery, or its
// Trigger / DependsOn fields can be adjusted first — the returned Spec is a
// plain value.
//
// Field mapping:
//   - Name, Description pass through verbatim; Name also becomes the tool name
//     stamped onto every finding the Detect adapter emits.
//   - ConfigFile becomes Inputs (the config file is what the tool reads).
//   - Analyze is wrapped as a finding.Detector that converts each ConfigIssue
//     via FindingFromIssue.
//   - A non-nil Repair is wrapped as a toolsdk.Repairer; a nil Repair stays nil,
//     the canonical signal for a suggest-only tool (suggestions still flow
//     through findings carrying FixStrategySuggest).
//   - Trigger and DependsOn have no ProviderSpec equivalent and stay zero;
//     set them on the returned Spec when needed.
//
// Validation errors name the offending field: Name, Description, and Analyze
// are all required.
func ProviderFromSpec(spec ProviderSpec) (toolsdk.Spec, error) {
	switch {
	case spec.Name == "":
		return toolsdk.Spec{}, errors.New("autoconfigure: ProviderFromSpec: Name must not be empty")
	case spec.Description == "":
		return toolsdk.Spec{}, errors.New("autoconfigure: ProviderFromSpec: Description must not be empty")
	case spec.Analyze == nil:
		return toolsdk.Spec{}, errors.New("autoconfigure: ProviderFromSpec: Analyze must not be nil")
	}

	converted := toolsdk.Spec{
		Name:        spec.Name,
		Description: spec.Description,
		Detect:      issueDetector{spec: spec},
	}
	if spec.ConfigFile != "" {
		converted.Inputs = []string{string(spec.ConfigFile)}
	}

	if spec.Repair != nil {
		converted.Repair = toolsdk.RepairerFunc(func(ctx context.Context) (toolsdk.RepairResult, error) {
			description, err := spec.Repair(ctx)
			if err != nil {
				return toolsdk.RepairResult{}, err
			}

			return toolsdk.RepairResult{Description: description}, nil
		})
	}

	return converted, nil
}

// issueDetector adapts a ProviderSpec's Analyze closure to the finding.Detector
// interface, emitting the reported config issues as findings attributed to the
// auto-configurer's tool name.
type issueDetector struct {
	spec ProviderSpec
}

func (d issueDetector) Name() string { return d.spec.Name }

func (d issueDetector) Detect(ctx context.Context) ([]finding.Finding, error) {
	issues, err := d.spec.Analyze(ctx)
	if err != nil {
		return nil, fmt.Errorf("analyze config: %w", err)
	}

	return FindingsFromIssues(finding.ToolName(d.spec.Name), issues)
}
