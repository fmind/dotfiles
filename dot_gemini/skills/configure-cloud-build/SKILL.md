---
name: configure-cloud-build
description: Guide for authoring cloudbuild.yaml — build steps, substitutions, secrets, artifacts, triggers, and pushing images to Artifact Registry.
---

# Configure Cloud Build (`cloudbuild.yaml`)

Cloud Build runs serverless CI/CD on Google Cloud — each build is a sequence of containerized steps with shared `/workspace`, `/builder`, and substitution variables.

## Minimal `cloudbuild.yaml`

```yaml
steps:
  - name: gcr.io/cloud-builders/docker
    args:
      - build
      - -t
      - $LOCATION-docker.pkg.dev/$PROJECT_ID/$_REPO/$_IMAGE:$SHORT_SHA
      - .

images:
  - $LOCATION-docker.pkg.dev/$PROJECT_ID/$_REPO/$_IMAGE:$SHORT_SHA

options:
  logging: CLOUD_LOGGING_ONLY
```

Run locally / one-off:

```bash
gcloud builds submit --config=cloudbuild.yaml \
  --substitutions=_REPO=my-repo,_IMAGE=my-app \
  --region=us-central1 .
```

## Built-in Substitutions

| Variable | Value |
|----------|-------|
| `$PROJECT_ID` | Active project |
| `$BUILD_ID` | Unique build ID |
| `$LOCATION` | Build region |
| `$SHORT_SHA` | First 7 chars of commit SHA (trigger only) |
| `$COMMIT_SHA`, `$REVISION_ID` | Full SHA |
| `$BRANCH_NAME`, `$TAG_NAME` | Source ref (trigger only) |
| `$REPO_NAME` | Source repo (trigger only) |

User-defined substitutions start with `_` (e.g. `$_REPO`).

## Realistic Multi-step Pipeline (Cloud Run)

```yaml
substitutions:
  _REGION: us-central1
  _REPO: app-repo
  _SERVICE: api

steps:
  # 1. Pull build cache (best-effort — don't fail the build).
  - id: pull-cache
    name: gcr.io/cloud-builders/docker
    entrypoint: bash
    args:
      - -c
      - |
        docker pull $_REGION-docker.pkg.dev/$PROJECT_ID/$_REPO/$_SERVICE:cache || true

  # 2. Build with buildkit + multi-tag (cache + sha + latest).
  - id: build
    name: gcr.io/cloud-builders/docker
    args:
      - build
      - --tag=$_REGION-docker.pkg.dev/$PROJECT_ID/$_REPO/$_SERVICE:$SHORT_SHA
      - --tag=$_REGION-docker.pkg.dev/$PROJECT_ID/$_REPO/$_SERVICE:latest
      - --cache-from=$_REGION-docker.pkg.dev/$PROJECT_ID/$_REPO/$_SERVICE:cache
      - --build-arg=BUILDKIT_INLINE_CACHE=1
      - .
    env: ["DOCKER_BUILDKIT=1"]

  # 3. Push all tags.
  - id: push
    name: gcr.io/cloud-builders/docker
    args:
      - push
      - --all-tags
      - $_REGION-docker.pkg.dev/$PROJECT_ID/$_REPO/$_SERVICE

  # 4. Deploy to Cloud Run.
  - id: deploy
    name: gcr.io/google.com/cloudsdktool/cloud-sdk
    entrypoint: gcloud
    args:
      - run
      - deploy
      - $_SERVICE
      - --image=$_REGION-docker.pkg.dev/$PROJECT_ID/$_REPO/$_SERVICE:$SHORT_SHA
      - --region=$_REGION
      - --project=$PROJECT_ID
      - --platform=managed

images:
  - $_REGION-docker.pkg.dev/$PROJECT_ID/$_REPO/$_SERVICE:$SHORT_SHA
  - $_REGION-docker.pkg.dev/$PROJECT_ID/$_REPO/$_SERVICE:latest

options:
  machineType: E2_HIGHCPU_8
  logging: CLOUD_LOGGING_ONLY
  dynamic_substitutions: true

timeout: 1200s
```

