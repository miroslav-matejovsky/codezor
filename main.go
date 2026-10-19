// Command codezor counts a source tree and estimates what producing that code
// with an LLM would cost, then writes the result as markdown.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
	"github.com/miroslav-matejovsky/codezor/internal/report"
)

func main() {
	path := flag.String("path", ".", "directory to analyze")
	out := flag.String("out", "codezor.md", "markdown report to write")
	preset := flag.String("locomo-preset", "medium",
		"LOCOMO model tier: "+strings.Join(analyze.Presets, ", "))
	flag.Parse()

	if err := run(*path, *out, *preset); err != nil {
		fmt.Fprintln(os.Stderr, "codezor:", err)
		os.Exit(1)
	}
}

// run analyzes the tree and writes the report, so main only handles flags and
// the exit code.
func run(path, out, preset string) error {
	a, err := analyze.Run(path, preset)
	if err != nil {
		return err
	}

	if err := os.WriteFile(out, []byte(report.Markdown(&a)), 0o600); err != nil {
		return fmt.Errorf("write %q: %w", out, err)
	}

	fmt.Println("codezor: wrote", out)

	return nil
}
