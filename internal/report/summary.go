package report

import (
	"strings"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
)

// Summary renders changes against HEAD and the merge base with main in one
// terminal table. Each numeric cell holds complexity and LOC, identified by
// the Language (cplx loc) header. Languages with zero complexity across
// both comparisons are omitted; totals include all languages. The comparisons
// must describe the same working tree, with c relative to HEAD and main
// relative to the merge base with main.
func Summary(c, main *analyze.Comparison) string {
	lines := combinedTable(c, main)

	return strings.Join(lines, "\n") + "\n"
}
