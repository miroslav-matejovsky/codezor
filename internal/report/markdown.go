package report

import (
	"strings"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
)

// Markdown renders the comparison as the body of codezor.md: a title and one
// table per metric, nothing else. What the two columns mean is documented
// rather than repeated in every report.
func Markdown(c *analyze.Comparison) string {
	lines := []string{
		"# Code Statistics",
		"",
		"## LOC",
		"",
	}
	lines = append(lines, metricTable(c, code, true)...)
	lines = append(lines, "", "## Complexity", "")
	lines = append(lines, metricTable(c, complexity, true)...)

	return strings.Join(lines, "\n") + "\n"
}
