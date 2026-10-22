// Package report renders an analysis, either as the markdown codezor report or
// as a summary for the console. Both renderings order languages by current LOC.
// The markdown report gives LOC and complexity a table each under a heading,
// with Baseline, Current, Delta and Change columns and emphasised totals.
// The console shows Delta and Change against HEAD and the merge base with main,
// with both metrics per cell and one row per language. The Complexity (LOC)
// label follows immediately below the table. Both complexity views omit
// languages with zero complexity across their comparisons;
// totals include every language. What the columns mean belongs in the
// documentation, so neither rendering repeats it.
//
// Everything in the output is derived from the analysis, so the same tree
// always renders the same bytes. Nothing dated or environment specific is
// added, which keeps the report diffable between runs.
package report
