package markdown

import (
	"strings"
	"unicode/utf8"
)

// minColumnWidth is the narrowest a separator cell can be and still carry the
// three dashes markdown wants.
const minColumnWidth = 3

// Alignment is which side of a cell the padding goes on.
type Alignment int

// The alignments a column can have: padding on the right, or on the left.
const (
	AlignLeft Alignment = iota
	AlignRight
)

// Column is one column of a table: its heading and how its cells are aligned.
type Column struct {
	Heading   string
	Alignment Alignment
}

// Pair is a cell holding two values, the second one in parentheses.
type Pair struct {
	Lead          string
	Parenthesised string
}

// Table lays out a markdown table with padded columns.
//
// Every row must hold one cell per column.
func Table(columns []Column, rows [][]string) []string {
	widths := columnWidths(columns, rows)

	headings := make([]string, len(columns))
	separators := make([]string, len(columns))
	for i, c := range columns {
		headings[i] = pad(c.Heading, widths[i], c.Alignment)
		separators[i] = separator(widths[i], c.Alignment)
	}

	lines := make([]string, 0, len(rows)+2)
	lines = append(lines, tableRow(headings), tableRow(separators))

	for _, row := range rows {
		cells := make([]string, len(row))
		for i, cell := range row {
			cells[i] = pad(cell, widths[i], columns[i].Alignment)
		}
		lines = append(lines, tableRow(cells))
	}

	return lines
}

// PairCells renders one column's pairs as cells. Each part is padded to the
// widest of its own kind, so the leading values line up under each other and
// the parenthesised ones do too: one column of pairs reads as two columns of
// digits.
//
// Both parts are right aligned, since a pair is a pair of counts. Every cell
// comes out the same width, so Table then pads the column only against its
// heading.
func PairCells(pairs []Pair) []string {
	var lead, parenthesised int
	for _, p := range pairs {
		lead = max(lead, utf8.RuneCountInString(p.Lead))
		parenthesised = max(parenthesised, utf8.RuneCountInString(p.Parenthesised))
	}

	cells := make([]string, len(pairs))
	for i, p := range pairs {
		cells[i] = pad(p.Lead, lead, AlignRight) + " (" + pad(p.Parenthesised, parenthesised, AlignRight) + ")"
	}

	return cells
}

// columnWidths is the width each column needs to hold its heading and every
// cell under it.
func columnWidths(columns []Column, rows [][]string) []int {
	widths := make([]int, len(columns))
	for i, c := range columns {
		widths[i] = max(minColumnWidth, utf8.RuneCountInString(c.Heading))
	}

	for _, row := range rows {
		for i, cell := range row {
			widths[i] = max(widths[i], utf8.RuneCountInString(cell))
		}
	}

	return widths
}

// pad widens cell to width, putting the spaces on the side the alignment asks
// for. Width is counted in runes, so a non-ASCII language name still lines up.
func pad(cell string, width int, a Alignment) string {
	gap := width - utf8.RuneCountInString(cell)
	if gap <= 0 {
		return cell
	}

	if a == AlignRight {
		return strings.Repeat(" ", gap) + cell
	}

	return cell + strings.Repeat(" ", gap)
}

// separator is the row under the headings: dashes as wide as the column, with
// the colon that marks a right aligned one.
func separator(width int, a Alignment) string {
	if a == AlignRight {
		return strings.Repeat("-", width-1) + ":"
	}

	return strings.Repeat("-", width)
}

// tableRow joins cells into one markdown row.
func tableRow(cells []string) string {
	return "| " + strings.Join(cells, " | ") + " |"
}
