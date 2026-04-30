---
name: use-pnpm-cli
description: Guide for using pnpm (fast, disk-efficient Node.js package manager) — install, add/remove deps, scripts, workspaces (monorepo), `dlx`, `pnpm-lock.yaml`, and `.npmrc`/`pnpm-workspace.yaml` config.
---

# Use pnpm CLI

[pnpm](https://pnpm.io) is a fast, content-addressed Node.js package manager. Unlike `npm` and `yarn`, it stores every package version once in a global content-addressed store (`~/.local/share/pnpm/store` in this dotfile setup) and hard-links into each project's `node_modules` — saving disk space and keeping `node_modules` shallow with strict peer-dependency resolution.

> mise installs and pins pnpm via `dot_config/mise/config.toml.tmpl`. The user-level config lives in `dot_config/pnpm/rc` (linked to `~/.config/pnpm/rc`).

## Install

```bash
# Recommended — already configured for this user via mise.
mise use -g pnpm@latest

# Or via the official install script.
curl -fsSL https://get.pnpm.io/install.sh | sh -

# Or via Corepack (ships with Node ≥ 16.13).
corepack enable
corepack prepare pnpm@latest --activate

# Sanity check.
pnpm --version
pnpm help
```

## Projects

```bash
# Bootstrap.
pnpm init                   # creates package.json
pnpm install                # install everything in package.json (alias: `pnpm i`)

# Add / remove deps.
pnpm add react react-dom            # runtime deps
pnpm add -D typescript vitest       # devDependencies
pnpm add -O @types/node             # optionalDependencies
pnpm add -E react@18.3.1            # exact version (no `^`)
pnpm remove lodash

# Update.
pnpm outdated
pnpm update                          # respects ranges in package.json
pnpm update --latest                 # bump beyond ranges (rewrites package.json)
```

## Scripts

```bash
# Run a script defined in package.json "scripts".
pnpm run build              # explicit
pnpm build                  # shorthand (works for any non-reserved script name)
pnpm test                   # built-in shorthand
pnpm start                  # built-in shorthand

# Pass args after `--`.
pnpm test -- --watch

# List defined scripts.
pnpm run
```

## Ephemeral Tool Runs (`dlx`)

```bash
# npx-equivalent: download + run without installing.
pnpm dlx create-vite my-app --template react-ts
pnpm dlx tsx script.ts
```

## Workspaces (monorepo)

```yaml
# pnpm-workspace.yaml at repo root
packages:
  - "apps/*"
  - "packages/*"
```

```bash
# Install all workspace deps + link internal packages.
pnpm install

# Run a script in every workspace package.
pnpm -r run build           # `-r` = recursive, all packages
pnpm -r run test --parallel

# Run in one specific package.
pnpm --filter @scope/api run dev
pnpm --filter "./apps/*" run lint

# Add a dep only to one package.
pnpm --filter @scope/api add fastify

# Add a workspace package as a dep of another.
pnpm --filter @scope/web add @scope/ui --workspace
```

`workspace:*` protocol in `package.json` resolves to the local workspace package version — pnpm rewrites it during publish.

## `.npmrc` / `~/.config/pnpm/rc`

This dotfile setup keeps user-level pnpm config at `~/.config/pnpm/rc` (chezmoi: `dot_config/pnpm/rc`):

```ini
auto-install-peers=true
prefer-frozen-lockfile=false
strict-peer-dependencies=false
store-dir=~/.local/share/pnpm/store
state-dir=~/.local/state/pnpm
cache-dir=~/.cache/pnpm
```

Common project-level overrides go in `.npmrc` at the repo root:

```ini
# Hoist nothing — keeps node_modules strict (default).
hoist=false

# Use a private registry for one scope.
@my-org:registry=https://npm.pkg.github.com

# Pin Node engine.
engine-strict=true
```

## Lockfile & Reproducibility

- `pnpm-lock.yaml` is the single source of truth — commit it.
- `pnpm install --frozen-lockfile` errors if the lockfile is stale (use in CI).
- `pnpm install --offline` skips the registry entirely (must be in store).
- `pnpm fetch` pre-populates the store from the lockfile (great for Docker layer caching).

## Common Workflows

**Bootstrap a new TypeScript project.**

```bash
pnpm init
pnpm add -D typescript tsx vitest @types/node
pnpm exec tsc --init
echo 'console.log("hi")' > src/index.ts
pnpm dlx tsx src/index.ts
```

**Reproduce CI locally.**

```bash
pnpm install --frozen-lockfile
pnpm -r run build
pnpm -r run test
```

**Migrate from npm / yarn.**

```bash
# Drop node_modules + lockfile, then re-install.
rm -rf node_modules package-lock.json yarn.lock
pnpm import           # converts package-lock.json / yarn.lock → pnpm-lock.yaml first if present
pnpm install
```

**Audit & dedupe.**

```bash
pnpm audit
pnpm audit --fix
pnpm dedupe           # rewrites lockfile to remove duplicate versions
```

## Important Notes

1. **`pnpm-lock.yaml` is the source of truth** — commit it; CI uses `--frozen-lockfile` for deterministic installs.
2. **Strict `node_modules` by default** — packages can only `require` deps they declared. This catches phantom dependencies that `npm`/`yarn` silently allow.
3. **`auto-install-peers=true`** (set in this dotfile) installs peer deps automatically. Disable per-project with `auto-install-peers=false` in repo `.npmrc` if you want explicit peer management.
4. **`pnpm exec` ≠ `pnpm dlx`** — `exec` runs a binary already installed in the project; `dlx` downloads + runs ephemerally.
5. **Use `--filter` aggressively in monorepos** — running `pnpm -r build` in a 30-package repo without filters is rarely what you want.
6. **The store is shared across projects** — `pnpm store prune` reclaims space from versions no longer referenced anywhere.

## Documentation

- [pnpm home](https://pnpm.io)
- [CLI reference](https://pnpm.io/cli/add)
- [Workspaces](https://pnpm.io/workspaces)
- [`.npmrc` settings](https://pnpm.io/npmrc)
- [`pnpm-lock.yaml` format](https://pnpm.io/git#lockfiles)
- [GitHub: pnpm/pnpm](https://github.com/pnpm/pnpm)
