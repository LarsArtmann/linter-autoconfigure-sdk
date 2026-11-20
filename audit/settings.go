package audit

import (
	"fmt"
	"maps"
	"slices"

	"github.com/larsartmann/go-finding"
	autoconfigure "github.com/larsartmann/linter-autoconfigure-sdk"
)

// RuleDeadSettings is the rule of the issues flagging a per-linter settings
// block for a linter that is neither enabled, disabled, nor enabled by
// default. Dead settings hide the fact that the linter they configure is
// no longer running.
const RuleDeadSettings finding.RuleName = "dead-settings-reference"

// settingsIssues returns one issue per settings block (linters and
// formatters sections) whose key is not in the valid set, in sorted key
// order so output is deterministic.
func (c golangciConfig) settingsIssues(configPath string, valid map[string]bool) []autoconfigure.ConfigIssue {
	var issues []autoconfigure.ConfigIssue

	appendIssue := func(section, name string) {
		if valid[name] {
			return
		}

		issue := autoconfigure.ConfigIssue{
			Rule: RuleDeadSettings,
			Message: fmt.Sprintf(
				"%s settings block for unknown linter %q: the linter it configures is not running",
				section,
				name,
			),
			Severity: finding.SeverityWarning,
			File:     finding.FilePath(configPath),
			Suggestion: fmt.Sprintf(
				"remove the %q settings block from %s, or enable the linter it belonged to",
				name,
				configPath,
			),
			Confidence: finding.ConfidenceHigh,
		}

		issues = append(issues, issue)
	}

	for _, name := range slices.Sorted(maps.Keys(c.Linters.Settings)) {
		appendIssue("linters", name)
	}

	for _, name := range slices.Sorted(maps.Keys(c.Formatters.Settings)) {
		appendIssue("formatters", name)
	}

	return issues
}
