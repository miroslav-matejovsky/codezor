package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseOptions(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
		path string
		out  string
	}{
		{name: "current directory", path: "."},
		{name: "positional path", args: []string{"../other"}, path: "../other"},
		{name: "path flag", args: []string{"-path", "../other"}, path: "../other"},
		{name: "explicit summary", args: []string{"-summary"}, path: "."},
		{name: "report", args: []string{"-out", "report.md"}, path: ".", out: "report.md"},
		{name: "report before path", args: []string{"-out", "report.md", "../other"}, path: "../other", out: "report.md"},
		{name: "report after path", args: []string{"../other", "-out", "report.md"}, path: "../other", out: "report.md"},
		{name: "escaped path", args: []string{"--", "-repo"}, path: "-repo"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOptions(tt.args, io.Discard)
			require.NoError(t, err)
			require.Equal(t, tt.path, got.path)
			require.Equal(t, tt.out, got.out)
		})
	}
}

func TestParseOptionsRejectsInvalidArguments(t *testing.T) {
	for _, args := range [][]string{
		{"one", "two"},
		{"-path", "one", "two"},
		{"one", "-path", "two"},
		{"-out"},
		{"-unknown"},
	} {
		t.Run("arguments "+strings.Join(args, " "), func(t *testing.T) {
			_, err := parseOptions(args, io.Discard)
			require.Error(t, err)
		})
	}
}

func TestRunOutput(t *testing.T) {
	repo := t.TempDir()
	for _, args := range [][]string{
		{"init", "-b", "main"},
		{"-c", "user.name=codezor test", "-c", "user.email=codezor@example.invalid", "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-m", "baseline"},
	} {
		cmd := exec.CommandContext(t.Context(), "git", args...)
		cmd.Dir = repo
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "%s", output)
	}
	require.NoError(t, os.WriteFile(filepath.Join(repo, "example.go"), []byte("package example\nfunc Pick(n int) int { if n > 0 { return 1 }; return 0 }\n"), 0o600))

	for _, name := range []string{"current directory", "positional path", "report"} {
		t.Run(name, func(t *testing.T) {
			t.Chdir(repo)
			capture, err := os.CreateTemp(t.TempDir(), "stdout")
			require.NoError(t, err)
			original := os.Stdout
			os.Stdout = capture
			t.Cleanup(func() {
				os.Stdout = original
				require.NoError(t, capture.Close())
			})

			var args []string
			if name == "positional path" {
				args = []string{repo}
			}
			out := filepath.Join(t.TempDir(), "report.md")
			if name == "report" {
				args = []string{repo, "-out", out}
			}
			opts, err := parseOptions(args, io.Discard)
			require.NoError(t, err)
			require.NoError(t, run(t.Context(), opts.path, opts.out, opts.summary))
			stdout, err := os.ReadFile(capture.Name())
			require.NoError(t, err)
			if name == "report" {
				data, err := os.ReadFile(out)
				require.NoError(t, err)
				require.Contains(t, string(data), "Baseline")
				require.Contains(t, string(stdout), "codezor: wrote "+out)
			} else {
				require.Contains(t, string(stdout), "Language (cplx loc)")
				require.Contains(t, string(stdout), "Delta (main)")
				require.NoFileExists(t, out)
			}
			entries, err := os.ReadDir(repo)
			require.NoError(t, err)
			require.Len(t, entries, 2, "only .git and the fixture source should exist")
		})
	}
}
