package gitrepo

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// devNull is the path git prints for the missing side of an add or a delete.
const devNull = "/dev/null"

// maxDiffLine bounds a single diff line. Generated and minified files hold
// very long lines, and the default scanner limit of 64 KiB rejects them.
const maxDiffLine = 16 * 1024 * 1024

// LineRange is a run of lines on one side of a diff. Start is 1 based, and a
// range of Count 0 is not emitted.
type LineRange struct {
	Start int64
	Count int64
}

// FileDiff is one file changed between two states. Added ranges address lines
// in the new state, Removed ranges lines in the old state. OldPath is empty on
// an added file, NewPath on a deleted one, and the two differ on a rename.
//
// Both range lists are empty for a file git reported as binary, and for a
// rename or a mode change that altered no line.
type FileDiff struct {
	OldPath string
	NewPath string
	Added   []LineRange
	Removed []LineRange
	Binary  bool
}

// parseDiff reads the output of `git diff -U0 --no-prefix`. Without context
// lines every hunk header describes exactly the changed lines, so the hunk
// bodies carry no information and are skipped.
//
// Body lines are not distinguishable from file headers by prefix alone: a
// removed line reading "-- x" prints as "--- x". Headers are therefore only
// accepted before the first hunk of each file, where git always puts them.
func parseDiff(r io.Reader) ([]FileDiff, error) {
	var (
		out      []FileDiff
		current  *FileDiff
		seenHunk bool
	)

	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), maxDiffLine)

	for s.Scan() {
		line := s.Text()

		switch {
		case strings.HasPrefix(line, "diff --git "):
			if current != nil {
				out = append(out, *current)
			}
			current = &FileDiff{}
			seenHunk = false

		case current == nil:
			// Preamble before the first file header holds nothing to collect.

		case strings.HasPrefix(line, "@@"):
			seenHunk = true

			removed, added, err := parseHunk(line)
			if err != nil {
				return nil, err
			}
			if removed.Count > 0 {
				current.Removed = append(current.Removed, removed)
			}
			if added.Count > 0 {
				current.Added = append(current.Added, added)
			}

		case seenHunk:
			// Hunk body, already described by its header.

		case strings.HasPrefix(line, "--- "):
			current.OldPath = diffPath(line[len("--- "):])

		case strings.HasPrefix(line, "+++ "):
			current.NewPath = diffPath(line[len("+++ "):])

		case strings.HasPrefix(line, "Binary files "), strings.HasPrefix(line, "GIT binary patch"):
			current.Binary = true
		}
	}

	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("read diff: %w", err)
	}

	if current != nil {
		out = append(out, *current)
	}

	return out, nil
}

// parseHunk reads a hunk header such as "@@ -20,2 +23,0 @@ func Example()".
func parseHunk(line string) (removed, added LineRange, err error) {
	const marker = "@@"

	end := strings.Index(line[len(marker):], marker)
	if end < 0 {
		return removed, added, fmt.Errorf("unterminated hunk header %q", line)
	}

	fields := strings.Fields(line[len(marker) : len(marker)+end])
	if len(fields) != 2 || !strings.HasPrefix(fields[0], "-") || !strings.HasPrefix(fields[1], "+") {
		return removed, added, fmt.Errorf("malformed hunk header %q", line)
	}

	if removed, err = parseRange(fields[0][1:]); err != nil {
		return removed, added, fmt.Errorf("hunk header %q: %w", line, err)
	}
	if added, err = parseRange(fields[1][1:]); err != nil {
		return removed, added, fmt.Errorf("hunk header %q: %w", line, err)
	}

	return removed, added, nil
}

// parseRange reads "start,count" or a bare "start", which means one line.
func parseRange(s string) (LineRange, error) {
	start, count := s, "1"
	if i := strings.IndexByte(s, ','); i >= 0 {
		start, count = s[:i], s[i+1:]
	}

	startN, err := strconv.ParseInt(start, 10, 64)
	if err != nil {
		return LineRange{}, fmt.Errorf("line number %q: %w", start, err)
	}

	countN, err := strconv.ParseInt(count, 10, 64)
	if err != nil {
		return LineRange{}, fmt.Errorf("line count %q: %w", count, err)
	}

	return LineRange{Start: startN, Count: countN}, nil
}

// diffPath normalises a path from a "---" or "+++" header. The missing side of
// an add or a delete becomes an empty path, and a path git quoted because of
// unusual characters is unquoted.
func diffPath(s string) string {
	s = strings.TrimSpace(s)
	if s == devNull {
		return ""
	}

	if strings.HasPrefix(s, `"`) {
		if unquoted, err := strconv.Unquote(s); err == nil {
			return unquoted
		}
	}

	return s
}
