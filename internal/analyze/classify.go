package analyze

import (
	"path"
	"slices"
	"sync"

	"github.com/boyter/scc/v3/processor"
)

// shebangLookahead is how much of a file scc inspects to resolve a #! line.
const shebangLookahead = 200

// ignoreLanguages are the languages scc leaves out of a count unless it is
// asked for them: a .gitignore steers a walk, it is not code. Excluding them
// on both sides keeps an unchanged ignore file from reading as deleted code.
var ignoreLanguages = []string{"ignore", "gitignore"}

// constants guards the one time load of scc's language database, which lives
// in package level state and is needed before any file can be counted.
var constants sync.Once

func ensureConstants() {
	constants.Do(processor.ProcessConstants)
}

// classified is one file as scc counted it.
type classified struct {
	language string
	code     int64
}

// classify counts content as if it were the file at repoPath, which is what
// lets a git blob and a file on disk go through the same counter. repoPath is
// a repository relative, slash separated path.
//
// ok is false for a binary file, for an ignore file, and for a path whose
// language scc does not recognise, so both sides of a comparison count the
// same set of files.
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

	filename := path.Base(repoPath)
	job := &processor.FileJob{
		Location:          repoPath,
		Extension:         extension,
		PossibleLanguages: languages,
		Content:           content,
		Bytes:             int64(len(content)),
	}

	job.Language = processor.DetermineLanguage(filename, job.Language, job.PossibleLanguages, job.Content)
	if job.Language == processor.SheBang {
		language, err := processor.DetectSheBang(string(content[:min(len(content), shebangLookahead)]))
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

	return classified{language: job.Language, code: job.Code}, true
}
