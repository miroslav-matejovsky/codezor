# codezor

codezor counts the code in a git repository and shows how big the current
change is, in lines and in complexity. It runs locally, reads the working tree
as it is on disk, and needs no service, no account, and no counting tool
installed.

Two views of the same numbers:

- a markdown report of everything the current branch changed against `main`
- a console summary of uncommitted work and the full change against `main`

## Install

```shell
go install github.com/miroslav-matejovsky/codezor@latest   # newest release
go install github.com/miroslav-matejovsky/codezor@v0.1.0   # a pinned release
```

Releases are semver git tags, so `@vX.Y.Z` pins an exact build and `@latest`
takes the newest tag. Building needs Go 1.27.3 or newer; running needs `git` on
`PATH` and nothing else.

## Usage

Run it anywhere inside the repository you want to measure.

```shell
codezor                              # writes codezor.md for this repository
codezor -summary                     # prints changes against HEAD and main
codezor -path ../other -out out.md   # another repository, another report
```

| Flag       | Default      | Meaning                                                 |
| ---------- | ------------ | ------------------------------------------------------- |
| `-path`    | `.`          | Directory inside the repository to analyze.             |
| `-out`     | `codezor.md` | Markdown report to write.                               |
| `-summary` | `false`      | Print changes against HEAD and main; write nothing.    |

Exit status is zero on success, one on failure, with the reason on stderr.

## Output

`-summary` writes nothing to disk, so it is the view for the inner loop:

```text
| Language (cplx loc) |   Delta |       Change | Delta (main) | Change (main) |
| ------------------- | ------: | -----------: | -----------: | ------------: |
| Go                  | +8 +113 | +6.3% +10.9% |      +8 +113 |  +6.3% +10.9% |
| Powershell          |  0    0 |  0.0%   0.0% |       0    0 |   0.0%   0.0% |
| Total               | +8 +122 | +5.5%  +5.5% |      +8 +122 |  +5.5%  +5.5% |
```

Every cell is the complexity with the LOC beside it, so a language is one row
and the whole change fits one screen. Both numbers are padded down the column,
the LOC ones among themselves. The `Language (cplx loc)` header identifies this order.

Without `-summary` the two metrics go to `codezor.md` as a table each, under
markdown headings and with the totals emphasised. Neither view explains itself:
the numbers are defined here and in `docs/measurements.md`, not repeated in
every report. The file carries no timestamp, so an unchanged tree rewrites the
same bytes and the report diffs cleanly between runs.

## Reading the tables

Every table orders languages by current LOC. The summary shows Language (cplx loc),
Delta, Change, Delta (main) and Change (main), with one row per language with
complexity in HEAD, the working tree or main's baseline. The report shows
Baseline, Current, Delta and Change in separate tables for each metric. Its
complexity table includes languages with complexity on either side. Totals
include all languages, including LOC from languages whose rows are hidden.

| Column     | Meaning                                                      |
| ---------- | ------------------------------------------------------------ |
| `Language` | Language of the counted files, plus a `Total` row.           |
| `Baseline` | The count in the committed baseline tree.                    |
| `Current`  | The count in the working tree as it is on disk.              |
| `Delta`    | `Current` minus `Baseline`.                                  |
| `Change`   | `Delta` relative to `Baseline`, or `n/a` for a zero baseline. |
| `Delta (main)` | Working tree minus the merge base with `main`.          |
| `Change (main)` | `Delta (main)` relative to that baseline, or `n/a` if zero. |

The baseline depends on the view. The report compares against the merge base
between `main` and `HEAD`, so the delta is everything this branch changed,
committed or not. In the summary, Delta and Change compare against `HEAD`, so
they show only uncommitted and untracked work. Delta (main) and Change (main)
use the report's merge base with `main`, including committed work as well.
Both views require a local `main` branch with a common ancestor with HEAD.

The current side is the files on disk: tracked files, with staged and unstaged
edits, plus untracked ones, and nothing `.gitignore` covers. LOC is code lines:
comments and blank lines are excluded, and so are binaries, files of no
recognised language, and ignore files such as `.gitignore`.

Complexity is scc's estimate over those same files: a count of the branching
keywords and operators its language database lists. It says how hard the code
is to follow, not how correct it is, and the keyword lists differ per language,
so read a language against its own baseline. Languages with no branches across
the compared revisions are omitted from the summary and the report's complexity
table. They still count toward the `Total` and keep their rows in the report's
LOC table.

`docs/measurements.md` gives the precise definitions.

## Documentation

| Document                                     | Contents                                               |
| -------------------------------------------- | ------------------------------------------------------ |
| [docs/measurements.md](docs/measurements.md) | What each number means, exactly.                       |
| [docs/architecture.md](docs/architecture.md) | Packages, their boundaries, and the counter behind them. |
| [docs/development.md](docs/development.md)   | Building, checks, tests, and cutting a release.        |

## License

EUPL-1.2. See [LICENSE](LICENSE).
