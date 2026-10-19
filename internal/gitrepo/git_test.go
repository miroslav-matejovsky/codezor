package gitrepo

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// gitConfig pins identity, signing and line endings so the fixtures behave the
// same on every machine. It is passed per command rather than written with
// `git config`, which would cost a process per setting.
var gitConfig = []string{
	"-c", "user.name=codezor test",
	"-c", "user.email=codezor@example.invalid",
	"-c", "commit.gpgsign=false",
	"-c", "core.autocrlf=false",
}

// newRepo builds an empty repository on a main branch.
func newRepo(t *testing.T) (repo *Repo, dir string) {
	t.Helper()

	dir = t.TempDir()
	git(t, dir, "init", "-b", "main")

	repo, err := Open(t.Context(), dir)
	require.NoError(t, err)

	return repo, dir
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

	path := filepath.Join(dir, filepath.FromSlash(name))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
}

func commit(t *testing.T, dir, message string) {
	t.Helper()

	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-m", message)
}

func TestOpenRejectsDirectoryOutsideRepository(t *testing.T) {
	_, err := Open(t.Context(), t.TempDir())
	require.ErrorContains(t, err, "not a git repository")
}

// The baseline must be where the branch left main, not main as it stands now,
// so commits pushed to main afterwards stay out of the measured change.
func TestMergeBaseIgnoresLaterMainCommits(t *testing.T) {
	repo, dir := newRepo(t)

	write(t, dir, "a.go", "package a\n")
	commit(t, dir, "first")
	forkPoint, err := repo.Head(t.Context())
	require.NoError(t, err)

	git(t, dir, "checkout", "-b", "feature")
	write(t, dir, "b.go", "package b\n")
	commit(t, dir, "on feature")

	git(t, dir, "checkout", "main")
	write(t, dir, "c.go", "package c\n")
	commit(t, dir, "on main after the fork")

	git(t, dir, "checkout", "feature")

	base, err := repo.MergeBase(t.Context(), "main")
	require.NoError(t, err)
	require.Equal(t, forkPoint, base)
}

func TestMergeBaseRejectsMissingBranch(t *testing.T) {
	repo, dir := newRepo(t)
	write(t, dir, "a.go", "package a\n")
	commit(t, dir, "first")

	_, err := repo.MergeBase(t.Context(), "trunk")
	require.ErrorContains(t, err, `"trunk" not found`)
}

func TestWalkTreeStreamsTrackedFilesOnly(t *testing.T) {
	repo, dir := newRepo(t)

	write(t, dir, "a.go", "package a\n")
	write(t, dir, "pkg/b.go", "package pkg\n")
	commit(t, dir, "first")
	write(t, dir, "untracked.go", "package untracked\n")

	head, err := repo.Head(t.Context())
	require.NoError(t, err)

	seen := map[string]string{}
	require.NoError(t, repo.WalkTree(t.Context(), head, func(path string, content []byte) error {
		seen[path] = string(content)

		return nil
	}))

	require.Equal(t, map[string]string{
		"a.go":     "package a\n",
		"pkg/b.go": "package pkg\n",
	}, seen)
}

func TestWalkTreeReportsUnknownRevision(t *testing.T) {
	repo, dir := newRepo(t)
	write(t, dir, "a.go", "package a\n")
	commit(t, dir, "first")

	err := repo.WalkTree(t.Context(), "no-such-rev", func(string, []byte) error { return nil })
	require.Error(t, err)
}

// The working tree is the tracked files plus the untracked ones, with
// everything .gitignore covers left out.
func TestWorkingFilesListsTrackedAndUntracked(t *testing.T) {
	repo, dir := newRepo(t)

	write(t, dir, "a.go", "package a\n")
	write(t, dir, ".gitignore", "ignored.go\n")
	commit(t, dir, "first")

	write(t, dir, "new.go", "package new\n")
	write(t, dir, "ignored.go", "package ignored\n")

	files, err := repo.WorkingFiles(t.Context())
	require.NoError(t, err)
	require.ElementsMatch(t, []string{".gitignore", "a.go", "new.go"}, files)
}

// A tracked file deleted on disk is still in the index, so it is still listed
// and the caller is the one that finds it missing.
func TestWorkingFilesStillListsDeletedTrackedFiles(t *testing.T) {
	repo, dir := newRepo(t)

	write(t, dir, "a.go", "package a\n")
	write(t, dir, "gone.go", "package gone\n")
	commit(t, dir, "first")
	require.NoError(t, os.Remove(filepath.Join(dir, "gone.go")))

	files, err := repo.WorkingFiles(t.Context())
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"a.go", "gone.go"}, files)
}

func TestHeadFailsWithoutCommits(t *testing.T) {
	repo, _ := newRepo(t)

	_, err := repo.Head(t.Context())
	require.Error(t, err)
}