## Secrets (Secret Manager)

```yaml
availableSecrets:
  secretManager:
    - versionName: projects/$PROJECT_ID/secrets/db-password/versions/latest
      env: DB_PASSWORD

steps:
  - name: gcr.io/cloud-builders/docker
    entrypoint: bash
    args: ["-c", "echo Connecting with $$DB_PASSWORD..."]
    secretEnv: [DB_PASSWORD]
```

Note the `$$VAR` escape inside shell — `$VAR` is consumed by Cloud Build's substitution layer first.

## Artifact Outputs

```yaml
artifacts:
  objects:
    location: gs://$PROJECT_ID-build-artifacts/$BUILD_ID
    paths: ["dist/**", "coverage/**"]
```

## Build Triggers

Defined separately (Console, Terraform, or `gcloud beta builds triggers create`). Triggers fire `cloudbuild.yaml` on push, PR, or schedule.

```bash
# Create a push trigger from a 2nd-gen Connection.
gcloud beta builds triggers create github \
  --name=api-on-main \
  --repo-owner=ACME --repo-name=api \
  --branch-pattern="^main$" \
  --build-config=cloudbuild.yaml \
  --region=us-central1
```

## Custom Builders & Caching

For builders not in `gcr.io/cloud-builders/*`:

```yaml
- name: us-central1-docker.pkg.dev/$PROJECT_ID/builders/playwright:latest
  args: [npm, test]
```

Build the custom builder image once, push to Artifact Registry, reuse forever.

## Provenance / SLSA

```yaml
options:
  requestedVerifyOption: VERIFIED   # build attestation written to AR
```

## Local Validation

```bash
# Lint cloudbuild.yaml.
gcloud builds submit --config=cloudbuild.yaml --no-source --dry-run=true

# Run a build with local source (no commit needed).
gcloud builds submit --config=cloudbuild.yaml --region=us-central1 .
```

## Common Workflows

**Bootstrap a Cloud Run deploy pipeline.**

1. Add `Dockerfile` and `cloudbuild.yaml` (see "Realistic" above).
2. `gcloud artifacts repositories create app-repo --repository-format=docker --location=us-central1 --project=$PROJECT`.
3. Grant the Cloud Build SA Cloud Run + Artifact Registry roles.
4. Create a trigger; push to `main` to fire the first build.

**Speed up a slow build.**

- Add a cache pull step + `--cache-from` (see above).
- Bump `options.machineType` (E2_HIGHCPU_8/32).
- Trim `.gcloudignore` so the upload stays small.
- Split into parallel steps with `waitFor: ["-"]`.

## Important Notes

1. **`gcr.io/cloud-builders/*`** images are maintained by Google but lag latest CLI versions — for cutting-edge tools, build a custom builder.
2. **Substitution syntax differs from shell**: `$VAR` is Cloud Build's substitution; `$$VAR` reaches the shell.
3. **The Cloud Build service account** needs roles for what each step does (Run admin, AR writer, Secret accessor). Lock down least-privilege.
4. **`logging: CLOUD_LOGGING_ONLY`** is required for some 2nd-gen logging configurations; pick one.
5. **Triggers and `cloudbuild.yaml` are separate** — a config without a trigger is just a manual `gcloud builds submit`.
6. **Never bake secrets into images** — use `availableSecrets` + Secret Manager, not env vars in `Dockerfile`.

## Documentation

- [Cloud Build configuration reference](https://docs.cloud.google.com/build/docs/build-config-file-schema)
- [Substitutions](https://docs.cloud.google.com/build/docs/configuring-builds/substitute-variable-values)
- [Secrets in builds](https://docs.cloud.google.com/build/docs/securing-builds/use-secrets)
- [Build triggers](https://docs.cloud.google.com/build/docs/automating-builds/create-manage-triggers)
- [Artifact Registry](https://docs.cloud.google.com/artifact-registry/docs)
- [Cloud Build builders catalog](https://github.com/GoogleCloudPlatform/cloud-builders)
- [SLSA build provenance](https://docs.cloud.google.com/build/docs/securing-builds/view-build-provenance)
