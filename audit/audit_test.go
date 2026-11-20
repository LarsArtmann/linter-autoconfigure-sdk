package audit

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/larsartmann/go-finding"
	autoconfigure "github.com/larsartmann/linter-autoconfigure-sdk"
)

// staticDefaults mirrors the golangci-lint standard default group plus the
// typecheck pseudo linter, so tests never shell out to a real binary.
func staticDefaults(context.Context) ([]string, error) {
	return []string{"errcheck", "govet", "ineffassign", "staticcheck", "unused", "typecheck"}, nil
}

// errLintersUnavailable stands in for a failed golangci-lint enumeration;
// package-level because dynamic errors are banned (err113).
var errLintersUnavailable = errors.New("golangci-lint unavailable")

// writeConfig persists a fixture config body and returns its path.
func writeConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), ".golangci.yml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture config: %v", err)
	}

	return path
}

// TestAuditReportsDeadReferences bites: the testdata config carries one
// dead reference per check (exclusion rule, linters settings block,
// formatters settings block) next to valid references; Audit must flag
// exactly those three, in that order, and none of the valid ones.
func TestAuditReportsDeadReferences(t *testing.T) {
	t.Parallel()

	checker := NewChecker("testdata/.golangci.yml", "testdata/empty", "unused-in-tests").
		WithDefaultLinters(staticDefaults)

	issues, err := checker.Audit(context.Background())
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}

	want := []struct {
		name        string
		rule        finding.RuleName
		file        string
		line        int
		messageHas  string
		suggestsFix bool
	}{
		{
			name:        "dead exclusion reference",
			rule:        RuleDeadExclusion,
			file:        "testdata/.golangci.yml",
			messageHas:  `exclusion rule for path "pkg/" references unknown linter "exhaustruct"`,
			suggestsFix: true,
		},
		{
			name:        "dead linters settings block",
			rule:        RuleDeadSettings,
			file:        "testdata/.golangci.yml",
			messageHas:  `linters settings block for unknown linter "exhaustruct"`,
			suggestsFix: true,
		},
		{
			name:        "dead formatters settings block",
			rule:        RuleDeadSettings,
			file:        "testdata/.golangci.yml",
			messageHas:  `formatters settings block for unknown linter "golines"`,
			suggestsFix: true,
		},
	}

	if len(issues) != len(want) {
		t.Fatalf("Audit reported %d issues, want %d: %+v", len(issues), len(want), issues)
	}

	for i, expected := range want {
		issue := issues[i]

		if issue.Rule != expected.rule {
			t.Errorf("issue %d (%s): rule = %q, want %q", i, expected.name, issue.Rule, expected.rule)
		}

		if string(issue.File) != expected.file {
			t.Errorf("issue %d (%s): file = %q, want %q", i, expected.name, issue.File, expected.file)
		}

		if issue.Line != expected.line {
			t.Errorf("issue %d (%s): line = %d, want %d", i, expected.name, issue.Line, expected.line)
		}

		if !strings.Contains(issue.Message, expected.messageHas) {
			t.Errorf(
				"issue %d (%s): message %q does not contain %q",
				i,
				expected.name,
				issue.Message,
				expected.messageHas,
			)
		}

		if expected.suggestsFix && issue.Suggestion == "" {
			t.Errorf("issue %d (%s): empty suggestion, want a fix hint", i, expected.name)
		}
	}
}

// TestAuditValidReferencesPass bites: references to enabled, explicitly
// disabled, default-enabled, and formatter linters must never be flagged,
// in exclusion rules or in settings blocks.
func TestAuditValidReferencesPass(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		config string
	}{
		{
			name: "enabled linter in exclusion rule",
			config: `
linters:
  enable:
    - gosec
  exclusions:
    rules:
      - path: pkg/
        linters:
          - gosec
`,
		},
		{
			name: "default linter in exclusion rule",
			config: `
linters:
  exclusions:
    rules:
      - path: pkg/
        linters:
          - staticcheck
`,
		},
		{
			name: "disabled linter in exclusion rule",
			config: `
linters:
  disable:
    - depguard
  exclusions:
    rules:
      - path: pkg/
        linters:
          - depguard
`,
		},
		{
			name: "enabled linter settings block",
			config: `
linters:
  enable:
    - gosec
  settings:
    gosec:
      excludes:
        - G304
`,
		},
		{
			name: "default linter settings block",
			config: `
linters:
  settings:
    staticcheck:
      checks:
        - all
`,
		},
		{
			name: "enabled formatter settings block",
			config: `
formatters:
  enable:
    - gci
  settings:
    gci:
      sections:
        - standard
`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			checker := NewChecker(writeConfig(t, testCase.config), "testdata/empty", "unused-in-tests").
				WithDefaultLinters(staticDefaults)

			issues, err := checker.Audit(context.Background())
			if err != nil {
				t.Fatalf("Audit: %v", err)
			}

			if len(issues) != 0 {
				t.Errorf("Audit reported %v, want none", issues)
			}
		})
	}
}

