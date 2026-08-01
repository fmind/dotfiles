# Repository Review Matrix

Use this matrix to plan a cross-cutting review. Skip a dimension only when it is genuinely out of scope, then name the omission and its effect on the proof ceiling.

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

## Proof boundaries

- **Source-ready**: The reviewed source and configuration satisfy inspection and focused checks. Mark this failed when a confirmed material finding violates the scoped source contract, and not proven when required source dimensions were skipped. This does not imply a complete local gate.
- **Local-green**: The complete repository-owned gate passed against one coherent materialized candidate.
- **Exact-head-CI**: Every required workflow passed for the exact reviewed commit SHA.
- **Runtime-proven**: Authorized runtime acceptance behavior was observed for the reviewed artifact/configuration, not merely service readiness.
- **Deployed**: The reviewed artifact or configuration is confirmed active in the named environment.
- **Release-published**: An immutable tag and public release artifacts are available and independently verified.

Never collapse these states into a single "done" label. State the strongest proven boundary and every material gap above it.
