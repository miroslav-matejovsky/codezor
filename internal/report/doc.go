// Package report renders an analysis, either as the markdown codezor report or
// as a summary for the console. Both renderings are tables and nothing else:
// one table per metric, LOC and complexity, with the same columns over the
// same rows; only the markdown one adds headings and emphasises the totals.
// What the columns mean belongs in the documentation, so neither rendering
// repeats it.
//
// Everything in the output is derived from the analysis, so the same tree
// always renders the same bytes. Nothing dated or environment specific is
// added, which keeps the report diffable between runs.
package report
