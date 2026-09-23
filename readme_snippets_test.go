package autoconfigure

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var goFence = regexp.MustCompile("(?s)```go\n(.*?)```")

var snippetDecl = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)\s*(?:,\s*([a-zA-Z_][a-zA-Z0-9_]*))?\s*:=`)

// snippetErrName matches the conventional `err` identifier in a := line or
// as a bare continuation argument, scoped to whole words.
var snippetErrName = regexp.MustCompile(`\berr\b`)

// snippetParts is the compilable projection of a README's go fences.
type snippetParts struct {
	typeDecls  []string
	statements []string
	declared   []string
}

// collectSnippetParts folds README go-fence blocks into compilable parts:
// type declarations (with their bodies) at file scope, statements in order,
// and the declared names needing blank uses. Per-block error-variable
// isolation rewrites the conventional `err` (block snippets each declare
// it, and concatenated their inferred types can collide).
func collectSnippetParts(blocks [][]string) snippetParts {
	var parts snippetParts

	inTypeBlock := false
	typeDepth := 0

	for blockIndex, block := range blocks {
		errName := "err" + strconv.Itoa(blockIndex)

		for line := range strings.SplitSeq(block[1], "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "//") {
				continue
			}

			if appendTypeLine(&parts, &inTypeBlock, &typeDepth, line, trimmed) {
				continue
			}

			parts.statements = append(parts.statements, snippetErrName.ReplaceAllString(line, errName))
			collectDeclaredNames(&parts.declared, trimmed, errName)
		}
	}

	return parts
}

// appendTypeLine hoists type declarations (consuming their brace-balanced
// bodies) to file scope; it reports whether the line was consumed.
func appendTypeLine(parts *snippetParts, inTypeBlock *bool, typeDepth *int, line, trimmed string) bool {
	if *inTypeBlock {
		parts.typeDecls = append(parts.typeDecls, line)

		*typeDepth += strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
		if *typeDepth == 0 {
			*inTypeBlock = false
		}

		return true
	}

	if strings.HasPrefix(trimmed, "type ") {
		parts.typeDecls = append(parts.typeDecls, line)

		*typeDepth = strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
		if *typeDepth > 0 {
			*inTypeBlock = true
		}

		return true
	}

	return false
}

// collectDeclaredNames records LHS names of := declarations (mapping the
// conventional err to its per-block isolated name).
func collectDeclaredNames(declared *[]string, trimmed, errName string) {
	names := snippetDecl.FindStringSubmatch(trimmed)
	if names == nil {
		return
	}

	for _, name := range names[1:] {
		if name == "" {
			continue
		}

		if name == "err" {
			name = errName
		}

		*declared = append(*declared, name)
	}
}

// renderSnippetProgram assembles the compilable main package from the
// collected parts: imports, hoisted types, one main carrying all snippet
// statements in order, and blank uses for every declared name.
func renderSnippetProgram(parts snippetParts) string {
	var program strings.Builder
	program.WriteString("package main\n\nimport (\n\t\"context\"\n\n" +
		"\t\"github.com/larsartmann/go-finding\"\n" +
		"\tautoconfigure \"github.com/larsartmann/linter-autoconfigure-sdk\"\n)\n\n")
	program.WriteString(strings.Join(parts.typeDecls, "\n"))
	program.WriteString("\n\nfunc main() {\n\t_ = context.Background()\n\t_ = finding.SeverityWarning\n")
	program.WriteString(strings.Join(parts.statements, "\n"))

	for _, name := range parts.declared {
		program.WriteString("\n\t_ = " + name)
	}

	program.WriteString("\n}\n")

	return program.String()
}

// TestREADMESnippetsCompile extracts every ```go fence from README.md,
// assembles them into one scratch main package, and compiles the result
// against the working tree via a replace directive. The guard trips when a
// README snippet references an API that no longer exists or no longer
// type-checks — documentation drift becomes a test failure instead of a
// lie that ships to pkg.go.dev.
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

	source := renderSnippetProgram(collectSnippetParts(blocks))

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
		cmd := exec.CommandContext(context.Background(), args[0], args[1:]...)
		cmd.Dir = workdir

		out, runErr := cmd.CombinedOutput()
		if runErr != nil {
			t.Fatalf("%v failed (README snippets no longer compile — documentation drift):\n%s",
				args, out)
		}
	}
}
