// Package markdown lays out markdown tables: it pads the cells so a table
// reads as a table in the file itself, and not only once a renderer has been
// through it.
//
// It knows nothing of what is being tabulated. Cells arrive as strings already
// formatted by the caller; all this package decides is how wide each column is
// and which side the padding goes on. Pairs are the one exception worth a type
// of their own: a cell holding a value and a second one in parentheses needs
// both parts aligned down the column, which cannot be done cell by cell.
package markdown
