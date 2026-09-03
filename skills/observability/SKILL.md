---
name: observability
description: Instrument a service or agent with JSON logs, OpenTelemetry traces and metrics, trace_id correlation, and GenAI spans on Google Cloud. Use when adding logs, traces, or metrics.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/observability
  created: "2026-09-03"
  updated: "2026-09-03"
---

# Observability

One telemetry stack for services and agents: JSON logs on stdout, OpenTelemetry traces and metrics over OTLP, and the trace id stamped on every log line so Cloud Logging, Cloud Trace, and Cloud Monitoring show one request end to end. The stacks ship the logging defaults ([go-stack](../go-stack/SKILL.md) `slog`, [python-stack](../python-stack/SKILL.md) `structlog`); this skill owns the wiring across signals and the conventions for LLM and agent spans.

## Workflow

1. **Structured logs**: JSON to stdout, one event per line. Go: `slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: gcpKeys})` where `gcpKeys` renames `level` to `severity` and `msg` to `message`, the keys Cloud Logging parses. Python: `structlog` with `JSONRenderer` and the same keys.
1. **Traces and metrics**: use the plain OTLP exporters configured only by the standard env vars (`OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_SERVICE_NAME`, `OTEL_RESOURCE_ATTRIBUTES`) so local runs stay silent and no vendor SDK enters the code; wrap HTTP servers and clients with `otelhttp` (Go) or the `opentelemetry-instrumentation-*` packages (Python).
1. **Correlate**: a `slog.Handler` or `structlog` processor reads the active span and adds `logging.googleapis.com/trace` = `projects/<project>/traces/<trace_id>` and `logging.googleapis.com/spanId`; Cloud Logging then links the line to its trace. Keep a plain `trace_id` field too for other backends.
1. **Export on Google Cloud**: run the Google-built OpenTelemetry Collector (`otelcol-google`) as a Cloud Run sidecar; the app exports to `http://localhost:4317`, the collector authenticates with ADC and forwards traces, metrics, and logs to `telemetry.googleapis.com`. ADK agents use `--otel_to_cloud` per [google-adk](../google-adk/SKILL.md).
1. **LLM and agent spans**: follow the GenAI semantic conventions: span `chat <model>` with `gen_ai.operation.name`, `gen_ai.provider.name`, `gen_ai.request.model`, `gen_ai.usage.input_tokens`, `gen_ai.usage.output_tokens`; `invoke_agent <name>` with `gen_ai.agent.name`; `execute_tool <name>` with `gen_ai.tool.name`. Never record prompt or completion bodies by default.
1. **Verify**: send one request, open its trace in Cloud Trace, then read the correlated lines per [gcloud](../gcloud/SKILL.md) and confirm the metric exists in Cloud Monitoring.
   ```bash
   gcloud logging read 'trace="projects/<project>/traces/<trace_id>"' --limit=20
   ```

## Gotchas

- **Cloud Logging keys**: `level` and `msg` are not parsed; without `severity` every line shows as `DEFAULT`.
- **Sampling**: `OTEL_TRACES_SAMPLER=parentbased_traceidratio` with a low `OTEL_TRACES_SAMPLER_ARG` in production; 100% only in dev.
- **Flush before exit**: defer the provider shutdown and honor `SIGTERM`; Cloud Run allows 10 seconds.
- **Cardinality**: never put user ids, ids in paths, or prompts in metric labels.
- **LLM platforms**: add Langfuse when you need prompt versions, sessions, scores, and per-trace evals, or MLflow tracing when experiments already live in MLflow; both ingest OpenTelemetry traces, so OpenTelemetry stays the source.

## Official Skills

Upstream: `langfuse/skills`, `mlflow/skills`, `pydantic/skills`, `grafana/skills`. List the current release, then install what the task needs at project scope after reviewing the snapshot (see [agent-skills](../agent-skills/SKILL.md)):

```bash
skills add langfuse/skills --list          # likewise mlflow/skills, pydantic/skills, grafana/skills
skills add <owner/repo> --skill <name> -y
```

## Documentation

- [OpenTelemetry](https://opentelemetry.io/docs/) · [GenAI semantic conventions](https://github.com/open-telemetry/semantic-conventions-genai) · [Cloud Logging structured logs](https://docs.cloud.google.com/logging/docs/structured-logging) · [Google-built OTel Collector](https://docs.cloud.google.com/stackdriver/docs/instrumentation/google-built-otel) · [Langfuse](https://langfuse.com/docs) · [MLflow Tracing](https://mlflow.org/docs/latest/genai/tracing/)
- Companion skills: [cloud-run](../cloud-run/SKILL.md) (sidecar deploy), [google-adk](../google-adk/SKILL.md) (agent tracing), [gcloud](../gcloud/SKILL.md) (reading logs), [benchmark](../benchmark/SKILL.md) (latency numbers).
