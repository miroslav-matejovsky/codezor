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

// Fixture files with hand counted lines and branches. Comments and blanks are
// present on purpose: they must stay out of every count below. Everything but
// b.go is straight-line code, so it contributes no complexity.
const (
	// 1 comment, 1 code, 1 blank, 1 code.
	aBaseline = "// Package a is fixture code.\npackage a\n\nconst One = 1\n"
	// The same, with the last code line replaced by a comment and a code line.
	aCurrent = "// Package a is fixture code.\npackage a\n\n// Two is a comment.\nconst Two = 2\n"
	// 7 code, 1 branch.
	bBaseline = "package b\n\nfunc Pick(n int) int {\n\tif n > 0 {\n\t\treturn 1\n\t}\n\n\treturn n\n}\n"
	// The same with a loop added: 10 code, 2 branches.
	bCurrent = "package b\n\nfunc Pick(n int) int {\n\tif n > 0 {\n\t\treturn 1\n\t}\n\tfor i := 0; i < n; i++ {\n\t\tn--\n\t}\n\n\treturn n\n}\n"
	// 1 comment, 1 code.
	dCommitted = "// Package d is fixture code.\npackage d\n"
	// 1 code.
	cUntracked = "package c\n"
)

// goStats is one Go row of a comparison, written out so the expectations below
// read as the four numbers they assert.
func goStats(codeBaseline, codeCurrent, complexityBaseline, complexityCurrent int64) analyze.LanguageStats {
	return analyze.LanguageStats{
		Name:       "Go",
		Code:       analyze.Metric{Baseline: codeBaseline, Current: codeCurrent},
		Complexity: analyze.Metric{Baseline: complexityBaseline, Current: complexityCurrent},
	}
}

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
// own, two uncommitted edits and one untracked file.
func branchFixture(t *testing.T) string {
	t.Helper()

	dir := newRepo(t)
	write(t, dir, "a.go", aBaseline)
	write(t, dir, "b.go", bBaseline)
	commit(t, dir, "baseline")

	git(t, dir, "checkout", "-b", "feature")
	write(t, dir, "d.go", dCommitted)
	commit(t, dir, "committed on the branch")

	write(t, dir, "a.go", aCurrent)
	write(t, dir, "b.go", bCurrent)
	write(t, dir, "c.go", cUntracked)

	return dir
}

// The baseline holds a.go and b.go; the current side adds the committed d.go
// and the untracked c.go, grows b.go by a branch, and counts neither the
// comments nor the blanks.
func TestCompareCountsBothSides(t *testing.T) {
	got, err := analyze.Compare(t.Context(), branchFixture(t), "main")
	require.NoError(t, err)

	require.Equal(t, "main", got.BaselineRef)
	require.NotEmpty(t, got.BaselineHash)
	require.True(t, got.BaselineIsMergeBase)

	require.Equal(t, []analyze.LanguageStats{goStats(9, 14, 1, 2)}, got.Languages)
	require.Equal(t, goStats(9, 14, 1, 2).Code, got.Total.Code)
	require.Equal(t, goStats(9, 14, 1, 2).Complexity, got.Total.Complexity)
	require.Equal(t, int64(5), got.Total.Code.Delta())
	require.Equal(t, int64(1), got.Total.Complexity.Delta())
	require.InDelta(t, 55.6, got.Total.Code.ChangePercent(), 0.1)
	require.InDelta(t, 100.0, got.Total.Complexity.ChangePercent(), 0.1)
}

// On main with nothing uncommitted the baseline is HEAD, so both sides match.
func TestCompareCleanTreeOnBaselineBranch(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", aBaseline)
	write(t, dir, "b.go", bBaseline)
	commit(t, dir, "baseline")

	got, err := analyze.Compare(t.Context(), dir, "main")
	require.NoError(t, err)

	require.Equal(t, []analyze.LanguageStats{goStats(9, 9, 1, 1)}, got.Languages)
	require.Zero(t, got.Total.Code.Delta())
	require.Zero(t, got.Total.Complexity.Delta())
}

// A file deleted on disk is still in the index, so the current side has to
// treat it as gone rather than fail reading it. Its branches go with it.
func TestCompareReportsDeletedCode(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", aBaseline)
	write(t, dir, "b.go", bBaseline)
	commit(t, dir, "baseline")

	require.NoError(t, os.Remove(filepath.Join(dir, "b.go")))

	got, err := analyze.Compare(t.Context(), dir, "main")
	require.NoError(t, err)

	require.Equal(t, []analyze.LanguageStats{goStats(9, 2, 1, 0)}, got.Languages)
	require.Equal(t, int64(-7), got.Total.Code.Delta())
	require.Equal(t, int64(-1), got.Total.Complexity.Delta())
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

	require.Equal(t, []analyze.LanguageStats{goStats(2, 2, 0, 0)}, got.Languages)
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

	require.Equal(t, []analyze.LanguageStats{goStats(2, 2, 0, 0)}, got.Languages)
}

// A language the baseline did not hold still gets a row, with a zero baseline.
func TestCompareReportsNewLanguages(t *testing.T) {
	dir := branchFixture(t)
	write(t, dir, "notes.md", "# Notes\n\nA line of prose.\n")

	got, err := analyze.Compare(t.Context(), dir, "main")
	require.NoError(t, err)

	require.Equal(t, []analyze.LanguageStats{
		goStats(9, 14, 1, 2),
		{Name: "Markdown", Code: analyze.Metric{Current: 2}},
	}, got.Languages)
	require.Zero(t, got.Languages[1].Code.ChangePercent())
}

// The baseline is the commit checked out, so the committed d.go counts on both
// sides and only the uncommitted edits and the untracked c.go move the numbers.
func TestCompareHeadCountsUncommittedWorkOnly(t *testing.T) {
	got, err := analyze.CompareHead(t.Context(), branchFixture(t))
	require.NoError(t, err)

	require.Equal(t, "HEAD", got.BaselineRef)
	require.NotEmpty(t, got.BaselineHash)
	require.False(t, got.BaselineIsMergeBase)

	require.Equal(t, []analyze.LanguageStats{goStats(10, 14, 1, 2)}, got.Languages)
	require.Equal(t, int64(4), got.Total.Code.Delta())
	require.Equal(t, int64(1), got.Total.Complexity.Delta())
}

// With nothing uncommitted the two sides match, whatever the branch holds over
// the baseline branch.
func TestCompareHeadCleanTreeHasNoDelta(t *testing.T) {
	dir := branchFixture(t)
	commit(t, dir, "everything committed")

	got, err := analyze.CompareHead(t.Context(), dir)
	require.NoError(t, err)

	require.Equal(t, []analyze.LanguageStats{goStats(14, 14, 2, 2)}, got.Languages)
	require.Zero(t, got.Total.Code.Delta())
	require.Zero(t, got.Total.Complexity.Delta())
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
