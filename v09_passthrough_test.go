package autoconfigure

import (
	"context"
	"errors"
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/toolsdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The v0.9.0 additive surface: ConfigIssue.Category/Tags/Metadata flow into
// the finding, RepairStructured replaces the prose-only repair contract,
// and Coverage feeds DetectorV2 evidence. Every field is optional and the
// zero value reproduces the pre-v0.9 behavior.

func TestFindingFromIssue_VocabularyPassthrough(t *testing.T) {
	t.Parallel()

	tool := finding.ToolName("test-tool")

	t.Run("category overrides the default when set", func(t *testing.T) {
		t.Parallel()

		f, err := FindingFromIssue(tool, ConfigIssue{
			Rule:     finding.RuleName("r"),
			Message:  "m",
			Severity: finding.SeverityWarning,
			File:     finding.FilePath("go.mod"),
			Line:     3,
			Category: finding.CategoryStyle,
		})
		require.NoError(t, err)
		assert.Equal(t, finding.CategoryStyle, f.Category)
	})

	t.Run("unset category keeps the configuration default", func(t *testing.T) {
		t.Parallel()

		f, err := FindingFromIssue(tool, ConfigIssue{
			Rule:     finding.RuleName("r"),
			Message:  "m",
			Severity: finding.SeverityWarning,
			File:     finding.FilePath("go.mod"),
			Line:     3,
		})
		require.NoError(t, err)
		assert.Equal(t, finding.CategoryConfiguration, f.Category)
	})

	t.Run("tags ride beside the default category", func(t *testing.T) {
		t.Parallel()

		f, err := FindingFromIssue(tool, ConfigIssue{
			Rule:     finding.RuleName("r"),
			Message:  "m",
			Severity: finding.SeverityWarning,
			File:     finding.FilePath("go.mod"),
			Line:     3,
			Tags:     []finding.Tag{finding.TagBuild, "alignment"},
		})
		require.NoError(t, err)
		assert.ElementsMatch(t, []finding.Tag{finding.TagBuild, "alignment"}, f.Tags)
	})

	t.Run("metadata is copied verbatim", func(t *testing.T) {
		t.Parallel()

		f, err := FindingFromIssue(tool, ConfigIssue{
			Rule:     finding.RuleName("r"),
			Message:  "m",
			Severity: finding.SeverityWarning,
			File:     finding.FilePath("go.mod"),
			Line:     3,
			Metadata: map[string]string{"dep-forced.floor": "1.27.1"},
		})
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"dep-forced.floor": "1.27.1"}, f.Metadata)
	})
}

func TestProviderFromSpec_RepairStructuredWins(t *testing.T) {
	t.Parallel()

	spec, err := ProviderFromSpec(ProviderSpec{
		Name:        "t",
		Description: "d",
		Analyze: func(ctx context.Context) ([]ConfigIssue, error) {
			return nil, nil
		},
		RepairStructured: func(ctx context.Context) (toolsdk.RepairResult, error) {
			return toolsdk.RepairResult{
				Description:   "rewrote go.mod",
				ModifiedFiles: []string{"go.mod"},
				Facts:         []toolsdk.Fact{{Key: "dep-forced-count", Value: "2"}},
			}, nil
		},
	})
	require.NoError(t, err)
	require.NotNil(t, spec.Repair)

	res, err := spec.Repair.Repair(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "rewrote go.mod", res.Description)
	assert.Equal(t, []string{"go.mod"}, res.ModifiedFiles)
	require.Len(t, res.Facts, 1)
	assert.Equal(t, "dep-forced-count", res.Facts[0].Key)
}

func TestProviderFromSpec_PlainRepairStillWorks(t *testing.T) {
	t.Parallel()

	spec, err := ProviderFromSpec(ProviderSpec{
		Name:        "t",
		Description: "d",
		Analyze: func(ctx context.Context) ([]ConfigIssue, error) {
			return nil, nil
		},
		Repair: func(ctx context.Context) (string, error) {
			return "prose only", nil
		},
	})
	require.NoError(t, err)
	require.NotNil(t, spec.Repair)

	res, err := spec.Repair.Repair(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "prose only", res.Description)
	assert.Empty(t, res.ModifiedFiles)
}

func TestProviderSpec_HasRepairCoversBothContracts(t *testing.T) {
	t.Parallel()

	assert.True(t, ProviderSpec{Repair: func(ctx context.Context) (string, error) { return "", nil }}.HasRepair())
	assert.True(t, ProviderSpec{RepairStructured: func(ctx context.Context) (toolsdk.RepairResult, error) {
		return toolsdk.RepairResult{}, nil
	}}.HasRepair())
	assert.False(t, ProviderSpec{}.HasRepair())
}

func TestIssueDetector_DetectV2Coverage(t *testing.T) {
	t.Parallel()

	t.Run("coverage feeds the evidence fields", func(t *testing.T) {
		t.Parallel()

		spec := ProviderSpec{
			Name:        "t",
			Description: "d",
			Analyze: func(ctx context.Context) ([]ConfigIssue, error) {
				return []ConfigIssue{{
					Rule:     finding.RuleName("r"),
					Message:  "m",
					Severity: finding.SeverityWarning,
					File:     finding.FilePath("go.mod"),
					Line:     3,
				}}, nil
			},
			Coverage: func(ctx context.Context) (int, []string) {
				return 7, []string{"vendor"}
			},
		}

		detector := issueDetector{spec: spec}
		res, err := detector.DetectV2(context.Background())
		require.NoError(t, err)
		assert.Len(t, res.Findings, 1)
		assert.Equal(t, 7, res.FilesScanned)
		assert.Equal(t, []string{"vendor"}, res.SkippedModules)
	})

	t.Run("no coverage hook keeps v1 fallback semantics", func(t *testing.T) {
		t.Parallel()

		spec := ProviderSpec{
			Name:        "t",
			Description: "d",
			Analyze: func(ctx context.Context) ([]ConfigIssue, error) {
				return nil, nil
			},
		}

		detector := issueDetector{spec: spec}
		res, err := detector.DetectV2(context.Background())
		require.NoError(t, err)
		assert.Empty(t, res.Findings)
		assert.Zero(t, res.FilesScanned)
		assert.Empty(t, res.SkippedModules)
	})

	t.Run("analyze errors pass through", func(t *testing.T) {
		t.Parallel()

		spec := ProviderSpec{
			Name:        "t",
			Description: "d",
			Analyze: func(ctx context.Context) ([]ConfigIssue, error) {
				return nil, errors.New("boom")
			},
		}

		_, err := issueDetector{spec: spec}.DetectV2(context.Background())
		assert.Error(t, err)
	})
}
