# Development

## Prerequisites

- Go 1.27.2 or newer, as `go.mod` requires.
- `git` on `PATH`: the analysis runs it, and so do the tests.
- [Task](https://taskfile.dev) for the targets below, and PowerShell 7 (`pwsh`)
  for the scripts in `taskfile/`.
- `task go-tools` installs the Go tools the targets use: `govulncheck`,
  `deadcode`, `gotestsum` and `go-arch-lint`. `golangci-lint` is installed
  separately, following its own instructions.

## Running it

```shell
task run                    # prints the summary for this repository
task run -- ../other        # arguments are passed after --
task run -- -out codezor.md  # writes a markdown report
task summary                # prints changes against HEAD and main, writes nothing
task file                   # writes this repository's report to codezor.md
```

## Checks

`task all` is the gate before any change is considered complete. It runs, in
order: `clean`, `tidy`, `fmt`, `vet`, `deadcode`, `arch-lint`, `lint`, `test`.
Each target also runs on its own.

| Target      | Checks                                                      |
| ----------- | ----------------------------------------------------------- |
| `vet`       | `go vet` over all packages.                                 |
| `deadcode`  | Unreachable functions, allowlisted symbols aside.           |
| `arch-lint` | The package boundaries in `.go-arch-lint.yml`.              |
| `lint`      | `golangci-lint` with the linters in `.golangci.yml`.        |
| `test`      | Unit tests through `gotestsum`.                             |

Test output is saved under `.test-results/`, which is not tracked. The git
fixtures in `internal/analyze` and `internal/gitrepo` build real repositories
in temporary directories, so the test run takes tens of seconds.

## Dependencies

```shell
task deps      # direct dependencies with available updates
task tidy      # go mod tidy
task vendor    # re-sync the vendor/ tree after a dependency change
```

Nothing in `vendor/` may be modified: third-party code is read-only.

## Releasing

A release is a semver git tag on `main`, which is all the Go module proxy needs.
GitHub releases are not used.

```shell
task release -- 0.1.0   # or v0.1.0
```

The script fails fast unless the version is valid semver, the current branch is
`main`, the working tree is clean, local `main` matches `origin/main`, and the
tag exists neither locally nor on the remote. It then creates an annotated tag
on `HEAD` and pushes it, which is what makes `go install ...@v0.1.0` resolve.
