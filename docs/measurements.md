# Measurements

Every number codezor prints or writes comes from one comparison: a committed
baseline tree against the working tree. This document defines both sides.

## Baseline

The report's baseline is the merge base between `main` and `HEAD`, not the tip
of `main`. The delta is therefore what this branch changed: commits that landed
on `main` after the branch started do not move the numbers. On `main` itself the
merge base is `HEAD`, so the report shows uncommitted work alone.

The summary's baseline is `HEAD`, the tip of the current branch. Whatever this
branch has committed counts on both sides, so the delta is the uncommitted and
untracked work alone: the change in the editor rather than the whole branch.

The baseline side is the committed tree at that revision. The current side is
the files on disk: tracked files, including staged and unstaged edits, plus
untracked ones, and nothing `.gitignore` covers.

Both sides are enumerated by git and counted by the same counter, so the two
columns are comparable by construction. Everything is measured from the work
tree root, so the result does not depend on which subdirectory codezor ran in.

A repository without commits has no baseline, and a run there fails rather than
reporting the working tree against nothing.

## LOC

LOC is code lines as scc counts them. Comment and blank lines are excluded.
Three kinds of file are counted on neither side: binaries, files whose language
scc does not recognise, and ignore files such as `.gitignore`, which steer a
walk rather than hold code.

Rows are ordered by current size, so the largest languages read first, and the
`Total` row is the sum of the rows above it. A language present on one side
only still gets a row, with zero on the other.

A language the baseline did not contain shows `n/a` for its relative change,
since a percentage of a zero baseline would read as no change.

## Determinism

Both renderings derive from the comparison alone, so the same tree always
produces the same output. The report carries no timestamp: an unchanged tree
renders an unchanged `codezor.md`, so the file diffs cleanly between runs.

`codezor.md` lives in the repository it measures, so it counts itself. Writing
it is a change to the tree, and a run that alters the report's own line count
settles on the run after it. The summary writes nothing, so it does not move
the numbers it reports.
