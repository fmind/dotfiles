---
name: duckdb
description: Query and transform CSV, Parquet, JSON, SQLite, and DuckDB files with the duckdb and sqlite3 CLIs for ad-hoc analysis and exports. Use for local SQL over files or an embedded database.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/duckdb
  created: "2026-09-02"
  updated: "2026-09-03"
---

# DuckDB and SQLite

DuckDB is the default engine for local analysis: it reads CSV, Parquet, and JSON directly, attaches SQLite files, and writes any of them back. SQLite stays the embedded store for applications; DuckDB is the tool you point at data.

## Commands

```bash
duckdb -c "SELECT name, score FROM 'people.csv' WHERE score > 80 ORDER BY score DESC"   # files are tables
duckdb -json -c "SELECT * FROM 'events/*.json'"                                          # JSON output for agents and jq
duckdb lab.duckdb -c "CREATE OR REPLACE TABLE people AS FROM 'people.csv'"               # persist into a database file
duckdb -c "COPY (FROM 'people.csv') TO 'people.parquet'"                                 # convert; Parquet is the durable format
duckdb -c "ATTACH 'app.sqlite' AS s (TYPE sqlite); SELECT count(*) FROM s.users"         # read an application database
duckdb -c "SUMMARIZE FROM 'people.parquet'"                                              # column stats in one call
sqlite3 app.sqlite ".schema" && sqlite3 app.sqlite "PRAGMA integrity_check"              # inspect the app store itself
```

The interactive shells load `~/.duckdbrc` and `~/.sqliterc` (box mode, headers, timer, `∅` for NULL); scripts pass `-json`, `-csv`, or `-markdown` explicitly so output does not depend on the rc file.

## Workflow

1. **Look before querying**: `DESCRIBE FROM '<file>'` and `SUMMARIZE` reveal types, nulls, and ranges; fix a wrong inference with `read_csv('<file>', types={'id': 'BIGINT'})`.
1. **Keep queries in files**: `duckdb < analysis.sql` or `duckdb -f analysis.sql` for anything longer than one line, committed next to the data description.
1. **Persist derived data as Parquet**: never commit `.duckdb` files (they change on every open); commit the SQL that rebuilds them.
1. **Check results**: row counts before and after joins, `count(*) FILTER (WHERE x IS NULL)` on keys, and a spot check against the source.
1. **Export for the reader**: `-markdown` for a report, `-json` for another tool, `COPY ... TO 'out.csv' (HEADER)` for a spreadsheet.

## Gotchas

- **Do not open a live SQLite database with DuckDB while the app writes to it**: attach a copy, or use `sqlite3` directly.
- **Glob paths quote as strings**: `'events/*.parquet'` works, unquoted paths do not.
- **Memory**: large joins spill to disk automatically; set `SET memory_limit='4GB'` and `SET threads=4` on a shared machine.
- **Extensions load on demand**: `httpfs`, `spatial`, `postgres` install once with `INSTALL <ext>; LOAD <ext>;` and need network the first time.
- **Secrets**: `CREATE SECRET` from environment variables or Application Default Credentials, never literal credentials in SQL.

## Official Skills

Upstream: `duckdb/duckdb-skills` (one skill per task: querying, file formats, attaching databases, spatial, S3, docs lookup). List the current release, then install what the task needs at project scope after reviewing the snapshot (see [agent-skills](../agent-skills/SKILL.md)):

```bash
skills add duckdb/duckdb-skills --list
skills add duckdb/duckdb-skills --skill <name> -y
```

## Documentation

- [DuckDB CLI](https://duckdb.org/docs/stable/clients/cli/overview) · [SQLite CLI](https://sqlite.org/cli.html)
- Companion skills: [python-script](../python-script/SKILL.md) (a one-file pipeline when SQL is not enough), [go-stack](../go-stack/SKILL.md) (`sqlc` for application SQL).
