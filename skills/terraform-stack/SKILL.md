---
name: terraform-stack
description: Canonical infrastructure-as-code stack — OpenTofu-first with tflint, trivy config scans, terraform-docs, native tests, and GCS state. Use for any Terraform or OpenTofu work.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/terraform-stack
  created: 2026-08-07
  updated: 2026-08-08
---

# Terraform / OpenTofu Stack Standard (OpenTofu 1.12+)

Canonical guidelines for infrastructure as code with **OpenTofu** (the open-source Terraform fork and default engine — the CLI binary is `tofu`). Reserve HashiCorp `terraform` for repositories that require a BSL-licensed feature or an employer mandate, and mention the deviation.

## 1. Core Stack

- **Engine**: OpenTofu 1.12+ via mise (`opentofu` tool, `tofu` binary). Every `terraform {}` block and `.tf` file works unchanged; OpenTofu adds client-side state encryption on top.
- **Task Runner & Hooks**: `mise.toml` ([mise.toml](references/mise.toml)) exposes the canonical vocabulary per the [mise skill](../mise/SKILL.md) — `install` (init + tflint rulesets + hooks), `format` (`tofu fmt` + dprint), `check` (fans out below), `test` (`tofu test`), `build` (plan artifact), `build:docs` (terraform-docs). `lefthook.yml` ([lefthook.yml](references/lefthook.yml)) wires pre-commit (format → leaks → check) and pre-push (test) per the [lefthook skill](../lefthook/SKILL.md).
- **Check Fan-Out** (all static, parallel, cloud-free): `check:format` (`tofu fmt -check` + dprint), `check:validate` (`tofu validate`), `check:lint` (`tflint --recursive` with the pinned rulesets in [tflint.hcl](references/tflint.hcl)), `check:scan` (`trivy config` for misconfigurations, per the vocabulary in the [mise skill](../mise/SKILL.md)), `check:leaks` (gitleaks).
- **Docs**: `terraform-docs` injects the module's inputs/outputs table into `README.md` between `<!-- BEGIN_TF_DOCS -->` / `<!-- END_TF_DOCS -->` markers, configured by [terraform-docs.yml](references/terraform-docs.yml).
- **Apply Is Manual**: `build` produces `tmp/plan.tfplan`; applying it is a deliberate human step (`tofu apply tmp/plan.tfplan`), never a task or hook — plans mutate infrastructure and money.

## 2. Project Scaffolding Workflow

1. **Information**: Define the project `Slug`, GCP `Project ID`, and default `Region`.
1. **Config Initialization**: Copy and customize:
   - `mise.toml` ([mise.toml](references/mise.toml)) and `lefthook.yml` ([lefthook.yml](references/lefthook.yml)).
   - `.tflint.hcl` ([tflint.hcl](references/tflint.hcl)) — pins the terraform preset and the GCP ruleset release.
   - `.terraform-docs.yml` ([terraform-docs.yml](references/terraform-docs.yml)) and add the `TF_DOCS` markers to `README.md`.
   - `dprint.json` per the [dprint skill](../dprint/SKILL.md), `.gitignore` ([gitignore](references/gitignore)), `LICENSE` per the [project-license skill](../project-license/SKILL.md).
1. **Scaffold Sources** (flat root module — no `modules/` tree until a unit is reused):
   - `versions.tf` ([versions.tf](references/versions.tf)) — version constraints, provider pins, and the commented GCS backend + encryption blocks.
   - `main.tf` ([main.tf](references/main.tf)), `variables.tf` ([variables.tf](references/variables.tf)) (typed, validated inputs), `outputs.tf` ([outputs.tf](references/outputs.tf)).
   - `terraform.example.tfvars` ([terraform.example.tfvars](references/terraform.example.tfvars)) — non-secret example and static-scan values; replace the project ID before planning.
   - `tests/main.tftest.hcl` ([main.tftest.hcl](references/main.tftest.hcl)) — plan-only native tests.
