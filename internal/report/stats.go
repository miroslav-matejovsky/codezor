package report

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
	"github.com/miroslav-matejovsky/codezor/internal/markdown"
)

// metricColumns is the shape of every table here. Counts are right aligned so
// the digits line up under each other.
var metricColumns = []markdown.Column{
	{Heading: "Language", Alignment: markdown.AlignLeft},
	{Heading: "Baseline", Alignment: markdown.AlignRight},
	{Heading: "Current", Alignment: markdown.AlignRight},
	{Heading: "Delta", Alignment: markdown.AlignRight},
	{Heading: "Change", Alignment: markdown.AlignRight},
}

// code and complexity select which measurement a table renders. Both tables
// have the same shape, so one renderer serves them with the metric passed in.
func code(l analyze.LanguageStats) analyze.Metric { return l.Code }

func complexity(l analyze.LanguageStats) analyze.Metric { return l.Complexity }

// metricTable renders one metric as a table: a row per language, ordered as
// the comparison ordered them, and the total last. emphasiseTotal wraps the
// total row in markdown emphasis, which only the markdown rendering wants.
//
// A language the metric is zero on both sides of carries no information and is
// left out, which is what keeps the languages of prose and data out of the
// complexity table. The total still counts them, and the total row is kept
// whatever it holds: it is what tells a reader the table is a table.
func metricTable(c *analyze.Comparison, pick func(analyze.LanguageStats) analyze.Metric, emphasiseTotal bool) []string {
	rows := make([][]string, 0, len(c.Languages)+1)
	for _, language := range c.Languages {
		metric := pick(language)
		if empty(metric) {
			continue
		}

		rows = append(rows, metricCells(language.Name, metric, false))
	}
	rows = append(rows, metricCells("Total", pick(c.Total), emphasiseTotal))

	return markdown.Table(metricColumns, rows)
}

// combinedTable renders both metrics as one table, a row per language, with
// each cell holding the complexity and the LOC in parentheses. The console
// wants the whole picture in one glance rather than two tables to scroll
// between.
//
// A language with nothing on either side is left out. A language with lines
// but no branches keeps its row, because its LOC still says something; only a
// language absent from both sides is noise. The total row is kept whatever it
// holds: it is what tells a reader the table is a table.
//
// The cells are built a column at a time, because aligning both numbers of a
// pair needs the whole column: one cell does not know how wide the widest LOC
// below it is.
func combinedTable(c *analyze.Comparison) []string {
	stats := make([]analyze.LanguageStats, 0, len(c.Languages)+1)
	for _, language := range c.Languages {
		if empty(language.Code) && empty(language.Complexity) {
			continue
		}

		stats = append(stats, language)
	}

	total := c.Total
	total.Name = "Total"
	stats = append(stats, total)

	baseline := make([]markdown.Pair, len(stats))
	current := make([]markdown.Pair, len(stats))
	delta := make([]markdown.Pair, len(stats))
	change := make([]markdown.Pair, len(stats))
	for i, s := range stats {
		baseline[i] = markdown.Pair{Lead: group(s.Complexity.Baseline), Parenthesised: group(s.Code.Baseline)}
		current[i] = markdown.Pair{Lead: group(s.Complexity.Current), Parenthesised: group(s.Code.Current)}
		delta[i] = markdown.Pair{Lead: signed(s.Complexity.Delta()), Parenthesised: signed(s.Code.Delta())}
		change[i] = markdown.Pair{Lead: changeCell(s.Complexity), Parenthesised: changeCell(s.Code)}
	}

	counts := [][]string{
		markdown.PairCells(baseline),
		markdown.PairCells(current),
		markdown.PairCells(delta),
		markdown.PairCells(change),
	}

	rows := make([][]string, len(stats))
	for i, s := range stats {
		row := make([]string, 0, len(metricColumns))
		row = append(row, s.Name)
		for _, cells := range counts {
			row = append(row, cells[i])
		}
		rows[i] = row
	}

	return markdown.Table(metricColumns, rows)
}

// empty says whether a metric carries nothing on either side.
func empty(m analyze.Metric) bool {
	return m.Baseline == 0 && m.Current == 0
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
