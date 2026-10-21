package report

import (
	"fmt"
	"strings"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
)

// Markdown renders the comparison as the body of codezor.md.
func Markdown(c *analyze.Comparison) string {
	lines := []string{
		"# Code Statistics",
		"",
		baselineLine(fmt.Sprintf("%#q", c.BaselineRef), c),
		currentLine,
		"",
		"## LOC",
		"",
	}
	lines = append(lines, metricTable(c, code, true)...)
	lines = append(lines, "", "## Complexity", "", complexityLine, "")
	lines = append(lines, metricTable(c, complexity, true)...)

	return strings.Join(lines, "\n") + "\n"
}
