package report_test

import (
	"strings"
	"testing"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
	"github.com/miroslav-matejovsky/codezor/internal/report"
	"github.com/stretchr/testify/require"
)

// The summary is the tables and their labels, with nothing else: the prose
// that described the two sides lives in the documentation.
func TestSummaryStartsWithTheLOCTable(t *testing.T) {
	s := report.Summary(fixture())

	require.True(t, strings.HasPrefix(s, "LOC\n\n|"), s)
	require.NotContains(t, s, "Baseline:")
	require.NotContains(t, s, "working tree")
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

	s := report.Summary(&analyze.Comparison{})

	require.Contains(t, s, want)
}
