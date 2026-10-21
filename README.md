# codezor

codezor counts the code in a git repository and shows how big the current
change is, in lines and in complexity. It runs locally, reads the working tree
as it is on disk, and needs no service, no account, and no counting tool
installed.

Two views of the same numbers:

- a markdown report of everything the current branch changed against `main`
- a console summary of the uncommitted and untracked work alone

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
codezor -summary                     # prints the uncommitted change, writes nothing
codezor -path ../other -out out.md   # another repository, another report
```

| Flag       | Default      | Meaning                                                 |
| ---------- | ------------ | ------------------------------------------------------- |
| `-path`    | `.`          | Directory inside the repository to analyze.             |
| `-out`     | `codezor.md` | Markdown report to write.                               |
| `-summary` | `false`      | Print the table for uncommitted work and write nothing. |

Exit status is zero on success, one on failure, with the reason on stderr.

## Output

`-summary` writes nothing to disk, so it is the view for the inner loop:

```text
Complexity (LOC)

| Language   |    Baseline |     Current |     Delta |         Change |
| ---------- | ----------: | ----------: | --------: | -------------: |
| Go         | 126 (1,041) | 134 (1,154) | +8 (+113) | +6.3% (+10.9%) |
| Markdown   |   0 (  504) |   0 (  512) |  0 (  +8) |   n/a ( +1.6%) |
| YAML       |   0 (  221) |   0 (  222) |  0 (  +1) |   n/a ( +0.5%) |
| License    |   0 (  216) |   0 (  216) |  0 (   0) |   n/a (  0.0%) |
| JSON       |   0 (  113) |   0 (  113) |  0 (   0) |   n/a (  0.0%) |
| Powershell |  20 (  111) |  20 (  111) |  0 (   0) |  0.0% (  0.0%) |
| Total      | 146 (2,206) | 154 (2,328) | +8 (+122) | +5.5% ( +5.5%) |
```

Every cell is the complexity with the LOC in parentheses, so a language is one
row and the whole change fits one screen. Both numbers are padded down the
column, the parenthesised ones among themselves.

Without `-summary` the two metrics go to `codezor.md` as a table each, under
markdown headings and with the totals emphasised. Neither view explains itself:
the numbers are defined here and in `docs/measurements.md`, not repeated in
every report. The file carries no timestamp, so an unchanged tree rewrites the
same bytes and the report diffs cleanly between runs.

## Reading the tables

Every table has the same columns and the same ordering, by current LOC. The
summary holds one row per language, dropping only a language absent from both
sides. The report splits the metrics, and its complexity table leaves out any
language with no branches on either side, so it is the shorter of its two.

| Column     | Meaning                                                      |
| ---------- | ------------------------------------------------------------ |
| `Language` | Language of the counted files, plus a `Total` row.           |
| `Baseline` | The count in the committed baseline tree.                    |
| `Current`  | The count in the working tree as it is on disk.              |
| `Delta`    | `Current` minus `Baseline`.                                  |
| `Change`   | `Delta` relative to `Baseline`, or `n/a` for a zero baseline. |

The baseline depends on the view. The report compares against the merge base
between `main` and `HEAD`, so the delta is everything this branch changed,
committed or not. `-summary` compares against `HEAD`, so committed work counts
on both sides and the delta is only what is still uncommitted.

The current side is the files on disk: tracked files, with staged and unstaged
edits, plus untracked ones, and nothing `.gitignore` covers. LOC is code lines:
comments and blank lines are excluded, and so are binaries, files of no
recognised language, and ignore files such as `.gitignore`.

Complexity is scc's estimate over those same files: a count of the branching
keywords and operators its language database lists. It says how hard the code
is to follow, not how correct it is, and the keyword lists differ per language,
so read a language against its own baseline. A language with no branches reads
as a zero and an `n/a`, and in the report's complexity table it has no row at
all; either way it still counts toward the `Total`.

`docs/measurements.md` gives the precise definitions.

## Documentation

| Document                                     | Contents                                               |
| -------------------------------------------- | ------------------------------------------------------ |
| [docs/measurements.md](docs/measurements.md) | What each number means, exactly.                       |
| [docs/architecture.md](docs/architecture.md) | Packages, their boundaries, and the counter behind them. |
| [docs/development.md](docs/development.md)   | Building, checks, tests, and cutting a release.        |

## License

EUPL-1.2. See [LICENSE](LICENSE).
