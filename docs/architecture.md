# Architecture

codezor is a single command over four internal packages: one reads git, one
counts, one renders, one lays out tables.

| Path                 | Responsibility                                              |
| -------------------- | ----------------------------------------------------------- |
| `main.go`            | Entry point: flags, wiring, exit code.                      |
| `internal/analyze/`  | Counts both sides and builds the comparison.                |
| `internal/gitrepo/`  | Reads the repository: baseline revision, trees, file lists. |
| `internal/report/`   | Renders a comparison as markdown or as a console summary.   |
| `internal/markdown/` | Lays out markdown tables: column widths and padding.        |
| `docs/`              | Technical documentation.                                    |
| `dev/`               | Working records: principles, studies, plans, bugs, backlog. |
| `taskfile/`          | PowerShell scripts behind the `Taskfile.yml` targets.       |

## Dependencies

`main` wires `analyze` and `report` together. The report uses the merge base with
main; the summary combines that comparison with a comparison against HEAD.
The CLI prints the summary by default, using the current directory or the
provided repository path. It writes a markdown report when `-out` is supplied;
an explicit `-summary` selects console output instead.
The summary runs the comparisons sequentially and joins their rows by language.
`analyze` reads the repository through `gitrepo` rather than
running git itself. `report` sees only codezor's own types, and hands its cells
to `markdown` to lay out. `gitrepo` and `markdown` import no component and no
third-party package: one talks to the `git` command line through the standard
library, the other works on strings.

The rules are declared in `.go-arch-lint.yml` and checked by `task arch-lint`,
so a new import that crosses a boundary fails the build rather than drifting in.

## Counting

Counting comes from [scc](https://github.com/boyter/scc), used as a library so
codezor needs no counting tool installed. `internal/analyze` is the only
package that touches scc. `internal/gitrepo` runs the `git` command line and
nothing else, so git is the one external tool required, which a tool that
measures a branch against another already implies.

Two numbers come out of each file: code lines and scc's complexity estimate, a
count of the branching keywords and operators in scc's language database.

scc keeps its language database in package level state, so the counter is
guarded by a one time load and the comparisons must not be run concurrently.

scc is pinned to `v4`. It adds git history reports that pull in `go-git` and
its ssh and crypto tail, which roughly doubles the indirect dependency list for
features codezor does not use. The cost is accepted because `v3`, which needed
none of it, no longer receives upstream fixes.

## Reading the tree

The baseline side is streamed from one `git archive` invocation and counted in
memory, so a whole committed tree costs a single git process and nothing is
written to disk. The current side is listed with `git ls-files` and read from
disk. Symbolic links, directories and submodules are skipped on both sides:
none of them hold code of their own.

Design notes and the principles behind these choices live in
[dev/](../dev/README.md).
