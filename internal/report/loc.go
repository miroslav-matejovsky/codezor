package report

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
)

// currentLine describes what the current column holds. Both renderings state
// it, so a reader of either one knows uncommitted and untracked code is in the
// numbers.
const currentLine = "Current: working tree, including uncommitted and untracked files"

// locColumns is the shape of the LOC table. Counts are right aligned so the
// digits line up under each other.
var locColumns = []column{
	{heading: "Language", alignment: alignLeft},
	{heading: "Baseline", alignment: alignRight},
	{heading: "Current", alignment: alignRight},
	{heading: "Delta", alignment: alignRight},
	{heading: "Change", alignment: alignRight},
}

// locRows is the body of the LOC table: one row per language, ordered as the
// comparison ordered them, and the total last. emphasiseTotal wraps the total
// row in markdown emphasis, which only the markdown rendering wants.
func locRows(c *analyze.Comparison, emphasiseTotal bool) [][]string {
	rows := make([][]string, 0, len(c.LOC)+1)
	for _, row := range c.LOC {
		rows = append(rows, locCells(row.Name, row, false))
	}

	return append(rows, locCells("Total", c.Total, emphasiseTotal))
}

// locCells is one row of the LOC table. The total row is emphasised, which is
// why the label is passed separately from the counts.
func locCells(label string, row analyze.LanguageLOC, emphasise bool) []string {
	cells := []string{
		label,
		group(row.Baseline),
		group(row.Current),
		signed(row.Delta()),
		changeCell(row),
	}

	if emphasise {
		for i, cell := range cells {
			cells[i] = "**" + cell + "**"
		}
	}

	return cells
}

// changeCell renders the relative change, or n/a for a language the baseline
// did not contain, where a percentage of zero would read as no change.
func changeCell(row analyze.LanguageLOC) string {
	if row.Baseline == 0 {
		return "n/a"
	}

	return signedPercent(row.ChangePercent())
}

// group writes n with commas between digit groups. Counts are never negative,
// so only non-negative input is grouped.
func group(n int64) string {
	s := strconv.FormatInt(n, 10)
	if n < 0 || len(s) <= 3 {
		return s
	}

	var b strings.Builder
	for i, digit := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteRune(',')
		}
		b.WriteRune(digit)
	}

	return b.String()
}

// signed writes a delta with an explicit sign, so a column of deltas reads as
// change rather than as a count.
func signed(n int64) string {
	switch {
	case n > 0:
		return "+" + group(n)
	case n < 0:
		return "-" + group(-n)
	default:
		return "0"
	}
}

// signedPercent writes a relative change with one decimal and an explicit
// sign.
func signedPercent(f float64) string {
	if f > 0 {
		return fmt.Sprintf("+%.1f%%", f)
	}

	return fmt.Sprintf("%.1f%%", f)
}
