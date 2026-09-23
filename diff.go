package autoconfigure

import (
	"encoding/json/v2"
	"fmt"
	"sort"
	"strings"
)

// --- Config diff engine ---

// Kind classifies a Change: a setting was added, removed, or modified. There
// is deliberately no "unchanged" kind: an unchanged setting is the absence of
// a Change, not a Change with a dead state.
type Kind string

const (
	// KindAdded marks a setting that exists only in the after config.
	KindAdded Kind = "added"
	// KindRemoved marks a setting that exists only in the before config.
	KindRemoved Kind = "removed"
	// KindModified marks a setting that exists in both configs with
	// different values.
	KindModified Kind = "modified"
)

// Change is one difference between two linter config versions. Path is the
// setting identifier (a dotted path like "rules.no-console" or a prefixed
// name like "plugin:import" — the comparator's caller chooses the scheme via
// its prefix argument). Old is empty for KindAdded; New is empty for
// KindRemoved; both are set for KindModified.
type Change struct {
	Kind Kind
	Path string
	Old  string
	New  string
}

// DiffMaps compares two string maps and returns one Change per key that was
// added, removed, or whose value differs. Keys are prefixed with prefix
// (pass "" for bare keys). The result is sorted by Path, so output is
// deterministic across runs regardless of map iteration order.
func DiffMaps(before, after map[string]string, prefix string) []Change {
	keys := unionKeys(before, after)
	sort.Strings(keys)

	var changes []Change

	for _, key := range keys {
		beforeVal, hadBefore := before[key]
		afterVal, hasAfter := after[key]
		path := prefix + key

		switch {
		case !hadBefore && hasAfter:
			changes = append(changes, Change{Kind: KindAdded, Path: path, New: afterVal})
		case hadBefore && !hasAfter:
			changes = append(changes, Change{Kind: KindRemoved, Path: path, Old: beforeVal})
		case hadBefore && hasAfter && beforeVal != afterVal:
			changes = append(changes, Change{Kind: KindModified, Path: path, Old: beforeVal, New: afterVal})
		}
	}

	return changes
}

// DiffSets compares two string slices with set semantics: duplicates
// collapse and order is ignored (a pure reorder is not a change). Each item
// present in only one side becomes a Change with Path = prefix + item and
// the item itself as the value. The result is sorted by Path.
func DiffSets(before, after []string, prefix string) []Change {
	beforeSet := sliceSet(before)
	afterSet := sliceSet(after)

	items := unionKeys(beforeSet, afterSet)
	sort.Strings(items)

	var changes []Change

	for _, item := range items {
		path := prefix + item

		_, inBefore := beforeSet[item]
		_, inAfter := afterSet[item]

		switch {
		case !inBefore:
			changes = append(changes, Change{Kind: KindAdded, Path: path, New: item})
		case !inAfter:
			changes = append(changes, Change{Kind: KindRemoved, Path: path, Old: item})
		}
	}

	return changes
}

// DiffBlobs compares two lists of pre-canonicalized blob strings (for
// example, each config overrides block marshalled to canonical JSON) with
// set semantics: order is ignored and exact duplicates collapse, because
// list order carries no policy meaning for these blocks. The canonical blob
// string itself serves as both the Path suffix and the displayed value. The
// result is sorted by Path.
func DiffBlobs(before, after []string, prefix string) []Change {
	return DiffSets(before, after, prefix)
}

// StringValue renders a config value for diff display: strings display
// bare; anything else displays as compact deterministic JSON (map keys
// sorted, so the same value always renders to the same bytes); values that
// cannot be marshaled fall back to a %v print.
func StringValue(v any) string {
	if s, ok := v.(string); ok {
		return s
	}

	data, err := json.Marshal(v, json.Deterministic(true))
	if err != nil {
		return fmt.Sprintf("%v", v)
	}

	return string(data)
}

// Summary renders a one-line human-readable tally of the changes:
// "Added: 1, Modified: 2, Removed: 3".
func Summary(changes []Change) string {
	var added, modified, removed int

	for _, c := range changes {
		switch c.Kind {
		case KindAdded:
			added++
		case KindModified:
			modified++
		case KindRemoved:
			removed++
		}
	}

	return fmt.Sprintf("Added: %d, Modified: %d, Removed: %d", added, modified, removed)
}

// FormatDiff renders changes as a unified-diff-flavored listing, one change
// per line, sorted by Path: "+ path: new", "- path: old", and
// "~ path: old → new". Returns "No changes." for an empty slice. Every line
// ends with a newline.
func FormatDiff(changes []Change) string {
	if len(changes) == 0 {
		return "No changes."
	}

	sorted := make([]Change, len(changes))
	copy(sorted, changes)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Path != sorted[j].Path {
			return sorted[i].Path < sorted[j].Path
		}

		return sorted[i].Kind < sorted[j].Kind
	})

	var b strings.Builder

	for _, c := range sorted {
		switch c.Kind {
		case KindAdded:
			fmt.Fprintf(&b, "+ %s: %s\n", c.Path, c.New)
		case KindRemoved:
			fmt.Fprintf(&b, "- %s: %s\n", c.Path, c.Old)
		case KindModified:
			fmt.Fprintf(&b, "~ %s: %s → %s\n", c.Path, c.Old, c.New)
		}
	}

	return b.String()
}

// sliceSet collapses a slice into a set (map) view.
func sliceSet(items []string) map[string]string {
	set := make(map[string]string, len(items))
	for _, item := range items {
		set[item] = item
	}

	return set
}

// unionKeys returns the union of the keys of two maps (each key once, order
// unspecified — callers sort).
func unionKeys[K comparable, V any](before, after map[K]V) []K {
	keys := make([]K, 0, len(before)+len(after))

	for key := range before {
		keys = append(keys, key)
	}

	for key := range after {
		if _, exists := before[key]; !exists {
			keys = append(keys, key)
		}
	}

	return keys
}