// nolintWant is one expected dead-directive issue position.
type nolintWant struct {
	file string
	line int
}

// assertNolintIssues fails the test unless issues is exactly the wanted
// dead-directive issues, each naming the linter and its file:line.
func assertNolintIssues(t *testing.T, issues []autoconfigure.ConfigIssue, want []nolintWant) {
	t.Helper()

	if len(issues) != len(want) {
		t.Fatalf("Audit reported %d issues, want %d: %+v", len(issues), len(want), issues)
	}

	for i, expected := range want {
		issue := issues[i]

		if issue.Rule != RuleDeadNolint {
			t.Errorf("issue %d: rule = %q, want %q", i, issue.Rule, RuleDeadNolint)
		}

		if string(issue.File) != expected.file || issue.Line != expected.line {
			t.Errorf("issue %d: position = %s:%d, want %s:%d",
				i, issue.File, issue.Line, expected.file, expected.line)
		}

		wantInMessage := filepath.ToSlash(expected.file + ":" + strconv.Itoa(expected.line))
		if !strings.Contains(issue.Message, wantInMessage) {
			t.Errorf("issue %d: message %q does not contain %q", i, issue.Message, wantInMessage)
		}

		if !strings.Contains(issue.Message, `"exhaustruct"`) {
			t.Errorf("issue %d: message %q does not name the dead linter", i, issue.Message)
		}
	}
}

// TestAuditNolintDirectives bites: dead directives are flagged with their
// file and line, while all, enabled, and default references pass; the
// vendor subtree is pruned and non-Go files are ignored unless the
// extension filter says otherwise.
func TestAuditNolintDirectives(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		extensions []string
		want       []nolintWant
	}{
		{
			name:       "default extensions",
			extensions: nil,
			want: []nolintWant{
				{file: "testdata/repo/dead.go", line: 6},
				{file: "testdata/repo/dead.go", line: 7},
			},
		},
		{
			name:       "custom extensions scan text files",
			extensions: []string{".txt"},
			want: []nolintWant{
				{file: "testdata/repo/notes.txt", line: 1},
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			checker := NewChecker("testdata/valid.golangci.yml", "testdata/repo", "unused-in-tests").
				WithDefaultLinters(staticDefaults)
			if testCase.extensions != nil {
				checker = checker.WithSourceExtensions(testCase.extensions...)
			}

			issues, err := checker.Audit(context.Background())
			if err != nil {
				t.Fatalf("Audit: %v", err)
			}

			assertNolintIssues(t, issues, testCase.want)
		})
	}
}

// TestAuditCleanRepoPasses bites: a fully valid config paired with a
// source tree whose directives only reference enabled, default, or all
// linters yields zero issues.
func TestAuditCleanRepoPasses(t *testing.T) {
	t.Parallel()

	checker := NewChecker("testdata/valid.golangci.yml", "testdata/cleanrepo", "unused-in-tests").
		WithDefaultLinters(staticDefaults)

	issues, err := checker.Audit(context.Background())
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("Audit reported %v, want none", issues)
	}
}

// TestAuditEnumerationFailurePropagates bites: when the default-linter
// source fails, Audit must fail with the cause wrapped, not report
// misleading issues against an empty defaults set.
func TestAuditEnumerationFailurePropagates(t *testing.T) {
	t.Parallel()

	boom := errLintersUnavailable
	failing := func(context.Context) ([]string, error) { return nil, boom }

	checker := NewChecker("testdata/valid.golangci.yml", "testdata/empty", "unused-in-tests").
		WithDefaultLinters(failing)

	if _, err := checker.Audit(context.Background()); !errors.Is(err, boom) {
		t.Errorf("Audit error = %v, want wrapped %v", err, boom)
	}
}

