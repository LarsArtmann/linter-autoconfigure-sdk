package autoconfigure

import (
	"reflect"
	"testing"
)

func TestDiffMaps_AddedRemovedModified(t *testing.T) {
	t.Parallel()

	before := map[string]string{"kept": "same", "changed": "old", "dropped": "x"}
	after := map[string]string{"kept": "same", "changed": "new", "added": "y"}

	want := []Change{
		{Kind: KindModified, Path: "prefix.changed", Old: "old", New: "new"},
		{Kind: KindRemoved, Path: "prefix.dropped", Old: "x"},
		{Kind: KindAdded, Path: "prefix.added", New: "y"},
	}

	got := DiffMaps(before, after, "prefix.")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("DiffMaps mismatch:\nwant: %+v\ngot:  %+v", want, got)
	}
}

func TestDiffMaps_EqualMaps_NoChanges(t *testing.T) {
	t.Parallel()

	got := DiffMaps(map[string]string{"a": "1"}, map[string]string{"a": "1"}, "")
	if len(got) != 0 {
		t.Errorf("expected no changes, got %+v", got)
	}
}

func TestDiffMaps_DeterministicAcrossRuns(t *testing.T) {
	t.Parallel()

	before := map[string]string{}
	after := map[string]string{}
	for _, key := range []string{"z", "a", "m", "b", "y"} {
		before[key] = "old"
		after[key] = "new"
	}

	first := DiffMaps(before, after, "")
	second := DiffMaps(before, after, "")

	if !reflect.DeepEqual(first, second) {
		t.Errorf("expected identical order across runs:\nfirst:  %+v\nsecond: %+v", first, second)
	}
}

func TestDiffSets_AddedRemovedDuplicateCollapsed(t *testing.T) {
	t.Parallel()

	before := []string{"import", "import", "unicorn", "node"}
	after := []string{"import", "promise", "node"}

	want := []Change{
		{Kind: KindRemoved, Path: "plugin:unicorn", Old: "unicorn"},
		{Kind: KindAdded, Path: "plugin:promise", New: "promise"},
	}

	got := DiffSets(before, after, "plugin:")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("DiffSets mismatch:\nwant: %+v\ngot:  %+v", want, got)
	}
}

func TestDiffSets_ReorderOnly_NoChanges(t *testing.T) {
	t.Parallel()

	before := []string{"a", "b", "c"}
	after := []string{"c", "a", "b"}

	if got := DiffSets(before, after, ""); len(got) != 0 {
		t.Errorf("expected reorder to produce no changes, got %+v", got)
	}
}

func TestDiffBlobs_ReorderOnly_NoChanges(t *testing.T) {
	t.Parallel()

	before := []string{`{"files":["a"],"rules":{"r":"off"}}`, `{"files":["b"]}`}
	after := []string{`{"files":["b"]}`, `{"files":["a"],"rules":{"r":"off"}}`}

	if got := DiffBlobs(before, after, "override:"); len(got) != 0 {
		t.Errorf("expected blob reorder to produce no changes, got %+v", got)
	}
}

func TestDiffBlobs_AddedRemovedBlock(t *testing.T) {
	t.Parallel()

	before := []string{`{"files":["a"]}`}
	after := []string{`{"files":["a"]}`, `{"files":["b"]}`}

	want := []Change{
		{Kind: KindAdded, Path: `override:{"files":["b"]}`, New: `{"files":["b"]}`},
	}

	got := DiffBlobs(before, after, "override:")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("DiffBlobs mismatch:\nwant: %+v\ngot:  %+v", want, got)
	}
}

func TestStringValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   any
		want string
	}{
		{"string bare", "error", "error"},
		{"map sorted keys", map[string]any{"z": 1, "a": 2}, `{"a":2,"z":1}`},
		{"slice", []any{"off", map[string]any{"x": true}}, `["off",{"x":true}]`},
		{"bool", true, "true"},
		{"int", 42, "42"},
		{"nil", nil, "null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := StringValue(tt.in); got != tt.want {
				t.Errorf("StringValue(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestStringValue_UnmarshalableFallsBackToPrint(t *testing.T) {
	t.Parallel()

	got := StringValue(struct{ F func() }{})
	if got == "" {
		t.Error("expected non-empty fallback rendering")
	}
}

func TestSummary(t *testing.T) {
	t.Parallel()

	changes := []Change{
		{Kind: KindAdded},
		{Kind: KindModified},
		{Kind: KindModified},
		{Kind: KindRemoved},
	}

	want := "Added: 1, Modified: 2, Removed: 1"
	if got := Summary(changes); got != want {
		t.Errorf("Summary = %q, want %q", got, want)
	}
}

func TestFormatDiff_Golden(t *testing.T) {
	t.Parallel()

	changes := []Change{
		{Kind: KindModified, Path: "rules.no-console", Old: "off", New: "warn"},
		{Kind: KindAdded, Path: "plugin:import", New: "import"},
		{Kind: KindRemoved, Path: "category:style", Old: "warn"},
	}

	want := "+ plugin:import: import\n" +
		"- category:style: warn\n" +
		"~ rules.no-console: off → warn\n"

	if got := FormatDiff(changes); got != want {
		t.Errorf("FormatDiff mismatch:\nwant: %q\ngot:  %q", want, got)
	}
}

func TestFormatDiff_Empty(t *testing.T) {
	t.Parallel()

	if got := FormatDiff(nil); got != "No changes." {
		t.Errorf("FormatDiff(nil) = %q, want %q", got, "No changes.")
	}
}
