# codezor

codezor counts the code in a git repository and compares the working tree
against the branch it forked from. It runs locally, reads the tree as it is on
disk, and writes a markdown report.

## Usage

```shell
go run .                              # writes codezor.md for this repository
go run . -path ../other -out out.md   # another repository, another report
```

| Flag    | Default      | Meaning                                     |
| ------- | ------------ | ------------------------------------------- |
| `-path` | `.`          | Directory inside the repository to analyze. |
| `-out`  | `codezor.md` | Markdown report to write.                   |

`task run` does the same and accepts the flags after `--`.

## Architecture

| Path                | Responsibility                                              |
| ------------------- | ----------------------------------------------------------- |
| `main.go`           | Entry point: flags, wiring, exit code.                      |
| `internal/analyze/` | Counts both sides and builds the comparison.                |
| `internal/gitrepo/` | Reads the repository: baseline revision, trees, file lists. |
| `internal/report/`  | Renders a comparison as markdown.                           |
| `dev/`              | Working records: principles, studies, plans, bugs, backlog.  |
| `taskfile/`         | PowerShell scripts behind the `Taskfile.yml` targets.       |

Counting comes from [scc](https://github.com/boyter/scc), used as a library so
codezor needs no counting tool installed. `internal/analyze` is the only
package that touches scc; `internal/report` sees only codezor's own types.
`internal/gitrepo` runs the `git` command line and nothing else, so git is the
one external tool required, which a tool that measures a branch against another
already implies.

scc is pinned to `v3`. `v4` adds git history reports that pull in `go-git` and
its ssh and crypto tail, which triples the dependency tree for features
codezor does not use. The tradeoff is that `v3` no longer receives upstream
fixes.

## Measurements

### Baseline

The baseline is the merge base between `main` and `HEAD`, not the tip of
`main`. The delta is therefore what this branch changed: commits that landed on
`main` after the branch started do not move the numbers. On `main` itself the
baseline is `HEAD`, so the report shows uncommitted work alone.

The baseline side is the committed tree at that revision. The current side is
the files on disk: tracked files, including staged and unstaged edits, plus
untracked ones, and nothing `.gitignore` covers.

Both sides are enumerated by git and counted by the same counter, so the two
columns are comparable by construction.

### LOC

LOC is code lines as scc counts them. Comment and blank lines are excluded.
Three kinds of file are counted on neither side: binaries, files whose language
scc does not recognise, and ignore files such as `.gitignore`, which steer a
walk rather than hold code.

A language the baseline did not contain shows `n/a` for its relative change,
since a percentage of a zero baseline would read as no change.

### Determinism

The report carries no timestamp: an unchanged tree renders an unchanged
`codezor.md`, so the file diffs cleanly between runs.

`codezor.md` lives in the repository it measures, so it counts itself. Writing
it is a change to the tree, and a run that alters the report's own line count
settles on the run after it.

## Links

- <https://github.com/boyter/scc>
