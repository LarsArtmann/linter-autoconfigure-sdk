// Bootstrap-mode providers: tools whose BuildFlow job is to GENERATE a
// missing config file (never overwrite an existing one) and to report drift
// advisorially. This file owns the lifecycle skeleton oxlint-auto-configure
// hand-rolled before v0.8.0, so future auto-configurers (biome, ...) get the
// safety invariants by construction instead of by discipline:
//
//   - Detect flags only a MISSING config (an existing config — under ANY
//     discovery name, including user-curated formats — is never flagged).
//   - Repair never overwrites: it writes only when no candidate exists, and
//     honors toolsdk.DryRunFromContext.
//   - HealthCheck never writes: config drift is reported as an advisory error
//     wrapping the unexported errConfigDrift sentinel.

package autoconfigure

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/toolsdk"
)

// errConfigDrift is wrapped by the bootstrap health check when an existing
// config no longer matches what the tool would generate. Advisory only:
// consumers must never treat it as a trigger to rewrite the config.
// Unexported by owner decision (2026-09-22): export on concrete demand.
var errConfigDrift = errors.New("config drift detected")

// defaultMissingRule is the Detect rule name when MissingRule is unset.
const defaultMissingRule finding.RuleName = "CONFIG_MISSING"

// BootstrapSpec describes a bootstrap-mode auto-configurer: a tool whose job
// is to generate its config file when it is missing, never to modify an
// existing one, and to report drift advisorially. BootstrapProviderFromSpec
// derives the full toolsdk lifecycle (Detect / Repair / HealthCheck) from it.
//
// It is deliberately a separate type from ProviderSpec rather than a Generate
// field on it: Analyze/Repair specs and bootstrap specs are two different
// lifecycles, and one struct offering both would invite split-brain specs
// where the chosen mode is ambiguous. A bootstrap tool uses this type; a
// config-auditing tool uses ProviderSpec.
//
// The zero value is not usable; Name, Description, ConfigFile, Generate, and
// Compare are required (validation returns exported sentinels matchable with
// errors.Is).
type BootstrapSpec[T any] struct {
	Name        string
	Description string
	// ConfigFile is the canonical config file the tool writes (the drift
	// health check also reads only this name: other discovery names are
	// user-curated formats the tool never writes).
	ConfigFile finding.FilePath
	// ConfigFiles lists every config filename the tool recognizes, in
	// priority order. An existing file under ANY of these names suppresses
	// Detect and Repair (generating next to a curated alternative-format
	// config could shadow it). Defaults to ConfigFile alone.
	ConfigFiles []finding.FilePath

	// MissingRule names the missing-config Detect finding (domain-specific,
	// e.g. "OXLOPT_CONFIG_MISSING"). Defaults to "CONFIG_MISSING".
	MissingRule finding.RuleName
	// FixCommand is the user-runnable command that regenerates the config
	// (e.g. "oxlint-auto-configure configure"); woven into Detect
	// suggestions and health-check messages. Empty keeps the messages
	// repair-generic.
	FixCommand string
	// CountLabel names Generate's count in repair descriptions (e.g.
	// "rules" renders "wrote .oxlintrc.json (870 rules)"). Empty omits the
	// count.
	CountLabel string

	// Recognizable gates Detect: when it reports false the project is not
	// one this tool configures and no finding is emitted. Nil means every
	// project counts.
	Recognizable func(ctx context.Context) (bool, error)
	// Generate produces the canonical config for the current project plus a
	// domain count for repair descriptions (pass 0 to omit).
	Generate func(ctx context.Context) (T, int, error)
	// Marshal renders T to the exact file bytes, including any trailing
	// newline the tool's format carries. Nil defaults to
	// MarshalJSONIndented (no trailing newline).
	Marshal func(T) ([]byte, *ConfigError)
	// Parse reads existing config bytes into T. Nil defaults to ParseJSON[T].
	Parse func(data []byte) (T, *ConfigError)
	// NormalizeExpected adjusts the freshly generated expected config to
	// honor user customizations carried by the existing config (e.g.
	// preserved external-plugin blocks), so deliberate customization does
	// not read as drift. Nil compares as generated.
	NormalizeExpected func(existing, expected T) T
	// Compare projects two parsed configs onto the diff engine and is what
	// the drift health check reports. Required: the projection is domain
	// knowledge (which fields carry policy meaning) the SDK cannot guess.
	Compare func(existing, expected T) []Change
}

