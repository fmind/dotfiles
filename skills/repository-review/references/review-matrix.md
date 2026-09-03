# Repository Review Matrix

Use this matrix to plan a cross-cutting review. Skip a dimension only when it is genuinely out of scope, then name the omission and the rung it caps on the [proof ladder](../production-readiness/SKILL.md).

| Dimension | Inspect | Strong evidence | Common residual gap |
| --- | --- | --- | --- |
| Architecture | package boundaries, dependencies, data/control flow, failure domains | source paths, generated graphs, dependency metadata | design intent not encoded in the repository |
| Source | correctness, types, errors, concurrency, portability, dead code | focused reproduction, compiler, linter, tests | paths requiring unavailable services |
| Tests | deterministic coverage, race behavior, fixtures, integration boundaries | exact commands, test count, coverage, race result | paid or destructive end-to-end scenarios |
| Tooling | pinned versions, tasks, hooks, clean-clone bootstrap | lockfiles, `mise tasks`, hook execution, isolated clone | host-only configuration masking a clean-run failure |
| Security | secrets, dependencies, IaC, licenses, image/provenance controls | complete scanner output with scope and database state | timeout, offline database, skipped image/history |
| CI/CD | syntax, permissions, task reuse, exact-head status, deployment gates | workflow source plus checks for the reviewed SHA | historical green run or inaccessible environment |
| Documentation | commands, links, paths, examples, agent instructions | live task/CLI/file comparison | behavior that is documented but not machine-checked |
| Generated state | committed artifacts, formatters, generators, lockfiles | clean isolated regeneration and `git diff` | non-hermetic generator or unavailable toolchain |
| Release | version source, changelog, immutable tags, artifact/SBOM/signature publication | exact tag SHA and public artifact verification | draft release, movable tag, or unverified artifact |
| Runtime | readiness, acceptance behavior, persistence, observability | bounded read-only probe against the authorized target | missing credential, ambiguous target, or mutation-only test |

Report the highest proven rung of the [proof ladder](../production-readiness/SKILL.md) and every material gap above it.