// TestAuditMissingConfigFails bites: an unreadable config surfaces as an
// error wrapping fs.ErrNotExist through the SDK's ConfigError chain, not
// as an empty issue list.
func TestAuditMissingConfigFails(t *testing.T) {
	t.Parallel()

	missing := filepath.Join(t.TempDir(), "missing.yml")
	checker := NewChecker(missing, "testdata/empty", "unused-in-tests").
		WithDefaultLinters(staticDefaults)

	_, err := checker.Audit(context.Background())
	if err == nil {
		t.Fatal("Audit error = nil, want fs.ErrNotExist")
	}

	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Audit error = %v, want it to wrap fs.ErrNotExist", err)
	}
}

// TestParseDefaultLinters bites: only the standard group counts as
// default, the typecheck pseudo linter is always included, and output
// without any standard-group linter is an error, not a silent empty set.
func TestParseDefaultLinters(t *testing.T) {
	t.Parallel()

	// fixtureJSON is a fake `golangci-lint linters --json` output; reading
	// it here keeps the table below pure data.
	fixtureJSON, err := os.ReadFile("testdata/linters.json")
	if err != nil {
		t.Fatalf("read linters.json: %v", err)
	}

	cases := []struct {
		name    string
		data    []byte
		want    []string
		wantErr bool
	}{
		{
			name: "fixture output",
			data: fixtureJSON,
			want: []string{"errcheck", "govet", "ineffassign", "staticcheck", "typecheck", "unused"},
		},
		{
			name:    "malformed JSON",
			data:    []byte(`{"Enabled":[}`),
			wantErr: true,
		},
		{
			name:    "no standard group linter",
			data:    []byte(`{"Enabled":[{"name":"gosec","groups":["all"]}]}`),
			wantErr: true,
		},
		{
			name:    "empty listing",
			data:    []byte(`{}`),
			wantErr: true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseDefaultLinters(testCase.data)
			if testCase.wantErr {
				if err == nil {
					t.Fatalf("parseDefaultLinters error = nil, want one (got %v)", got)
				}

				return
			}

			if err != nil {
				t.Fatalf("parseDefaultLinters: %v", err)
			}

			if !slices.Equal(got, testCase.want) {
				t.Errorf("parseDefaultLinters = %v, want %v", got, testCase.want)
			}
		})
	}
}

// TestNewCheckerAppliesDefaults bites: empty constructor arguments fall
// back to .golangci.yml, the working directory, and the golangci-lint
// binary on PATH, with Go sources as the nolint scan default.
func TestNewCheckerAppliesDefaults(t *testing.T) {
	t.Parallel()

	defaults := NewChecker("", "", "")
	if defaults.configPath != defaultConfigPath ||
		defaults.repoRoot != defaultRepoRoot ||
		defaults.golangciBin != defaultGolangciBinary ||
		!slices.Equal(defaults.sourceExtensions, []string{".go"}) {
		t.Errorf("NewChecker(\"\", \"\", \"\") = %+v, want all defaults", defaults)
	}

	explicit := NewChecker("ci.yml", "src", "/usr/local/bin/golangci-lint")
	if explicit.configPath != "ci.yml" || explicit.repoRoot != "src" ||
		explicit.golangciBin != "/usr/local/bin/golangci-lint" {
		t.Errorf("NewChecker explicit args = %+v, want them honored", explicit)
	}

	overridden := NewChecker("", "", "").WithSourceExtensions(".go", ".templ")
	if !slices.Equal(overridden.sourceExtensions, []string{".go", ".templ"}) {
		t.Errorf("WithSourceExtensions = %v, want [.go .templ]", overridden.sourceExtensions)
	}

	reset := overridden.WithSourceExtensions()
	if !slices.Equal(reset.sourceExtensions, []string{".go"}) {
		t.Errorf("WithSourceExtensions() = %v, want [.go]", reset.sourceExtensions)
	}
}
