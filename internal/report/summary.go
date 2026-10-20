package report

import (
	"strings"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
)

// Summary renders the comparison for a terminal: the baseline it was measured
// against and the LOC table, with no markdown emphasis, so the columns line up
// as printed.
//
// The baseline is whatever the comparison was built against, which the first
// line names: the summary does not assume the branch comparison the markdown
// report uses.
func Summary(c *analyze.Comparison) string {
	lines := []string{
		baselineLine(c.BaselineRef, c),
		currentLine,
		"",
	}
	lines = append(lines, renderTable(locColumns, locRows(c, false))...)

	return strings.Join(lines, "\n") + "\n"
}
