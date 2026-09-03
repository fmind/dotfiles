---
name: atlas
description: Manage database schema migrations with the Atlas CLI, declarative schema apply or versioned migrate diff, lint, and apply, and sqlc pairing. Use for any schema migration.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/atlas
  created: "2026-09-03"
  updated: "2026-09-03"
---

# Atlas

Manage the database schema as code with the Atlas CLI: a declarative `schema.sql` (or HCL) is the desired state, `atlas` plans changes against a throwaway dev database, and versioned migration files carry them to Cloud SQL. Query code generation stays with `sqlc` in [go-stack](../go-stack/SKILL.md); provisioning Cloud SQL belongs to [google-cloud](../google-cloud/SKILL.md).

## Workflow

1. **Project file**: commit `atlas.hcl` with one env per target; `dev` is a disposable Docker database Atlas uses to normalize and lint.
   ```hcl
   env "local" {
     src = "file://schema.sql"
     url = getenv("DATABASE_URL")
     dev = "docker://postgres/17/dev?search_path=public"
     migration { dir = "file://migrations" }
     lint { latest = 1 }
   }
   ```
1. **Declarative (dev and prototypes)**: edit `schema.sql`, preview, then apply.
   ```bash
   atlas schema apply --env local --dry-run
   atlas schema apply --env local --auto-approve
   ```
1. **Versioned (shared and production)**: generate a migration from the schema change, lint it, and apply it as a deploy step.
   ```bash
   atlas migrate diff add_users --env local
   atlas migrate lint --env local
   atlas migrate apply --env local --dry-run
   atlas migrate apply --env local
   ```
1. **Pair with sqlc**: point `sqlc.yaml` `schema:` at the same `schema.sql` (or at `migrations/`, whose Atlas files sqlc parses) so generated Go types and migrations never disagree; run `sqlc generate` after every diff.
1. **Cloud SQL**: connect through the Cloud SQL Auth Proxy or private IP with a least-privilege migration user; run `atlas migrate apply` from CI or a Cloud Run job before the new revision takes traffic, never at application start.
1. **Verify**: `atlas migrate status --env <env>` shows no pending files and `atlas schema inspect --env <env>` matches `schema.sql`.

## Gotchas

- **`atlas.sum` is a checksum**: change migration files only through `atlas migrate edit` or re-hash with `atlas migrate hash`; a stale sum aborts `apply`.
- **Dev database is mandatory**: `diff` and `lint` need `dev`, so Docker must be running.
- **Destructive changes**: `migrate lint` flags drops and non-additive changes; keep them in a separate, reviewed migration.
- **No `schema apply` in production**: it plans against the live database; production takes reviewed, versioned files only.
- **Secrets**: keep `url` behind `getenv(...)`, never a literal.

## Documentation

- [Atlas](https://atlasgo.io/docs) · [Versioned workflow](https://atlasgo.io/versioned/intro) · [Declarative workflow](https://atlasgo.io/declarative/apply) · [Atlas agent skill](https://atlasgo.io/guides/ai-tools/agent-skills) (vendor `SKILL.md` shipped as docs, copy on demand)
- Companion skills: [go-stack](../go-stack/SKILL.md) (`sqlc`, `pgx`), [google-cloud](../google-cloud/SKILL.md) (Cloud SQL), [cloud-run](../cloud-run/SKILL.md) (jobs and revisions), [sops-secrets](../sops-secrets/SKILL.md) (`DATABASE_URL`).
