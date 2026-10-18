# Principles

- **Local-first** — Repository statistics require only the local repository. No network, CI service, remote API, telemetry, or external database is required.

- **Fast** — Statistics belong in the inner development loop. Running the tool should be cheap enough to do repeatedly while making changes.

- **Current** — The local working tree is the authoritative current state, including committed branch changes, staged changes, unstaged changes, and untracked files.

- **Comparative** — Describe both the repository and the current change. Absolute measurements provide repository context; comparison against `main` provides a sense of change scale.

- **Informative, not judgmental** — Report observable facts such as LOC, churn, complexity, dependencies, and architectural reach. Avoid opaque quality scores, grades, or unsupported risk classifications.

- **Deterministic** — The same repository state and tool version should produce the same meaningful output. Avoid timestamps, machine-specific paths, unstable ordering, and other incidental differences.

- **Explainable** — Every metric must have a precise and inspectable definition. A developer or agent should be able to understand where a number came from and what it means.

- **Stable semantics** — Metric definitions must remain consistent. Rules for generated code, tests, comments, vendor code, build artifacts, submodules, and other exclusions must be explicit.

- **Repository-aware** — Measure structures meaningful to the repository's technology and architecture, such as projects, packages, modules, dependencies, public interfaces, or bounded contexts.

- **Change-aware** — Measure gross change, not only net growth. Large rewrites must remain visible even when additions and deletions approximately cancel each other out.

- **Architecturally aware** — Prefer information about the structural reach of a change over metric abundance. Touching multiple projects, modules, bounded contexts, dependencies, or public contracts can be more significant than raw LOC.

- **Minimal** — Prefer a small set of decision-useful measurements over comprehensive static-analysis reporting. `stats.md` should provide situational awareness within seconds.

- **Diff-friendly** — Keep structure, ordering, formatting, and precision stable so changes to `stats.md` produce meaningful Git diffs.

- **Gracefully degradable** — Basic repository and Git statistics should remain available even when language-specific or complexity analyzers are unavailable.

- **Human- and agent-readable** — Use simple Markdown, stable headings, explicit units, and predictable structure. Humans and development agents should consume the same representation.

- **No history database** — Describe the repository as it exists now and its deviation from the baseline. Git remains the source of historical information; `stats.md` is not an analytics warehouse.

- **Repository-centric** — `stats.md` describes the current repository. Developer assessment, code review, project management, historical analytics, and quality gates are separate concerns.