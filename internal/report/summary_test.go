package report_test

import (
	"strings"
	"testing"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
	"github.com/miroslav-matejovsky/codezor/internal/report"
	"github.com/stretchr/testify/require"
)

func TestSummaryNamesTheBaseline(t *testing.T) {
	s := report.Summary(fixture())

	require.Contains(t, s, "Baseline: main (merge-base a1b2c3d)")
	require.Contains(t, s, "Current: working tree, including uncommitted and untracked files")
}

// Measured against the commit checked out there is no merge base to name, only
// the baseline commit itself.
func TestSummaryNamesTheHeadBaseline(t *testing.T) {
	c := fixture()
	c.BaselineRef = "HEAD"
	c.BaselineIsMergeBase = false

	require.Contains(t, report.Summary(c), "Baseline: HEAD (a1b2c3d)")
}

// The console table carries no markdown emphasis, so the total row sets no
// column width of its own and the numbers still line up in a terminal.
func TestSummaryAlignsTheLOCTable(t *testing.T) {
	want := strings.Join([]string{
		"LOC",
		"",
		"| Language   | Baseline | Current |  Delta | Change |",
		"| ---------- | -------: | ------: | -----: | -----: |",
		"| Go         |   38,420 |  39,601 | +1,181 |  +3.1% |",
		"| PowerShell |    1,204 |   1,261 |    +57 |  +4.7% |",
		"| Markdown   |        0 |     120 |   +120 |    n/a |",
		"| YAML       |      842 |     766 |    -76 |  -9.0% |",
		"| Total      |   40,466 |  41,748 | +1,282 |  +3.2% |",
	}, "\n")

	require.Contains(t, report.Summary(fixture()), want)
}

// The console shows complexity under its own label, with the columns sized
// independently of the LOC table above it and the branchless languages left
// out.
func TestSummaryAlignsTheComplexityTable(t *testing.T) {
	want := strings.Join([]string{
		"Complexity",
		"",
		"Branching keywords and operators as scc counts them, not a cyclomatic",
		"complexity. Languages with none on either side are left out.",
		"",
		"| Language   | Baseline | Current | Delta | Change |",
		"| ---------- | -------: | ------: | ----: | -----: |",
		"| Go         |    4,201 |   4,388 |  +187 |  +4.5% |",
		"| PowerShell |       96 |      96 |     0 |   0.0% |",
		"| YAML       |       31 |      28 |    -3 |  -9.7% |",
		"| Total      |    4,328 |   4,512 |  +184 |  +4.3% |",
	}, "\n")

	require.Contains(t, report.Summary(fixture()), want)
}

func TestSummaryLinesHaveNoTrailingWhitespace(t *testing.T) {
	s := report.Summary(fixture())

	for i, line := range strings.Split(strings.TrimSuffix(s, "\n"), "\n") {
		require.Equal(t, strings.TrimRight(line, " \t"), line, "line %d", i+1)
	}
}

// The summary is derived only from the comparison, like the markdown report.
func TestSummaryIsDeterministic(t *testing.T) {
	require.Equal(t, report.Summary(fixture()), report.Summary(fixture()))
}

// With no rows at all the headings still set the column widths.
func TestSummaryHandlesEmptyRepository(t *testing.T) {
	want := strings.Join([]string{
		"| Language | Baseline | Current | Delta | Change |",
		"| -------- | -------: | ------: | ----: | -----: |",
		"| Total    |        0 |       0 |     0 |    n/a |",
	}, "\n")

	s := report.Summary(&analyze.Comparison{BaselineRef: "main", BaselineHash: "0000000"})

	require.Contains(t, s, want)
}
