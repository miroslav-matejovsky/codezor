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

// LanguageLOC is one language's code line count on each side of a comparison.
// A language that exists on one side only still appears, with zero on the
// other.
type LanguageLOC struct {
	Name     string
	Baseline int64
	Current  int64
}

// Delta is the change in code lines, negative when code was removed.
func (l LanguageLOC) Delta() int64 {
	return l.Current - l.Baseline
}

// ChangePercent is Delta relative to the baseline. It is zero for a language
// the baseline did not contain, where a percentage would mean nothing.
func (l LanguageLOC) ChangePercent() float64 {
	if l.Baseline == 0 {
		return 0
	}

	return float64(l.Delta()) / float64(l.Baseline) * 100
}

// Comparison is the working tree measured against a baseline revision. Total
// carries no name; it is the sum of the LOC rows.
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
	LOC                 []LanguageLOC
	Total               LanguageLOC
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

	loc, total := locTable(baseline, current)

	return Comparison{
		Root:                root,
		BaselineRef:         baselineRef,
		BaselineHash:        baseShort,
		BaselineIsMergeBase: isMergeBase,
		LOC:                 loc,
		Total:               total,
	}, nil
}

// countTree counts code lines per language in rev's committed tree.
func countTree(ctx context.Context, repo *gitrepo.Repo, rev string) (map[string]int64, error) {
	loc := map[string]int64{}

	err := repo.WalkTree(ctx, rev, func(repoPath string, content []byte) error {
		if file, ok := classify(repoPath, content); ok {
			loc[file.language] += file.code
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return loc, nil
}

// countWorkingTree counts code lines per language in the files on disk.
func countWorkingTree(ctx context.Context, repo *gitrepo.Repo, root string) (map[string]int64, error) {
	paths, err := repo.WorkingFiles(ctx)
	if err != nil {
		return nil, err
	}

	loc := map[string]int64{}
	for _, repoPath := range paths {
		content, present, err := readWorkingFile(root, repoPath)
		if err != nil {
			return nil, err
		}
		if !present {
			continue
		}

		if file, ok := classify(repoPath, content); ok {
			loc[file.language] += file.code
		}
	}

	return loc, nil
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

// locTable merges the baseline and current language sets, ordered by current
// size so the largest languages read first.
func locTable(baseline, current map[string]int64) (rows []LanguageLOC, total LanguageLOC) {
	byName := make(map[string]LanguageLOC, len(baseline)+len(current))
	for name, loc := range baseline {
		byName[name] = LanguageLOC{Name: name, Baseline: loc}
	}
	for name, loc := range current {
		row := byName[name]
		row.Name = name
		row.Current = loc
		byName[name] = row
	}

	rows = make([]LanguageLOC, 0, len(byName))
	for _, row := range byName {
		rows = append(rows, row)
		total.Baseline += row.Baseline
		total.Current += row.Current
	}

	slices.SortFunc(rows, func(a, b LanguageLOC) int {
		if c := cmp.Compare(b.Current, a.Current); c != 0 {
			return c
		}
		if c := cmp.Compare(b.Baseline, a.Baseline); c != 0 {
			return c
		}

		return strings.Compare(a.Name, b.Name)
	})

	return rows, total
}
