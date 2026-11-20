package audit

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"os/exec"
	"slices"
)

// defaultGroup is the linter group golangci-lint enables by default
// ("standard" since golangci-lint v2).
const defaultGroup = "standard"

// pseudoLinters never appear in linters enumeration output but may be
// legitimately referenced: typecheck is the pseudo linter that carries
// compilation errors in reports and exclusion rules.
const pseudoLinters = "typecheck"

// DefaultLinters returns the linters the given golangci-lint binary
// enables by default: the standard group from
// "golangci-lint linters --no-config --json", plus the typecheck pseudo
// linter.
//
// Deriving the set from the live binary is the point of this package: a
// pinned list would rot on the next golangci-lint upgrade and silently
// un-know the very defaults the audit compares against.
func DefaultLinters(ctx context.Context, binary string) ([]string, error) {
	out, err := exec.CommandContext(ctx, binary, "linters", "--no-config", "--json").Output()
	if err != nil {
		return nil, enumerationError(err)
	}

	return parseDefaultLinters(out)
}

// ErrNoDefaults is returned when a linters listing contains no linter
// from the default group: the output shape is unexpected (a golangci-lint
// without JSON enumeration support, or a schema change), and an empty
// defaults set would wrongly flag every default linter as dead.
var ErrNoDefaults = errors.New("audit: no default-group linter in listing")

// lintersListing models the JSON shape of
// "golangci-lint linters --json": the entries golangci-lint reports as
// enabled and disabled, each carrying the groups it belongs to.
type lintersListing struct {
	Enabled  []linterEntry `json:"Enabled"`  //nolint:tagliatelle // golangci-lint's listing JSON uses PascalCase keys
	Disabled []linterEntry `json:"Disabled"` //nolint:tagliatelle // golangci-lint's listing JSON uses PascalCase keys
}

type linterEntry struct {
	Name   string   `json:"name"`
	Groups []string `json:"groups"`
}

// parseDefaultLinters extracts the sorted default-enabled linter names
// (standard group plus the pseudo linters) from linters --json output.
func parseDefaultLinters(data []byte) ([]string, error) {
	var listing lintersListing
	if err := json.Unmarshal(data, &listing); err != nil {
		return nil, fmt.Errorf("parse linters JSON: %w", err)
	}

	var defaults []string

	for _, entry := range listing.Enabled {
		if slices.Contains(entry.Groups, defaultGroup) {
			defaults = append(defaults, entry.Name)
		}
	}

	if len(defaults) == 0 {
		return nil, fmt.Errorf("%w: no enabled linter belongs to the %q group", ErrNoDefaults, defaultGroup)
	}

	defaults = append(defaults, pseudoLinters)
	slices.Sort(defaults)

	return defaults, nil
}

// enumerationError decorates a failed linters enumeration with the
// binary's stderr when available, so a version that rejects the flags
// explains itself.
func enumerationError(err error) error {
	exitErr, ok := errors.AsType[*exec.ExitError](err)
	if !ok {
		return fmt.Errorf("run linters enumeration: %w", err)
	}

	return fmt.Errorf("run linters enumeration: %w: %s", err, bytes.TrimSpace(exitErr.Stderr))
}
