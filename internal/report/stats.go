package report

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
	"github.com/miroslav-matejovsky/codezor/internal/markdown"
)

// metricColumns is the shape of each report table. Counts are right aligned so
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

// combinedTable renders Delta and Change for both metrics as one table, a row
// per language, with complexity and LOC side by side in each cell. The console
// wants the whole picture in one glance rather than two tables to scroll
// between.
//
// Languages with zero complexity across both comparisons are left out. The
// total includes all languages and is always kept, even when both metrics are zero.
//
// The cells are built a column at a time, because aligning both numbers of a
// pair needs the whole column: one cell does not know how wide the widest LOC
// below it is.
func combinedTable(c, main *analyze.Comparison) []string {
	mainByName := make(map[string]analyze.LanguageStats, len(main.Languages))
	for _, language := range main.Languages {
		mainByName[language.Name] = language
	}

	stats := make([]analyze.LanguageStats, 0, len(c.Languages)+1)
	mainStats := make([]analyze.LanguageStats, 0, len(main.Languages)+1)
	seen := make(map[string]bool, len(c.Languages))
	for _, language := range c.Languages {
		seen[language.Name] = true
		mainLanguage := mainByName[language.Name]
		if empty(language.Complexity) && empty(mainLanguage.Complexity) {
			continue
		}

		stats = append(stats, language)
		mainStats = append(mainStats, mainLanguage)
	}
	// Languages present only in main's baseline have zero current LOC and go
	// last, preserving the comparison's ordering for the remaining rows.
	for _, language := range main.Languages {
		if seen[language.Name] || empty(language.Complexity) {
			continue
		}
		stats = append(stats, analyze.LanguageStats{Name: language.Name})
		mainStats = append(mainStats, language)
	}

	total := c.Total
	total.Name = "Total"
	stats = append(stats, total)
	mainStats = append(mainStats, main.Total)

	delta := make([]markdown.Pair, len(stats))
	change := make([]markdown.Pair, len(stats))
	mainDelta := make([]markdown.Pair, len(stats))
	mainChange := make([]markdown.Pair, len(stats))
	for i, s := range stats {
		delta[i] = markdown.Pair{Lead: signed(s.Complexity.Delta()), Trailing: signed(s.Code.Delta())}
		change[i] = markdown.Pair{Lead: changeCell(s.Complexity), Trailing: changeCell(s.Code)}
		m := mainStats[i]
		mainDelta[i] = markdown.Pair{Lead: signed(m.Complexity.Delta()), Trailing: signed(m.Code.Delta())}
		mainChange[i] = markdown.Pair{Lead: changeCell(m.Complexity), Trailing: changeCell(m.Code)}
	}

	deltaCells := markdown.PairCells(delta)
	changeCells := markdown.PairCells(change)
	mainDeltaCells := markdown.PairCells(mainDelta)
	mainChangeCells := markdown.PairCells(mainChange)

	rows := make([][]string, len(stats))
	for i, s := range stats {
		rows[i] = []string{s.Name, deltaCells[i], changeCells[i], mainDeltaCells[i], mainChangeCells[i]}
	}

	columns := []markdown.Column{
		{Heading: "Language (cplx loc)", Alignment: markdown.AlignLeft},
		{Heading: "Delta", Alignment: markdown.AlignRight},
		{Heading: "Change", Alignment: markdown.AlignRight},
		{Heading: "Delta (main)", Alignment: markdown.AlignRight},
		{Heading: "Change (main)", Alignment: markdown.AlignRight},
	}

	return markdown.Table(columns, rows)
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
