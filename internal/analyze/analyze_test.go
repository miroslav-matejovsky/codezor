package analyze_test

import (
	"testing"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
	"github.com/stretchr/testify/require"
)

// The sample tree holds a single Go file with hand verified counts, so the
// numbers below are the contract between scc and this package.
func TestRunCountsSampleTree(t *testing.T) {
	a, err := analyze.Run("testdata/sample", "medium")
	require.NoError(t, err)

	require.Equal(t, "testdata/sample", a.Path)
	require.Len(t, a.Languages, 1)
	require.Equal(t, "Go", a.Languages[0].Name)

	want := analyze.Counts{
		Files:      1,
		Lines:      10,
		Code:       7,
		Comments:   2,
		Blanks:     1,
		Complexity: 1,
	}
	require.Equal(t, want, a.Languages[0].Counts)
	require.Equal(t, want, a.Totals)
}

func TestRunEstimatesLocomo(t *testing.T) {
	a, err := analyze.Run("testdata/sample", "small")
	require.NoError(t, err)

	require.Equal(t, "small", a.Locomo.Preset)
	require.Positive(t, a.Locomo.OutputTokens)
	require.Greater(t, a.Locomo.InputTokens, a.Locomo.OutputTokens)
	require.Positive(t, a.Locomo.Cost)
	require.Positive(t, a.Locomo.GenerationSeconds)
	require.Positive(t, a.Locomo.ReviewHours)
	require.Greater(t, a.Locomo.IterationFactor, 1.0)
}

func TestRunRejectsUnknownPreset(t *testing.T) {
	_, err := analyze.Run("testdata/sample", "enormous")
	require.ErrorContains(t, err, "enormous")
}

func TestRunRejectsMissingPath(t *testing.T) {
	_, err := analyze.Run("testdata/does-not-exist", "medium")
	require.Error(t, err)
}
