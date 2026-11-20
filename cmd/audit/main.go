// Command audit runs the SDK's golangci-lint config drift audit against a
// repository and prints one block per dead linter reference:
//
//	go run github.com/larsartmann/linter-autoconfigure-sdk/cmd/audit
//
// The -config, -root, and -golangci flags override the golangci-lint
// config file path (default .golangci.yml), the repository root scanned
// for nolint directives (default .), and the golangci-lint binary used to
// enumerate the default linters (default golangci-lint on PATH).
//
// Exit code mirrors go vet: non-zero when any issue is reported, 2 when
// the audit itself fails.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/larsartmann/linter-autoconfigure-sdk/audit"
)

// exitAuditFailure marks an audit that could not run (config unreadable,
// linters enumeration failed), distinct from exit 1 which means the audit
// ran and found dead references.
const exitAuditFailure = 2

// emitf writes to stdout, ignoring write errors: the exit code already
// carries the outcome, and a broken pipe must not mask the issues.
func emitf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stdout, format, args...)
}

func main() {
	configPath := flag.String("config", ".golangci.yml", "golangci-lint config file path")
	root := flag.String("root", ".", "repository root scanned for nolint directives")
	binary := flag.String("golangci", "golangci-lint", "golangci-lint binary path")

	flag.Parse()

	issues, err := audit.NewChecker(*configPath, *root, *binary).Audit(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "audit:", err)
		os.Exit(exitAuditFailure)
	}

	for _, issue := range issues {
		emitf("%s: %s\n", issue.Rule, issue.Message)

		if issue.Suggestion != "" {
			emitf("  fix: %s\n", issue.Suggestion)
		}
	}

	emitf("%d issue(s)\n", len(issues))

	if len(issues) > 0 {
		os.Exit(1)
	}
}
