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

// Churn is the code lines a change added and removed. Comment and blank lines
// are excluded, so churn is measured in the same unit as the LOC counts.
type Churn struct {
	Added   int64
	Removed int64
}

// Gross is the total code lines touched. A rewritten line counts twice, once
// on each side, which is what makes gross churn a measure of work done rather
// than of growth.
func (c Churn) Gross() int64 {
	return c.Added + c.Removed
}

// Net is the growth in code lines.
func (c Churn) Net() int64 {
	return c.Added - c.Removed
}

// Footprint is Gross as a percentage of baseline: how large the change is
// against the code it started from. It is zero when the baseline holds no
// code.
func (c Churn) Footprint(baseline int64) float64 {
	if baseline == 0 {
		return 0
	}

	return float64(c.Gross()) / float64(baseline) * 100
}

// Comparison is the working tree measured against a baseline revision.
// TotalLOC carries no name; it is the sum of the LOC rows.
type Comparison struct {
	Root         string
	BaselineRef  string
	BaselineHash string
	LOC          []LanguageLOC
	TotalLOC     LanguageLOC
	Current      Churn
	LatestCommit Churn
	Locomo       Locomo
}

// Compare measures the working tree of the repository holding dir against the
// merge base with baselineRef.
//
// The baseline is the merge base, not the tip of the branch, so commits that
// landed on baselineRef after this branch started are not reported as part of
// this change. Both sides are counted by scc: the baseline from the committed
// tree, the current side from the files on disk, which includes uncommitted
// and untracked code.
//
// Everything is measured from the work tree root, so the report does not
// depend on which subdirectory codezor ran in.
func Compare(ctx context.Context, dir, baselineRef, locomoPreset string) (Comparison, error) {
	repo, err := gitrepo.Open(ctx, dir)
	if err != nil {
		return Comparison{}, err
	}

	root, err := repo.Root(ctx)
	if err != nil {
		return Comparison{}, err
	}

	if _, err := repo.Head(ctx); err != nil {
		return Comparison{}, fmt.Errorf("repository at %q has no commits to compare against: %w", root, err)
	}

	base, err := repo.MergeBase(ctx, baselineRef)
	if err != nil {
		return Comparison{}, err
	}

	baseShort, err := repo.ShortHash(ctx, base)
	if err != nil {
		return Comparison{}, err
	}

	current, err := Run(root, locomoPreset)
	if err != nil {
		return Comparison{}, err
	}

	// The diff names the baseline files whose line classification the churn
	// count needs, so the single pass over the baseline tree below can keep
	// just those.
	changes, err := repo.Diff(ctx, base, "")
	if err != nil {
		return Comparison{}, err
	}

	baselineLOC, baselineLines, err := countTree(ctx, repo, base, removedPaths(changes))
	if err != nil {
		return Comparison{}, err
	}

	working, err := workingChurn(ctx, repo, root, changes, baselineLines)
	if err != nil {
		return Comparison{}, err
	}

	latest, err := latestCommitChurn(ctx, repo)
	if err != nil {
		return Comparison{}, err
	}

	loc, total := locTable(baselineLOC, current.Languages)

	return Comparison{
		Root:         root,
		BaselineRef:  baselineRef,
		BaselineHash: baseShort,
		LOC:          loc,
		TotalLOC:     total,
		Current:      working,
		LatestCommit: latest,
		Locomo:       current.Locomo,
	}, nil
}

// removedPaths is the set of baseline paths whose lines the churn count has to
// classify, which is every path the diff removes lines from.
func removedPaths(changes []gitrepo.FileDiff) map[string]bool {
	keep := make(map[string]bool, len(changes))
	for _, change := range changes {
		if !change.Binary && change.OldPath != "" && len(change.Removed) > 0 {
			keep[change.OldPath] = true
		}
	}

	return keep
}

