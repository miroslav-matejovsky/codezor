// Package analyze counts the code lines in a repository on both sides of a
// baseline revision.
//
// Files are enumerated by git and counted by scc. scc keeps its language
// database in package level state, so the counter is guarded by a one time
// load and Compare must not be called concurrently.
package analyze
