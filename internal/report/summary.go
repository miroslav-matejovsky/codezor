package report

import (
	"strings"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
)

// Summary renders the comparison for a terminal: the LOC and complexity
// tables under bare labels, with no markdown emphasis, so the columns line up
// as printed.
func Summary(c *analyze.Comparison) string {
	lines := []string{
		"LOC",
		"",
	}
	lines = append(lines, metricTable(c, code, false)...)
	lines = append(lines, "", "Complexity", "")
	lines = append(lines, metricTable(c, complexity, false)...)

	return strings.Join(lines, "\n") + "\n"
}
