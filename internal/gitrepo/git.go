package gitrepo

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// EmptyTree is git's object hash for an empty tree. Diffing the first commit
// in a repository against it is the only way to say "everything here is new".
const EmptyTree = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

// Repo is a local git repository read through the git command line.
type Repo struct {
	dir string
}

// Open checks that dir is inside a git work tree.
func Open(ctx context.Context, dir string) (*Repo, error) {
	r := &Repo{dir: dir}
	if _, err := r.text(ctx, "rev-parse", "--git-dir"); err != nil {
		return nil, fmt.Errorf("%q is not a git repository: %w", dir, err)
	}

	return r, nil
}

// Root is the top of the work tree. Everything is measured from there, so a
// report does not change depending on which subdirectory it was run from.
func (r *Repo) Root(ctx context.Context) (string, error) {
	return r.text(ctx, "rev-parse", "--show-toplevel")
}

// Head is the commit currently checked out. It fails in a repository that has
// no commits yet, which is where a baseline comparison stops making sense.
func (r *Repo) Head(ctx context.Context) (string, error) {
	return r.text(ctx, "rev-parse", "--verify", "HEAD^{commit}")
}

// MergeBase returns the commit where ref and HEAD last agreed. That is the
// baseline: commits that landed on ref after this branch started are not part
// of the measured change.
func (r *Repo) MergeBase(ctx context.Context, ref string) (string, error) {
	if _, err := r.text(ctx, "rev-parse", "--verify", "--quiet", ref+"^{commit}"); err != nil {
		return "", fmt.Errorf("baseline branch %q not found in this repository", ref)
	}

	return r.text(ctx, "merge-base", ref, "HEAD")
}

// ShortHash abbreviates a revision to the repository's short form.
func (r *Repo) ShortHash(ctx context.Context, rev string) (string, error) {
	return r.text(ctx, "rev-parse", "--short", rev)
}

// Parent returns the first parent of rev, or EmptyTree when rev is the first
// commit, so that a diff against the result always describes the whole commit.
func (r *Repo) Parent(ctx context.Context, rev string) (string, error) {
	out, err := r.text(ctx, "rev-list", "--parents", "-n", "1", rev)
	if err != nil {
		return "", err
	}

	// The line is "<commit> <parent>..." so a single field means no parent.
	fields := strings.Fields(out)
	if len(fields) < 2 {
		return EmptyTree, nil
	}

	return fields[1], nil
}

// Diff reports the files changed between base and target. An empty target
// means the working tree, so staged and unstaged edits are both included;
// untracked files are not, see Untracked.
func (r *Repo) Diff(ctx context.Context, base, target string) ([]FileDiff, error) {
	args := []string{
		// Paths stay literal UTF-8 instead of being escaped per byte.
		"-c", "core.quotePath=false",
		"diff",
		// No context lines: each hunk header then describes only changed lines.
		"-U0",
		"--no-prefix",
		"--no-color",
		"--no-ext-diff",
		"--no-textconv",
		"--find-renames",
		base,
	}
	if target != "" {
		args = append(args, target)
	}
	args = append(args, "--")

	out, err := r.raw(ctx, args...)
	if err != nil {
		return nil, err
	}

	return parseDiff(bytes.NewReader(out))
}

// Untracked lists the files git neither tracks nor ignores, relative to the
// work tree root. They are part of the current state of the code.
func (r *Repo) Untracked(ctx context.Context) ([]string, error) {
	out, err := r.raw(ctx, "-c", "core.quotePath=false", "ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return nil, err
	}

	var paths []string
	for p := range strings.SplitSeq(string(out), "\x00") {
		if p != "" {
			paths = append(paths, p)
		}
	}

	return paths, nil
}

// Blob returns the contents of path as of rev.
func (r *Repo) Blob(ctx context.Context, rev, path string) ([]byte, error) {
	return r.raw(ctx, "cat-file", "blob", rev+":"+path)
}

// WalkTree streams every regular file in rev's tree and calls fn with the
// path and the contents. It reads one `git archive` stream, so the whole tree
// costs a single git invocation and nothing is written to disk.
//
// Symbolic links, directories and submodules are skipped: none of them hold
// code of their own.
func (r *Repo) WalkTree(ctx context.Context, rev string, fn func(path string, content []byte) error) error {
	cmd := r.command(ctx, "archive", "--format=tar", rev)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("git archive %s: %w", rev, err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("git archive %s: %w", rev, err)
	}

	walkErr := walkTar(stdout, fn)
	if walkErr != nil {
		// Nothing more is wanted from the stream, and the error already
		// explains why.
		_ = cmd.Process.Kill()
	}

	// A tar ends with a marker followed by padding, so the walk above returns
	// while git is still writing. The rest has to be drained: waiting on a
	// child that is blocked writing into a full pipe never returns.
	_, _ = io.Copy(io.Discard, stdout)

	waitErr := cmd.Wait()

	switch {
	case walkErr != nil:
		return fmt.Errorf("walk tree %s: %w", rev, walkErr)
	case waitErr != nil:
		return fmt.Errorf("git archive %s: %w: %s", rev, waitErr, strings.TrimSpace(stderr.String()))
	}

	return nil
}

// walkTar hands every regular file in the archive to fn.
func walkTar(r io.Reader, fn func(path string, content []byte) error) error {
	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read archive: %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		content, err := io.ReadAll(tr)
		if err != nil {
			return fmt.Errorf("read %q from archive: %w", header.Name, err)
		}

		if err := fn(header.Name, content); err != nil {
			return err
		}
	}
}

// command builds a git invocation rooted in the repository.
func (r *Repo) command(ctx context.Context, args ...string) *exec.Cmd {
	// git is resolved from PATH and every argument is built in this package,
	// never taken from the analyzed repository.
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.dir

	return cmd
}

// raw runs git and returns stdout unchanged, keeping stderr in the error so a
// failure explains itself.
func (r *Repo) raw(ctx context.Context, args ...string) ([]byte, error) {
	cmd := r.command(ctx, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}

	return stdout.Bytes(), nil
}

// text runs git and returns stdout with surrounding whitespace removed, for
// the commands that answer with a single value.
func (r *Repo) text(ctx context.Context, args ...string) (string, error) {
	out, err := r.raw(ctx, args...)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}
