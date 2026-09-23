package autoconfigure

import (
	"os"
	"os/exec"
	"strconv"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var goFence = regexp.MustCompile("(?s)```go\n(.*?)```")

// snippetErrName matches the conventional `err` identifier in a := line or
// as a bare continuation argument, scoped to whole words.
var snippetErrName = regexp.MustCompile(`\berr\b`)

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
	inTypeBlock := false
	typeDepth := 0
	for blockIndex, block := range blocks {
		// Per-block error-variable isolation: independent README snippets
		// each declare `err`, but concatenated their inferred types can
		// collide (e.g. *ConfigError vs error). Rewrite the conventional
		// name per block so every snippet compiles standalone-true.
		errName := "err" + strconv.Itoa(blockIndex)
		for _, line := range strings.Split(block[1], "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "//") {
				continue
			}

			// A hoisted type declaration consumes its own body lines up to
			// the balanced closing brace.
			if inTypeBlock {
				typeDecls = append(typeDecls, line)
				typeDepth += strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
				if typeDepth == 0 {
					inTypeBlock = false
				}

				continue
			}

			if strings.HasPrefix(trimmed, "type ") {
				typeDecls = append(typeDecls, line)
				typeDepth = strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
				if typeDepth > 0 {
					inTypeBlock = true
				}

				continue
			}

			statements = append(statements, snippetErrName.ReplaceAllString(line, errName))

			if names := snippetDecl.FindStringSubmatch(trimmed); names != nil {
				for _, name := range names[1:] {
					if name == "" {
						continue
					}

					if name == "err" {
						name = errName
					}
					declared = append(declared, name)
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
