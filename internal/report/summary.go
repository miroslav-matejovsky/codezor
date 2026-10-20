package report

import (
	"fmt"
	"strings"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
)

// Summary renders the comparison for a terminal: the baseline it was measured
// against and the LOC table, with no markdown emphasis, so the columns line up
// as printed.
//
// It is the same comparison the markdown report carries, written for a reader
// who wants the numbers in the shell rather than in a file.
func Summary(c *analyze.Comparison) string {
	lines := []string{
		fmt.Sprintf("Baseline: %s (merge-base %s)", c.BaselineRef, c.BaselineHash),
		currentLine,
		"",
	}
	lines = append(lines, renderTable(locColumns, locRows(c, false))...)

	return strings.Join(lines, "\n") + "\n"
}
