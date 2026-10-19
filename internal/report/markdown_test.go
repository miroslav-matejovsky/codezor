package report_test

import (
	"strings"
	"testing"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
	"github.com/miroslav-matejovsky/codezor/internal/report"
	"github.com/stretchr/testify/require"
)

func fixture() *analyze.Comparison {
	return &analyze.Comparison{
		Root:         "/repo",
		BaselineRef:  "main",
		BaselineHash: "a1b2c3d",
		LOC: []analyze.LanguageLOC{
			{Name: "Go", Baseline: 38420, Current: 39601},
			{Name: "PowerShell", Baseline: 1204, Current: 1261},
			{Name: "Markdown", Baseline: 0, Current: 120},
			{Name: "YAML", Baseline: 842, Current: 766},
		},
		TotalLOC:     analyze.LanguageLOC{Baseline: 40466, Current: 41748},
		Current:      analyze.Churn{Added: 2941, Removed: 1659},
		LatestCommit: analyze.Churn{Added: 412, Removed: 173},
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

func TestMarkdownNamesTheBaseline(t *testing.T) {
	md := report.Markdown(fixture())

	require.Contains(t, md, "# Code Statistics")
	require.Contains(t, md, "Baseline: `main` (merge-base a1b2c3d)")
}

func TestMarkdownRendersTheLOCTable(t *testing.T) {
	md := report.Markdown(fixture())

	require.Contains(t, md, "| Go | 38,420 | 39,601 | +1,181 | +3.1% |")
	require.Contains(t, md, "| PowerShell | 1,204 | 1,261 | +57 | +4.7% |")
	require.Contains(t, md, "| YAML | 842 | 766 | -76 | -9.0% |")
	require.Contains(t, md, "| **Total** | **40,466** | **41,748** | **+1,282** | **+3.2%** |")
}

// A language the baseline did not contain has no percentage: reporting zero
// would read as no change.
func TestMarkdownMarksNewLanguages(t *testing.T) {
	md := report.Markdown(fixture())

	require.Contains(t, md, "| Markdown | 0 | 120 | +120 | n/a |")
}

func TestMarkdownRendersChurn(t *testing.T) {
	md := report.Markdown(fixture())

	require.Contains(t, md, "## Current Change")
	require.Contains(t, md, "- Added LOC: 2,941")
	require.Contains(t, md, "- Removed LOC: 1,659")
	require.Contains(t, md, "- Gross churn: 4,600 LOC")
	require.Contains(t, md, "- Net change: +1,282 LOC")
	require.Contains(t, md, "- Change footprint: 11.4% of baseline LOC")

	require.Contains(t, md, "## Latest Commit")
	require.Contains(t, md, "- Added LOC: 412")
	require.Contains(t, md, "- Net change: +239 LOC")
}

// The footprint of a commit against the whole baseline would be a second,
// confusing denominator, so the commit block leaves it out.
func TestMarkdownOmitsFootprintForTheLatestCommit(t *testing.T) {
	md := report.Markdown(fixture())

	_, commit, found := strings.Cut(md, "## Latest Commit")
	require.True(t, found)
	require.NotContains(t, commit, "Change footprint")
}

func TestMarkdownRendersLocomoEstimate(t *testing.T) {
	md := report.Markdown(fixture())

	require.Contains(t, md, "| Model tier | medium |")
	require.Contains(t, md, "| Input tokens | 1,234,567 |")
	require.Contains(t, md, "| Cost | $4.25 |")
	require.Contains(t, md, "| Iteration factor | 1.79 |")
	require.Contains(t, md, "| Generation time | 2.0 h |")
	require.Contains(t, md, "| Review time | 30.0 min |")
}

// The report is derived only from the comparison, so the same tree always
// renders the same bytes.
func TestMarkdownIsDeterministic(t *testing.T) {
	require.Equal(t, report.Markdown(fixture()), report.Markdown(fixture()))
}

func TestMarkdownHandlesEmptyBaseline(t *testing.T) {
	md := report.Markdown(&analyze.Comparison{
		BaselineRef:  "main",
		BaselineHash: "0000000",
		Locomo:       analyze.Locomo{Preset: "medium"},
	})

	require.Contains(t, md, "| **Total** | **0** | **0** | **0** | **n/a** |")
	require.Contains(t, md, "- Gross churn: 0 LOC")
	require.NotContains(t, md, "Change footprint")
	require.Contains(t, md, "| Generation time | 0 s |")
}
