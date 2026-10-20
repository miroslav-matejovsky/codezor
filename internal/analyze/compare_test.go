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

// The baseline holds a.go alone; the current side adds the committed d.go and
// the untracked c.go, and counts neither the comments nor the blanks.
func TestCompareCountsBothSides(t *testing.T) {
	got, err := analyze.Compare(t.Context(), branchFixture(t), "main")
	require.NoError(t, err)

	require.Equal(t, "main", got.BaselineRef)
	require.NotEmpty(t, got.BaselineHash)
	require.True(t, got.BaselineIsMergeBase)

	require.Equal(t, []analyze.LanguageLOC{{Name: "Go", Baseline: 2, Current: 4}}, got.LOC)
	require.Equal(t, analyze.LanguageLOC{Baseline: 2, Current: 4}, got.Total)
	require.Equal(t, int64(2), got.Total.Delta())
	require.InDelta(t, 100.0, got.Total.ChangePercent(), 0.001)
}

// On main with nothing uncommitted the baseline is HEAD, so both sides match.
func TestCompareCleanTreeOnBaselineBranch(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", aBaseline)
	commit(t, dir, "baseline")

	got, err := analyze.Compare(t.Context(), dir, "main")
	require.NoError(t, err)

	require.Equal(t, []analyze.LanguageLOC{{Name: "Go", Baseline: 2, Current: 2}}, got.LOC)
	require.Zero(t, got.Total.Delta())
}

// A file deleted on disk is still in the index, so the current side has to
// treat it as gone rather than fail reading it.
func TestCompareReportsDeletedCode(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", aBaseline)
	write(t, dir, "d.go", dCommitted)
	commit(t, dir, "baseline")

	require.NoError(t, os.Remove(filepath.Join(dir, "d.go")))

	got, err := analyze.Compare(t.Context(), dir, "main")
	require.NoError(t, err)

	require.Equal(t, []analyze.LanguageLOC{{Name: "Go", Baseline: 3, Current: 2}}, got.LOC)
	require.Equal(t, int64(-1), got.Total.Delta())
}

// Files git ignores are not part of the current state of the code.
func TestCompareSkipsIgnoredFiles(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", aBaseline)
	write(t, dir, ".gitignore", "generated.go\n")
	commit(t, dir, "baseline")

	write(t, dir, "generated.go", "package generated\n")

	got, err := analyze.Compare(t.Context(), dir, "main")
	require.NoError(t, err)

	require.Equal(t, []analyze.LanguageLOC{{Name: "Go", Baseline: 2, Current: 2}}, got.LOC)
}

// scc leaves ignore files out of a count, so neither side may report the
// .gitignore above as code.
func TestCompareExcludesIgnoreFilesThemselves(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", aBaseline)
	write(t, dir, ".gitignore", "bin/\n*.tmp\n")
	commit(t, dir, "baseline")

	got, err := analyze.Compare(t.Context(), dir, "main")
	require.NoError(t, err)

	require.Equal(t, []analyze.LanguageLOC{{Name: "Go", Baseline: 2, Current: 2}}, got.LOC)
}

// A language the baseline did not hold still gets a row, with a zero baseline.
func TestCompareReportsNewLanguages(t *testing.T) {
	dir := branchFixture(t)
	write(t, dir, "notes.md", "# Notes\n\nA line of prose.\n")

	got, err := analyze.Compare(t.Context(), dir, "main")
	require.NoError(t, err)

	require.Equal(t, []analyze.LanguageLOC{
		{Name: "Go", Baseline: 2, Current: 4},
		{Name: "Markdown", Baseline: 0, Current: 2},
	}, got.LOC)
	require.Zero(t, got.LOC[1].ChangePercent())
}

// The baseline is the commit checked out, so the committed d.go counts on both
// sides and only the uncommitted edit to a.go and the untracked c.go move the
// numbers.
func TestCompareHeadCountsUncommittedWorkOnly(t *testing.T) {
	got, err := analyze.CompareHead(t.Context(), branchFixture(t))
	require.NoError(t, err)

	require.Equal(t, "HEAD", got.BaselineRef)
	require.NotEmpty(t, got.BaselineHash)
	require.False(t, got.BaselineIsMergeBase)

	require.Equal(t, []analyze.LanguageLOC{{Name: "Go", Baseline: 3, Current: 4}}, got.LOC)
	require.Equal(t, int64(1), got.Total.Delta())
}

// With nothing uncommitted the two sides match, whatever the branch holds over
// the baseline branch.
func TestCompareHeadCleanTreeHasNoDelta(t *testing.T) {
	dir := branchFixture(t)
	commit(t, dir, "everything committed")

	got, err := analyze.CompareHead(t.Context(), dir)
	require.NoError(t, err)

	require.Equal(t, []analyze.LanguageLOC{{Name: "Go", Baseline: 4, Current: 4}}, got.LOC)
	require.Zero(t, got.Total.Delta())
}

func TestCompareHeadRejectsNonRepository(t *testing.T) {
	_, err := analyze.CompareHead(t.Context(), t.TempDir())
	require.ErrorContains(t, err, "not a git repository")
}

func TestCompareHeadRejectsRepositoryWithoutCommits(t *testing.T) {
	_, err := analyze.CompareHead(t.Context(), newRepo(t))
	require.ErrorContains(t, err, "no commits")
}

func TestCompareRejectsNonRepository(t *testing.T) {
	_, err := analyze.Compare(t.Context(), t.TempDir(), "main")
	require.ErrorContains(t, err, "not a git repository")
}

func TestCompareRejectsMissingBaselineBranch(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", aBaseline)
	commit(t, dir, "baseline")

	_, err := analyze.Compare(t.Context(), dir, "trunk")
	require.ErrorContains(t, err, `"trunk" not found`)
}

func TestCompareRejectsRepositoryWithoutCommits(t *testing.T) {
	_, err := analyze.Compare(t.Context(), newRepo(t), "main")
	require.ErrorContains(t, err, "no commits")
}
