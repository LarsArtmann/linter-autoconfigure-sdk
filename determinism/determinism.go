// Package determinism provides a go/analysis analyzer that flags
// encoding/json/v2 Marshal calls lacking an explicit determinism option.
//
// The bug class it prevents is real and shipped: encoding/json/v2 serializes
// map keys in an order that CHANGES between calls on the same map, so an
// output-facing json.Marshal without json.Deterministic(true) produces
// byte-unstable configs (silent VCS diff churn, flaky drift comparisons,
// defeated write-if-changed idempotency). The original linter-autoconfigure-sdk
// SaveJSON carried exactly this bug until v0.3.1.
//
// Rule: a call to encoding/json/v2's Marshal is flagged when none of its
// option arguments statically references json.Deterministic. Passing
// json.Deterministic(true) satisfies the rule; passing
// json.Deterministic(false) is the explicit, self-documenting opt-out for
// marshals that genuinely tolerate unstable bytes. Calls that spread an
// opaque options slice (opts...) cannot be verified statically and are not
// flagged — the analyzer catches the accidental class, not every hazard.
//
// Run it via the module's cmd/jsondeterminism (go vet -vettool compatible):
//
//	go run github.com/larsartmann/linter-autoconfigure-sdk/cmd/jsondeterminism ./...
package determinism

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// jsonV2Path is the package the rule applies to. The legacy encoding/json
// (v1) is deliberately out of scope: this module's policy targets the v2
// marshal paths its SDK owns.
const jsonV2Path = "encoding/json/v2"

// NewAnalyzer returns the jsondeterminism analyzer. A fresh value per call
// so callers can attach their own flags without cross-talk.
func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "jsondeterminism",
		Doc:  "flag encoding/json/v2 Marshal calls without an explicit json.Deterministic option",
		Run:  run,
	}
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			if !isJSONV2Marshal(call, pass) {
				return true
			}

			// An opaque options spread (opts...) cannot be verified
			// statically: never flag it (conservative — the analyzer
			// catches the accidental class, not every hazard).
			if call.Ellipsis.IsValid() {
				return true
			}

			if hasDeterministicOption(optionArgs(call)) {
				return true
			}

			pass.Reportf(
				call.Pos(),
				"json.Marshal without an explicit json.Deterministic option: map-bearing values "+
					"marshal in an order that changes between calls, so output is byte-unstable "+
					"(pass json.Deterministic(true); pass json.Deterministic(false) only as a "+
					"deliberate opt-out)",
			)

			return true
		})
	}

	return nil, nil //nolint:nilnil // the canonical go/analysis Run result: no fact graph, no error
}

// isJSONV2Marshal reports whether call is pkg.Marshal where pkg resolves to
// encoding/json/v2.
func isJSONV2Marshal(call *ast.CallExpr, pass *analysis.Pass) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Marshal" {
		return false
	}

	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	pkgName, ok := pass.TypesInfo.Uses[ident].(*types.PkgName)
	if !ok {
		return false
	}

	return pkgName.Imported().Path() == jsonV2Path
}

// optionArgs returns the arguments after the marshalled value (the
// json.Options parameters), or nil when the call has none.
func optionArgs(call *ast.CallExpr) []ast.Expr {
	if len(call.Args) <= 1 {
		return nil
	}

	return call.Args[1:]
}

// hasDeterministicOption reports whether any option argument statically
// references json.Deterministic (true or false — both express intent).
func hasDeterministicOption(optionArgs []ast.Expr) bool {
	for _, arg := range optionArgs {
		call, ok := arg.(*ast.CallExpr)
		if !ok {
			continue
		}

		if selector, ok := call.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "Deterministic" {
			return true
		}
	}

	return false
}
