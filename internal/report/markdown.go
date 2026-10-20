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
	lines = append(lines, renderTable(locColumns, locRows(c, true))...)

	return strings.Join(lines, "\n") + "\n"
}
