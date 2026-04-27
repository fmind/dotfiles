---
name: use-terraform-cli
description: Guide for the terraform CLI — init/plan/apply discipline, workspaces, state operations, modules, providers, and CI patterns.
---

# Use Terraform CLI

`terraform` is the IaC CLI for HashiCorp Terraform / OpenTofu. The skill bundle `hashicorp/agent-skills` (installed via `install-terraform-skills`) covers HCL authoring; this skill focuses on the **plan/apply discipline** and operational patterns.

The agent should:
1. Always `terraform plan` and surface the diff before any `apply`.
2. Always reuse a `.terraform.lock.hcl` (commit it).
3. Never run `terraform destroy` without explicit user confirmation.

## One-time Setup

```bash
# Install via mise.
mise use -g terraform@1.10

# Or OpenTofu (drop-in replacement).
mise use -g opentofu@latest

# Sanity check.
terraform version
```

## Init & Providers

```bash
# Pull providers + modules into .terraform/.
terraform init

# Refresh after editing required_providers / module sources.
terraform init -upgrade

# Migrate state to a new backend.
terraform init -migrate-state
```

`.terraform.lock.hcl` records exact provider versions — commit it.

## Plan / Apply Loop

```bash
# Read-only diff. Always run before apply.
terraform plan -out=tfplan
terraform plan -out=tfplan -var-file=staging.tfvars

# Apply a saved plan (no surprises — exactly what you reviewed).
terraform apply tfplan

# Targeted apply (use sparingly — bypasses dependency graph).
terraform apply -target=google_cloud_run_service.svc

# Destroy (require explicit confirmation).
terraform plan -destroy -out=destroy.tfplan
terraform apply destroy.tfplan
```

## Workspaces (multi-env state)

```bash
terraform workspace list
terraform workspace new staging
terraform workspace select prod
terraform workspace show
```

Each workspace has its own state file under the same backend. Use `terraform.workspace` in HCL to branch on environment.

## State Inspection & Surgery

```bash
# Inspect.
terraform show                                # human-readable plan/state
terraform show -json | jq                     # machine-readable
terraform state list
terraform state show google_storage_bucket.foo

# Surgery (mutates state — risky).
terraform state mv 'module.foo.bar' 'module.bar.foo'
terraform state rm 'aws_s3_bucket.deleted'
terraform state pull > backup.tfstate         # always back up before surgery
```

## Modules

```bash
# Reference a registry module.
module "vpc" {
  source  = "terraform-google-modules/network/google"
  version = "~> 9.0"
}

# Reference a Git module (with a ref).
module "shared" {
  source = "git::https://github.com/org/repo.git//modules/shared?ref=v1.4.0"
}

# Local refactor.
terraform fmt -recursive
terraform validate
```

## Test (Terraform 1.6+)

```bash
# Native HCL tests live in tests/*.tftest.hcl.
terraform test
terraform test -filter=tests/network.tftest.hcl
```

## Common Workflows

**Bootstrap a new module.**
```bash
mkdir -p modules/network && cd modules/network
cat > main.tf <<'EOF'
terraform { required_version = ">= 1.6.0" }
EOF
terraform init
terraform fmt
terraform validate
```

**Promote staging → prod.**
```bash
terraform workspace select staging
terraform plan -var-file=staging.tfvars -out=staging.tfplan
terraform apply staging.tfplan

terraform workspace select prod
terraform plan -var-file=prod.tfvars -out=prod.tfplan
# Human review of the diff…
terraform apply prod.tfplan
```

**CI gate.**
```bash
terraform fmt -check -recursive
terraform validate
terraform plan -lock-timeout=60s -out=tfplan
# Post the plan summary to PR; gate apply on approval.
```

## MCP Companion

For agent-driven plan/apply lookups, install the HashiCorp Terraform MCP server (Docker-based) via `install-terraform-mcp`. It complements the CLI, doesn't replace it — `terraform apply` still runs locally with your provider creds.

## Important Notes

1. **Never edit state by hand** — use `terraform state` subcommands; back up first.
2. **Commit `.terraform.lock.hcl`** but not `.terraform/` or `*.tfstate`.
3. **Use `-out=tfplan`** for the plan/apply pair so the apply matches exactly what was reviewed.
4. **`-target` is an emergency tool**, not a workflow — it skips the dependency graph and can leave drift.
5. **Treat workspaces as environments**, not feature branches; PR previews belong in a separate state backend or use `terraform-cdk`/Terragrunt.
6. **Never run `terraform destroy` against shared state without a separate, explicit user confirmation.**

## Documentation

- [Terraform CLI overview](https://developer.hashicorp.com/terraform/cli)
- [Terraform language docs](https://developer.hashicorp.com/terraform/language)
- [Backends](https://developer.hashicorp.com/terraform/language/backend)
- [State commands](https://developer.hashicorp.com/terraform/cli/commands/state)
- [`terraform test`](https://developer.hashicorp.com/terraform/language/tests)
- [Google Provider](https://registry.terraform.io/providers/hashicorp/google/latest/docs)
- [HashiCorp Agent Skills (companion bundle)](https://github.com/hashicorp/agent-skills)
- [OpenTofu (Terraform fork)](https://opentofu.org)
