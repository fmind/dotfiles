---
name: run-firebase-emulators
description: Run the Firebase Emulator Suite — start/stop emulators, seed and export state, connect client SDKs, and drive Firestore/Storage rules and Cloud Functions tests in CI.
---

# Run Firebase Emulators

The Firebase Emulator Suite runs Auth, Firestore, Realtime Database, Cloud Storage, Cloud Functions, Hosting, Pub/Sub, Eventarc, App Hosting, Data Connect, and Extensions locally — no cloud project, no billing, no risk to prod data.

Operational counterpart to `configure-firebase-rules` (rules authoring) and `create-firebase-function` (function scaffolding). The general CLI surface lives in `use-firebase-cli`; this skill is specifically about **driving the emulators** for dev and tests.

## When to Trigger

- The user wants to develop or test against Firebase locally without hitting prod.
- A `firebase.json` declares an `emulators` block but the user doesn't know how to use it.
- The user wants to run security-rules unit tests, Cloud Functions tests, or end-to-end tests in CI.
- The user mentions seeding/exporting Firestore data, importing fixtures, or "demo-" project IDs.

## Bootstrap

```bash
# The emulators ship with firebase-tools — no separate install.
firebase --version
java -version                         # Firestore/Pub-Sub/Storage emulators need a JRE (>= 11)

# Add (or update) the emulators block interactively.
firebase init emulators
```

A typical `firebase.json` block:

```json
{
  "emulators": {
    "auth":      { "port": 9099 },
    "functions": { "port": 5001 },
    "firestore": { "port": 8080 },
    "database":  { "port": 9000 },
    "hosting":   { "port": 5000 },
    "storage":   { "port": 9199 },
    "pubsub":    { "port": 8085 },
    "eventarc":  { "port": 9299 },
    "ui":        { "enabled": true, "port": 4000 },
    "singleProjectMode": true
  }
}
```

## Quick Reference

```bash
# Start everything declared in firebase.json.
firebase emulators:start

# Scope to specific emulators (faster boot, narrower test surface).
firebase emulators:start --only firestore,auth,functions

# Run a one-shot command against fresh emulators (great for CI).
firebase emulators:exec --only firestore "npm test"
firebase emulators:exec "npm run test:e2e"

# Seed: import on start; persist on shutdown.
firebase emulators:start --import=./seed --export-on-exit=./seed
firebase emulators:export ./seed       # snapshot a running suite

# Use a deterministic, offline-safe project id (no auth needed).
firebase emulators:start --project=demo-local
```

## Standard Workflows

### 1. Local dev loop (Hosting + Functions + Firestore)

```bash
firebase emulators:start --import=./seed --export-on-exit
# UI: http://localhost:4000
```

In your app, point SDKs at the emulators (only in dev):

```ts
import { connectFirestoreEmulator, getFirestore } from 'firebase/firestore';
import { connectAuthEmulator, getAuth } from 'firebase/auth';
import { connectFunctionsEmulator, getFunctions } from 'firebase/functions';

const db = getFirestore();
if (location.hostname === 'localhost') {
  connectFirestoreEmulator(db, '127.0.0.1', 8080);
  connectAuthEmulator(getAuth(), 'http://127.0.0.1:9099');
  connectFunctionsEmulator(getFunctions(), '127.0.0.1', 5001);
}
```

Server-side / admin SDK auto-connects when these env vars are set:

| Variable | Example |
|----------|---------|
| `FIRESTORE_EMULATOR_HOST` | `127.0.0.1:8080` |
| `FIREBASE_AUTH_EMULATOR_HOST` | `127.0.0.1:9099` |
| `FIREBASE_DATABASE_EMULATOR_HOST` | `127.0.0.1:9000` |
| `FIREBASE_STORAGE_EMULATOR_HOST` | `127.0.0.1:9199` |
| `PUBSUB_EMULATOR_HOST` | `127.0.0.1:8085` |
| `FUNCTIONS_EMULATOR` | `true` (set by the Functions emulator itself) |

### 2. Security-rules unit tests

```bash
npm install --save-dev @firebase/rules-unit-testing
firebase emulators:exec --only firestore,storage "npx vitest run"
```

`@firebase/rules-unit-testing` ignores rules when using `withSecurityRulesDisabled` for seeding, then re-enables them per test. See `configure-firebase-rules` for the test scaffolding.

### 3. Cloud Functions tests

Two flavors:

- **Unit** — `firebase-functions-test` (offline, mocked context):

  ```bash
  npm install --save-dev firebase-functions-test
  npx vitest run functions/test/unit
  ```

- **Integration** — emulators:exec with the real Functions runtime:

  ```bash
  firebase emulators:exec --only functions,firestore,auth \
    "npx vitest run functions/test/integration"
  ```

### 4. End-to-end / Playwright

```bash
# Boot emulators + seed + run e2e in one command, then tear down.
firebase emulators:exec --import=./seed \
  "npm run dev:wait-on && npx playwright test"
```

