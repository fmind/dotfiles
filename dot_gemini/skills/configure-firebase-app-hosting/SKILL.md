---
name: configure-firebase-app-hosting
description: Guide for authoring apphosting.yaml — backend config, env vars, secrets, scaling, regions, build settings, and rollouts for Next.js / Angular / SvelteKit / Astro on Firebase App Hosting.
---

# Configure Firebase App Hosting (`apphosting.yaml`)

[Firebase App Hosting](https://firebase.google.com/docs/app-hosting) is a managed runtime for modern web frameworks (Next.js, Angular, SvelteKit, Astro). It auto-detects your framework, builds with Cloud Build, and serves from a globally-distributed Cloud Run backend with HTTPS, CDN, and integrated Firebase services.

The `apphosting.yaml` file (at the repo root, or per-app for monorepos) declares runtime config; the rest is handled by the platform.

## Minimal `apphosting.yaml`

```yaml
runConfig:
  minInstances: 0
  maxInstances: 100
  cpu: 1
  memoryMiB: 512

env:
  - variable: NODE_ENV
    value: production
    availability:
      - BUILD
      - RUNTIME
```

## Production-grade `apphosting.yaml`

```yaml
# Build phase (Cloud Build).
buildConfig:
  rootDirectory: .                    # path to the framework project (monorepo: e.g. apps/web)
  runtime: nodejs22                   # nodejs20 | nodejs22
  commands:
    install: pnpm install --frozen-lockfile
    build: pnpm build

# Runtime (Cloud Run).
runConfig:
  minInstances: 1                     # set ≥1 for low-latency
  maxInstances: 50
  concurrency: 80                     # requests per instance
  cpu: 2
  memoryMiB: 1024
  timeoutSeconds: 60
  vpcAccess:
    connector: projects/PROJECT/locations/us-central1/connectors/web
    egress: PRIVATE_RANGES_ONLY

# Env vars and secrets.
env:
  - variable: NEXT_PUBLIC_API_URL
    value: https://api.example.com
    availability: [BUILD, RUNTIME]    # injected at both build and runtime

  - variable: NODE_ENV
    value: production
    availability: [BUILD, RUNTIME]

  - variable: STRIPE_SECRET_KEY
    secret: stripe-secret-key         # Secret Manager: secrets/stripe-secret-key
    availability: [RUNTIME]           # NEVER expose secrets at build time

  - variable: DATABASE_URL
    secret: database-url
    availability: [RUNTIME]
```

`availability` controls when the variable is materialized:

| Availability | Use for |
|--------------|---------|
| `BUILD` | Public config baked into client bundles (`NEXT_PUBLIC_*`, `VITE_*`) |
| `RUNTIME` | Server-only secrets and runtime config |
| Both | Stable values needed at build *and* runtime (e.g. `NODE_ENV`) |

## Backend Lifecycle

```bash
# Create the backend (interactive).
firebase apphosting:backends:create \
  --project=$PROJECT \
  --app=my-app \
  --location=us-central1

# List backends.
firebase apphosting:backends:list --project=$PROJECT

# Trigger a rollout.
firebase apphosting:rollouts:create my-app --git-branch=main

# Inspect rollouts.
firebase apphosting:rollouts:list my-app
firebase apphosting:rollouts:describe my-app <rollout-id>
```

## Monorepo Layout

For a multi-app repo, set `buildConfig.rootDirectory` to the framework subfolder:

```yaml
# apphosting.yaml at the repo root, scoped to apps/web.
buildConfig:
  rootDirectory: apps/web
  runtime: nodejs22
  commands:
    install: pnpm install --frozen-lockfile
    build: pnpm --filter web build

env:
  - variable: NEXT_PUBLIC_API_URL
    value: https://api.example.com
    availability: [BUILD, RUNTIME]
```

Or run multiple backends with one config file each:

```
apps/
├── web/
│   └── apphosting.yaml
└── admin/
    └── apphosting.yaml
```

## Local Emulator

```bash
firebase emulators:start --only apphosting
# Mirrors the App Hosting build + serve flow locally.
```

## Framework Notes

| Framework | Detection | Notes |
|-----------|-----------|-------|
| Next.js | `next.config.{js,ts,mjs}` | App Router, RSC, ISR all supported. `images.unoptimized = false` works (uses Cloud CDN). |
| Angular | `angular.json` | SSR/Universal builds detected automatically. |
| SvelteKit | `svelte.config.{js,ts}` | Use `@sveltejs/adapter-node` (App Hosting auto-configures it). |
| Astro | `astro.config.{js,mjs,ts}` | SSR mode required for non-static content. |

App Hosting wraps each framework's standard build output and runs it on Cloud Run; you don't need to write a Dockerfile.

## Domains & Rollback

```bash
# Custom domain (Firebase Hosting integration).
firebase apphosting:custom-domains:create my-app www.example.com

# Roll back to a previous rollout.
firebase apphosting:rollouts:list my-app
firebase apphosting:rollouts:create my-app --rollback-to=<previous-id>
```

## Secrets Setup

```bash
# Create a secret in Secret Manager (one-time).
gcloud secrets create stripe-secret-key \
  --data-file=stripe-key.txt --project=$PROJECT

# Grant the App Hosting runtime SA access.
gcloud secrets add-iam-policy-binding stripe-secret-key \
  --member="serviceAccount:firebase-app-hosting-compute@$PROJECT.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

Reference it in `apphosting.yaml` via `secret: stripe-secret-key`.

## Common Workflows

**Bootstrap.**
1. `firebase init apphosting` (or use the Console).
2. Connect the GitHub repo (Developer Connect).
3. Drop `apphosting.yaml` at the repo root (or per-app).
4. Push to the connected branch — first rollout fires automatically.

**Cut over a custom domain.**
1. Verify ownership in the Console / via DNS TXT.
2. `firebase apphosting:custom-domains:create my-app www.example.com`.
3. Update DNS to the App Hosting CNAME / A records as instructed.
4. Wait for cert provisioning (~minutes).

**Roll back fast.**
```bash
firebase apphosting:rollouts:list my-app
firebase apphosting:rollouts:create my-app --rollback-to=<id>
```

## Important Notes

1. **`apphosting.yaml` lives at the build root** (set via `rootDirectory`). For monorepos, prefer one config per app.
2. **`availability: [BUILD]` for `NEXT_PUBLIC_*` and `VITE_*`** — values inlined into client bundles. Don't put secrets here.
3. **Secrets always `availability: [RUNTIME]`** — exposing them at build time bakes them into static assets.
4. **Rollouts are immutable** — to change config, push a commit; you can roll back to any prior rollout.
5. **Runtime SA is `firebase-app-hosting-compute@<project>.iam.gserviceaccount.com`** — grant it Secret Accessor and any other roles your runtime needs.
6. **Cold starts depend on `minInstances`** — set ≥1 for SSR/RSC apps where TTFB matters.

## Documentation

- [App Hosting overview](https://firebase.google.com/docs/app-hosting)
- [`apphosting.yaml` reference](https://firebase.google.com/docs/app-hosting/configure)
- [Environment variables and secrets](https://firebase.google.com/docs/app-hosting/configure#user-defined-environment)
- [Frameworks support](https://firebase.google.com/docs/app-hosting/frameworks-tooling)
- [Custom domains](https://firebase.google.com/docs/app-hosting/custom-domain)
- [Rollouts and rollback](https://firebase.google.com/docs/app-hosting/rollouts)
- [Companion: `firebase-app-hosting-basics` skill](https://github.com/firebase/agent-skills)
