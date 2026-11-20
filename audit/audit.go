// Package audit detects dead linter references in a golangci-lint-consuming
// repository: exclusion rules, per-linter settings blocks, and //nolint
// directives that name a linter which is neither enabled, explicitly
// disabled, nor enabled by default.
//
// When golangci-lint renames a linter (for example exhaustruct to
// exhaustruct_v5), such references stop matching without any error.
// Exclusion rules become dead config that lets real findings through,
// settings blocks hide that their linter is no longer running, and nolint
// directives suppress nothing, so the finding they were meant to hide comes
// back. Checker.Audit reports one autoconfigure.ConfigIssue per dead
// reference so any repository can fail loudly instead.
//
// The set of default-enabled linters is derived from the live golangci-lint
// binary (DefaultLinters), not a hard-coded list: a pinned list is exactly
// the rot this package exists to catch, and it would rot itself on the next
// golangci-lint upgrade.
//
// Run the checks as a command via the module's cmd/audit:
//
//	go run github.com/larsartmann/linter-autoconfigure-sdk/cmd/audit
package audit

import (
	"context"
	"fmt"
	"path/filepath"

	autoconfigure "github.com/larsartmann/linter-autoconfigure-sdk"
)

// defaultConfigPath, defaultRepoRoot, and defaultGolangciBinary are the
// Checker fallbacks when NewChecker receives empty arguments.
const (
	defaultConfigPath     = ".golangci.yml"
	defaultRepoRoot       = "."
	defaultGolangciBinary = "golangci-lint"
)

// Checker audits one repository's golangci-lint configuration for dead
// linter references. Build it with NewChecker; the zero value is not
// usable.
type Checker struct {
	configPath  string
	repoRoot    string
	golangciBin string
	// defaultLinters supplies the linters golangci-lint enables by
	// default. Nil means derive them from the golangci-lint binary.
	defaultLinters func(context.Context) ([]string, error)
	// sourceExtensions lists the file extensions scanned for //nolint
	// directives.
	sourceExtensions []string
}

// NewChecker returns a Checker for the given golangci-lint config file
// path, repository root (scanned for //nolint directives), and
// golangci-lint binary path (used to enumerate the default-enabled
// linters). Empty arguments fall back to ".golangci.yml", ".", and
// "golangci-lint".
func NewChecker(configPath, repoRoot, golangciBinary string) *Checker {
	return &Checker{
		configPath:       orDefault(configPath, defaultConfigPath),
		repoRoot:         orDefault(repoRoot, defaultRepoRoot),
		golangciBin:      orDefault(golangciBinary, defaultGolangciBinary),
		sourceExtensions: []string{".go"},
	}
}

// WithDefaultLinters overrides where the checker learns the linters that
// golangci-lint enables by default. The default source runs the
// golangci-lint binary (DefaultLinters); tests inject a static list to stay
// hermetic. Returns the receiver for chaining.
func (c *Checker) WithDefaultLinters(source func(context.Context) ([]string, error)) *Checker {
	c.defaultLinters = source

	return c
}

// WithSourceExtensions overrides the file extensions scanned for //nolint
// directives (default ".go"). Calling it with no arguments restores the
// default. Returns the receiver for chaining.
func (c *Checker) WithSourceExtensions(extensions ...string) *Checker {
	if len(extensions) == 0 {
		extensions = []string{".go"}
	}

	c.sourceExtensions = extensions

	return c
}

// Audit runs all three drift checks and returns one ConfigIssue per dead
// linter reference: exclusion rules first, then settings blocks, then
// //nolint directives in scan order. Relative paths given to NewChecker
// are resolved against the working directory carried by ctx
// (autoconfigure.WorkingDir), matching the SDK's provider convention.
//
// A non-nil error means the audit could not run at all: the config file is
// unreadable or unparsable, or the default-linter enumeration failed. Dead
// references are reported as issues, never as errors.
func (c *Checker) Audit(ctx context.Context) ([]autoconfigure.ConfigIssue, error) {
	base := autoconfigure.WorkingDir(ctx)
	configPath := resolve(base, c.configPath)
	repoRoot := resolve(base, c.repoRoot)

	defaults, err := c.defaults(ctx)
	if err != nil {
		return nil, fmt.Errorf("audit %s: enumerate default linters: %w", configPath, err)
	}

	cfg, cfgErr := loadConfig(configPath)
	if cfgErr != nil {
		return nil, fmt.Errorf("audit %s: %w", configPath, cfgErr)
	}

	valid := cfg.validLinters(defaults)

	var issues []autoconfigure.ConfigIssue
	issues = append(issues, cfg.exclusionIssues(configPath, valid)...)
	issues = append(issues, cfg.settingsIssues(configPath, valid)...)

	nolintIssues, err := nolintIssues(repoRoot, c.sourceExtensions, withNolintNames(valid))
	if err != nil {
		return nil, fmt.Errorf("audit %s: %w", configPath, err)
	}

	return append(issues, nolintIssues...), nil
}

// defaults returns the default-enabled linter names from the injected
// source, falling back to the golangci-lint binary.
func (c *Checker) defaults(ctx context.Context) ([]string, error) {
	if c.defaultLinters != nil {
		return c.defaultLinters(ctx)
	}

	return DefaultLinters(ctx, c.golangciBin)
}

// orDefault returns value, or fallback when value is empty.
func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}

// resolve joins a relative path onto base; absolute paths pass through.
func resolve(base, path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(base, path)
}
