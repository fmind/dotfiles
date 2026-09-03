---
name: secure
description: "Run the repository security pass: secret leaks, dependency and IaC scans, workflow audit, image signing, encrypted secrets, threat model. Use for a security review or checklist."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/secure
  created: "2026-07-04"
  updated: "2026-09-03"
---

# Secure a Repository

One ordered checklist that composes the tool skills; each step names the skill that owns the detail, this file only decides the order and the gate.

## Checklist

1. **Secrets in git**: run the full-history scan and wire the staged pre-commit hook per [gitleaks](../gitleaks/SKILL.md); rotate anything found before touching anything else.
1. **Secrets at rest**: move plaintext credentials into environment variables or encrypted `*.enc.*` files per [sops-secrets](../sops-secrets/SKILL.md); runtime secrets come from Secret Manager per [cloud-run](../cloud-run/SKILL.md).
1. **Dependencies, IaC, licenses**: run `check:vuln` (native scanner) and `check:scan` per [trivy](../trivy/SKILL.md); fix or justify every `HIGH`/`CRITICAL`.
1. **Code patterns (opt-in)**: adopt `check:sast` per [opengrep](../opengrep/SKILL.md) when the project handles untrusted input, authentication, or agents with tools.
1. **Workflows**: run `check:actions` and fix findings per [zizmor](../zizmor/SKILL.md): least-privilege `permissions`, no template injection, pinned actions, `persist-credentials: false`.
1. **Automated updates**: enable [dependabot](../dependabot/SKILL.md) for every ecosystem so the scans above stay green without manual bumps.
1. **Supply chain**: for shipped images, scan, sign, and attest the digest per [containerize](../containerize/SKILL.md) and [cosign](../cosign/SKILL.md); verification pins identity and issuer.
1. **Runtime posture**: private services by default (`--no-allow-unauthenticated`), dedicated runtime service accounts, keyless CI identity via Workload Identity Federation per [cloud-run](../cloud-run/SKILL.md).
1. **Threat model**: when the project handles authentication, personal data, agents with tools, or public exposure, run [threat-model](../threat-model/SKILL.md); scanners do not find design flaws.

## Gate

Every offline scan lives in `mise run check` so hooks and CI share it: `check:leaks`, `check:vuln`, `check:scan`, `check:actions`, and `check:sast` once adopted (names from [mise](../mise/SKILL.md)). Full-history and image scans run in the scheduled `security.yml` per [github-actions](../github-actions/SKILL.md).

## Report

- List findings by severity with the fix applied or the justified ignore (`.trivyignore`, `.gitleaks.toml`, `.semgrepignore` or `nosemgrep`, `.github/zizmor.yml`), each with a reason.
- State what each proof covers: a green local gate proves the checkout and recent history; only the scheduled workflow proves full history, and only a verified signature proves the shipped image.
- Never describe a suppressed finding as fixed.

## Documentation

- [OpenSSF Scorecard](https://securityscorecards.dev) is the checklist behind these steps; the one deliberate deviation is major-tag pinning of actions instead of hash pins (see [zizmor](../zizmor/SKILL.md)).
- Third-party audit skills: `skills add trailofbits/skills --list` (supply-chain and agentic-actions auditors; CC-BY-SA, install at project scope only) and `gh skill preview cli/cli dependabot-triager --allow-hidden-dirs`.
- Companion skills: [skill-security-review](../skill-security-review/SKILL.md) (vet third-party skills before installing), [incident-response](../incident-response/SKILL.md) (when something already leaked).
