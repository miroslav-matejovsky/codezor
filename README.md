# codezor

codezor reports what a codebase contains and what producing that code with an
LLM would cost. It runs locally, reads the working tree as it is on disk, and
writes a markdown report.

## Usage

```shell
go run .                              # writes codezor.md for this repository
go run . -path ../other -out out.md   # another tree, another report
go run . -locomo-preset large         # price against a frontier model tier
```

| Flag             | Default      | Meaning                                                |
| ---------------- | ------------ | ------------------------------------------------------ |
| `-path`          | `.`          | Directory to count.                                    |
| `-out`           | `codezor.md` | Markdown report to write.                              |
| `-locomo-preset` | `medium`     | LOCOMO model tier: `large`, `medium`, `small`, `local`. |

`task run` does the same and accepts the flags after `--`.

## Architecture

| Path                | Responsibility                                                       |
| ------------------- | -------------------------------------------------------------------- |
| `main.go`           | Entry point: flags, wiring, exit code.                               |
| `internal/analyze/` | Counts a tree with scc and estimates its LOCOMO cost.                |
| `internal/report/`  | Renders an analysis as markdown.                                     |
| `dev/`              | Working records: principles, studies, plans, bugs, backlog.           |
| `taskfile/`         | PowerShell scripts behind the `Taskfile.yml` targets.                |

Counting and cost estimation both come from
[scc](https://github.com/boyter/scc), used as a library so codezor stays a
single binary with no external tool to install. `internal/analyze` is the only
package that touches scc; `internal/report` sees only codezor's own types.

### Measurements

- **Counts** per language: files, lines, code, comments, blanks, and
  complexity, where complexity is the number of branching tokens scc
  recognised.
- **LOCOMO** (LLM Output COst MOdel): scc's estimate of the input and output
  tokens, dollar cost, generation time, and human review time needed to produce
  the counted code. It scales with code volume and complexity and leaves out
  context reuse, test generation, and debugging, so it is a ballpark figure in
  the way COCOMO is for human development.

Model tiers describe a price and throughput class rather than a named model, so
they do not go stale as models are retired or renamed. An unknown tier is an
error instead of a silent fallback.

The report carries no timestamp: an unchanged tree renders an unchanged
`codezor.md`, so the file diffs cleanly between runs.

## Links

- <https://github.com/boyter/scc>
- <https://boyter.org/posts/sloc-cloc-code-locomo-llm-output-cost-model/>
