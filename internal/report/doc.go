// Package report renders an analysis, either as the markdown codezor report or
// as a summary for the console. Both renderings name the baseline the
// comparison was built against and then show one table per metric, LOC and
// complexity, with the same columns over the same rows; only the markdown one
// adds headings and emphasises the totals.
//
// Everything in the output is derived from the analysis, so the same tree
// always renders the same bytes. Nothing dated or environment specific is
// added, which keeps the report diffable between runs.
package report
