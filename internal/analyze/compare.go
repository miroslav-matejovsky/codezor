package analyze

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/miroslav-matejovsky/codezor/internal/gitrepo"
)

// Metric is one measured quantity on each side of a comparison.
type Metric struct {
	Baseline int64
	Current  int64
}

// Delta is the change, negative when the metric fell.
func (m Metric) Delta() int64 {
	return m.Current - m.Baseline
}

// ChangePercent is Delta relative to the baseline. It is zero when the baseline
// was zero, where a percentage would mean nothing.
func (m Metric) ChangePercent() float64 {
	if m.Baseline == 0 {
		return 0
	}

	return float64(m.Delta()) / float64(m.Baseline) * 100
}

// add folds another measurement of the same quantity into m, which is how the
// per-language rows are summed into a total.
func (m *Metric) add(o Metric) {
	m.Baseline += o.Baseline
	m.Current += o.Current
}

// LanguageStats is one language's measurements on each side of a comparison. A
// language that exists on one side only still appears, with zero on the other.
type LanguageStats struct {
	Name string
	// Code is code lines, comments and blank lines excluded.
	Code Metric
	// Complexity is scc's complexity estimate: a count of the branching
	// keywords and operators its language database lists for the language. It
	// is an indicator of how hard the code is to follow, not a measured
	// cyclomatic complexity, and it is comparable across languages only
	// loosely.
	Complexity Metric
}

// Comparison is the working tree measured against a baseline revision. Total
// carries no name; it is the sum of the language rows.
type Comparison struct {
	// Root is the work tree root everything was measured from.
	Root string
	// BaselineRef is the reference the baseline was derived from: a branch for
	// Compare, HEAD for CompareHead.
	BaselineRef string
	// BaselineHash is the short hash of the baseline commit itself.
	BaselineHash string
	// BaselineIsMergeBase says whether BaselineHash is the merge base with
	// BaselineRef rather than BaselineRef itself, which is what tells a reader
	// whether committed work on this branch is part of the delta.
	BaselineIsMergeBase bool
	// Languages is one row per language, ordered by current code size.
	Languages []LanguageStats
	Total     LanguageStats
}

// Compare counts the code lines in the repository holding dir, on both sides
// of the merge base with baselineRef.
//
// The baseline is the merge base, not the tip of the branch, so commits that
// landed on baselineRef after this branch started are not reported as part of
// this change. The delta therefore covers everything this branch did: its
// commits and its uncommitted work.
func Compare(ctx context.Context, dir, baselineRef string) (Comparison, error) {
	repo, root, _, err := openRepo(ctx, dir)
	if err != nil {
		return Comparison{}, err
	}

	base, err := repo.MergeBase(ctx, baselineRef)
	if err != nil {
		return Comparison{}, err
	}

	return compare(ctx, repo, root, baselineRef, base, true)
}

// CompareHead counts the code lines in the repository holding dir against the
// commit currently checked out.
//
// The delta is the uncommitted and untracked work alone: whatever this branch
// has committed counts on both sides. That is the question the inner loop asks,
// where what matters is the change in the editor rather than the whole branch.
func CompareHead(ctx context.Context, dir string) (Comparison, error) {
	repo, root, head, err := openRepo(ctx, dir)
	if err != nil {
		return Comparison{}, err
	}

	return compare(ctx, repo, root, "HEAD", head, false)
}

// openRepo opens the repository holding dir and reports its work tree root and
// the commit checked out. A repository without commits is rejected here: there
// is no baseline to compare against.
func openRepo(ctx context.Context, dir string) (repo *gitrepo.Repo, root, head string, err error) {
	repo, err = gitrepo.Open(ctx, dir)
	if err != nil {
		return nil, "", "", err
	}

	root, err = repo.Root(ctx)
	if err != nil {
		return nil, "", "", err
	}

	head, err = repo.Head(ctx)
	if err != nil {
		return nil, "", "", fmt.Errorf("repository at %q has no commits to compare against: %w", root, err)
	}

	return repo, root, head, nil
}

