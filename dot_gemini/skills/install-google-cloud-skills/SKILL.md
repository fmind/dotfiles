---
name: install-google-cloud-skills
description: Install Google's official Agent Skills repository so the agent gains expert knowledge for Cloud Run, BigQuery, Cloud SQL, AlloyDB, GKE, Gemini API, and the Well-Architected Framework.
---

# Install Google Cloud Skills

Google publishes the official [`google/skills`](https://github.com/google/skills) repository (announced at Cloud Next 2026), covering core Google Cloud products and well-architected pillars. This skill explains when and how to install it; the actual GCP expertise lives in the bundled skills.

## When to Trigger

- The user is about to write code that touches Google Cloud (Cloud Run, BigQuery, Cloud SQL, AlloyDB, GKE, Gemini API, Vertex AI, IAM).
- The user mentions deploying / authenticating to GCP, designing for security/reliability/cost, or asks for `gcloud` idioms.
- Verify first: `ls .agents/skills/ ~/.gemini/skills/ 2>/dev/null | grep -E 'cloud-run|bigquery|gke|alloydb|firebase|gemini-api'`. If skills are already present, skip installation.

## Install

```bash
# Interactive: pick the skills you want from the bundle.
npx skills add google/skills
```

`npx skills` writes to `.agents/skills/` for project scope (default — repo-pinned, commits with the codebase) or to `~/.gemini/skills/` for global scope when prompted.

For a single skill rather than the whole bundle:

```bash
npx skills add google/skills --skill cloud-run-basics
npx skills add google/skills --skill bigquery-basics
```

## What Gets Installed

13 skills at `skills/cloud/` at the time of writing (verify with `npx skills add google/skills --list`):

**Cloud product basics**
- `alloydb-basics`, `bigquery-basics`, `cloud-run-basics`, `cloud-sql-basics`, `gke-basics`.
- `gemini-api` — Gemini API on Google Cloud.
- `firebase-basics` — also published separately at `firebase/agent-skills` with more depth.

**Well-Architected Framework pillars**
- `google-cloud-waf-security`, `google-cloud-waf-reliability`, `google-cloud-waf-cost-optimization`.

**Recipes**
- `google-cloud-recipe-onboarding` — first-time GCP onboarding.
- `google-cloud-recipe-auth` — authenticating clients to Google Cloud.
- `google-cloud-networking-observability` — VPC / Cloud Logging / NetMon hands-on.

## Related: `gcloud` CLI

The skills assume `gcloud` is configured. Quick sanity check:

```bash
gcloud auth login
gcloud auth application-default login    # for client-library code
gcloud config set project <PROJECT_ID>
gcloud config list
```

Common idioms the skills will use / generate:

```bash
# Always pass --project explicitly in scripts and CI.
gcloud run deploy svc --source . --project $PROJECT --region us-central1
gcloud iam service-accounts create deployer --project $PROJECT
gcloud secrets create db-password --data-file=- --project $PROJECT < pw.txt
```

## After Install

1. Restart the agent so the new skill descriptions are picked up by progressive disclosure.
2. Project-scope installs (`.agents/skills/`) commit naturally with the repo. Global-scope installs are machine-local — `chezmoi add ~/.gemini/skills/<slug>` to track them.
3. Update later via `npx skills update`.

## Important Notes

1. This skill does **not** duplicate the bundled content — it's an installer guide. The agent should defer Cloud-specific guidance to the official skills once installed.
2. Firebase has its own dedicated skill bundle (`firebase/agent-skills`) — see `install-firebase-skills`. The `firebase-basics` skill in `google/skills` is lighter than that bundle.
3. There's also `GoogleCloudPlatform/gemini-cloud-assist-mcp` with an `operating-google-cloud` skill — useful complement for live cloud operations via MCP.
4. `npx skills` requires Node.js (already provided by mise).

## Documentation

- [google/skills repo](https://github.com/google/skills)
- [Cloud Next 2026 announcement](https://cloud.google.com/blog/topics/developers-practitioners/level-up-your-agents-announcing-googles-official-skills-repository)
- [`npx skills` CLI](https://github.com/vercel-labs/skills)
- [gcloud CLI reference](https://docs.cloud.google.com/sdk/gcloud/reference)
