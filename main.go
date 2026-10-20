// Command codezor counts the code lines of a git repository against the
// branch it forked from, and writes the result as markdown or prints it as a
// summary table.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

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
	summary := flag.Bool("summary", false, "print the summary table to stdout and write no report")
	flag.Parse()

	if err := run(context.Background(), *path, *out, *summary); err != nil {
		fmt.Fprintln(os.Stderr, "codezor:", err)
		os.Exit(1)
	}
}

// run compares the tree and renders the comparison, so main only handles flags
// and the exit code.
//
// With summary set the table goes to stdout and nothing is written, which keeps
// a quick look at the current change out of the repository it measures.
func run(ctx context.Context, path, out string, summary bool) error {
	comparison, err := analyze.Compare(ctx, path, baselineBranch)
	if err != nil {
		return err
	}

	if summary {
		fmt.Print(report.Summary(&comparison))

		return nil
	}

	if err := os.WriteFile(out, []byte(report.Markdown(&comparison)), 0o600); err != nil {
		return fmt.Errorf("write %q: %w", out, err)
	}

	fmt.Println("codezor: wrote", out)

	return nil
}
