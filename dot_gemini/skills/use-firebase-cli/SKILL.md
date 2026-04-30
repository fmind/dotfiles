---
name: use-firebase-cli
description: Guide for using the firebase CLI — login, project context, init/deploy, emulators, App Hosting, security rules tests, and Functions workflows.
---

# Use Firebase CLI

`firebase` is the primary CLI for Firebase projects (Firestore, Auth, Hosting, App Hosting, Functions, Storage, Realtime Database, FCM, Remote Config, Crashlytics, Data Connect, AI Logic). The skill bundle `firebase/agent-skills` (installed via `install-firebase-skills`) covers product expertise; this skill focuses on the **CLI workflow**.

## One-time Setup

```bash
# Browser login (opens a window).
firebase login

# Headless / CI: pass a `FIREBASE_TOKEN` from `firebase login:ci`.
firebase login:ci                # generate a long-lived CI token (interactive)
export FIREBASE_TOKEN=<token>    # then non-interactive in CI

# Sanity checks.
firebase login:list
firebase projects:list
```

## Project Context

```bash
# Bind a local repo to a Firebase project.
firebase use <project-id>
firebase use --add               # interactive: alias projects (default, staging, prod)

# Inspect.
firebase projects:list
cat .firebaserc                  # alias map written by `firebase use`
```

## Init / Layout

```bash
# Interactive product picker (Firestore, Functions, Hosting, Storage, Emulators, …).
firebase init

# A typical local layout after init:
#   firebase.json          # service config (Hosting rewrites, Functions source, …)
#   .firebaserc            # project aliases
#   firestore.rules        # Firestore security rules
#   firestore.indexes.json # Firestore composite indexes
#   storage.rules          # Storage security rules
#   apphosting.yaml        # App Hosting config (Next.js / Angular)
#   functions/             # Cloud Functions source (Node, Python)
```

## Emulators (local dev)

```bash
# Start all configured emulators (Auth, Firestore, Functions, Hosting, …).
firebase emulators:start

# Run a one-shot command against fresh emulators (great for CI).
firebase emulators:exec --only firestore "npm test"

# Import / export emulator state.
firebase emulators:start --import=./seed --export-on-exit=./seed
```

## Deploy

```bash
# Deploy everything in firebase.json.
firebase deploy

# Scope to specific products.
firebase deploy --only hosting
firebase deploy --only functions
firebase deploy --only firestore:rules,storage:rules

# Multi-target hosting (per-site rewrites).
firebase target:apply hosting marketing my-marketing-site
firebase deploy --only hosting:marketing
```

## Cloud Functions

```bash
# Local dev.
cd functions && npm install
firebase emulators:start --only functions

# Deploy a single function (faster than the whole bundle).
firebase deploy --only functions:onUserCreate

# Inspect logs.
firebase functions:log
firebase functions:log --only onUserCreate
```

## App Hosting (modern web frameworks)

```bash
# Create a new App Hosting backend.
firebase apphosting:backends:create

# List rollouts.
firebase apphosting:rollouts:list <backend>

# Local preview that mirrors App Hosting.
firebase emulators:start --only apphosting
```

## Security Rules

```bash
# Test rules locally (uses the emulator).
firebase emulators:exec --only firestore "npm run test:rules"

# Deploy only rules.
firebase deploy --only firestore:rules
firebase deploy --only storage:rules
```

## Hosting Channels (preview deploys)

```bash
firebase hosting:channel:deploy preview-pr-42 --expires=7d
firebase hosting:channel:list
firebase hosting:channel:delete preview-pr-42
```

## Common Workflows

**Bootstrap a Hosting + Functions project.**

1. `firebase login`
2. `firebase init hosting functions emulators`
3. `cd functions && npm install`
4. `firebase emulators:start` to develop locally.
5. `firebase deploy` to ship.

**Promote dev → prod.**

1. `firebase use --add` to alias `default`, `staging`, `prod`.
2. `firebase deploy -P staging --only hosting,functions`
3. After validation: `firebase deploy -P prod --only hosting,functions`

**Iterate on Firestore rules safely.**

```bash
firebase emulators:exec --only firestore "npx jest tests/rules"
firebase deploy --only firestore:rules -P staging
firebase deploy --only firestore:rules -P prod
```

## MCP Mode

The Firebase CLI ships an MCP server (`firebase mcp` / `firebase-tools mcp`). For agent-driven Firebase work in nearly every session, install it via `install-firebase-mcp`; otherwise use the `firebase` subagent.

## Important Notes

1. **Always specify the project** in CI/scripts: `-P <alias>` or `--project=<id>`. Don't rely on the active alias.
2. **Test security rules in the emulator** before deploying — production rule bugs lock real users out.
3. **Deploy by product** (`--only ...`) — full deploys are slow and risk shipping unrelated changes.
4. **Hosting channels** are how you do PR previews; they auto-expire and don't replace prod.
5. **Functions deploys go through Cloud Build** — first deploy enables `cloudbuild.googleapis.com`; check build logs in the Console on failure.

## Documentation

- [Firebase CLI reference](https://firebase.google.com/docs/cli)
- [Firebase MCP server](https://firebase.google.com/docs/cli/mcp-server)
- [Firebase emulators](https://firebase.google.com/docs/emulator-suite)
- [Hosting deploy targets & channels](https://firebase.google.com/docs/hosting/multisites)
- [Cloud Functions for Firebase](https://firebase.google.com/docs/functions)
- [App Hosting](https://firebase.google.com/docs/app-hosting)
- [Security Rules unit testing](https://firebase.google.com/docs/rules/unit-tests)
