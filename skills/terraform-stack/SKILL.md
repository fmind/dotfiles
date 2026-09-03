---
name: terraform-stack
description: "Canonical infrastructure-as-code stack: OpenTofu-first with tflint, trivy config scans, terraform-docs, native tests, GCS state. Use for any Terraform or OpenTofu work."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/terraform-stack
  created: "2026-08-07"
  updated: "2026-09-03"
---

# Terraform / OpenTofu Stack Standard

Canonical infrastructure as code with OpenTofu (the open-source Terraform fork; the binary is `tofu`). Reserve HashiCorp `terraform` for repositories that require a BSL-licensed feature or an employer mandate, and mention the deviation.

## 1. Core Stack

- **Engine**: OpenTofu via mise (`opentofu` tool, `tofu` binary); every `terraform {}` block and `.tf` file works unchanged, and OpenTofu adds client-side state encryption.
- **Tasks and hooks**: [mise.toml](references/mise.toml) exposes the canonical vocabulary per [mise](../mise/SKILL.md) — `check` fans out to format, validate, lint (tflint), scan (trivy), and leaks; [lefthook.yml](references/lefthook.yml) wires the hooks per [lefthook](../lefthook/SKILL.md).
- **Docs**: `terraform-docs` injects the inputs/outputs table into `README.md` between `<!-- BEGIN_TF_DOCS -->` / `<!-- END_TF_DOCS -->` markers, configured by [terraform-docs.yml](references/terraform-docs.yml).
- **Apply is manual**: `build` produces `tmp/plan.tfplan`; applying it is a deliberate human step (`tofu apply tmp/plan.tfplan`), never a task or hook.

## 2. Project Scaffolding Workflow

1. **Information**: define the project `Slug`, GCP `Project ID`, and default `Region`.
1. **Config files**:
   - [mise.toml](references/mise.toml) and [lefthook.yml](references/lefthook.yml).
   - `.tflint.hcl` from [tflint.hcl](references/tflint.hcl) — pins the terraform preset and the GCP ruleset release.
   - `.terraform-docs.yml` from [terraform-docs.yml](references/terraform-docs.yml), plus the `TF_DOCS` markers in `README.md`.
   - `dprint.json` per [dprint](../dprint/SKILL.md), `.gitignore` from [gitignore](references/gitignore), `LICENSE` per [project-license](../project-license/SKILL.md).
1. **Sources** (flat root module; no `modules/` tree until a unit is reused):
   - [versions.tf](references/versions.tf) — version constraints, provider pins, and the commented GCS backend and encryption blocks.
   - [main.tf](references/main.tf), [variables.tf](references/variables.tf) (typed, validated inputs), [outputs.tf](references/outputs.tf).
   - [terraform.example.tfvars](references/terraform.example.tfvars) — non-secret example and static-scan values; replace the project ID before planning.
   - `tests/main.tftest.hcl` from [main.tftest.hcl](references/main.tftest.hcl) — plan-only native tests.
1. **Validate**: `git init --initial-branch=main`, then `mise run install`, `mise run format`, `mise run check`, `mise run test` — green without cloud access because the backend ships commented and tests run in plan mode.
1. **Lock providers**: commit `.terraform.lock.hcl`; on multi-platform teams run `tofu providers lock -platform=linux_amd64 -platform=darwin_arm64`.
1. **Promote the backend**: once real state exists, create the versioned GCS bucket (commands in [versions.tf](references/versions.tf)), uncomment `backend "gcs"`, and re-run `mise run install`; `tofu init` migrates local state after a prompt.

## 3. State & Secrets

- **State is secret**: state stores every attribute in plaintext — never in git, always in the versioned GCS bucket, ideally wrapped by OpenTofu's `encryption` block (GCP KMS, sketched in [versions.tf](references/versions.tf)).
- **Variable files**: `*.tfvars` is gitignored; commit only `*.example.tfvars`. Feed secrets at plan time per [sops-secrets](../sops-secrets/SKILL.md): `sops exec-env secrets.enc.env 'tofu plan -out=tmp/plan.tfplan'` with `TF_VAR_<name>` entries.
- **Credentials**: Application Default Credentials locally; Workload Identity Federation in CI per [github-actions](../github-actions/SKILL.md) — no service-account keys.

## 4. Testing Standard

- **Native tests first**: `tofu test` over `tests/*.tftest.hcl` with `command = plan` asserts on planned attributes ([main.tftest.hcl](references/main.tftest.hcl)) — deterministic and credential-free.
- **Apply-mode tests**: `command = apply` creates real (then destroyed) resources; reserve it for behavior a plan cannot prove, with approved project access and cost.
- **Policy checks**: encode "must never happen" rules (public buckets, missing labels) as `check:scan` trivy findings or plan-test assertions.

## Gotchas

- **`tofu init` before everything**: `check:validate`, `check:lint`, and `test` need providers downloaded; a fresh clone runs `mise run install` first.
- **tflint rulesets are downloaded**: `tflint --init` (in `install:lint`) fetches the plugins pinned in [tflint.hcl](references/tflint.hcl); export `GITHUB_TOKEN` in CI to avoid API rate limits.
- **terraform-docs needs markers**: `inject` mode only rewrites between the `TF_DOCS` markers; add them to `README.md` once at scaffold time.
- **Provider majors move fast**: [versions.tf](references/versions.tf) pins the google provider to one major with `~>`; bump majors deliberately with [upgrade-tools](../upgrade-tools/SKILL.md) and read the upgrade guide, because resources rename across majors.
- **One state per concern**: several small root modules (one per GCS `prefix`) beat one monolithic state — smaller blast radius, faster plans.

## Official Skills

Upstream: `hashicorp/agent-skills`. List the current release, then install what the task needs at project scope after reviewing the snapshot (see [agent-skills](../agent-skills/SKILL.md)):

```bash
skills add hashicorp/agent-skills --list
skills add hashicorp/agent-skills --skill <name> -y
```

## Documentation

- [OpenTofu](https://opentofu.org/docs/) · [tflint](https://github.com/terraform-linters/tflint) · [trivy config](https://trivy.dev/latest/docs/scanner/misconfiguration/) · [terraform-docs](https://terraform-docs.io) · [State encryption](https://opentofu.org/docs/language/state/encryption/)
- Companion skills: [mise](../mise/SKILL.md) (task vocabulary), [sops-secrets](../sops-secrets/SKILL.md) (encrypted variables), [github-actions](../github-actions/SKILL.md) (CI), [secure](../secure/SKILL.md) (full-repo scans), [google-cloud](../google-cloud/SKILL.md) (product skills).
