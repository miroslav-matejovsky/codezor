// Package report renders an analysis, either as the markdown codezor report or
// as a summary table for the console. Both renderings carry the same numbers
// and share the table layout; only the markdown one emphasises the total and
// wraps the baseline in a heading.
//
// Everything in the output is derived from the analysis, so the same tree
// always renders the same bytes. Nothing dated or environment specific is
// added, which keeps the report diffable between runs.
package report
