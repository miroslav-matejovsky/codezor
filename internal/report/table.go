package report

import (
	"strings"
	"unicode/utf8"
)

// minColumnWidth is the narrowest a separator cell can be and still carry the
// three dashes markdown wants.
const minColumnWidth = 3

// alignment is which side of a cell the padding goes on.
type alignment int

const (
	alignLeft alignment = iota
	alignRight
)

// column is one column of a table: its heading and how its cells are aligned.
type column struct {
	heading   string
	alignment alignment
}

// renderTable lays out a markdown table with padded columns, so the table
// reads as a table in the file itself and not only once a renderer has been
// through it.
//
// Every row must hold one cell per column.
func renderTable(columns []column, rows [][]string) []string {
	widths := columnWidths(columns, rows)

	headings := make([]string, len(columns))
	separators := make([]string, len(columns))
	for i, c := range columns {
		headings[i] = pad(c.heading, widths[i], c.alignment)
		separators[i] = separator(widths[i], c.alignment)
	}

	lines := make([]string, 0, len(rows)+2)
	lines = append(lines, tableRow(headings), tableRow(separators))

	for _, row := range rows {
		cells := make([]string, len(row))
		for i, cell := range row {
			cells[i] = pad(cell, widths[i], columns[i].alignment)
		}
		lines = append(lines, tableRow(cells))
	}

	return lines
}

// columnWidths is the width each column needs to hold its heading and every
// cell under it.
func columnWidths(columns []column, rows [][]string) []int {
	widths := make([]int, len(columns))
	for i, c := range columns {
		widths[i] = max(minColumnWidth, utf8.RuneCountInString(c.heading))
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
func pad(cell string, width int, a alignment) string {
	gap := width - utf8.RuneCountInString(cell)
	if gap <= 0 {
		return cell
	}

	if a == alignRight {
		return strings.Repeat(" ", gap) + cell
	}

	return cell + strings.Repeat(" ", gap)
}

// separator is the row under the headings: dashes as wide as the column, with
// the colon that marks a right aligned one.
func separator(width int, a alignment) string {
	if a == alignRight {
		return strings.Repeat("-", width-1) + ":"
	}

	return strings.Repeat("-", width)
}

// tableRow joins cells into one markdown row.
func tableRow(cells []string) string {
	return "| " + strings.Join(cells, " | ") + " |"
}
