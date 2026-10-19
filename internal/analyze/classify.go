package analyze

import (
	"path"
	"slices"
	"sync"

	"github.com/boyter/scc/v4/processor"
)

// shebangLookahead is how much of a file scc inspects to resolve a #! line.
const shebangLookahead = 200

// ignoreLanguages are the languages scc leaves out of a directory walk unless
// it is asked for them: a .gitignore steers the walk, it is not code. The same
// exclusion has to apply here, or an ignore file present on both sides would
// read as deleted code.
var ignoreLanguages = []string{"ignore", "gitignore"}

// constants guards the one time load of scc's language database, which lives
// in package level state and is needed before any file can be counted.
var constants sync.Once

func ensureConstants() {
	constants.Do(processor.ProcessConstants)
}

// lineTypes holds the type scc gave each line of one file, indexed from zero
// for line one.
type lineTypes []processor.LineType

// isCode reports whether the 1 based line n was counted as code. A line
// outside the file is not code, so a range that no longer matches the content
// cannot inflate a count.
func (l lineTypes) isCode(n int64) bool {
	if n < 1 || n > int64(len(l)) {
		return false
	}

	return l[n-1] == processor.LINE_CODE
}

// lineCollector records scc's verdict for each line as it counts.
type lineCollector struct {
	types lineTypes
}

// ProcessLine implements processor.FileJobCallback. Lines arrive in order, but
// the reported line number is used rather than assumed, so a gap cannot shift
// every later line.
func (c *lineCollector) ProcessLine(_ *processor.FileJob, line int64, lineType processor.LineType) bool {
	for int64(len(c.types)) < line {
		c.types = append(c.types, processor.LINE_BLANK)
	}
	c.types[line-1] = lineType

	return true
}

// classified is one file as scc counted it.
type classified struct {
	language string
	code     int64
	lines    lineTypes
}

// classify counts content as if it were the file at repoPath, which is what
// lets a git blob be measured without being written to disk. repoPath is a
// repository relative, slash separated path.
//
// ok is false for a binary file and for a path whose language scc does not
// recognise. Those are the same two exclusions scc applies while walking a
// directory, so a baseline tree and a working tree count the same set of
// files.
func classify(repoPath string, content []byte) (classified, bool) {
	ensureConstants()

	languages, extension := processor.DetectLanguage(repoPath)
	if len(languages) == 0 {
		return classified{}, false
	}
	for _, l := range languages {
		if slices.Contains(ignoreLanguages, l) {
			return classified{}, false
		}
	}
	for _, l := range languages {
		processor.LoadLanguageFeature(l)
	}

	collector := &lineCollector{}
	job := &processor.FileJob{
		Location:          repoPath,
		Filename:          path.Base(repoPath),
		Extension:         extension,
		PossibleLanguages: languages,
		Content:           content,
		Bytes:             int64(len(content)),
		Callback:          collector,
	}

	job.Language = processor.DetermineLanguage(job.Filename, job.Language, job.PossibleLanguages, job.Content)
	if job.Language == processor.SheBang {
		language, err := processor.DetectSheBang(content[:min(len(content), shebangLookahead)])
		if err != nil {
			return classified{}, false
		}

		job.Language = language
		processor.LoadLanguageFeature(language)
	}

	processor.CountStats(job)
	if job.Binary {
		return classified{}, false
	}

	return classified{language: job.Language, code: job.Code, lines: collector.types}, true
}