// Validation sentinels returned by BootstrapProviderFromSpec when a required
// field is missing. Match with errors.Is.
var (
	ErrConfigFileRequired = errors.New("autoconfigure: BootstrapProviderFromSpec: ConfigFile must not be empty")
	ErrGenerateRequired   = errors.New("autoconfigure: BootstrapProviderFromSpec: Generate must not be nil")
	ErrCompareRequired    = errors.New("autoconfigure: BootstrapProviderFromSpec: Compare must not be nil")
)

// BootstrapProviderFromSpec converts a BootstrapSpec into the canonical
// BuildFlow provider contract, deriving the full bootstrap lifecycle:
//
//   - Detect emits exactly one warning when the config is missing under every
//     discovery name AND Recognizable (when set) accepts the project.
//   - Repair generates and writes the config only when it is missing,
//     honoring the dry-run flag; an existing config yields a keep description.
//   - HealthCheck parses the existing ConfigFile, regenerates the expected
//     config, applies NormalizeExpected, and reports Compare's changes as an
//     advisory error wrapping the drift sentinel with the fix command.
//
// Trigger and DependsOn stay zero; set them on the returned Spec when needed.
// Inputs derive from ConfigFiles (falling back to ConfigFile), matching
// ProviderFromSpec.
func BootstrapProviderFromSpec[T any](spec BootstrapSpec[T]) (toolsdk.Spec, error) {
	switch {
	case spec.Name == "":
		return toolsdk.Spec{}, ErrNameRequired
	case spec.Description == "":
		return toolsdk.Spec{}, ErrDescriptionRequired
	case spec.ConfigFile == "":
		return toolsdk.Spec{}, ErrConfigFileRequired
	case spec.Generate == nil:
		return toolsdk.Spec{}, ErrGenerateRequired
	case spec.Compare == nil:
		return toolsdk.Spec{}, ErrCompareRequired
	}

	converted := toolsdk.Spec{
		Name:        spec.Name,
		Description: spec.Description,
		Detect:      bootstrapDetector[T]{spec: spec},
		Repair: toolsdk.RepairerFunc(func(ctx context.Context) (toolsdk.RepairResult, error) {
			description, err := spec.repair(ctx)
			if err != nil {
				return toolsdk.RepairResult{}, err
			}

			return toolsdk.RepairResult{Description: description}, nil
		}),
		HealthCheck: spec.healthCheck,
	}
	if inputs := discoveryCandidateStrings(spec.ConfigFile, spec.ConfigFiles); len(inputs) > 0 {
		converted.Inputs = inputs
	}

	return converted, nil
}

// bootstrapDetector adapts a BootstrapSpec's missing-config detection to the
// finding.Detector interface.
type bootstrapDetector[T any] struct {
	spec BootstrapSpec[T]
}

func (d bootstrapDetector[T]) Name() string { return d.spec.Name }

func (d bootstrapDetector[T]) Detect(ctx context.Context) ([]finding.Finding, error) {
	issues, err := d.spec.detectIssues(ctx)
	if err != nil {
		return nil, fmt.Errorf("analyze config: %w", err)
	}

	return FindingsFromIssues(finding.ToolName(d.spec.Name), issues)
}

// detectIssues is the missing-only Detect: one warning when no discovery
// candidate exists and the project is recognizable.
func (s BootstrapSpec[T]) detectIssues(ctx context.Context) ([]ConfigIssue, error) {
	root := WorkingDir(ctx)
	if _, found := s.existingConfig(root); found {
		return nil, nil
	}

	if s.Recognizable != nil {
		recognizable, err := s.Recognizable(ctx)
		if err != nil {
			return nil, fmt.Errorf("%s recognize project: %w", s.Name, err)
		}

		if !recognizable {
			return nil, nil
		}
	}

	rule := s.MissingRule
	if rule == "" {
		rule = defaultMissingRule
	}

	suggestion := "apply this repair to write " + string(s.ConfigFile)
	if s.FixCommand != "" {
		suggestion = "run `" + s.FixCommand + "` or " + suggestion
	}

	return []ConfigIssue{
		{
			Rule:       rule,
			Message:    "No " + string(s.ConfigFile) + " found; generate the optimal config",
			Severity:   finding.SeverityWarning,
			File:       s.ConfigFile,
			Suggestion: suggestion,
			Confidence: finding.ConfidenceHigh,
		},
	}, nil
}

