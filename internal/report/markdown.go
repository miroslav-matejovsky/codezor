package report

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
)

// Markdown renders the analysis as the body of codezor.md.
func Markdown(a *analyze.Analysis) string {
	lines := []string{
		"# codezor",
		"",
		fmt.Sprintf("Source counts for %#q, measured with [scc](https://github.com/boyter/scc).", a.Path),
		"The report describes the working tree as it is on disk and carries no timestamp,",
		"so an unchanged tree renders an unchanged report.",
		"",
		"## Code",
		"",
		"| Language | Files | Lines | Code | Comments | Blanks | Complexity |",
		"| --- | ---: | ---: | ---: | ---: | ---: | ---: |",
	}

	for _, l := range a.Languages {
		lines = append(lines, countsRow(l.Name, l.Counts))
	}
	lines = append(lines,
		countsRow("**Total**", a.Totals),
		"",
		"## LOCOMO",
		"",
		"LOCOMO (LLM Output COst MOdel) estimates what producing this code with an LLM",
		"would cost. It scales with code volume and complexity, and leaves out context",
		"reuse, test generation and debugging. It is a ballpark figure, as COCOMO is for",
		"human development.",
		"",
		"| Estimate | Value |",
		"| --- | ---: |",
		fmt.Sprintf("| Model tier | %s |", a.Locomo.Preset),
		fmt.Sprintf("| Input tokens | %s |", group(int64(a.Locomo.InputTokens))),
		fmt.Sprintf("| Output tokens | %s |", group(int64(a.Locomo.OutputTokens))),
		fmt.Sprintf("| Cost | $%.2f |", a.Locomo.Cost),
		fmt.Sprintf("| Iteration factor | %.2f |", a.Locomo.IterationFactor),
		fmt.Sprintf("| Generation time | %s |", duration(a.Locomo.GenerationSeconds)),
		fmt.Sprintf("| Review time | %s |", duration(a.Locomo.ReviewHours*3600)),
	)

	return strings.Join(lines, "\n") + "\n"
}

// countsRow renders one row of the code table, label included verbatim so the
// total row can carry markdown emphasis.
func countsRow(label string, c analyze.Counts) string {
	return fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |",
		label,
		group(c.Files),
		group(c.Lines),
		group(c.Code),
		group(c.Comments),
		group(c.Blanks),
		group(c.Complexity),
	)
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
