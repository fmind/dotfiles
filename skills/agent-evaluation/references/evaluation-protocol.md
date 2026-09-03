# Evaluation Protocol

Execution rules behind the [agent-evaluation](SKILL.md) workflow: isolation, retention, statistics, invalid-run handling, and the brief that closes an evaluation.

## Sandbox and tool gateway

- Exercise external actions through a fake or deny-by-default tool gateway; record attempted calls, including forbidden attempts, instead of granting production access.
- Run each tool-using trial in a disposable per-run sandbox with read-only source fixtures, unique writable state, bounded CPU, memory, disk, process, and time budgets, and deny-by-default network access.
- Fake or block destructive local tools, verify cleanup and teardown, and retain attempted-action evidence without granting the action.

## Trace retention

- Redact secrets, personal data, tenant identifiers, and sensitive prompts before persisting traces.
- Predeclare storage, access owners, retention, and verified deletion for sanitized artifacts; retain deletion and exceptional-access receipts.
- Preserve a re-identification mapping only when authorized and necessary, under a separate stricter lifecycle.

## Statistical adequacy

- Choose the number of trials, seeds, minimum detectable difference or sensitivity, and uncertainty method before execution, in proportion to decision consequence and outcome variance.
- Predeclare stopping rules and adjustments for multiple metrics or repeated looks; keep paired and clustered samples intact in uncertainty estimates.
- Distinguish `pass@k` (at least one of `k` attempts succeeds) from `pass^k` (all `k` attempts succeed) and use the measure that matches the retry policy and harm model.
- Report distributions and uncertainty rather than only a best run or a mean; for zero observed severe failures, report an upper confidence bound rather than claiming zero risk.

## Invalid-run handling

- Define harness-invalid criteria, candidate failures, timeouts, missing traces, exclusions, maximum replacement runs, and adjudication before execution, and apply the taxonomy symmetrically to baseline and candidate.
- Report attempted, valid, invalid, excluded, replaced, and analyzed denominators; a candidate error is not a harness-invalid run merely because it lowers the score.

## Sealed holdout

- Assign named access principals, keep the sealed set encrypted where storage permits, and maintain an append-only exposure log.
- Assert immediately before execution that candidate authors and graders have not accessed its contents; tune only on development cases.
- Invalidate and rotate an exposed holdout; when the holdout invalidates the candidate, return to development with a new future holdout rather than tuning against the exposed set.

## Evaluation brief

- **Decision and confidence**: recommendation, intended release decision, and unresolved uncertainty.
- **Baseline and candidate identity**: every pinned model, prompt, tool, retrieval, code, environment, sampling, and retry component.
- **Corpus**: mode, development counts, segments, provenance, adversarial coverage, and hashes; in release mode also sealed counts, access, exposure, and contamination checks.
- **Graders**: deterministic, rule, model, and human responsibilities, versions, calibration, and disagreement path.
- **Execution**: trial count and adequacy rationale, uncertainty method, stopping and multiplicity controls, sandbox and fresh-state controls, invalid-run taxonomy, attempted-versus-analyzed denominators, fake-tool boundary, trace schema, redaction and retention receipts, holdout exposure evidence, token and cost ceiling.
- **Scorecard**: baseline, candidate, delta, distribution, and threshold for capability, reliability, safety, latency, tokens, and cost by segment.
- **Missing evidence**: runtime, paid-service, human, deployment, or public proof still absent.
- **Next step**: one bounded change or additional test, without silently authorizing implementation or external execution.
