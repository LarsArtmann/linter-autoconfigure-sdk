package audit

import (
	"bytes"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/larsartmann/go-finding"
	autoconfigure "github.com/larsartmann/linter-autoconfigure-sdk"
)

// RuleDeadNolint is the rule of the issues flagging a //nolint directive
// that names a linter neither enabled, disabled, nor enabled by default
// (the special name all is always legitimate). An orphaned directive
// suppresses nothing: the finding it was meant to hide comes back.
const RuleDeadNolint finding.RuleName = "dead-nolint-reference"

// nolintDirective matches nolint colon directives of the form
// "nolint:name1,name2" (trailing explanation comment optional). The shape
// is ported verbatim from the repository-local test this package
// generalizes.
var nolintDirective = regexp.MustCompile(`//\s*nolint:([a-zA-Z0-9_,-]+)`)

// nolintReference is one linter name referenced by a nolint directive,
// with the 1-based line the directive sits on.
type nolintReference struct {
	name string
	line int
}

// nolintIssues walks root for source files matching one of the extensions
// and returns one issue per directive referencing a name outside the valid
// set, in lexical walk order.
func nolintIssues(
	root string,
	extensions []string,
	valid map[string]bool,
) ([]autoconfigure.ConfigIssue, error) {
	var issues []autoconfigure.ConfigIssue

	walk := func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if skipDir(entry.Name()) {
				return filepath.SkipDir
			}

			return nil
		}

		if !matchesAnySuffix(path, extensions) {
			return nil
		}

		data, err := os.ReadFile(path) //nolint:gosec // G122 TOCTOU accepted; the walk root is trusted repo content
		if err != nil {
			return err
		}

		for _, ref := range nolintReferences(data) {
			if valid[ref.name] {
				continue
			}

			issue := autoconfigure.ConfigIssue{
				Rule: RuleDeadNolint,
				Message: fmt.Sprintf(
					"nolint directive at %s:%d references unknown linter %q: it suppresses nothing",
					path,
					ref.line,
					ref.name,
				),
				Severity: finding.SeverityWarning,
				File:     finding.FilePath(path),
				Line:     ref.line,
				Suggestion: fmt.Sprintf(
					"update the nolint:%s directive at %s:%d to the linter's current name, or remove it",
					ref.name,
					path,
					ref.line,
				),
				Confidence: finding.ConfidenceHigh,
			}

			issues = append(issues, issue)
		}

		return nil
	}

	if err := filepath.WalkDir(root, walk); err != nil {
		return nil, fmt.Errorf("walk %s: %w", root, err)
	}

	return issues, nil
}

// nolintReferences returns every linter name referenced by nolint
// directives in the given file bytes, each with the 1-based line the
// directive starts on.
func nolintReferences(data []byte) []nolintReference {
	var refs []nolintReference

	for _, match := range nolintDirective.FindAllSubmatchIndex(data, -1) {
		line := lineAt(data, match[0])

		for name := range strings.SplitSeq(string(data[match[2]:match[3]]), ",") {
			refs = append(refs, nolintReference{name: strings.TrimSpace(name), line: line})
		}
	}

	return refs
}

// lineAt returns the 1-based line number of the byte offset in data.
func lineAt(data []byte, offset int) int {
	return 1 + bytes.Count(data[:offset], []byte("\n"))
}

// skipDir reports whether a walked directory must be pruned entirely. The
// list is ported from the repository-local test this package generalizes.
func skipDir(base string) bool {
	switch base {
	case ".git", "vendor", "_tmp", "result", "node_modules", "dist":
		return true
	default:
		return false
	}
}

// matchesAnySuffix reports whether path ends with one of the extensions.
func matchesAnySuffix(path string, extensions []string) bool {
	for _, extension := range extensions {
		if strings.HasSuffix(path, extension) {
			return true
		}
	}

	return false
}

// withNolintNames returns valid plus "all", the special name that
// suppresses every linter in a directive and is therefore always
// legitimate there.
func withNolintNames(valid map[string]bool) map[string]bool {
	withAll := maps.Clone(valid)
	if withAll == nil {
		withAll = make(map[string]bool)
	}

	withAll["all"] = true

	return withAll
}
