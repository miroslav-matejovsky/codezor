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
		Total: analyze.LanguageLOC{Baseline: 40466, Current: 41748},
	}
}

func TestMarkdownNamesTheBaseline(t *testing.T) {
	md := report.Markdown(fixture())

	require.Contains(t, md, "# Code Statistics")
	require.Contains(t, md, "Baseline: `main` (merge-base a1b2c3d)")
}

// Every column is padded to its widest cell, so the table lines up in the file
// and not only once a renderer has been through it.
func TestMarkdownAlignsTheLOCTable(t *testing.T) {
	want := strings.Join([]string{
		"| Language   |   Baseline |    Current |      Delta |    Change |",
		"| ---------- | ---------: | ---------: | ---------: | --------: |",
		"| Go         |     38,420 |     39,601 |     +1,181 |     +3.1% |",
		"| PowerShell |      1,204 |      1,261 |        +57 |     +4.7% |",
		"| Markdown   |          0 |        120 |       +120 |       n/a |",
		"| YAML       |        842 |        766 |        -76 |     -9.0% |",
		"| **Total**  | **40,466** | **41,748** | **+1,282** | **+3.2%** |",
	}, "\n")

	require.Contains(t, report.Markdown(fixture()), want)
}

// No line may carry trailing whitespace, which editors and formatters strip
// and which would then show up as a change to the report.
func TestMarkdownLinesHaveNoTrailingWhitespace(t *testing.T) {
	md := report.Markdown(fixture())

	for i, line := range strings.Split(strings.TrimSuffix(md, "\n"), "\n") {
		require.Equal(t, strings.TrimRight(line, " \t"), line, "line %d", i+1)
	}
}

// The report is derived only from the comparison, so the same tree always
// renders the same bytes.
func TestMarkdownIsDeterministic(t *testing.T) {
	require.Equal(t, report.Markdown(fixture()), report.Markdown(fixture()))
}

// A language the baseline did not contain has no percentage: reporting zero
// would read as no change. With no rows at all the headings still set the
// column widths.
func TestMarkdownHandlesEmptyRepository(t *testing.T) {
	want := strings.Join([]string{
		"| Language  | Baseline | Current | Delta |  Change |",
		"| --------- | -------: | ------: | ----: | ------: |",
		"| **Total** |    **0** |   **0** | **0** | **n/a** |",
	}, "\n")

	md := report.Markdown(&analyze.Comparison{BaselineRef: "main", BaselineHash: "0000000"})

	require.Contains(t, md, want)
}
