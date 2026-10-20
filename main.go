// Command codezor counts the code lines of a git repository. The report it
// writes compares the working tree against the branch it forked from; the
// summary it prints compares the working tree against the current branch.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
	"github.com/miroslav-matejovsky/codezor/internal/report"
)

// baselineBranch is the branch the written report is compared against. The
// baseline commit is the merge base with it, not its tip, so work pushed to it
// after this branch started is not reported as part of this change.
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

// run picks the rendering, so main only handles flags and the exit code.
func run(ctx context.Context, path, out string, summary bool) error {
	if summary {
		return printSummary(ctx, path)
	}

	return writeReport(ctx, path, out)
}

// printSummary writes the table for the uncommitted and untracked work to
// stdout. The baseline is the current branch and nothing is written to disk,
// which keeps a look at the change in the editor out of the repository it
// measures.
func printSummary(ctx context.Context, path string) error {
	comparison, err := analyze.CompareHead(ctx, path)
	if err != nil {
		return err
	}

	fmt.Print(report.Summary(&comparison))

	return nil
}

// writeReport writes the markdown report for everything this branch changed
// against baselineBranch.
func writeReport(ctx context.Context, path, out string) error {
	comparison, err := analyze.Compare(ctx, path, baselineBranch)
	if err != nil {
		return err
	}

	if err := os.WriteFile(out, []byte(report.Markdown(&comparison)), 0o600); err != nil {
		return fmt.Errorf("write %q: %w", out, err)
	}

	fmt.Println("codezor: wrote", out)

	return nil
}
