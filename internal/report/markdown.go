package report

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
)

// Markdown renders the comparison as the body of codezor.md.
func Markdown(c *analyze.Comparison) string {
	lines := []string{
		"# Code Statistics",
		"",
		fmt.Sprintf("Baseline: %#q (merge-base %s)", c.BaselineRef, c.BaselineHash),
		"Current: working tree, including uncommitted and untracked files",
		"",
		"LOC is code lines as counted by [scc](https://github.com/boyter/scc):",
		"comment and blank lines are left out of every number below, churn",
		"included. The report carries no timestamp, so an unchanged tree renders",
		"an unchanged report.",
		"",
		"## LOC",
		"",
		"| Language | Baseline | Current | Delta | Change |",
		"| --- | ---: | ---: | ---: | ---: |",
	}

	for _, row := range c.LOC {
		lines = append(lines, locRow(row.Name, row, false))
	}

	lines = append(lines,
		locRow("Total", c.TotalLOC, true),
		"",
		"## Current Change",
		"",
	)
	lines = append(lines, churnLines(c.Current, c.TotalLOC.Baseline)...)

	lines = append(lines,
		"",
		"## Latest Commit",
		"",
	)
	lines = append(lines, churnLines(c.LatestCommit, 0)...)

	lines = append(lines,
		"",
		"## LOCOMO",
		"",
		"LOCOMO (LLM Output COst MOdel) estimates what producing the current",
		"code with an LLM would cost. It scales with code volume and complexity,",
		"and leaves out context reuse, test generation and debugging. It is a",
		"ballpark figure, as COCOMO is for human development.",
		"",
		"| Estimate | Value |",
		"| --- | ---: |",
		fmt.Sprintf("| Model tier | %s |", c.Locomo.Preset),
		fmt.Sprintf("| Input tokens | %s |", group(int64(c.Locomo.InputTokens))),
		fmt.Sprintf("| Output tokens | %s |", group(int64(c.Locomo.OutputTokens))),
		fmt.Sprintf("| Cost | $%.2f |", c.Locomo.Cost),
		fmt.Sprintf("| Iteration factor | %.2f |", c.Locomo.IterationFactor),
		fmt.Sprintf("| Generation time | %s |", duration(c.Locomo.GenerationSeconds)),
		fmt.Sprintf("| Review time | %s |", duration(c.Locomo.ReviewHours*3600)),
	)

	return strings.Join(lines, "\n") + "\n"
}

// locRow renders one row of the LOC table. The total row is emphasised, which
// is why the label is passed separately from the counts.
func locRow(label string, row analyze.LanguageLOC, emphasise bool) string {
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

	return "| " + strings.Join(cells, " | ") + " |"
}

// changeCell renders the relative change, or n/a for a language the baseline
// did not contain, where a percentage of zero would read as no change.
func changeCell(row analyze.LanguageLOC) string {
	if row.Baseline == 0 {
		return "n/a"
	}

	return signedPercent(row.ChangePercent())
}

// churnLines renders one churn block. A baseline of zero leaves out the
// footprint, which only means something relative to existing code.
func churnLines(c analyze.Churn, baseline int64) []string {
	lines := []string{
		fmt.Sprintf("- Added LOC: %s", group(c.Added)),
		fmt.Sprintf("- Removed LOC: %s", group(c.Removed)),
		fmt.Sprintf("- Gross churn: %s LOC", group(c.Gross())),
		fmt.Sprintf("- Net change: %s LOC", signed(c.Net())),
	}

	if baseline > 0 {
		lines = append(lines, fmt.Sprintf("- Change footprint: %.1f%% of baseline LOC", c.Footprint(baseline)))
	}

	return lines
}

// group writes n with commas between digit groups. Counts and token estimates
// are never negative, so only non-negative input is grouped.
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

// duration renders seconds in the largest unit that keeps the number small,
// because LOCOMO estimates span seconds for a file and days for a codebase.
func duration(seconds float64) string {
	switch {
	case seconds < 60:
		return fmt.Sprintf("%.0f s", seconds)
	case seconds < 3600:
		return fmt.Sprintf("%.1f min", seconds/60)
	case seconds < 86400:
		return fmt.Sprintf("%.1f h", seconds/3600)
	default:
		return fmt.Sprintf("%.1f days", seconds/86400)
	}
}
