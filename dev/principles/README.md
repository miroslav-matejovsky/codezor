# Principles

- **Local-first** — Everything needed to understand the codebase and its current changes is available locally, without network services, CI, telemetry, or external databases.

- **Fast** — Analysis belongs in the inner development loop and must be inexpensive enough to run repeatedly while coding.

- **Contextual** — Show the codebase as it exists now and measure the current working tree, including uncommitted and untracked code, against a clear baseline such as `main`.

- **Informative, not judgmental** — Report observable facts such as size, churn, complexity, structure, dependencies, and architectural reach. Do not reduce them to opaque scores, grades, or risk labels.

- **Explainable and deterministic** — Every metric has a precise definition, stable semantics, and reproducible output. Equivalent code states should produce equivalent results.

- **Focused** — Prefer a small set of high-information measurements that provide immediate orientation. The goal is understanding the codebase and the scale of the current intervention, not comprehensive static analysis or historical analytics.
