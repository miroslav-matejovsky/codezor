// Command codezor counts the code lines of a git repository. The report it
// writes compares the working tree against the branch it forked from; the
// summary compares it against both HEAD and the merge base with main.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
	"github.com/miroslav-matejovsky/codezor/internal/report"
)

// baselineBranch is the branch the report and summary compare against. The
// baseline commit is the merge base with it, not its tip, so work pushed to it
// after this branch started is not reported as part of this change.
const baselineBranch = "main"

func main() {
	opts, err := parseOptions(os.Args[1:], os.Stderr)
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err == nil {
		err = run(context.Background(), opts.path, opts.out, opts.summary)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "codezor:", err)
		os.Exit(1)
	}
}

type options struct {
	path    string
	out     string
	summary bool
}

// parseOptions accepts one repository path, with flags before or after it.
func parseOptions(args []string, output io.Writer) (options, error) {
	var opts options
	flags := flag.NewFlagSet("codezor", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.StringVar(&opts.path, "path", ".", "directory inside the repository to analyze")
	flags.StringVar(&opts.out, "out", "", "markdown report to write (default: print summary)")
	flags.BoolVar(&opts.summary, "summary", false, "print the summary table to stdout and write no report")

	var positional []string
	for {
		if err := flags.Parse(args); err != nil {
			return opts, err
		}
		args = flags.Args()
		if len(args) == 0 {
			break
		}
		positional = append(positional, args[0])
		args = args[1:]
	}
	pathFlag := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "path" {
			pathFlag = true
		}
	})
	if len(positional) > 1 || (len(positional) == 1 && pathFlag) {
		return opts, errors.New("provide one repository path, either as an argument or with -path")
	}
	if len(positional) == 1 {
		opts.path = positional[0]
	}

	return opts, nil
}

// run picks the rendering, so main only handles flags and the exit code.
func run(ctx context.Context, path, out string, summary bool) error {
	if summary || out == "" {
		return printSummary(ctx, path)
	}

	return writeReport(ctx, path, out)
}

// printSummary writes changes against HEAD and the merge base with main to
// stdout, including uncommitted and untracked work in both comparisons.
func printSummary(ctx context.Context, path string) error {
	comparison, err := analyze.CompareHead(ctx, path)
	if err != nil {
		return err
	}

	mainComparison, err := analyze.Compare(ctx, path, baselineBranch)
	if err != nil {
		return err
	}

	fmt.Print(report.Summary(&comparison, &mainComparison))

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
