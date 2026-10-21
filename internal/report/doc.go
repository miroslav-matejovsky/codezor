// Package report renders an analysis, either as the markdown codezor report or
// as a summary for the console. Both renderings are tables and nothing else,
// over the same columns and the same ordering: the markdown one gives each
// metric, LOC and complexity, a table under a heading and emphasises the
// totals, while the console one puts both metrics in every cell to keep a
// language to a single row. What the columns mean belongs in the
// documentation, so neither rendering repeats it.
//
// Everything in the output is derived from the analysis, so the same tree
// always renders the same bytes. Nothing dated or environment specific is
// added, which keeps the report diffable between runs.
package report
