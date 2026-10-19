package gitrepo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// newRepo builds a repository with a fixed identity and fixed line endings so
// the fixtures behave the same on every machine.
func newRepo(t *testing.T) (repo *Repo, dir string) {
	t.Helper()

	dir = t.TempDir()
	git(t, dir, "init", "-b", "main")
	git(t, dir, "config", "user.name", "codezor test")
	git(t, dir, "config", "user.email", "codezor@example.invalid")
	git(t, dir, "config", "commit.gpgsign", "false")
	git(t, dir, "config", "core.autocrlf", "false")

	repo, err := Open(t.Context(), dir)
	require.NoError(t, err)

	return repo, dir
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), "git", args...)
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

func TestDiffAgainstWorkingTreeSeesUncommittedEdits(t *testing.T) {
	repo, dir := newRepo(t)

	write(t, dir, "a.go", "package a\n\nconst One = 1\n")
	commit(t, dir, "first")
	base, err := repo.Head(t.Context())
	require.NoError(t, err)

	write(t, dir, "a.go", "package a\n\nconst One = 1\nconst Two = 2\n")
	write(t, dir, "new.go", "package new\n")

	changes, err := repo.Diff(t.Context(), base, "")
	require.NoError(t, err)
	require.Len(t, changes, 1)
	require.Equal(t, "a.go", changes[0].NewPath)
	require.Equal(t, []LineRange{{Start: 4, Count: 1}}, changes[0].Added)
	require.Empty(t, changes[0].Removed)

	untracked, err := repo.Untracked(t.Context())
	require.NoError(t, err)
	require.Equal(t, []string{"new.go"}, untracked)
}

func TestDiffBetweenCommits(t *testing.T) {
	repo, dir := newRepo(t)

	write(t, dir, "a.go", "package a\nconst One = 1\n")
	commit(t, dir, "first")
	write(t, dir, "a.go", "package a\n")
	commit(t, dir, "second")

	parent, err := repo.Parent(t.Context(), "HEAD")
	require.NoError(t, err)

	changes, err := repo.Diff(t.Context(), parent, "HEAD")
	require.NoError(t, err)
	require.Len(t, changes, 1)
	require.Equal(t, []LineRange{{Start: 2, Count: 1}}, changes[0].Removed)
	require.Empty(t, changes[0].Added)
}

func TestParentOfFirstCommitIsEmptyTree(t *testing.T) {
	repo, dir := newRepo(t)
	write(t, dir, "a.go", "package a\n")
	commit(t, dir, "first")

	parent, err := repo.Parent(t.Context(), "HEAD")
	require.NoError(t, err)
	require.Equal(t, EmptyTree, parent)
}

func TestBlobReadsCommittedContent(t *testing.T) {
	repo, dir := newRepo(t)
	write(t, dir, "a.go", "package a\n")
	commit(t, dir, "first")

	content, err := repo.Blob(t.Context(), "HEAD", "a.go")
	require.NoError(t, err)
	require.Equal(t, "package a\n", string(content))

	_, err = repo.Blob(t.Context(), "HEAD", "missing.go")
	require.Error(t, err)
}

func TestHeadFailsWithoutCommits(t *testing.T) {
	repo, _ := newRepo(t)

	_, err := repo.Head(t.Context())
	require.Error(t, err)
}
