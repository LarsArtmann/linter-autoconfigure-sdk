package audit

import (
	"fmt"

	"github.com/larsartmann/go-finding"
	autoconfigure "github.com/larsartmann/linter-autoconfigure-sdk"
)

// RuleDeadExclusion is the rule of the issues flagging an exclusion rule
// that names a linter neither enabled, disabled, nor enabled by default.
// The rule is dead config: it matches nothing and lets real findings
// through (the drift class of a linter rename such as exhaustruct to
// exhaustruct_v5).
const RuleDeadExclusion finding.RuleName = "dead-exclusion-reference"

// exclusionIssues returns one issue per exclusion-rule linter reference
// that is not in the valid set.
func (c golangciConfig) exclusionIssues(configPath string, valid map[string]bool) []autoconfigure.ConfigIssue {
	var issues []autoconfigure.ConfigIssue

	for _, rule := range c.Linters.Exclusions.Rules {
		for _, name := range rule.Linters {
			if valid[name] {
				continue
			}

			issue := autoconfigure.ConfigIssue{
				Rule: RuleDeadExclusion,
				Message: fmt.Sprintf(
					"exclusion rule for path %q references unknown linter %q: it is dead config that matches nothing and lets real findings through",
					rule.Path,
					name,
				),
				Severity: finding.SeverityWarning,
				File:     finding.FilePath(configPath),
				Suggestion: fmt.Sprintf(
					"remove %q from the exclusion rule in %s, or update it to the linter's current name (a linter rename likely orphaned it)",
					name,
					configPath,
				),
				Confidence: finding.ConfidenceHigh,
			}

			issues = append(issues, issue)
		}
	}

	return issues
}
