# Measurements

Every number codezor prints or writes comes from one comparison: a committed
baseline tree against the working tree. Two quantities are reported over the
same files, LOC and complexity. This document defines both sides and both
quantities.

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

Neither rendering names the baseline it used, so which of the two applies
follows from how codezor was run.

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

## Complexity

Complexity is scc's complexity estimate: a count of the branching keywords and
operators its language database lists for the language, summed over exactly the
files LOC counts. In Go an `if` and a `for` each add one, so a function holding
both scores two.

It is an indicator of how hard the code is to follow, not a measured cyclomatic
complexity. The keyword lists differ per language, so read a language against
its own baseline rather than against another language.

Ordering is the LOC table's. A zero baseline shows `n/a` for the relative
change, for the same reason LOC does.

The report gives complexity a table of its own, and leaves out a language whose
complexity is zero on both sides: its row would be a line of zeroes and an
`n/a`, which is what a language of plain prose or data looks like. It still
counts toward the `Total`, and it keeps its LOC row, so the two tables can hold
a different number of rows.

## Renderings

The report writes a table per metric. The summary writes one, with both metrics
in every cell: the complexity, then the LOC in parentheses. Both numbers are
padded down the column, the parenthesised ones among themselves, so a column
reads as two columns of digits. The numbers are the same either way; one row
per language is what fits a terminal.

The summary leaves out only a language that is zero on both sides of both
metrics, so a language of prose keeps its row there and shows its LOC against a
complexity of zero.

## Determinism

Both renderings derive from the comparison alone, so the same tree always
produces the same output. The report carries no timestamp: an unchanged tree
renders an unchanged `codezor.md`, so the file diffs cleanly between runs.

`codezor.md` lives in the repository it measures, so it counts itself. Writing
it is a change to the tree, and a run that alters the report's own line count
settles on the run after it. The summary writes nothing, so it does not move
the numbers it reports.
