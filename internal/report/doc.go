// Package report renders an analysis, either as the markdown codezor report or
// as a summary table for the console. Both renderings share the table layout
// and name the baseline the comparison was built against; only the markdown
// one adds a heading and emphasises the total.
//
// Everything in the output is derived from the analysis, so the same tree
// always renders the same bytes. Nothing dated or environment specific is
// added, which keeps the report diffable between runs.
package report
