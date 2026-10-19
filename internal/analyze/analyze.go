package analyze

import (
	"fmt"
	"slices"

	"github.com/boyter/scc/v4/processor"
)

// Presets are the LOCOMO model tiers. They describe a price and throughput
// class rather than a named model, so they do not go stale when models are
// retired or renamed.
var Presets = []string{"large", "medium", "small", "local"}

// Counts holds the line counts scc produced for one language or for a whole
// tree. Lines is the sum of Code, Comments and Blanks. Complexity is the count
// of branching tokens scc recognised, not a normalised score.
type Counts struct {
	Files      int64
	Lines      int64
	Code       int64
	Comments   int64
	Blanks     int64
	Complexity int64
}

// Language is the count of a single detected language.
type Language struct {
	Name string
	Counts
}

// Locomo is the LOCOMO (LLM Output COst MOdel) estimate of producing the
// counted code with an LLM. Token counts, Cost in US dollars and
// GenerationSeconds all scale with code volume and complexity; ReviewHours
// assumes a human reads every line.
type Locomo struct {
	Preset            string
	InputTokens       float64
	OutputTokens      float64
	Cost              float64
	GenerationSeconds float64
	ReviewHours       float64
	IterationFactor   float64
}

// Analysis is the counted state of one tree, as it exists on disk now.
type Analysis struct {
	Path      string
	Languages []Language
	Totals    Counts
	Locomo    Locomo
}

// Run counts the tree at path and estimates its LOCOMO cost using the named
// model tier. Files ignored by .gitignore are left out, matching scc defaults.
//
// It fails on an unknown tier instead of falling back to a different one, so a
// typo cannot silently change the reported cost.
func Run(path, locomoPreset string) (Analysis, error) {
	if !slices.Contains(Presets, locomoPreset) {
		return Analysis{}, fmt.Errorf("unknown LOCOMO preset %q, want one of %v", locomoPreset, Presets)
	}

	processor.DirFilePaths = []string{path}
	processor.LocomoPresetName = locomoPreset

	summaries, err := processor.ProcessResult()
	if err != nil {
		return Analysis{}, fmt.Errorf("count %q: %w", path, err)
	}

	a := Analysis{Path: path, Languages: make([]Language, 0, len(summaries))}
	for i := range summaries {
		s := &summaries[i]
		c := Counts{
			Files:      s.Count,
			Lines:      s.Lines,
			Code:       s.Code,
			Comments:   s.Comment,
			Blanks:     s.Blank,
			Complexity: s.Complexity,
		}
		a.Languages = append(a.Languages, Language{Name: s.Name, Counts: c})
		a.Totals.Files += c.Files
		a.Totals.Lines += c.Lines
		a.Totals.Code += c.Code
		a.Totals.Comments += c.Comments
		a.Totals.Blanks += c.Blanks
		a.Totals.Complexity += c.Complexity
	}

	e := processor.LocomoEstimate(a.Totals.Code, a.Totals.Complexity)
	a.Locomo = Locomo{
		Preset:            e.Preset,
		InputTokens:       e.InputTokens,
		OutputTokens:      e.OutputTokens,
		Cost:              e.Cost,
		GenerationSeconds: e.GenerationSeconds,
		ReviewHours:       e.ReviewHours,
		IterationFactor:   e.IterationFactor,
	}

	return a, nil
}
