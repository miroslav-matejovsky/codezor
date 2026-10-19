# Code Statistics

Baseline: `main` (merge-base 4f6607d)
Current: working tree, including uncommitted and untracked files

LOC is code lines as counted by [scc](https://github.com/boyter/scc):
comment and blank lines are left out of every number below, churn
included. The report carries no timestamp, so an unchanged tree renders
an unchanged report.

## LOC

| Language | Baseline | Current | Delta | Change |
| --- | ---: | ---: | ---: | ---: |
| Go | 28 | 1,416 | +1,388 | +4957.1% |
| Markdown | 319 | 358 | +39 | +12.2% |
| License | 216 | 216 | 0 | 0.0% |
| YAML | 206 | 216 | +10 | +4.9% |
| JSON | 108 | 113 | +5 | +4.6% |
| Powershell | 111 | 111 | 0 | 0.0% |
| **Total** | **988** | **2,430** | **+1,442** | **+146.0%** |

## Current Change

- Added LOC: 1,558
- Removed LOC: 116
- Gross churn: 1,674 LOC
- Net change: +1,442 LOC
- Change footprint: 169.4% of baseline LOC

## Latest Commit

- Added LOC: 1,210
- Removed LOC: 79
- Gross churn: 1,289 LOC
- Net change: +1,131 LOC

## LOCOMO

LOCOMO (LLM Output COst MOdel) estimates what producing the current
code with an LLM would cost. It scales with code volume and complexity,
and leaves out context reuse, test generation and debugging. It is a
ballpark figure, as COCOMO is for human development.

| Estimate | Value |
| --- | ---: |
| Model tier | medium |
| Input tokens | 251,149 |
| Output tokens | 50,771 |
| Cost | $1.52 |
| Iteration factor | 2.09 |
| Generation time | 16.9 min |
| Review time | 24.3 min |
