package report_test

import (
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
		Total: analyze.LanguageLOC{Baseline: 40466, Current: 41748},
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

// The report is derived only from the comparison, so the same tree always
// renders the same bytes.
func TestMarkdownIsDeterministic(t *testing.T) {
	require.Equal(t, report.Markdown(fixture()), report.Markdown(fixture()))
}

func TestMarkdownHandlesEmptyRepository(t *testing.T) {
	md := report.Markdown(&analyze.Comparison{BaselineRef: "main", BaselineHash: "0000000"})

	require.Contains(t, md, "| **Total** | **0** | **0** | **0** | **n/a** |")
}
