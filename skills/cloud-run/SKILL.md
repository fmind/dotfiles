---
name: cloud-run
description: Deploy container images to Google Cloud Run with Artifact Registry, keyless CI identity, Secret Manager, ko, or Dockerfiles. Use to ship a service or agent to GCP.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/cloud-run
  created: "2026-08-07"
  updated: "2026-09-03"
---

# Cloud Run Deployment

Cloud Run is the default deploy target for web apps, APIs, and agents: serverless containers, scale-to-zero, per-request billing. [containerize](../containerize/SKILL.md) builds the image (`ko` for Go, Dockerfile for Python); this skill owns the registry, identities, deploy command, configuration, and CD job.

## Workflow

1. **Pick the tier**: `gcloud run deploy` for one service; [service.yaml](references/service.yaml) via `gcloud run services replace` once settings accumulate; [terraform-stack](../terraform-stack/SKILL.md) for a fleet.
1. **Set up the project once**: enable the APIs, create the Artifact Registry repository, and create a dedicated runtime service account (never the default compute SA; grant it only what the app reads).

   ```bash
   gcloud services enable run.googleapis.com artifactregistry.googleapis.com iamcredentials.googleapis.com
   gcloud artifacts repositories create <slug> --repository-format=docker --location=<region>
   gcloud iam service-accounts create <slug>-runtime   # runtime identity, least privilege
   ```

1. **Create the keyless CI identity**: one Workload Identity Federation pool and provider per project, plus a deployer service account that GitHub Actions impersonates over OIDC, so no key is ever exported.

   ```bash
   gcloud iam workload-identity-pools create github --location=global
   gcloud iam workload-identity-pools providers create-oidc github-actions \
     --location=global --workload-identity-pool=github \
     --issuer-uri="https://token.actions.githubusercontent.com" \
     --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository" \
     --attribute-condition="assertion.repository_owner=='<owner>'"
   gcloud iam service-accounts create <slug>-deployer
   gcloud iam service-accounts add-iam-policy-binding "<slug>-deployer@<project>.iam.gserviceaccount.com" \
     --role=roles/iam.workloadIdentityUser \
     --member="principalSet://iam.googleapis.com/projects/<project_number>/locations/global/workloadIdentityPools/github/attribute.repository/<owner>/<repo>"
   ```

   The deployer needs `roles/run.admin` and `roles/artifactregistry.writer` on the project plus `roles/iam.serviceAccountUser` on the runtime SA (that grant authorizes deploying as it).

1. **Build and deploy**: the `build:image` task from [containerize](../containerize/SKILL.md) pushes to `KO_DOCKER_REPO=<region>-docker.pkg.dev/<project>/<slug>` and prints the digest reference on its last line.

   ```bash
   IMAGE=$(mise run build:image | tail -n 1)
   gcloud run deploy <slug> --image="$IMAGE" --region=<region> \
     --service-account="<slug>-runtime@<project>.iam.gserviceaccount.com" \
     --no-allow-unauthenticated
   ```

1. **Configure the service**: plain config rides `--set-env-vars=LOG_LEVEL=info`; secrets ride `--set-secrets=API_KEY=api-key:latest`, so the value never enters the spec or the revision history.

   ```bash
   sops -d secrets.enc.yaml | yq -r .api_key | gcloud secrets versions add api-key --data-file=-   # seed from sops-secrets
   ```

1. **Expose a `deploy` task**: run it on demand only, never from a git hook, because a deploy spends money and mutates production (see [mise](../mise/SKILL.md)).

   ```toml
   [tasks.deploy]
   description = "Deploy an image to Cloud Run — pass the digest ref as the argument"
   # mise appends CLI args to the last command: `mise run deploy <image-ref>`.
   run = "gcloud run deploy <slug> --region=<region> --image"
   ```

1. **Wire CD**: copy [deploy.yml](references/deploy.yml) into the `cd.yml` from [github-actions](../github-actions/SKILL.md), set the `GCP_*` and `CLOUDRUN_SERVICE` variables, and opt in with `ENABLE_DEPLOY_CLOUDRUN=true`.

## Gotchas

- **Digest, not tag**: deploy the `@sha256:` reference; `gcloud run services update-traffic` then rolls back to an exact revision.
- **Private by default**: keep `--no-allow-unauthenticated`; grant `roles/run.invoker` to callers or front the service with an authenticating load balancer.
- **Listen on `$PORT`**: Cloud Run injects `PORT` (default 8080); a hardcoded port fails the revision health check.
- **CPU is request-scoped**: background goroutines stall between requests; use `--no-cpu-throttling` for always-on work or a Cloud Run job for batch.
- **Cold starts**: set `--min-instances=1` only when first-hit latency matters more than the idle cost.
- **One region**: the Artifact Registry repository and the service share one region and project (cross-region pulls add latency and egress); WIF pools are global (`--location=global`).

## Official Skills

Upstream: `google/skills` (`skills/cloud`), listed and installed through [google-cloud](../google-cloud/SKILL.md); pick the Cloud Run and CLI guardrail skills.

## Documentation

- [Cloud Run](https://cloud.google.com/run/docs) · [ko](https://ko.build) · [google-github-actions/auth](https://github.com/google-github-actions/auth) · [deploy-cloudrun](https://github.com/google-github-actions/deploy-cloudrun)
- Companion skills: [containerize](../containerize/SKILL.md) (image), [github-actions](../github-actions/SKILL.md) (CD), [sops-secrets](../sops-secrets/SKILL.md) (secrets), [gcloud](../gcloud/SKILL.md) (context), [terraform-stack](../terraform-stack/SKILL.md) (IaC).
