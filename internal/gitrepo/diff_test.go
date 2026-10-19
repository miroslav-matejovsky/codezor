package gitrepo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDiffCollectsHunkRanges(t *testing.T) {
	// git diff -U0 --no-prefix output: the hunk headers carry every changed
	// line, so a modified file needs no body parsing.
	diff := strings.Join([]string{
		"diff --git internal/a.go internal/a.go",
		"index 1111111..2222222 100644",
		"--- internal/a.go",
		"+++ internal/a.go",
		"@@ -12,0 +13,3 @@ func Example() {",
		"+a",
		"+b",
		"+c",
		"@@ -20,2 +23,0 @@",
		"-old",
		"-old",
		"@@ -31 +34 @@",
		"-replaced",
		"+replacement",
	}, "\n")

	got, err := parseDiff(strings.NewReader(diff))
	require.NoError(t, err)
	require.Len(t, got, 1)

	require.Equal(t, "internal/a.go", got[0].OldPath)
	require.Equal(t, "internal/a.go", got[0].NewPath)
	require.Equal(t, []LineRange{{Start: 13, Count: 3}, {Start: 34, Count: 1}}, got[0].Added)
	require.Equal(t, []LineRange{{Start: 20, Count: 2}, {Start: 31, Count: 1}}, got[0].Removed)
}

func TestParseDiffHandlesAddDeleteAndRename(t *testing.T) {
	diff := strings.Join([]string{
		"diff --git added.go added.go",
		"new file mode 100644",
		"--- /dev/null",
		"+++ added.go",
		"@@ -0,0 +1,2 @@",
		"+one",
		"+two",
		"diff --git gone.go gone.go",
		"deleted file mode 100644",
		"--- gone.go",
		"+++ /dev/null",
		"@@ -1,4 +0,0 @@",
		"-a",
		"-b",
		"-c",
		"-d",
		"diff --git from.go to.go",
		"similarity index 90%",
		"rename from from.go",
		"rename to to.go",
		"--- from.go",
		"+++ to.go",
		"@@ -5,1 +5,1 @@",
		"-before",
		"+after",
	}, "\n")

	got, err := parseDiff(strings.NewReader(diff))
	require.NoError(t, err)
	require.Len(t, got, 3)

	require.Empty(t, got[0].OldPath)
	require.Equal(t, "added.go", got[0].NewPath)
	require.Equal(t, []LineRange{{Start: 1, Count: 2}}, got[0].Added)
	require.Empty(t, got[0].Removed)

	require.Equal(t, "gone.go", got[1].OldPath)
	require.Empty(t, got[1].NewPath)
	require.Equal(t, []LineRange{{Start: 1, Count: 4}}, got[1].Removed)
	require.Empty(t, got[1].Added)

	require.Equal(t, "from.go", got[2].OldPath)
	require.Equal(t, "to.go", got[2].NewPath)
}

// A removed line whose content starts with "-- " prints as "--- ", and an
// added "++ " line prints as "+++ ". Only the headers before the first hunk
// may be read as paths.
func TestParseDiffIgnoresBodyLinesThatLookLikeHeaders(t *testing.T) {
	diff := strings.Join([]string{
		"diff --git schema.sql schema.sql",
		"--- schema.sql",
		"+++ schema.sql",
		"@@ -1,1 +1,2 @@",
		"--- old sql comment",
		"+++ new sql comment",
		"+Binary files are not here",
	}, "\n")

	got, err := parseDiff(strings.NewReader(diff))
	require.NoError(t, err)
	require.Len(t, got, 1)

	require.Equal(t, "schema.sql", got[0].OldPath)
	require.Equal(t, "schema.sql", got[0].NewPath)
	require.False(t, got[0].Binary)
	require.Equal(t, []LineRange{{Start: 1, Count: 2}}, got[0].Added)
}

func TestParseDiffMarksBinaryFiles(t *testing.T) {
	diff := strings.Join([]string{
		"diff --git logo.png logo.png",
		"index 1111111..2222222 100644",
		"Binary files logo.png and logo.png differ",
	}, "\n")

	got, err := parseDiff(strings.NewReader(diff))
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.True(t, got[0].Binary)
	require.Empty(t, got[0].Added)
	require.Empty(t, got[0].Removed)
}

func TestParseDiffUnquotesPaths(t *testing.T) {
	diff := strings.Join([]string{
		`diff --git "has space.go" "has space.go"`,
		`--- "has space.go"`,
		`+++ "has space.go"`,
		"@@ -1 +1 @@",
		"-a",
		"+b",
	}, "\n")

	got, err := parseDiff(strings.NewReader(diff))
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "has space.go", got[0].NewPath)
}

func TestParseDiffEmptyInput(t *testing.T) {
	got, err := parseDiff(strings.NewReader(""))
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestParseDiffRejectsBrokenHunkHeader(t *testing.T) {
	diff := "diff --git a.go a.go\n--- a.go\n+++ a.go\n@@ nonsense @@\n"

	_, err := parseDiff(strings.NewReader(diff))
	require.ErrorContains(t, err, "hunk")
}
