---
name: trivy
description: Run Trivy scans for dependency vulnerabilities, IaC misconfigurations, secrets, licenses, SBOMs, and container images with the shared trivy.yaml. Use for any trivy command.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/trivy
  created: "2026-09-02"
  updated: "2026-09-03"
---

# Trivy

One scanner for the whole repository: dependencies, infrastructure as code, secrets, licenses, container images, and SBOMs. The policy lives in `trivy.yaml` (`HIGH`/`CRITICAL`, `ignore-unfixed`, scanners: vuln, misconfig, secret, license) so local runs and CI report the same findings.

## Commands

```bash
trivy --config trivy.yaml fs .                                    # repository: deps, IaC, secrets, licenses
trivy --config trivy.yaml config .                                # IaC only: Dockerfile, Terraform, Kubernetes, GitHub Actions
trivy --config trivy.yaml image --input tmp/image.tar             # local image tarball, before any push (check:image)
trivy --config trivy.yaml image <registry>/<slug>@<digest>        # pushed image, by immutable digest
trivy --config trivy.yaml image --format cyclonedx -o sbom.json <registry>/<slug>@<digest>
trivy --config trivy.yaml fs --tf-vars terraform.example.tfvars . # OpenTofu modules need variables to evaluate
```

## Mise Task

Expose the repository scan as `check:scan` in `mise.toml` per [mise](../mise/SKILL.md); `mise run check` then runs it locally and in CI:

```toml
[tasks."check:scan"]
description = "Scan dependencies, IaC, licenses, and secrets (Trivy)"
run = "trivy --config trivy.yaml fs ." # mise appends extra path arguments
```

Language-native scanners stay separate as `check:vuln` (`govulncheck`, `uv audit`, `pnpm audit`) because they know the lockfile semantics better; keep both tasks, `trivy fs` adds IaC, secrets, and licenses on top.

## Triage

1. Group findings by severity, then split fixable from `unfixed`.
1. Prefer the minimal upgrade of the affected dependency or base image; re-run the scan to prove the fix.
1. Record an accepted risk in `.trivyignore` (one CVE or path per line, with a `#` reason) instead of lowering the global severity bar.
1. Treat a secret finding as compromised: rotate it, then clean the history per [gitleaks](../gitleaks/SKILL.md).

## Gotchas

- **Always pass `--config trivy.yaml`**: precedence is `--config` > `TRIVY_CONFIG` > `./trivy.yaml`, and the owner's shell exports a global `TRIVY_CONFIG`, so a bare `trivy fs .` silently uses the global policy.
- **Commit a project `trivy.yaml`**: copy the global one so CI, hooks, and agents share one policy.
- **Scan digests, not tags**: a tag can move after the scan; the digest is what ships.
- **Databases update on first run**: a scan needs network access once per day for the vulnerability database; use `--skip-db-update` in offline reruns.

## Documentation

- [Trivy](https://trivy.dev)
- Companion skills: [secure](../secure/SKILL.md) (the repository checklist), [containerize](../containerize/SKILL.md) (image scans), [github-actions](../github-actions/SKILL.md) (`security.yml` scheduled scan).
