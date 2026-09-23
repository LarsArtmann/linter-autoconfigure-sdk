package determinism_test

import (
	"testing"

	"github.com/larsartmann/linter-autoconfigure-sdk/determinism"
	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer bites: the `want` comments in testdata/src/a encode the
// ORIGINAL SDK SaveJSON bug (a bare json.Marshal over a map) — the analyzer
// must flag exactly that shape. Removing the report (or the
// json.Deterministic reference detection) fails this test.
func TestAnalyzer(t *testing.T) {
	t.Parallel()

	analysistest.Run(t, analysistest.TestData(), determinism.NewAnalyzer(), "a")
}