// countTree counts code lines per language in rev's tree, keeping the line
// classification of the paths in keep for the churn count.
func countTree(
	ctx context.Context,
	repo *gitrepo.Repo,
	rev string,
	keep map[string]bool,
) (loc map[string]int64, kept map[string]lineTypes, err error) {
	loc = map[string]int64{}
	kept = make(map[string]lineTypes, len(keep))

	err = repo.WalkTree(ctx, rev, func(repoPath string, content []byte) error {
		file, ok := classify(repoPath, content)
		if !ok {
			return nil
		}

		loc[file.language] += file.code
		if keep[repoPath] {
			kept[repoPath] = file.lines
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return loc, kept, nil
}

// workingChurn counts the code lines the working tree adds to and removes from
// the baseline, including the files git does not track yet.
//
// A file scc does not count, such as a binary or an unrecognised extension,
// contributes nothing, which keeps churn consistent with the LOC counts.
func workingChurn(
	ctx context.Context,
	repo *gitrepo.Repo,
	root string,
	changes []gitrepo.FileDiff,
	baseline map[string]lineTypes,
) (Churn, error) {
	var churn Churn

	for _, change := range changes {
		if change.Binary {
			continue
		}

		churn.Removed += codeLines(baseline[change.OldPath], change.Removed)

		if change.NewPath == "" || len(change.Added) == 0 {
			continue
		}

		content, err := readWorkingFile(root, change.NewPath)
		if err != nil {
			return Churn{}, err
		}

		if file, ok := classify(change.NewPath, content); ok {
			churn.Added += codeLines(file.lines, change.Added)
		}
	}

	untracked, err := repo.Untracked(ctx)
	if err != nil {
		return Churn{}, err
	}

	for _, repoPath := range untracked {
		content, err := readWorkingFile(root, repoPath)
		if errors.Is(err, fs.ErrNotExist) {
			// Listed a moment ago and already gone, so it is not part of the
			// tree being measured.
			continue
		}
		if err != nil {
			return Churn{}, err
		}

		if file, ok := classify(repoPath, content); ok {
			churn.Added += file.code
		}
	}

	return churn, nil
}

// latestCommitChurn counts the code lines the newest commit added and removed.
// The first commit in a repository is compared against the empty tree, so all
// of it counts as added.
func latestCommitChurn(ctx context.Context, repo *gitrepo.Repo) (Churn, error) {
	parent, err := repo.Parent(ctx, "HEAD")
	if err != nil {
		return Churn{}, err
	}

	changes, err := repo.Diff(ctx, parent, "HEAD")
	if err != nil {
		return Churn{}, err
	}

	var churn Churn
	for _, change := range changes {
		if change.Binary {
			continue
		}

		if change.OldPath != "" && len(change.Removed) > 0 {
			removed, err := blobCodeLines(ctx, repo, parent, change.OldPath, change.Removed)
			if err != nil {
				return Churn{}, err
			}
			churn.Removed += removed
		}

		if change.NewPath != "" && len(change.Added) > 0 {
			added, err := blobCodeLines(ctx, repo, "HEAD", change.NewPath, change.Added)
			if err != nil {
				return Churn{}, err
			}
			churn.Added += added
		}
	}

	return churn, nil
}

// blobCodeLines counts the lines inside ranges that are code in repoPath as of
// rev.
func blobCodeLines(
	ctx context.Context,
	repo *gitrepo.Repo,
	rev, repoPath string,
	ranges []gitrepo.LineRange,
) (int64, error) {
	content, err := repo.Blob(ctx, rev, repoPath)
	if err != nil {
		return 0, err
	}

	file, ok := classify(repoPath, content)
	if !ok {
		return 0, nil
	}

	return codeLines(file.lines, ranges), nil
}

// codeLines counts the lines inside ranges that scc classified as code.
func codeLines(types lineTypes, ranges []gitrepo.LineRange) int64 {
	var n int64
	for _, r := range ranges {
		for line := r.Start; line < r.Start+r.Count; line++ {
			if types.isCode(line) {
				n++
			}
		}
	}

	return n
}

// readWorkingFile reads a file by its repository relative path.
func readWorkingFile(root, repoPath string) ([]byte, error) {
	full := filepath.Join(root, filepath.FromSlash(repoPath))

	// The path comes from git's own listing of this repository, not from
	// anything a caller supplied.
	content, err := os.ReadFile(full)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", repoPath, err)
	}

	return content, nil
}

// locTable merges the baseline and current language sets, ordered by current
// size so the largest languages read first.
func locTable(baseline map[string]int64, current []Language) (rows []LanguageLOC, total LanguageLOC) {
	byName := make(map[string]LanguageLOC, len(baseline)+len(current))
	for name, loc := range baseline {
		byName[name] = LanguageLOC{Name: name, Baseline: loc}
	}
	for _, language := range current {
		row := byName[language.Name]
		row.Name = language.Name
		row.Current = language.Code
		byName[language.Name] = row
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
