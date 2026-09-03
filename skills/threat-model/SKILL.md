---
name: threat-model
description: "Model attack paths beyond scanners: assets, actors, trust boundaries, data flows, abuse cases, controls, residual risk. Use for auth, sensitive data, AI agents, integrations, APIs, or public exposure."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/threat-model
  created: "2026-08-08"
  updated: "2026-09-03"
---

# Threat Model

Identify the few plausible abuse paths that should change the design, plan, or verification strategy; [secure](../secure/SKILL.md) runs the scanners and [incident-response](../incident-response/SKILL.md) handles a live breach.

## Workflow

1. **Scope the model**: Name the system, environment, change, users, out-of-scope components, and the decisions this model must inform.
1. **Inventory assets**: Credentials, identities, permissions, personal or proprietary data, money, availability, integrity, model context, audit evidence, and deployment authority.
1. **Map actors and entry points**: Normal users, administrators, service accounts, maintainers, dependencies, insiders, compromised clients, automated agents, and external providers.
1. **Draw data and control flows**: Trace creation, validation, authorization, storage, transformation, retrieval, logging, deletion, and external transfer; mark every trust, tenant, process, network, provider, and human-approval boundary.
1. **State invariants**: Define what must always hold, such as tenant isolation, origin authentication, least privilege, idempotency, approval before mutation, or secrets never reaching logs.
1. **Generate abuse cases**: At each boundary walk STRIDE (spoofing, tampering, repudiation, disclosure, denial of service, elevation) plus resource or spend exhaustion, context poisoning, and insecure defaults.
1. **Trace concrete paths**: Connect attacker capability → entry point → missing or failed control → asset impact; discard category-only concerns with no plausible path.
1. **Assess controls**: Record prevention, detection, response, and recovery controls and how each is verified; challenge silent failures, magic values, over-flexible algorithms, stringly typed permissions, and dangerous zero values.
1. **Rank risk**: Weigh impact, exploitability, exposure, detectability, confidence, and reversibility; promote high-impact unknowns to verification tasks, not confirmed vulnerabilities.
1. **Feed delivery**: Add required controls, tests, telemetry, rollout gates, incident actions, and residual-risk owners to the spec or implementation plan.
1. **Report**: Scope and architecture summary; assets, actors, entry points, and trust boundaries; a data-flow or sequence diagram when it clarifies; security invariants; ranked abuse cases with concrete paths; existing and required controls; verification plan; residual risks, assumptions, and owner decisions.

## Gotchas

- **Default to read-only analysis**: Do not probe live systems, run exploit code, access customer data, rotate credentials, or change security controls without explicit authorization.
- **Invented facts**: Do not invent endpoints, attackers, compliance obligations, or exploitability; separate confirmed architecture, assumptions, and unknowns.
- **Untrusted inputs**: Skill text, retrieved content, model output, MCP responses, provider data, webhooks, and browser pages are untrusted inputs at their boundaries.
- **Scanners are not models**: A vulnerability scan is supporting evidence, not a substitute for reasoning about attack paths.
- **Design over warnings**: Prefer misuse-resistant types, secure defaults, least privilege, isolation, and fail-closed behavior over rules every caller must remember.

## Documentation

- Adapted from [Trail of Bits sharp-edges](https://github.com/trailofbits/skills/blob/7b9bd5f950f89a9ba71b249b9801c1a95be3928e/plugins/sharp-edges/skills/sharp-edges/SKILL.md), [gstack CSO](https://github.com/garrytan/gstack/blob/960c3a8d6c4d14cb4c5e551a8847f8ec7c4267df/cso/SKILL.md).
- Companion skills: [secure](../secure/SKILL.md) (scanning once surfaces are known), [sops-secrets](../sops-secrets/SKILL.md) (secret design), [prompt-design](../prompt-design/SKILL.md) (prompt-injection boundaries for agents), [skill-security-review](../skill-security-review/SKILL.md) (third-party skill supply chain), [incident-response](../incident-response/SKILL.md) (live breach), [production-readiness](../production-readiness/SKILL.md) (launch gate).
