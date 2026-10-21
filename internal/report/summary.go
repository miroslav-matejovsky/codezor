package report

import (
	"strings"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
)

// Summary renders the comparison for a terminal: one table holding both
// metrics, with no markdown emphasis, so the columns line up as printed. The
// label says which number is which, since the column headings cannot.
func Summary(c *analyze.Comparison) string {
	lines := []string{
		"Complexity (LOC)",
		"",
	}
	lines = append(lines, combinedTable(c)...)

	return strings.Join(lines, "\n") + "\n"
}
