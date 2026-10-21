package report_test

import (
	"strings"
	"testing"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
	"github.com/miroslav-matejovsky/codezor/internal/report"
	"github.com/stretchr/testify/require"
)

// The console shows both metrics in one table, one row per language, each cell
// holding the complexity with the LOC in parentheses. Both numbers are aligned
// down the column, the parenthesised ones among themselves, so a column reads
// as two columns of digits. The table carries no markdown emphasis, so the
// total row sets no column width of its own.
func TestSummaryIsOneTableOfBothMetrics(t *testing.T) {
	want := strings.Join([]string{
		"Complexity (LOC)",
		"",
		"| Language   |       Baseline |        Current |         Delta |        Change |",
		"| ---------- | -------------: | -------------: | ------------: | ------------: |",
		"| Go         | 4,201 (38,420) | 4,388 (39,601) | +187 (+1,181) | +4.5% (+3.1%) |",
		"| PowerShell |    96 ( 1,204) |    96 ( 1,261) |    0 (   +57) |  0.0% (+4.7%) |",
		"| Markdown   |     0 (     0) |     0 (   120) |    0 (  +120) |   n/a (  n/a) |",
		"| YAML       |    31 (   842) |    28 (   766) |   -3 (   -76) | -9.7% (-9.0%) |",
		"| Total      | 4,328 (40,466) | 4,512 (41,748) | +184 (+1,282) | +4.3% (+3.2%) |",
	}, "\n")

	s := report.Summary(fixture())

	require.True(t, strings.HasPrefix(s, want+"\n"), s)
	require.NotContains(t, s, "Baseline:")
	require.NotContains(t, s, "working tree")
}

// A language with nothing on either side is left out. A language with lines but
// no branches, such as Markdown, keeps its row: its LOC is not noise.
func TestSummaryLeavesOutEmptyLanguages(t *testing.T) {
	c := fixture()
	c.Languages = append(c.Languages, language("Plain Text", 0, 0, 0, 0))

	s := report.Summary(c)

	require.NotContains(t, s, "Plain Text")
	require.Contains(t, s, "| Markdown   |")
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
		"| Language | Baseline | Current | Delta |    Change |",
		"| -------- | -------: | ------: | ----: | --------: |",
		"| Total    |    0 (0) |   0 (0) | 0 (0) | n/a (n/a) |",
	}, "\n")

	s := report.Summary(&analyze.Comparison{})

	require.Contains(t, s, want)
}
