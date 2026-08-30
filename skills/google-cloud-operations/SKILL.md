---
name: google-cloud-operations
description: "Operate Google Cloud with gcloud: pin account, configuration, project, and billing context for IAM, APIs, logs, and audits outside deployment."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/google-cloud-operations
  created: 2026-08-30
  updated: 2026-08-30
---

# Operate Google Cloud

Use `gcloud` for bounded account, project, IAM, API, billing, logging, and audit operations. Use [cloud-run](../cloud-run/SKILL.md) for application deployment, [terraform-stack](../terraform-stack/SKILL.md) for provisioned infrastructure, and [incident-response](../incident-response/SKILL.md) for a live outage.

## Workflow

1. **Resolve identity and scope:** Inspect the active named configuration, account, project, and quota or billing project. Do not activate another configuration or change defaults merely to make a command work.

   ```bash
   gcloud config configurations list --format=json
   gcloud config configurations describe <configuration> --format=json
   gcloud config get auth/impersonate_service_account --configuration <configuration>
   gcloud config get auth/access_token_file --configuration <configuration>
   gcloud projects describe <project-id> --format=json
   ```

1. **Resolve the effective principal:** Check `CLOUDSDK_AUTH_IMPERSONATE_SERVICE_ACCOUNT`, `CLOUDSDK_AUTH_ACCESS_TOKEN_FILE`, and invocation flags as well as the named configuration. Impersonation and access-token-file overrides replace the selected account; before consequential calls, pin the intended impersonation chain explicitly or prove that neither override is active.
1. **Pin every consequential call:** Pass `--configuration`, `--account`, `--project`, and `--billing-project` when applicable so terminal defaults cannot redirect the operation. Pass `--impersonate-service-account` when the approved effective principal is an impersonation chain. Prefer a named configuration per environment.
1. **Start read-only:** Describe the resource, current IAM policy, enabled services, billing linkage, relevant logs, quotas, and dependencies. Bound list and log queries by project, resource, and time window.
1. **Plan the mutation:** State the exact resource, before and after state, permissions, cost or quota impact, dependent services, rollback, and verification command. Enabling APIs, changing IAM or billing, deleting resources, and production operations require explicit authority.
1. **Use short-lived identity:** Prefer user OAuth, Workload Identity Federation, or explicit service-account impersonation over downloaded keys. Never print access tokens, credential files, or debug HTTP headers.
1. **Apply minimally:** Change only the named resource. `--quiet` suppresses prompts; it does not make an operation safe or authorize it.
1. **Verify independently:** Re-read the resource and its audit or operation status, then confirm behavior at the intended boundary. Separate local configuration, accepted API request, completed operation, and user-visible outcome.

## Failure Rules

- On authentication failure, report the missing effective principal or credential boundary; do not switch profiles, broaden IAM, or create keys automatically.
- On permission failure, identify the denied permission and target resource before proposing the narrowest role.
- On quota or billing failure, stop before enabling spend or moving the charge to another project.
