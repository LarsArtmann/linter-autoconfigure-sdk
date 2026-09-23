// Command jsondeterminism runs the SDK's jsondeterminism analyzer as a
// go vet-compatible vettool, so repositories can enforce byte-stable
// json.Marshal output without a golangci-lint plugin build:
//
//	go run github.com/larsartmann/linter-autoconfigure-sdk/cmd/jsondeterminism ./...
//
// Exit code mirrors go vet: non-zero when any finding is reported.
package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/larsartmann/linter-autoconfigure-sdk/determinism"
)

func main() {
	singlechecker.Main(determinism.NewAnalyzer())
}
