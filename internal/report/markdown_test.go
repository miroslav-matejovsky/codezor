package report_test

import (
	"testing"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
	"github.com/miroslav-matejovsky/codezor/internal/report"
	"github.com/stretchr/testify/require"
)

func fixture() *analyze.Analysis {
	return &analyze.Analysis{
		Path: ".",
		Languages: []analyze.Language{
			{Name: "Go", Counts: analyze.Counts{Files: 3, Lines: 1200, Code: 1000, Comments: 120, Blanks: 80, Complexity: 42}},
			{Name: "Markdown", Counts: analyze.Counts{Files: 2, Lines: 300, Code: 240, Comments: 0, Blanks: 60}},
		},
		Totals: analyze.Counts{Files: 5, Lines: 1500, Code: 1240, Comments: 120, Blanks: 140, Complexity: 42},
		Locomo: analyze.Locomo{
			Preset:            "medium",
			InputTokens:       1234567,
			OutputTokens:      23456,
			Cost:              4.25,
			GenerationSeconds: 7200,
			ReviewHours:       0.5,
			IterationFactor:   1.79,
		},
	}
}

func TestMarkdownReportsLanguagesAndTotals(t *testing.T) {
	md := report.Markdown(fixture())

	require.Contains(t, md, "| Go | 3 | 1,200 | 1,000 | 120 | 80 | 42 |")
	require.Contains(t, md, "| Markdown | 2 | 300 | 240 | 0 | 60 | 0 |")
	require.Contains(t, md, "| **Total** | 5 | 1,500 | 1,240 | 120 | 140 | 42 |")
}

func TestMarkdownReportsLocomoEstimate(t *testing.T) {
	md := report.Markdown(fixture())

	require.Contains(t, md, "| Model tier | medium |")
	require.Contains(t, md, "| Input tokens | 1,234,567 |")
	require.Contains(t, md, "| Output tokens | 23,456 |")
	require.Contains(t, md, "| Cost | $4.25 |")
	require.Contains(t, md, "| Iteration factor | 1.79 |")
	require.Contains(t, md, "| Generation time | 2.0 h |")
	require.Contains(t, md, "| Review time | 30.0 min |")
}

// The report is derived only from the analysis, so the same tree always
// renders the same bytes.
func TestMarkdownIsDeterministic(t *testing.T) {
	require.Equal(t, report.Markdown(fixture()), report.Markdown(fixture()))
}

func TestMarkdownHandlesEmptyTree(t *testing.T) {
	md := report.Markdown(&analyze.Analysis{Path: "empty", Locomo: analyze.Locomo{Preset: "medium"}})

	require.Contains(t, md, "| **Total** | 0 | 0 | 0 | 0 | 0 | 0 |")
	require.Contains(t, md, "| Generation time | 0 s |")
}
