// Command codezor measures the working tree of a git repository against the
// branch it forked from, and writes the result as markdown.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
	"github.com/miroslav-matejovsky/codezor/internal/report"
)

// baselineBranch is the branch every comparison is made against. The baseline
// commit is the merge base with it, not its tip, so work pushed to it after
// this branch started is not reported as part of this change.
const baselineBranch = "main"

func main() {
	path := flag.String("path", ".", "directory inside the repository to analyze")
	out := flag.String("out", "codezor.md", "markdown report to write")
	preset := flag.String("locomo-preset", "medium",
		"LOCOMO model tier: "+strings.Join(analyze.Presets, ", "))
	flag.Parse()

	if err := run(context.Background(), *path, *out, *preset); err != nil {
		fmt.Fprintln(os.Stderr, "codezor:", err)
		os.Exit(1)
	}
}

// run compares the tree and writes the report, so main only handles flags and
// the exit code.
func run(ctx context.Context, path, out, preset string) error {
	comparison, err := analyze.Compare(ctx, path, baselineBranch, preset)
	if err != nil {
		return err
	}

	if err := os.WriteFile(out, []byte(report.Markdown(&comparison)), 0o600); err != nil {
		return fmt.Errorf("write %q: %w", out, err)
	}

	fmt.Println("codezor: wrote", out)

	return nil
}
