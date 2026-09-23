package autoconfigure

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var goFence = regexp.MustCompile("(?s)```go\n(.*?)```")

var snippetDecl = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)\s*(?:,\s*([a-zA-Z_][a-zA-Z0-9_]*))?\s*:=`)

// TestREADMESnippetsCompile extracts every ```go fence from README.md,
// assembles them into one scratch main package (type declarations at file
// scope, statements in order inside main), appends blank uses for declared
// names, and compiles the result against the working tree via a replace
// directive. The guard trips when a README snippet references an API that no
// longer exists or no longer type-checks — documentation drift becomes a
// test failure instead of a lie that ships to pkg.go.dev.
func TestREADMESnippetsCompile(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}

	blocks := goFence.FindAllStringSubmatch(string(raw), -1)
	if len(blocks) == 0 {
		t.Fatal("README.md contains no go fences; the guard must have something to check")
	}

	var (
		typeDecls  []string
		statements []string
		declared   []string
	)
	for _, block := range blocks {
		for _, line := range strings.Split(block[1], "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "//") {
				continue
			}

			if strings.HasPrefix(trimmed, "type ") {
				typeDecls = append(typeDecls, line)

				continue
			}

			statements = append(statements, line)

			if names := snippetDecl.FindStringSubmatch(trimmed); names != nil {
				for _, name := range names[1:] {
					if name != "" {
						declared = append(declared, name)
					}
				}
			}
		}
	}

	var blankUses strings.Builder
	for _, name := range declared {
		blankUses.WriteString("\n\t_ = " + name)
	}

	source := "package main\n\nimport (\n\t\"context\"\n\n\t\"github.com/larsartmann/go-finding\"\n\t" +
		"autoconfigure \"github.com/larsartmann/linter-autoconfigure-sdk\"\n)\n\n" +
		strings.Join(typeDecls, "\n") +
		"\n\nfunc main() {\n\t_ = context.Background()\n\t_ = finding.SeverityWarning\n" +
		strings.Join(statements, "\n") + blankUses.String() + "\n}\n"

	moduleRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("working dir: %v", err)
	}

	workdir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workdir, "main.go"), []byte(source), 0o600); err != nil {
		t.Fatalf("write scratch main.go: %v", err)
	}

	goMod := "module snippetcheck\n\ngo 1.27.1\n\n" +
		"replace github.com/larsartmann/linter-autoconfigure-sdk => " + moduleRoot + "\n"
	if err := os.WriteFile(filepath.Join(workdir, "go.mod"), []byte(goMod), 0o600); err != nil {
		t.Fatalf("write scratch go.mod: %v", err)
	}

	for _, args := range [][]string{{"go", "mod", "tidy"}, {"go", "build", "./..."}} {
		cmd := exec.Command(args[0], args[1:]...) //nolint:gosec // fixed test-controlled argv
		cmd.Dir = workdir
		out, runErr := cmd.CombinedOutput()
		if runErr != nil {
			t.Fatalf("%v failed (README snippets no longer compile — documentation drift):\n%s",
				args, out)
		}
	}
}
