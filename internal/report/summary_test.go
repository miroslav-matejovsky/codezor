package report_test

import (
	"strings"
	"testing"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
	"github.com/miroslav-matejovsky/codezor/internal/report"
	"github.com/stretchr/testify/require"
)

// The console shows both metrics in one table, one row per language, each cell
// holding the complexity and the LOC side by side. Both numbers are aligned
// down the column, the LOC ones among themselves, so a column reads as two
// columns of digits. The table carries no markdown emphasis, so the total row
// sets no column width of its own.
func TestSummaryIsOneTableOfBothMetrics(t *testing.T) {
	want := strings.Join([]string{
		"| Language   |       Delta |      Change | Delta (main) | Change (main) |",
		"| ---------- | ----------: | ----------: | -----------: | ------------: |",
		"| Go         | +187 +1,181 | +4.5% +3.1% |  +187 +1,181 |   +4.5% +3.1% |",
		"| PowerShell |    0    +57 |  0.0% +4.7% |     0    +57 |    0.0% +4.7% |",
		"| YAML       |   -3    -76 | -9.7% -9.0% |    -3    -76 |   -9.7% -9.0% |",
		"| Total      | +184 +1,282 | +4.3% +3.2% |  +184 +1,282 |   +4.3% +3.2% |",
		"Complexity (LOC)",
	}, "\n")

	s := report.Summary(fixture(), fixture())

	require.Equal(t, want+"\n", s)
	require.NotContains(t, s, "Baseline")
	require.NotContains(t, s, "Current")
	require.NotContains(t, s, "working tree")
}

// Languages with no complexity on either side are hidden, even with LOC.
// Languages with complexity on either side keep their row.
func TestSummaryLeavesOutLanguagesWithoutComplexity(t *testing.T) {
	c := fixture()
	c.Languages = append(c.Languages,
		language("Plain Text", 0, 0, 0, 0),
		language("Added", 0, 10, 0, 1),
		language("Removed", 10, 0, 1, 0),
	)

	s := report.Summary(c, c)

	require.NotContains(t, s, "Plain Text")
	require.NotContains(t, s, "Markdown")
	require.Contains(t, s, "| Added")
	require.Contains(t, s, "| Removed")
}

func TestSummaryLinesHaveNoTrailingWhitespace(t *testing.T) {
	s := report.Summary(fixture(), fixture())

	for i, line := range strings.Split(strings.TrimSuffix(s, "\n"), "\n") {
		require.Equal(t, strings.TrimRight(line, " \t"), line, "line %d", i+1)
	}
}

// The summary is derived only from the comparison, like the markdown report.
func TestSummaryIsDeterministic(t *testing.T) {
	require.Equal(t, report.Summary(fixture(), fixture()), report.Summary(fixture(), fixture()))
}

// With no rows at all the headings still set the column widths.
func TestSummaryHandlesEmptyRepository(t *testing.T) {
	want := strings.Join([]string{
		"| Language | Delta |  Change | Delta (main) | Change (main) |",
		"| -------- | ----: | ------: | -----------: | ------------: |",
		"| Total    |   0 0 | n/a n/a |          0 0 |       n/a n/a |",
		"Complexity (LOC)",
	}, "\n")

	s := report.Summary(&analyze.Comparison{}, &analyze.Comparison{})

	require.Equal(t, want+"\n", s)
}

// Rows are joined by language, since their order and membership may differ.
// Complexity in any of HEAD, main's baseline or the working tree keeps a row.
func TestSummaryCombinesHeadAndMainChanges(t *testing.T) {
	head := &analyze.Comparison{
		Languages: []analyze.LanguageStats{
			language("Go", 20, 25, 2, 3),
			language("HeadOnly", 5, 0, 1, 0),
			language("NoBranches", 10, 15, 0, 0),
		},
		Total: language("", 35, 40, 3, 3),
	}
	main := &analyze.Comparison{
		Languages: []analyze.LanguageStats{
			language("MainOnly", 10, 0, 2, 0),
			language("Go", 10, 25, 1, 3),
			language("NoBranches", 10, 15, 0, 0),
		},
		Total: language("", 30, 40, 3, 3),
	}

	s := report.Summary(head, main)

	var rows [][]string
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(line, "|") {
			cells := strings.Split(strings.Trim(line, "|"), "|")
			for i := range cells {
				cells[i] = strings.Join(strings.Fields(cells[i]), " ")
			}
			rows = append(rows, cells)
		}
	}
	require.Equal(t, [][]string{
		{"Go", "+1 +5", "+50.0% +25.0%", "+2 +15", "+200.0% +150.0%"},
		{"HeadOnly", "-1 -5", "-100.0% -100.0%", "0 0", "n/a n/a"},
		{"MainOnly", "0 0", "n/a n/a", "-2 -10", "-100.0% -100.0%"},
		{"Total", "0 +5", "0.0% +14.3%", "0 +10", "0.0% +33.3%"},
	}, rows[2:])
}
