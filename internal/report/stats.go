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

// complexityLine says what the complexity numbers are, so a reader does not
// take them for a measured cyclomatic complexity.
const complexityLine = "Branching keywords and operators as scc counts them, not a cyclomatic complexity."

// baselineLine names the revision the current column is measured against. The
// merge base is called out as such, because there the delta covers the commits
// of this branch too and not only the work on disk.
//
// ref is already formatted, so markdown can quote it and the console does not
// have to.
func baselineLine(ref string, c *analyze.Comparison) string {
	if c.BaselineIsMergeBase {
		return fmt.Sprintf("Baseline: %s (merge-base %s)", ref, c.BaselineHash)
	}

	return fmt.Sprintf("Baseline: %s (%s)", ref, c.BaselineHash)
}

// metricColumns is the shape of a metric table. Counts are right aligned so
// the digits line up under each other.
var metricColumns = []column{
	{heading: "Language", alignment: alignLeft},
	{heading: "Baseline", alignment: alignRight},
	{heading: "Current", alignment: alignRight},
	{heading: "Delta", alignment: alignRight},
	{heading: "Change", alignment: alignRight},
}

// code and complexity select which measurement a table renders. Both tables
// have the same shape, so one renderer serves them with the metric passed in.
func code(l analyze.LanguageStats) analyze.Metric { return l.Code }

func complexity(l analyze.LanguageStats) analyze.Metric { return l.Complexity }

// metricTable renders one metric as a table: a row per language, ordered as
// the comparison ordered them, and the total last. emphasiseTotal wraps the
// total row in markdown emphasis, which only the markdown rendering wants.
func metricTable(c *analyze.Comparison, pick func(analyze.LanguageStats) analyze.Metric, emphasiseTotal bool) []string {
	rows := make([][]string, 0, len(c.Languages)+1)
	for _, language := range c.Languages {
		rows = append(rows, metricCells(language.Name, pick(language), false))
	}
	rows = append(rows, metricCells("Total", pick(c.Total), emphasiseTotal))

	return renderTable(metricColumns, rows)
}

// metricCells is one row of a metric table. The total row is emphasised, which
// is why the label is passed separately from the counts.
func metricCells(label string, m analyze.Metric, emphasise bool) []string {
	cells := []string{
		label,
		group(m.Baseline),
		group(m.Current),
		signed(m.Delta()),
		changeCell(m),
	}

	if emphasise {
		for i, cell := range cells {
			cells[i] = "**" + cell + "**"
		}
	}

	return cells
}

// changeCell renders the relative change, or n/a for a zero baseline, where a
// percentage of zero would read as no change.
func changeCell(m analyze.Metric) string {
	if m.Baseline == 0 {
		return "n/a"
	}

	return signedPercent(m.ChangePercent())
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
