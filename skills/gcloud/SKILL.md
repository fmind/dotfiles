---
name: gcloud
description: Operate Google Cloud with gcloud by pinning account, configuration, project, and billing context for IAM, API, log, and audit work. Use for any gcloud call outside a deployment.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/gcloud
  created: "2026-08-30"
  updated: "2026-09-03"
---

# Google Cloud CLI

Use `gcloud` for bounded account, project, IAM, API, billing, logging, and audit operations. [cloud-run](../cloud-run/SKILL.md) owns deployment, [terraform-stack](../terraform-stack/SKILL.md) owns provisioned infrastructure, and [incident-response](../incident-response/SKILL.md) owns a live outage.

## Workflow

1. **Resolve identity and scope**: inspect the named configuration, account, project, and billing project; never activate another configuration just to make a command work.

   ```bash
   gcloud config configurations list --format=json
   gcloud config configurations describe <configuration> --format=json
   gcloud config get auth/impersonate_service_account --configuration <configuration>
   gcloud config get auth/access_token_file --configuration <configuration>
   gcloud projects describe <project-id> --format=json
   ```

1. **Pin every consequential call**: pass `--configuration`, `--account`, `--project`, and `--billing-project` (plus `--impersonate-service-account` for an approved chain) so terminal defaults cannot redirect the operation.
1. **Start read-only**: describe the resource, IAM policy, enabled services, billing linkage, quotas, and logs, bounded by project, resource, and time window.
1. **Plan the mutation**: state the resource, before and after state, permissions, cost or quota impact, rollback, and verification command; API enablement, IAM, billing, deletion, and production changes need explicit authority.
1. **Apply minimally and verify**: change only the named resource, then re-read it and its operation or audit status; separate local configuration, accepted request, completed operation, and user-visible outcome.

## Gotchas

- **`--quiet` is not authority**: it suppresses prompts; it neither makes an operation safe nor approves it.
- **Overrides replace the account**: `CLOUDSDK_AUTH_IMPERSONATE_SERVICE_ACCOUNT`, `CLOUDSDK_AUTH_ACCESS_TOKEN_FILE`, and `--impersonate-service-account` change the effective principal; prove which is active first.
- **Failures are findings**: on an auth, permission, or quota error report the missing principal, denied permission, or exhausted quota; do not switch configurations, broaden IAM, create keys, or move spend to another project.

## Official Skills

Upstream: `google/skills` (`skills/cloud`), listed and installed through [google-cloud](../google-cloud/SKILL.md); its CLI guardrail skill applies to every `gcloud` call.

## Documentation

- [gcloud reference](https://cloud.google.com/sdk/gcloud/reference) · [Authorize the gcloud CLI](https://cloud.google.com/sdk/docs/authorizing)
- Companion skills: [google-cloud](../google-cloud/SKILL.md) (which upstream skill), [cloud-run](../cloud-run/SKILL.md) (deploy), [terraform-stack](../terraform-stack/SKILL.md) (provision), [incident-response](../incident-response/SKILL.md) (outage).
