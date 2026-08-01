---
name: k8s-review
description: Review Kubernetes manifests and a specifically verified cluster using bounded, sanitized evidence, then develop evidence-backed optimization hypotheses without changing cluster state. Use for Kubernetes audits, workload health reviews, capacity and scheduling analysis, upgrade-readiness reviews, or optimization assessments; delegate cluster lifecycle and deployment work to k8s-local.
license: MIT
---

# Kubernetes Review

Review one explicit target without taking ownership of it. A review request never authorizes apply, delete, restart, rollout, scale, patch, drain, cordon, namespace creation, cluster lifecycle, or volume mutation.

Compose [k8s-local](../k8s-local/SKILL.md) for lifecycle or deployment work. Use the repository-owned `dot cluster diagnose` collector and its versioned `dot.cluster.diagnostics/v1` schema instead of assembling ad hoc `kubectl` output. Read the [review matrix](references/review-matrix.md) for dimension-specific evidence and exercise the [behavioral evaluations](tests/behavioral-evaluations.md) when changing this workflow. Agent metadata is in [openai.yaml](agents/openai.yaml).

## Target and authority gate

1. Read repository instructions and identify an explicit or repository-derived kubeconfig, context, namespace, and ownership boundary. Never fall back to the user's default current context.
1. Display the intended kubeconfig, context, namespace, and permitted evidence scope. If derivation is ambiguous, stop before any cluster probe.
1. Verify the selected target through `dot cluster --kubeconfig <path> --context <context> status`. Treat a missing tool, absent cluster, unavailable API, context mismatch, namespace mismatch, or authorization failure as a bounded result, not permission to switch targets.
1. For review-only work, keep every operation read-only. If the user later authorizes a mutation, hand lifecycle, deployment, or teardown to `k8s-local` and re-verify the exact target immediately before each state-changing command.

## Evidence collection

1. Review static manifests through repository-owned `mise` checks. Validate schema and deprecated APIs, security contexts, resource requests and limits, scheduling constraints, autoscaling, probes, disruption budgets, networking policy, storage semantics, controller ownership, and upgrade risk without applying manifests.
1. Collect runtime evidence with `dot cluster --kubeconfig <path> --context <context> diagnose --namespace <namespace> --output <owner-only-path>`. Keep the default positive timeout, time-window, log-tail, line, byte, and pod bounds unless a smaller scope is sufficient; add `--redact-pattern` for known project identifiers.
1. Validate that the bundle schema is exactly `dot.cluster.diagnostics/v1`, the target fingerprint and namespace match the verified target, limits are present and positive, and every result is `ok` or an explicit partial error. Never upload the bundle automatically.
1. Do not query Secret objects, dump kubeconfigs or environment variables, request unlimited logs, or bypass collector sanitization. Inspect additional read-only fields only when the bundle proves a specific evidence gap, the same target is re-verified, and output is bounded and redacted.

## Analysis and optimization

1. Correlate manifest intent with runtime state. Distinguish observed workload behavior from absent metrics and from static risk.
1. Identify bottlenecks, idle components, hot paths, and optimization hypotheses using requests versus observed CPU or memory, placement and pending reasons, controller conditions, recent events, storage pressure, restart patterns, probe behavior, disruption coverage, network exposure, and upgrade evidence.
1. Reject a recommendation when one snapshot cannot establish representative demand, when metrics are unavailable, or when the change would add unjustified complexity. Prefer a measurable experiment over permanent policy.
1. For every retained optimization, record the observation, hypothesis, proposed change, risk, rollback boundary, workload owner, and acceptance measurement. A proposed optimization is not runtime-proven until separately authorized execution produces representative before-and-after evidence.

## Proof boundaries

Report these layers independently:

1. **Static manifest proof**: Validated source paths, tools, versions, and unresolved risks.
1. **Runtime behavior**: Exact fingerprint, namespace, collection window, healthy probes, partial errors, and missing metrics.
1. **Proposed optimization**: Evidence-backed hypothesis and smallest reversible experiment; no claim that it was applied.
1. **Acceptance evidence**: Only after separately authorized mutation and representative measurement; include before and after signals and rollback result.

## Output contract

Return:

1. **Key findings**: Rank correctness, security, reliability, and efficiency findings by impact with direct manifest or bundle evidence.
1. **Evidence matrix**: Cover every review dimension from the reference and mark each `verified`, `partial`, `not present`, or `not checked`.
1. **Optimization hypotheses**: State the evidence, expected benefit, experiment, acceptance threshold, and rollback boundary.
1. **Proof boundaries**: Separate static, runtime, proposal, and acceptance layers.
1. **Actions**: Route any authorized lifecycle or deployment work to `k8s-local`; otherwise stop at review and recommendations.