Use `wait-on http://127.0.0.1:5000` (or similar) inside the script so Playwright doesn't race the dev server.

### 5. Snapshot prod data for local repro (read-only!)

```bash
# 1. Export a *small, scrubbed* dataset from a non-prod environment.
gcloud firestore export gs://<bucket>/dump --project=<staging-project>

# 2. Convert to emulator format with the official tool, or mirror via Admin SDK.
#    The emulator import format is NOT identical to gcloud firestore export.
#    Prefer running a small Node/Python script that reads from staging and writes
#    to the local emulator (FIRESTORE_EMULATOR_HOST=127.0.0.1:8080).

# 3. Persist for next runs.
firebase emulators:export ./seed
```

Never import real prod PII into a developer machine — scrub or synthesize first.

## CI Patterns

```yaml
# .github/workflows/test.yml
name: test
on: [pull_request]
jobs:
  emulator-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: "20" }
      - uses: actions/setup-java@v4
        with: { distribution: temurin, java-version: "17" }
      - run: npm ci
      - run: npx firebase emulators:exec --only firestore,auth,functions "npm test"
        env:
          # Demo project id: no creds, no network egress to GCP.
          GCLOUD_PROJECT: demo-ci
```

Key flags for CI:

- `--project=demo-<anything>` — disables outbound calls; required for hermetic runs.
- `--import=./seed` — deterministic state.
- Don't use `--export-on-exit` in CI; exports are noise for ephemeral runners.

## Diagnosing Common Failures

**"Port 8080 is not open" / "Could not start emulator."**
Another emulator (or stray Java process) is holding the port:

```bash
lsof -iTCP:8080 -sTCP:LISTEN
firebase emulators:start --inspect-functions     # if Node debugger is the culprit
```

Or change the port in `firebase.json`.

**"Cloud Functions emulator: Failed to load function definition."**
The functions code didn't compile. Check `functions/lib` (TS) or that `package.json` `main` points to the built entry. Run `cd functions && npm run build` first.

**Tests pass locally, fail in CI.**

- Missing `--project=demo-...` flag → emulator tries real GCP and times out without creds.
- Missing JRE in CI image → use `actions/setup-java`.
- Race condition on emulator boot → wrap with `emulators:exec` instead of background `start`.

**Admin SDK writes to prod instead of the emulator.**
The env var isn't set in the test process. Verify:

```bash
echo $FIRESTORE_EMULATOR_HOST       # must be 127.0.0.1:8080 (or your port)
```

`emulators:exec` injects these automatically; manual `start` does not.

**Imported state is empty.**
The export directory must come from `emulators:export` or `--export-on-exit` — `gcloud firestore export` produces a different format and won't load.

**Functions can't reach Firestore.**
Inside the Functions emulator, use `127.0.0.1` (not `localhost`) for the Firestore host on Linux/CI; Node's IPv6 resolution can pick `::1` and miss the v4 listener.

## Useful Flags

| Flag | Purpose |
|------|---------|
| `--only <a,b>` | Subset of emulators (faster) |
| `--project demo-<x>` | Hermetic project id, no GCP egress |
| `--import <dir>` | Load a previous export at boot |
| `--export-on-exit[=<dir>]` | Persist state on Ctrl-C |
| `--inspect-functions[=port]` | Attach Node debugger to Functions |
| `--ui` / `singleProjectMode` (in firebase.json) | Tame the Emulator UI |

## Important Notes

1. **Use `demo-` project IDs for hermetic runs.** Any project id starting with `demo-` disables network egress to real Google services — required for offline dev and trustworthy CI.
2. **`emulators:exec` is the test contract.** It boots, runs your command, tears down, and exits non-zero on failure. Prefer it over backgrounding `emulators:start` in scripts.
3. **The Emulator UI (port 4000) is read/write.** Anyone with localhost access can mutate state — fine for dev, but don't expose the port over a tunnel.
4. **Imports are emulator-format only.** `gcloud firestore export` artifacts are NOT directly importable; round-trip via the Admin SDK or `emulators:export`.
5. **Java is a hard dep** for Firestore, Storage, Pub/Sub, and Database emulators. JDK 11+ in CI images.
6. **Persisted seed dirs belong in git** for reproducible tests; gitignore them only if they contain real user data (which they shouldn't — see workflow #5).

## Documentation

- [Firebase Local Emulator Suite](https://firebase.google.com/docs/emulator-suite)
- [Install, configure, and integrate](https://firebase.google.com/docs/emulator-suite/install_and_configure)
- [Connect your app to the emulators](https://firebase.google.com/docs/emulator-suite/connect_and_prototype)
- [`emulators:exec` reference](https://firebase.google.com/docs/emulator-suite/install_and_configure#startup)
- [Security Rules unit testing](https://firebase.google.com/docs/rules/unit-tests)
- [Cloud Functions: unit testing](https://firebase.google.com/docs/functions/unit-testing)
- Companion skills: `use-firebase-cli`, `configure-firebase-rules`, `create-firebase-function`.
