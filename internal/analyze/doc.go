// Package analyze measures a repository on both sides of a baseline revision.
// Two quantities are counted per language: code lines and scc's complexity
// estimate.
//
// Two baselines are offered, and they answer different questions. Compare uses
// the merge base with a branch, so the delta is everything the current branch
// changed. CompareHead uses the commit checked out, so the delta is the
// uncommitted and untracked work alone.
//
// Files are enumerated by git and counted by scc. scc keeps its language
// database in package level state, so the counter is guarded by a one time
// load and the comparisons must not be called concurrently.
package analyze