// repair is the never-overwrite, dry-run-aware Repair.
func (s BootstrapSpec[T]) repair(ctx context.Context) (string, error) {
	root := WorkingDir(ctx)
	dryRun := toolsdk.DryRunFromContext(ctx)

	if _, found := s.existingConfig(root); found {
		return string(s.ConfigFile) + " already exists; keeping the existing configuration", nil
	}

	generated, count, err := s.Generate(ctx)
	if err != nil {
		return "", fmt.Errorf("%s repair: generate config: %w", s.Name, err)
	}

	if dryRun {
		return "held back by dry-run: would write " + string(s.ConfigFile) + s.countSuffix(count), nil
	}

	data, configErr := s.marshal(generated)
	if configErr != nil {
		return "", fmt.Errorf("%s repair: marshal %s: %w", s.Name, s.ConfigFile, configErr)
	}

	path := filepath.Join(root, string(s.ConfigFile))
	if _, configErr := SaveJSONBytes(path, data); configErr != nil {
		return "", fmt.Errorf("%s repair: write %s: %w", s.Name, s.ConfigFile, configErr)
	}

	return "wrote " + string(s.ConfigFile) + s.countSuffix(count), nil
}

// healthCheck is the advisory drift HealthCheck. A missing config is healthy
// (Detect owns that finding); a config under a non-canonical discovery name
// is out of scope (the tool never writes those files).
func (s BootstrapSpec[T]) healthCheck(ctx context.Context) error {
	root := WorkingDir(ctx)
	path := filepath.Join(root, string(s.ConfigFile))

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("%s health: read %s: %w", s.Name, s.ConfigFile, err)
	}

	existing, configErr := s.parse(data)
	if configErr != nil {
		return fmt.Errorf(
			"%s health: %s is malformed (%w); %s",
			s.Name, s.ConfigFile, configErr, s.fixPhrase(),
		)
	}

	expected, _, err := s.Generate(ctx)
	if err != nil {
		return fmt.Errorf("%s health: generate expected config: %w", s.Name, err)
	}

	if s.NormalizeExpected != nil {
		expected = s.NormalizeExpected(existing, expected)
	}

	changes := s.Compare(existing, expected)
	if len(changes) == 0 {
		return nil
	}

	return fmt.Errorf(
		"%w: %s has drifted from the generated config (%s); %s (advisory only — nothing was modified)",
		errConfigDrift, s.ConfigFile, Summary(changes), s.fixPhrase(),
	)
}

// existingConfig resolves the first existing discovery candidate in root.
func (s BootstrapSpec[T]) existingConfig(root string) (string, bool) {
	return FirstExisting(root, discoveryCandidateStrings(s.ConfigFile, s.ConfigFiles)...)
}

// marshal renders T via the spec's Marshal hook or the SDK default.
func (s BootstrapSpec[T]) marshal(v T) ([]byte, *ConfigError) {
	if s.Marshal != nil {
		return s.Marshal(v)
	}

	return MarshalJSONIndented(v)
}

// parse reads config bytes via the spec's Parse hook or the SDK default.
func (s BootstrapSpec[T]) parse(data []byte) (T, *ConfigError) {
	if s.Parse != nil {
		return s.Parse(data)
	}

	parsed, configErr := ParseJSON[T](data)
	if configErr != nil {
		var zero T

		return zero, configErr
	}

	return *parsed, nil
}

// countSuffix renders Generate's count for repair descriptions.
func (s BootstrapSpec[T]) countSuffix(count int) string {
	if count <= 0 || s.CountLabel == "" {
		return ""
	}

	return fmt.Sprintf(" (%d %s)", count, s.CountLabel)
}

// fixPhrase names the regeneration path for health-check messages.
func (s BootstrapSpec[T]) fixPhrase() string {
	if s.FixCommand == "" {
		return "regenerate it via this tool's repair"
	}

	return "run `" + s.FixCommand + "` to regenerate"
}
