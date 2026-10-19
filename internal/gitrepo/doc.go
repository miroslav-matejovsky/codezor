// Package gitrepo reads a local git repository through the git command line.
//
// Only read operations are used and nothing reaches the network: the baseline
// revision, the tree behind it, and the diffs against it all come from the
// repository on disk. git is expected on PATH, which a tool that measures a
// working tree against a branch already implies.
package gitrepo
