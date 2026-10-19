# codezor

codezor reports what a codebase contains, how much of it this branch changed,
and what producing that code with an LLM would cost. It runs locally against a
git repository, reads the working tree as it is on disk, and writes a markdown
report.

## Usage

```shell
go run .                              # writes codezor.md for this repository
go run . -path ../other -out out.md   # another repository, another report
go run . -locomo-preset large         # price against a frontier model tier
```

| Flag             | Default      | Meaning                                                 |
| ---------------- | ------------ | ------------------------------------------------------- |
| `-path`          | `.`          | Directory inside the repository to analyze.             |
| `-out`           | `codezor.md` | Markdown report to write.                               |
| `-locomo-preset` | `medium`     | LOCOMO model tier: `large`, `medium`, `small`, `local`. |

`task run` does the same and accepts the flags after `--`.

## Architecture

| Path                | Responsibility                                                   |
| ------------------- | ---------------------------------------------------------------- |
| `main.go`           | Entry point: flags, wiring, exit code.                           |
| `internal/analyze/` | Counts both sides, measures churn, estimates LOCOMO cost.        |
| `internal/gitrepo/` | Reads the repository: baseline revision, trees, diffs.           |
| `internal/report/`  | Renders a comparison as markdown.                                |
| `dev/`              | Working records: principles, studies, plans, bugs, backlog.      |
| `taskfile/`         | PowerShell scripts behind the `Taskfile.yml` targets.            |

Counting and cost estimation both come from
[scc](https://github.com/boyter/scc), used as a library so codezor needs no
counting tool installed. `internal/analyze` is the only package that touches
scc; `internal/report` sees only codezor's own types. `internal/gitrepo` runs
the `git` command line and nothing else, so git is the one external tool
required, which a tool that measures a branch against another already implies.

## Measurements

### Baseline

The baseline is the merge base between `main` and `HEAD`, not the tip of
`main`. The delta is therefore what this branch changed: commits that landed on
`main` after the branch started do not move the numbers. On `main` itself the
baseline is `HEAD`, so the report shows uncommitted work alone.

The baseline side is the committed tree at that revision. The current side is
the files on disk, which includes staged, unstaged and untracked code, and
excludes anything `.gitignore` covers.

### LOC

LOC is code lines as scc counts them. Comment and blank lines are excluded from
every number in the report, churn included, so one unit runs through all of it.
Files scc does not recognise, and binary files, are left out of both sides.

### Churn

Churn comes from a zero context diff: each changed line is classified by scc on
the side it belongs to, and only the code lines are counted. Added and removed
are therefore comparable with the LOC table, and net churn equals the LOC
delta. The one case that breaks the equality is a tracked file that the current
`.gitignore` covers: git still diffs it, while the walk over the working tree
does not see it.

Gross churn counts a rewritten line twice, once on each side, which makes it a
measure of work done rather than of growth. The change footprint is gross churn
as a share of baseline LOC.

### LOCOMO

LOCOMO (LLM Output COst MOdel) is scc's estimate of the input and output
tokens, dollar cost, generation time and human review time needed to produce
the current code. It scales with code volume and complexity and leaves out
context reuse, test generation and debugging, so it is a ballpark figure in the
way COCOMO is for human development.

Model tiers describe a price and throughput class rather than a named model, so
they do not go stale as models are retired or renamed. An unknown tier is an
error instead of a silent fallback.

### Determinism

The report carries no timestamp: an unchanged tree renders an unchanged
`codezor.md`, so the file diffs cleanly between runs.

`codezor.md` lives in the repository it measures, so it counts itself. Writing
it is a change to the tree, and a run that alters the report's own line count
settles on the run after it.

## Links

- <https://github.com/boyter/scc>
- <https://boyter.org/posts/sloc-cloc-code-locomo-llm-output-cost-model/>
