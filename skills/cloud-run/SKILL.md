---
name: cloud-run
description: Deploy container apps to Google Cloud Run — ko or Dockerfile images to Artifact Registry, keyless CI via Workload Identity Federation, Secret Manager wiring. Use when shipping a service, web app, or ADK agent to GCP.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/cloud-run
  created: 2026-08-07
  updated: 2026-08-07
---

# Cloud Run Deployment Standard

Canonical path from a container image to a running GCP service: **Cloud Run** is the default deploy target for web apps, APIs, and ADK agents — serverless containers, scale-to-zero, per-request billing. The image comes from the [containerize skill](../containerize/SKILL.md) (`ko` for Go, Dockerfile for Python); this skill covers registry, identity, deploy, and CD.

## 1. Deployment Ladder

1. **`gcloud run deploy` (default)**: one command, ideal for a single service iterated from the CLI.
1. **Declarative manifest**: [service.yaml](references/service.yaml) applied with `gcloud run services replace` — reviewable in PRs once settings accumulate (scaling, secrets, resources).
1. **Full IaC**: `google_cloud_run_v2_service` in the [terraform-stack skill](../terraform-stack/SKILL.md) — when the service is one resource among many (domains, IAM, schedulers).

## 2. One-Time Project Setup

```bash
gcloud services enable run.googleapis.com artifactregistry.googleapis.com iamcredentials.googleapis.com
gcloud artifacts repositories create <slug> --repository-format=docker --location=<region>
gcloud iam service-accounts create <slug>-runtime   # runtime identity, least privilege
```

- **Runtime identity**: always deploy with a dedicated service account; grant it only what the app reads (e.g. `roles/secretmanager.secretAccessor` on specific secrets) — never run on the default compute SA.
- **CI identity (keyless)**: create a Workload Identity Federation pool + provider once per project, then a deployer SA that GitHub Actions impersonates via OIDC — no exported keys:

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

The deployer needs `roles/run.admin` and `roles/artifactregistry.writer` on the project, plus `roles/iam.serviceAccountUser` on the **runtime** SA (that last grant is what authorizes deploying _as_ it).

## 3. Deploy Workflow

```bash
export KO_DOCKER_REPO=<region>-docker.pkg.dev/<project>/<slug>
IMAGE=$(go tool ko build ./cmd/<slug> --bare)   # builds, pushes, prints the digest ref
gcloud run deploy <slug> --image="$IMAGE" --region=<region> \
  --service-account="<slug>-runtime@<project>.iam.gserviceaccount.com" \
  --no-allow-unauthenticated
```

- **Digest, not tag**: `ko` outputs an `@sha256:` reference — deploys are immutable and `gcloud run services update-traffic` rolls back to an exact revision.
- **Private by default**: keep `--no-allow-unauthenticated` unless the service is deliberately public; front private services with IAM (`roles/run.invoker`) or an authenticating load balancer.
- **Python / Dockerfile apps**: build and push per the [containerize skill](../containerize/SKILL.md), then the same `gcloud run deploy` with the pushed digest.
- **Mise task**: expose it as an explicit `deploy` task next to the canonical vocabulary of the [mise skill](../mise/SKILL.md) — run on demand only, never from a git hook, since deploys spend money and mutate prod:

```toml
[tasks.deploy]
description = "Deploy an image to Cloud Run — pass the digest ref as the argument"
# mise auto-appends CLI args to the last command: `mise run deploy <image-ref>`.
run = "gcloud run deploy <slug> --region=<region> --image"
```

## 4. Configuration & Secrets

- **Plain config** rides env vars (`--set-env-vars=LOG_LEVEL=info`), matching the typed config packages of the [go-stack](../go-stack/SKILL.md) and [python-stack](../python-stack/SKILL.md).
- **Secrets** ride Secret Manager references (`--set-secrets=API_KEY=api-key:latest`) — the value never appears in the service spec, revision history, or console. Seed them from the repo's encrypted files per the [sops-secrets skill](../sops-secrets/SKILL.md): `sops -d secrets.enc.yaml | yq -r .api_key | gcloud secrets versions add api-key --data-file=-`.
- **Observability**: Cloud Run captures stdout/stderr as structured logs — the JSON `slog` handlers from the stacks are parsed automatically; OpenTelemetry traces flow with `OTEL_EXPORTER_OTLP_ENDPOINT`, and ADK agents can pass `--otel_to_cloud` for Cloud Trace.

## 5. Continuous Deployment

Copy [deploy.yml](references/deploy.yml) (or append its job to the existing `cd.yml` from the [github-actions skill](../github-actions/SKILL.md)): tag push → `mise run build:image` (ko) → `deploy-cloudrun@v3` with the digest, authenticated by `auth@v3` over WIF. Configure the `GCP_*`/`CLOUDRUN_SERVICE` repo variables and opt in with `ENABLE_DEPLOY_CLOUDRUN=true`, mirroring the other CD jobs.

## 6. Gotchas

- **Listen on `$PORT`**: Cloud Run injects `PORT` (default 8080) and routes traffic only there — the stacks' config packages already read it; a hardcoded port fails the revision health check.
- **CPU is request-scoped**: outside a request the CPU is throttled — background goroutines stall. Use `--no-cpu-throttling` for always-on work, or a Cloud Run **job** for batch.
- **Cold starts**: scale-to-zero means first-hit latency; set `--min-instances=1` only when latency matters more than the idle cost.
- **Region is triple-pinned**: Artifact Registry repo, service region, and WIF resources should share one region/project — cross-region pulls add latency and egress cost.

## Documentation

- [Cloud Run](https://cloud.google.com/run/docs) · [ko](https://ko.build) · [google-github-actions/auth](https://github.com/google-github-actions/auth) · [deploy-cloudrun](https://github.com/google-github-actions/deploy-cloudrun)
- Companion skills:
  - [containerize](../containerize/SKILL.md) — builds the image this skill ships.
  - [github-actions](../github-actions/SKILL.md) — the CD pipeline this job extends.
  - [sops-secrets](../sops-secrets/SKILL.md) — seeds Secret Manager from the repo.
  - [terraform-stack](../terraform-stack/SKILL.md) — full-IaC alternative for fleets.