// compare counts both sides against base and builds the comparison. The
// baseline side is the committed tree at base; the current side is the files on
// disk, which includes uncommitted and untracked code and excludes whatever
// .gitignore covers.
//
// Both sides are enumerated by git and counted by the same counter, so the two
// columns are comparable by construction. Everything is measured from the work
// tree root, so the result does not depend on which subdirectory codezor ran
// in.
func compare(ctx context.Context, repo *gitrepo.Repo, root, baselineRef, base string, isMergeBase bool) (Comparison, error) {
	baseShort, err := repo.ShortHash(ctx, base)
	if err != nil {
		return Comparison{}, err
	}

	baseline, err := countTree(ctx, repo, base)
	if err != nil {
		return Comparison{}, err
	}

	current, err := countWorkingTree(ctx, repo, root)
	if err != nil {
		return Comparison{}, err
	}

	languages, total := statsTable(baseline, current)

	return Comparison{
		Root:                root,
		BaselineRef:         baselineRef,
		BaselineHash:        baseShort,
		BaselineIsMergeBase: isMergeBase,
		Languages:           languages,
		Total:               total,
	}, nil
}

// tally is what one side of a comparison counted for one language.
type tally struct {
	code       int64
	complexity int64
}

// tallies is one side of a comparison: the counted totals per language.
type tallies map[string]tally

// add folds one counted file into its language's tally.
func (t tallies) add(file classified) {
	sum := t[file.language]
	sum.code += file.code
	sum.complexity += file.complexity
	t[file.language] = sum
}

// countTree counts rev's committed tree.
func countTree(ctx context.Context, repo *gitrepo.Repo, rev string) (tallies, error) {
	counted := tallies{}

	err := repo.WalkTree(ctx, rev, func(repoPath string, content []byte) error {
		if file, ok := classify(repoPath, content); ok {
			counted.add(file)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return counted, nil
}

// countWorkingTree counts the files on disk.
func countWorkingTree(ctx context.Context, repo *gitrepo.Repo, root string) (tallies, error) {
	paths, err := repo.WorkingFiles(ctx)
	if err != nil {
		return nil, err
	}

	counted := tallies{}
	for _, repoPath := range paths {
		content, present, err := readWorkingFile(root, repoPath)
		if err != nil {
			return nil, err
		}
		if !present {
			continue
		}

		if file, ok := classify(repoPath, content); ok {
			counted.add(file)
		}
	}

	return counted, nil
}

// readWorkingFile reads a file by its repository relative path. present is
// false for a path that holds no file to count: one git still has in the index
// after it was deleted on disk, and one that is a directory or a symbolic
// link, which is what a submodule and a link look like here.
func readWorkingFile(root, repoPath string) (content []byte, present bool, err error) {
	full := filepath.Join(root, filepath.FromSlash(repoPath))

	info, err := os.Lstat(full)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("inspect %q: %w", repoPath, err)
	}
	if !info.Mode().IsRegular() {
		return nil, false, nil
	}

	// The path comes from git's own listing of this repository, not from
	// anything a caller supplied.
	content, err = os.ReadFile(full)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("read %q: %w", repoPath, err)
	}

	return content, true, nil
}

// statsTable merges the baseline and current language sets, ordered by current
// code size so the largest languages read first.
func statsTable(baseline, current tallies) (rows []LanguageStats, total LanguageStats) {
	byName := make(map[string]LanguageStats, len(baseline)+len(current))
	for name, sum := range baseline {
		byName[name] = LanguageStats{
			Name:       name,
			Code:       Metric{Baseline: sum.code},
			Complexity: Metric{Baseline: sum.complexity},
		}
	}
	for name, sum := range current {
		row := byName[name]
		row.Name = name
		row.Code.Current = sum.code
		row.Complexity.Current = sum.complexity
		byName[name] = row
	}

	rows = make([]LanguageStats, 0, len(byName))
	for _, row := range byName {
		rows = append(rows, row)
		total.Code.add(row.Code)
		total.Complexity.add(row.Complexity)
	}

	slices.SortFunc(rows, func(a, b LanguageStats) int {
		if c := cmp.Compare(b.Code.Current, a.Code.Current); c != 0 {
			return c
		}
		if c := cmp.Compare(b.Code.Baseline, a.Code.Baseline); c != 0 {
			return c
		}

		return strings.Compare(a.Name, b.Name)
	})

	return rows, total
}
