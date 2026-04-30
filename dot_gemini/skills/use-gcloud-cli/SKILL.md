---
name: use-gcloud-cli
description: Guide for using the gcloud CLI — auth, project context, deploys, IAM, services, secrets, and structured JSON output for scripting.
---

# Use gcloud CLI

`gcloud` is the primary CLI for Google Cloud. The agent should always run with explicit project context, prefer `--format=json` for parseable output, and read before writing IAM / billing / service-enable changes.

## One-time Setup

```bash
# Login (browser-based; on a remote box use --no-launch-browser).
gcloud auth login
gcloud auth login --no-launch-browser

# Application Default Credentials — used by client libraries.
gcloud auth application-default login

# Pick the active project + region/zone.
gcloud config set project <PROJECT_ID>
gcloud config set compute/region us-central1
gcloud config set compute/zone us-central1-a
gcloud config list

# Enable an API on a project (idempotent).
gcloud services enable run.googleapis.com bigquery.googleapis.com --project=<PROJECT_ID>
```

## Project, Account, and Configuration Switching

```bash
# Named configurations (separate accounts/projects).
gcloud config configurations list
gcloud config configurations create work
gcloud config configurations activate work

# Switch identity within a config.
gcloud auth list
gcloud config set account user@example.com

# Always pass --project explicitly in scripts/CI to avoid surprises.
gcloud run services list --project=$PROJECT
```

## IAM

```bash
# Read.
gcloud projects get-iam-policy $PROJECT --format=json
gcloud iam service-accounts list --project=$PROJECT

# Add a binding (idempotent at the policy level).
gcloud projects add-iam-policy-binding $PROJECT \
  --member="serviceAccount:deployer@$PROJECT.iam.gserviceaccount.com" \
  --role="roles/run.admin"

# Service accounts.
gcloud iam service-accounts create deployer --project=$PROJECT
gcloud iam service-accounts keys create key.json \
  --iam-account=deployer@$PROJECT.iam.gserviceaccount.com
```

## Cloud Run

```bash
# Deploy from source (Buildpacks / Dockerfile auto-detected).
gcloud run deploy svc --source . \
  --region=$REGION --project=$PROJECT --allow-unauthenticated

# Update env, traffic, revisions.
gcloud run services update svc --update-env-vars KEY=val --region=$REGION
gcloud run services update-traffic svc --to-revisions=svc-00012-abc=10 --region=$REGION
gcloud run revisions list --service=svc --region=$REGION
```

## Cloud Storage (`gcloud storage` — replaces `gsutil`)

```bash
gcloud storage ls
gcloud storage cp ./out.tar.gz gs://my-bucket/releases/
gcloud storage rsync -r ./site/ gs://my-bucket/site/
gcloud storage objects describe gs://my-bucket/key
gcloud storage buckets create gs://my-new-bucket --location=us-central1
```

## Secret Manager

```bash
echo -n "$DB_PASS" | gcloud secrets create db-password --data-file=- --project=$PROJECT
gcloud secrets versions access latest --secret=db-password --project=$PROJECT
gcloud secrets list --project=$PROJECT
```

## BigQuery (use the `bq` CLI; `gcloud` has no GA `bigquery` group)

```bash
# List / inspect datasets.
bq ls --project_id=$PROJECT
bq show --project_id=$PROJECT $DATASET

# Run SQL.
bq query --use_legacy_sql=false 'SELECT COUNT(*) FROM `proj.ds.tbl`'
```

`gcloud` exposes BigQuery only under `gcloud alpha bq` (alpha, may break). For day-to-day work, use the dedicated `bq` CLI which ships with the Google Cloud SDK.

## Logs, Trace, Errors

```bash
gcloud logging read 'severity>=ERROR AND resource.type="cloud_run_revision"' \
  --project=$PROJECT --limit=20 --format=json

gcloud beta trace traces list --project=$PROJECT --limit=10
```

## JSON / Scripting

```bash
# Always pin --format and pipe through jq when scripting.
gcloud run services list --project=$PROJECT --format=json \
  | jq '.[] | {name, region: .metadata.labels."cloud.googleapis.com/location"}'

# Or use --format='value(...)' for one column.
gcloud projects list --format='value(projectId)'
```

## Common Workflows

**Bootstrap a new project for a deploy.**

1. `gcloud projects create $PROJECT --name="$NAME"`
2. `gcloud beta billing projects link $PROJECT --billing-account=$BILLING`
3. `gcloud services enable run.googleapis.com cloudbuild.googleapis.com artifactregistry.googleapis.com --project=$PROJECT`
4. `gcloud iam service-accounts create deployer --project=$PROJECT`
5. `gcloud projects add-iam-policy-binding $PROJECT --member=... --role=roles/run.admin`

**Deploy + tail logs.**

```bash
gcloud run deploy svc --source . --region=$REGION --project=$PROJECT
gcloud logging tail 'resource.type="cloud_run_revision" AND resource.labels.service_name="svc"' \
  --project=$PROJECT
```

## Important Notes

1. **Always pass `--project=...` in scripts** — relying on `gcloud config set project` makes scripts fragile across configurations.
2. **Read IAM before writing** (`get-iam-policy --format=json`); `add-iam-policy-binding` mutates a live policy and races on concurrent edits.
3. **Use `--format=json` + `jq`** for everything programmatic; `--format=value(...)` for one-shot column extraction.
4. **`gcloud storage` is the modern replacement for `gsutil`** — same operations, faster, gcloud-native.
5. **`alpha` / `beta`** subcommands ship pre-GA features; pin or accept the warning rather than wrapping in `yes`.

## Documentation

- [gcloud CLI reference](https://docs.cloud.google.com/sdk/gcloud/reference)
- [gcloud cheat sheet](https://docs.cloud.google.com/sdk/docs/cheatsheet)
- [Authorization overview](https://docs.cloud.google.com/sdk/docs/authorizing)
- [Configurations](https://docs.cloud.google.com/sdk/docs/configurations)
- [Filtering & formatting output](https://docs.cloud.google.com/sdk/gcloud/reference/topic/formats)
