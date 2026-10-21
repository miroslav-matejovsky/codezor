package markdown_test

import (
	"strings"
	"testing"

	"github.com/miroslav-matejovsky/codezor/internal/markdown"
	"github.com/stretchr/testify/require"
)

// Every column is padded to its widest cell, and the separator carries the
// colon of a right aligned column, so the table lines up in the file and not
// only once a renderer has been through it.
func TestTablePadsEveryColumnToItsWidestCell(t *testing.T) {
	want := strings.Join([]string{
		"| Language   | Count |",
		"| ---------- | ----: |",
		"| Go         | 1,041 |",
		"| PowerShell |   111 |",
	}, "\n")

	got := markdown.Table(
		[]markdown.Column{
			{Heading: "Language", Alignment: markdown.AlignLeft},
			{Heading: "Count", Alignment: markdown.AlignRight},
		},
		[][]string{
			{"Go", "1,041"},
			{"PowerShell", "111"},
		},
	)

	require.Equal(t, want, strings.Join(got, "\n"))
}

// A heading narrower than three characters is still widened: markdown wants
// three dashes under it.
func TestTableKeepsTheSeparatorThreeDashesWide(t *testing.T) {
	want := strings.Join([]string{
		"| A   |   B |",
		"| --- | --: |",
		"| x   |   y |",
	}, "\n")

	got := markdown.Table(
		[]markdown.Column{
			{Heading: "A", Alignment: markdown.AlignLeft},
			{Heading: "B", Alignment: markdown.AlignRight},
		},
		[][]string{{"x", "y"}},
	)

	require.Equal(t, want, strings.Join(got, "\n"))
}

// With no rows the headings alone set the widths.
func TestTableRendersHeadingsWithoutRows(t *testing.T) {
	want := []string{
		"| Language |",
		"| -------- |",
	}

	got := markdown.Table([]markdown.Column{{Heading: "Language", Alignment: markdown.AlignLeft}}, nil)

	require.Equal(t, want, got)
}

// Both parts of a pair are padded to the widest of their own kind, so a column
// of pairs reads as two columns of digits. Every cell comes out the same width,
// which is what lets Table align the column as a whole.
func TestPairCellsAlignBothPartsSeparately(t *testing.T) {
	want := []string{
		"4,201 (38,420)",
		"   96 ( 1,204)",
		"    0 (     0)",
	}

	got := markdown.PairCells([]markdown.Pair{
		{Lead: "4,201", Parenthesised: "38,420"},
		{Lead: "96", Parenthesised: "1,204"},
		{Lead: "0", Parenthesised: "0"},
	})

	require.Equal(t, want, got)
}

func TestPairCellsHandlesNoPairs(t *testing.T) {
	require.Empty(t, markdown.PairCells(nil))
}
