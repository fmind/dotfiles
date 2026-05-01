---
name: install-gcp-mcp
description: Install one or more Google Cloud / Google MCP servers (BigQuery, Cloud Run, Cloud Storage, Cloud Logging, Cloud Monitoring, Cloud Trace, Compute Engine, Firestore, Pub/Sub, Resource Manager, Error Reporting, Agent Registry, Agent Search, Developer Knowledge, Vertex AI, CX Agent Studio, Gemini Cloud Assist, AlloyDB, Bigtable, Cloud SQL, Spanner, GKE, ...) into .gemini/settings.json. Use when one or more GCP services are central to the project.
---

# Install GCP MCP

Drops one or more **Google Cloud MCP servers** into `.gemini/settings.json` for the current project. Every server in this family shares the same install shape — an `httpUrl` plus `authProviderType: "google_credentials"` — and authenticates via Application Default Credentials (ADC).

Use this skill when one or more GCP services are central to the project (their tools are needed in nearly every session). Otherwise prefer the per-product subagent at `~/.gemini/agents/<service>.md`, which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

The repo, configuration, or user request indicates regular use of a GCP service in the catalog below. Always:

1. Pick the **minimum** set actually central to the project — do not install everything.
2. Skip entries already installed: `grep -q '"<server-name>"' .gemini/settings.json 2>/dev/null`.
3. If the user mentions a GCP service that is not in the catalog below, fetch the canonical list at <https://docs.cloud.google.com/mcp/supported-products> — Google adds servers regularly and that page is authoritative for endpoints, regions, and Preview status.

## Catalog (most-used GCP MCPs)

| Server name (JSON key) | `httpUrl` | API to enable | Trigger heuristics | Notes |
|---|---|---|---|---|
| `bigquery` | `https://bigquery.googleapis.com/mcp` | `bigquery.googleapis.com` | `*.sql`, `bq` scripts, `schema.json`, imports `google-cloud-bigquery` / `@google-cloud/bigquery` | |
| `cloud-run` | `https://run.googleapis.com/mcp` | `run.googleapis.com` | `service.yaml`, Cloud Run `Dockerfile`, `cloudbuild.yaml` | |
| `cloud-storage` | `https://storage.googleapis.com/storage/mcp` | `storage.googleapis.com` | imports `google-cloud-storage` / `@google-cloud/storage`, GCS-bound assets | |
| `cloud-logging` | `https://logging.googleapis.com/mcp` | `logging.googleapis.com` | log queries, log sinks/buckets | |
| `cloud-monitoring` | `https://monitoring.googleapis.com/mcp` | `monitoring.googleapis.com` | MQL/PromQL queries, alert policies, SLOs | |
| `cloud-trace` | `https://cloudtrace.googleapis.com/mcp` | `cloudtrace.googleapis.com` | latency / perf debugging across services | |
| `compute-engine` | `https://compute.googleapis.com/mcp` | `compute.googleapis.com` | `google_compute_instance`, `gcloud compute …` scripts | |
| `firestore` | `https://firestore.googleapis.com/mcp` | `firestore.googleapis.com` | `firestore.rules`, `firestore.indexes.json`, imports `@google-cloud/firestore` / `firebase-admin/firestore` | |
| `pubsub` | `https://pubsub.googleapis.com/mcp` | `pubsub.googleapis.com` | Pub/Sub Terraform, imports `google-cloud-pubsub` / `@google-cloud/pubsub` | |
| `resource-manager` | `https://cloudresourcemanager.googleapis.com/mcp` | `cloudresourcemanager.googleapis.com` | project / folder / IAM mgmt across an org | |
| `error-reporting` | `https://clouderrorreporting.googleapis.com/mcp` | `clouderrorreporting.googleapis.com` | crash / exception triage | |
| `agent-registry` | `https://agentregistry.googleapis.com/mcp` | `agentregistry.googleapis.com` | publishes / consumes agent definitions | Preview |
| `agent-search` | `https://discoveryengine.googleapis.com/mcp` | `discoveryengine.googleapis.com` | grounded enterprise search, RAG over GCS / BigQuery corpora | Same endpoint as the legacy `vertex-ai-search`; pick one name. |
| `developer-knowledge` | `https://developerknowledge.googleapis.com/mcp` | `developerknowledge.googleapis.com` | grounded GCP doc lookup | |
| `cx-agent-studio` | `https://ces.us.rep.googleapis.com/mcp` | `dialogflow.googleapis.com` + `ces.googleapis.com` | Dialogflow CX flows / intents / webhooks / eval sets | Regional endpoint (swap `us` for the right region). Caller needs `roles/mcp.toolUser` plus a Dialogflow CX role (typically `roles/dialogflow.admin`). |
| `gemini-cloud-assist` | `https://geminicloudassist.googleapis.com/mcp` | `geminicloudassist.googleapis.com` | GCP architecture design, troubleshooting, cost analysis | Distinct from Gemini *Code* Assist. Preview. |