1. **Git & Validation**: `git init --initial-branch=main`, then `mise run install`, `format`, `check`, `test` — all green without any cloud access, because the backend ships commented and tests run in plan mode. Commit `.terraform.lock.hcl` (provider pins); on multi-platform teams run `tofu providers lock -platform=linux_amd64 -platform=darwin_arm64` so the lockfile carries hashes for both.
1. **Backend Promotion**: Once real state exists, create the versioned GCS bucket (commands inline in [versions.tf](references/versions.tf)), uncomment the `backend "gcs"` block, and re-run `mise run install` — `tofu init` migrates local state after an explicit prompt.

## 3. State & Secrets

- **State Is Secret**: State files store every attribute in plaintext — never in git ([gitignore](references/gitignore) blocks them), always in the versioned GCS bucket, ideally wrapped by OpenTofu's client-side `encryption` block (GCP KMS key provider, sketched in [versions.tf](references/versions.tf)).
- **Variable Files**: `*.tfvars` is gitignored by default; commit only `*.example.tfvars`. Feed secret variables at plan time from a [sops-secrets](../sops-secrets/SKILL.md) file — `sops exec-env secrets.enc.env 'tofu plan -out=tmp/plan.tfplan'` with `TF_VAR_<name>` entries — so plaintext never touches disk.
- **Credentials**: Application Default Credentials locally (`gcloud auth application-default login`); Workload Identity Federation in CI (see the [github-actions skill](../github-actions/SKILL.md)) — no service-account keys.

## 4. Testing Standard

- **Native Tests First**: `tofu test` over `tests/*.tftest.hcl` with `command = plan` asserts on planned attributes ([main.tftest.hcl](references/main.tftest.hcl)) — deterministic, credential-free, and fast, matching the global testing standard.
- **Apply-Mode Tests**: `command = apply` runs create real (then destroyed) resources — reserve for modules whose behavior a plan cannot prove, and only with explicitly approved project access and cost.
- **Policy Checks**: Encode "must never happen" rules (public buckets, missing labels) as `check:scan` trivy findings or plan-test assertions, not code review memory.

## 5. Gotchas

- **`tofu init` Before Everything**: `check:validate`, `check:lint` (deep checks), and `test` all need providers downloaded — a fresh clone must run `mise run install` first (`run_auto_install = false` makes missing tools fail fast, per the [mise skill](../mise/SKILL.md)).
- **tflint Rulesets Are Installed, Not Bundled**: `tflint --init` (wired into `install:lint`) downloads the plugins pinned in [tflint.hcl](references/tflint.hcl); CI needs `GITHUB_TOKEN` exported to avoid GitHub API rate limits on that download.
- **terraform-docs Needs Markers**: `inject` mode only rewrites between the `TF_DOCS` comment markers — add them to `README.md` once at scaffold time.
- **Provider Majors Move Fast**: The google provider pins `~> 7.0` in [versions.tf](references/versions.tf); bump majors deliberately with the [upgrade-tools skill](../upgrade-tools/SKILL.md) and read the upgrade guide — resources rename across majors.
- **One State Per Concern**: Prefer several small root modules (one per `prefix` in the GCS backend) over one monolithic state — smaller blast radius, faster plans.

## Documentation

- [OpenTofu](https://opentofu.org/docs/) · [tflint](https://github.com/terraform-linters/tflint) · [trivy config](https://trivy.dev/latest/docs/scanner/misconfiguration/) · [terraform-docs](https://terraform-docs.io) · [OpenTofu state encryption](https://opentofu.org/docs/language/state/encryption/)
- Companion skills:
  - [mise](../mise/SKILL.md) — task vocabulary this stack implements.
  - [sops-secrets](../sops-secrets/SKILL.md) — encrypted variable files.
  - [github-actions](../github-actions/SKILL.md) — CI running the same tasks.
  - [security-scan](../security-scan/SKILL.md) — full-repo trivy + gitleaks sweeps.
