# codezor

Source counts for `.`, measured with [scc](https://github.com/boyter/scc).
The report describes the working tree as it is on disk and carries no timestamp,
so an unchanged tree renders an unchanged report.

## Code

| Language | Files | Lines | Code | Comments | Blanks | Complexity |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Markdown | 12 | 405 | 316 | 0 | 89 | 0 |
| Go | 8 | 384 | 289 | 44 | 51 | 19 |
| Powershell | 6 | 171 | 111 | 36 | 24 | 20 |
| YAML | 3 | 258 | 214 | 13 | 31 | 0 |
| JSON | 2 | 110 | 110 | 0 | 0 | 0 |
| License | 1 | 287 | 216 | 0 | 71 | 0 |
| **Total** | 32 | 1,615 | 1,256 | 93 | 266 | 39 |

## LOCOMO

LOCOMO (LLM Output COst MOdel) estimates what producing this code with an LLM
would cost. It scales with code volume and complexity, and leaves out context
reuse, test generation and debugging. It is a ballpark figure, as COCOMO is for
human development.

| Estimate | Value |
| --- | ---: |
| Model tier | medium |
| Input tokens | 87,531 |
| Output tokens | 23,266 |
| Cost | $0.61 |
| Iteration factor | 1.85 |
| Generation time | 7.8 min |
| Review time | 12.6 min |