### Vertex AI (multi-toolset)

The `aiplatform.googleapis.com` MCP gateway has **no bare `/mcp` endpoint** — each toolset is a separate MCP server. Register one entry per toolset you need.

| Server name | `httpUrl` | Use when |
|---|---|---|
| `vertex-ai-models` | `https://aiplatform.googleapis.com/mcp/models` | model registry |
| `vertex-ai-predict` | `https://aiplatform.googleapis.com/mcp/predict` | online inference |
| `vertex-ai-generate` | `https://aiplatform.googleapis.com/mcp/generate` | Gemini / foundation-model generation |
| `vertex-ai-notebook` | `https://aiplatform.googleapis.com/mcp/notebook` | managed notebooks |
| `vertex-ai-endpoints` | `https://aiplatform.googleapis.com/mcp/endpoints` | deploy / undeploy |
| `vertex-ai-tuning` | `https://aiplatform.googleapis.com/mcp/tuning` | custom-model training |

For data residency, swap to a regional host (e.g. `https://europe-west4-aiplatform.googleapis.com/mcp/models`). The agent-platform toolsets (`/mcp/retrieval`, `/mcp/evaluation`, `/mcp/prompts`) live in `install-gemini-enterprise-mcp` — not here.

### Other ADC MCPs (catalog page only)

The supported-products page lists many more ADC servers without dedicated coverage in this skill — e.g. AlloyDB (`alloydb.googleapis.com`), Cloud SQL (`sqladmin.googleapis.com`), Spanner (`spanner.googleapis.com`), Bigtable (`bigtableadmin.googleapis.com`), GKE (`container.googleapis.com`), Memorystore (`redis.googleapis.com` / `memorystore.googleapis.com`), Network Intelligence Center (`networkmanagement.googleapis.com`), Cloud Asset Inventory (`cloudasset.googleapis.com`), Datastream, Database Migration Service, Knowledge Catalog (Dataplex), Managed Service for Apache Spark / Kafka / Airflow, Chronicle / Security Operations, Oracle Database@GCP. They use the same install pattern below — fetch the catalog page for the exact endpoint, then proceed.

## Install

Merge one entry per chosen server into `.gemini/settings.json` (create the file if missing). Combine all entries under a single `mcpServers` object:

```json
{
  "mcpServers": {
    "<server-name>": {
      "httpUrl": "<endpoint from the catalog>",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc <server-name>
```

## Authentication

All ADC GCP MCPs share one auth flow:

```bash
gcloud auth application-default login
gcloud config set project <PROJECT_ID>
```

Enable any APIs the chosen servers need (see the `API to enable` column):

```bash
gcloud services enable <api1> <api2> --project=<PROJECT_ID>
```

If a server's API is not enabled, the first tool call fails with `SERVICE_DISABLED`. Some servers also require additional IAM (e.g. `cx-agent-studio` needs `roles/mcp.toolUser` plus a Dialogflow CX role).

## Companion Agents

Most of these MCPs have a sibling subagent at `~/.gemini/agents/<service>.md` (e.g. `cloud-run`, `firestore`, `cloud-logging`). The subagent wraps the same MCP and loads it lazily — keep the subagent installed for cross-project ad-hoc work even after pinning the MCP at the project level.

## Documentation

- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products) — canonical, always-current list of endpoints, regions, and Preview status.
- [Application Default Credentials](https://docs.cloud.google.com/docs/authentication/provide-credentials-adc)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
- Per-product MCP references typically live at `https://docs.cloud.google.com/<product>/docs/reference/mcp`.
