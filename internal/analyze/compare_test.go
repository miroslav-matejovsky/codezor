package analyze_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/miroslav-matejovsky/codezor/internal/analyze"
	"github.com/stretchr/testify/require"
)

// Fixture files with hand counted lines. Comments and blanks are present on
// purpose: they must stay out of every count below.
const (
	// 1 comment, 1 code, 1 blank, 1 code.
	aBaseline = "// Package a is fixture code.\npackage a\n\nconst One = 1\n"
	// The same, with the last code line replaced by a comment and a code line.
	aCurrent = "// Package a is fixture code.\npackage a\n\n// Two is a comment.\nconst Two = 2\n"
	// 1 comment, 1 code.
	dCommitted = "// Package d is fixture code.\npackage d\n"
	// 1 code.
	cUntracked = "package c\n"
)

// gitConfig pins identity, signing and line endings so the fixtures count the
// same on every machine. It is passed per command rather than written with
// `git config`, which would cost a process per setting.
var gitConfig = []string{
	"-c", "user.name=codezor test",
	"-c", "user.email=codezor@example.invalid",
	"-c", "commit.gpgsign=false",
	"-c", "core.autocrlf=false",
}

// newRepo builds an empty repository on a main branch.
func newRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	git(t, dir, "init", "-b", "main")

	return dir
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), "git", slices.Concat(gitConfig, args)...) //nolint:gosec // test fixture: arguments are literals from this file
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %s: %s", strings.Join(args, " "), out)
}

func write(t *testing.T, dir, name, content string) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
}

func commit(t *testing.T, dir, message string) {
	t.Helper()

	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-m", message)
}

// branchFixture is a repository on a feature branch with one commit of its
// own, one uncommitted edit and one untracked file.
func branchFixture(t *testing.T) string {
	t.Helper()

	dir := newRepo(t)
	write(t, dir, "a.go", aBaseline)
	commit(t, dir, "baseline")

	git(t, dir, "checkout", "-b", "feature")
	write(t, dir, "d.go", dCommitted)
	commit(t, dir, "committed on the branch")

	write(t, dir, "a.go", aCurrent)
	write(t, dir, "c.go", cUntracked)

	return dir
}

func TestCompareCountsBothSides(t *testing.T) {
	got, err := analyze.Compare(t.Context(), branchFixture(t), "main", "medium")
	require.NoError(t, err)

	require.Equal(t, "main", got.BaselineRef)
	require.NotEmpty(t, got.BaselineHash)

	// Baseline holds a.go alone, the working tree adds d.go and c.go.
	require.Equal(t, []analyze.LanguageLOC{{Name: "Go", Baseline: 2, Current: 4}}, got.LOC)
	require.Equal(t, analyze.LanguageLOC{Baseline: 2, Current: 4}, got.TotalLOC)
	require.Equal(t, int64(2), got.TotalLOC.Delta())
	require.InDelta(t, 100.0, got.TotalLOC.ChangePercent(), 0.001)
}

// Churn counts code lines only: the comment that replaced a code line in a.go
// is not an addition, and the comment in d.go is not either.
func TestCompareChurnExcludesCommentsAndBlanks(t *testing.T) {
	got, err := analyze.Compare(t.Context(), branchFixture(t), "main", "medium")
	require.NoError(t, err)

	require.Equal(t, analyze.Churn{Added: 3, Removed: 1}, got.Current)
	require.Equal(t, int64(4), got.Current.Gross())
	require.InDelta(t, 200.0, got.Current.Footprint(got.TotalLOC.Baseline), 0.001)

	require.Equal(t, analyze.Churn{Added: 1}, got.LatestCommit)
}

// Net churn and the LOC delta measure the same change in the same unit, so
// they have to agree.
func TestCompareNetChurnMatchesLOCDelta(t *testing.T) {
	got, err := analyze.Compare(t.Context(), branchFixture(t), "main", "medium")
	require.NoError(t, err)

	require.Equal(t, got.TotalLOC.Delta(), got.Current.Net())
}

// On main with nothing uncommitted the baseline is HEAD, so the tree matches
// the baseline and only the commit itself shows any churn.
func TestCompareCleanTreeOnBaselineBranch(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", aBaseline)
	commit(t, dir, "baseline")

	got, err := analyze.Compare(t.Context(), dir, "main", "medium")
	require.NoError(t, err)

	require.Equal(t, []analyze.LanguageLOC{{Name: "Go", Baseline: 2, Current: 2}}, got.LOC)
	require.Equal(t, analyze.Churn{}, got.Current)
	require.Zero(t, got.TotalLOC.Delta())

	// The first commit is compared against the empty tree, so all of it is new.
	require.Equal(t, analyze.Churn{Added: 2}, got.LatestCommit)
}

func TestCompareReportsDeletedCode(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", aBaseline)
	write(t, dir, "d.go", dCommitted)
	commit(t, dir, "baseline")

	require.NoError(t, os.Remove(filepath.Join(dir, "d.go")))

	got, err := analyze.Compare(t.Context(), dir, "main", "medium")
	require.NoError(t, err)

	require.Equal(t, analyze.Churn{Removed: 1}, got.Current)
	require.Equal(t, int64(-1), got.TotalLOC.Delta())
	require.Equal(t, got.TotalLOC.Delta(), got.Current.Net())
}

// scc leaves ignore files out of a directory walk, so the baseline side has to
// leave them out too: otherwise an unchanged .gitignore reads as deleted code
// and the LOC delta stops matching the churn.
func TestCompareExcludesIgnoreFiles(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", aBaseline)
	write(t, dir, ".gitignore", "bin/\n*.tmp\n")
	commit(t, dir, "baseline")

	got, err := analyze.Compare(t.Context(), dir, "main", "medium")
	require.NoError(t, err)

	require.Equal(t, []analyze.LanguageLOC{{Name: "Go", Baseline: 2, Current: 2}}, got.LOC)
	require.Equal(t, analyze.Churn{Added: 2}, got.LatestCommit)
	require.Equal(t, got.TotalLOC.Delta(), got.Current.Net())
}

func TestCompareRejectsNonRepository(t *testing.T) {
	_, err := analyze.Compare(t.Context(), t.TempDir(), "main", "medium")
	require.ErrorContains(t, err, "not a git repository")
}

func TestCompareRejectsMissingBaselineBranch(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", aBaseline)
	commit(t, dir, "baseline")

	_, err := analyze.Compare(t.Context(), dir, "trunk", "medium")
	require.ErrorContains(t, err, `"trunk" not found`)
}

func TestCompareRejectsRepositoryWithoutCommits(t *testing.T) {
	_, err := analyze.Compare(t.Context(), newRepo(t), "main", "medium")
	require.ErrorContains(t, err, "no commits")
}

func TestCompareRejectsUnknownPreset(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", aBaseline)
	commit(t, dir, "baseline")

	_, err := analyze.Compare(t.Context(), dir, "main", "enormous")
	require.ErrorContains(t, err, "enormous")
}
