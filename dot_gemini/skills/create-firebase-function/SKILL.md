---
name: create-firebase-function
description: Scaffold a Cloud Functions for Firebase function (Node.js or Python) — HTTP, callable, Firestore/Storage triggers, scheduled, and pub/sub variants — with deploy and emulator workflow.
---

# Create Firebase Function

Cloud Functions for Firebase are managed serverless functions that integrate natively with Firebase services (Auth, Firestore, Storage, FCM, App Hosting). Use this skill to scaffold a single function — HTTP, callable, trigger-based, or scheduled — in either Node.js or Python.

The `firebase/agent-skills` bundle covers Firestore / App Hosting / AI Logic in depth, but **does not** ship a Functions-scaffolding skill. This skill fills that gap.

## When to Trigger

- The user wants to add server-side logic to a Firebase project (webhook, callable, trigger, cron).
- The user mentions Cloud Functions, `onRequest`, `onCall`, `onDocumentCreated`, scheduled functions, or Pub/Sub triggers.

## One-time Project Setup

```bash
# Init Functions in an existing Firebase project (interactive — pick language).
firebase init functions

# Pick: TypeScript (recommended), JavaScript, or Python.
# Pick: ESLint or not.
# Pick: install deps now or later.
```

A typical layout after init:

```text
functions/
├── package.json (or requirements.txt for Python)
├── tsconfig.json (Node only)
├── src/                    # Node entry: src/index.ts
│   └── index.ts
└── main.py                 # Python entry
```

## Node.js — v2 SDK (recommended)

```typescript
// functions/src/index.ts
import { onRequest, onCall, HttpsError } from "firebase-functions/v2/https";
import { onDocumentCreated } from "firebase-functions/v2/firestore";
import { onSchedule } from "firebase-functions/v2/scheduler";
import { onMessagePublished } from "firebase-functions/v2/pubsub";
import { logger } from "firebase-functions/v2";
import { initializeApp } from "firebase-admin/app";
import { getFirestore } from "firebase-admin/firestore";

initializeApp();
const db = getFirestore();

// HTTP function — public webhook.
export const webhook = onRequest(
  { region: "us-central1", maxInstances: 10, cors: false },
  async (req, res) => {
    if (req.method !== "POST") {
      res.status(405).send("Method not allowed");
      return;
    }
    logger.info("webhook received", { headers: req.headers });
    res.json({ ok: true });
  },
);

// Callable function — auto-authenticated via Firebase Auth.
export const helloAuthed = onCall(
  { region: "us-central1", enforceAppCheck: true },
  async (request) => {
    if (!request.auth) {
      throw new HttpsError("unauthenticated", "Sign in required");
    }
    return { uid: request.auth.uid, message: `Hello ${request.auth.token.email}` };
  },
);

// Firestore trigger.
export const onUserCreated = onDocumentCreated(
  { document: "users/{uid}", region: "us-central1" },
  async (event) => {
    const data = event.data?.data();
    logger.info("new user", { uid: event.params.uid, data });
  },
);

// Scheduled function (cron).
export const dailyCleanup = onSchedule(
  { schedule: "every day 03:00", timeZone: "Europe/Paris", region: "us-central1" },
  async () => {
    const cutoff = Date.now() - 30 * 24 * 60 * 60 * 1000;
    const stale = await db.collection("sessions").where("updatedAt", "<", cutoff).get();
    await Promise.all(stale.docs.map((d) => d.ref.delete()));
    logger.info("cleanup done", { deleted: stale.size });
  },
);

// Pub/Sub trigger.
export const onOrderEvent = onMessagePublished(
  { topic: "orders", region: "us-central1" },
  async (event) => {
    const payload = event.data.message.json;
    logger.info("order event", payload);
  },
);
```

`functions/package.json` (key parts):

```json
{
  "main": "lib/index.js",
  "scripts": {
    "build": "tsc",
    "build:watch": "tsc --watch",
    "serve": "npm run build && firebase emulators:start --only functions",
    "deploy": "firebase deploy --only functions"
  },
  "dependencies": {
    "firebase-admin": "^12",
    "firebase-functions": "^6"
  },
  "devDependencies": {
    "typescript": "^5",
    "@types/node": "^22"
  },
  "engines": { "node": "22" }
}
```

## Python (Preview, GA in 2025)

```python
# functions/main.py
from firebase_functions import https_fn, firestore_fn, scheduler_fn, options, logger
from firebase_admin import initialize_app, firestore

initialize_app()

@https_fn.on_request(region="us-central1", max_instances=10)
def webhook(req: https_fn.Request) -> https_fn.Response:
    if req.method != "POST":
        return https_fn.Response("Method not allowed", status=405)
    logger.info("webhook received", {"headers": dict(req.headers)})
    return https_fn.Response('{"ok": true}', headers={"Content-Type": "application/json"})

@https_fn.on_call(region="us-central1", enforce_app_check=True)
def hello_authed(req: https_fn.CallableRequest):
    if req.auth is None:
        raise https_fn.HttpsError(https_fn.FunctionsErrorCode.UNAUTHENTICATED, "Sign in required")
    return {"uid": req.auth.uid, "message": f"Hello {req.auth.token.get('email')}"}

@firestore_fn.on_document_created(document="users/{uid}", region="us-central1")
def on_user_created(event: firestore_fn.Event[firestore_fn.DocumentSnapshot]) -> None:
    data = event.data.to_dict() if event.data else None
    logger.info("new user", {"uid": event.params["uid"], "data": data})

@scheduler_fn.on_schedule(schedule="every day 03:00", timezone="Europe/Paris", region="us-central1")
def daily_cleanup(event: scheduler_fn.ScheduledEvent) -> None:
    db = firestore.client()
    cutoff_ms = int(__import__("time").time() * 1000) - 30 * 24 * 60 * 60 * 1000
    stale = db.collection("sessions").where("updatedAt", "<", cutoff_ms).stream()
    for doc in stale:
        doc.reference.delete()
```

`functions/requirements.txt`:

```text
firebase-functions>=0.4
firebase-admin>=6.5
```

## Local Dev (Emulators)

```bash
# Build TS once, then start emulators (or use `npm run serve`).
cd functions && npm run build
firebase emulators:start --only functions,firestore,auth

# Dev loop with TypeScript watch.
npm run build:watch                 # in one shell
firebase emulators:start --only functions   # in another
```

The Functions emulator surfaces logs at <http://localhost:4000>.

## Deploy

```bash
# Single function (faster than the whole bundle).
firebase deploy --only functions:webhook

# Multiple.
firebase deploy --only functions:webhook,functions:helloAuthed

# Full bundle.
firebase deploy --only functions

# Per-environment via aliases.
firebase deploy --only functions -P staging
firebase deploy --only functions -P prod
```

## Calling From Clients

```typescript
// Client-side (web SDK).
import { getFunctions, httpsCallable } from "firebase/functions";
const fns = getFunctions();
const hello = httpsCallable(fns, "helloAuthed");
const { data } = await hello({});
```

## Function Variants Cheat Sheet

| Trigger | Node v2 import | Python import |
|---------|---------------|---------------|
| HTTP webhook | `https.onRequest` | `https_fn.on_request` |
| Authed callable | `https.onCall` | `https_fn.on_call` |
| Firestore | `firestore.onDocumentCreated/Updated/Deleted/Written` | `firestore_fn.on_document_*` |
| Storage | `storage.onObjectFinalized/Deleted` | `storage_fn.on_object_*` |
| Auth | `identity.beforeUserCreated/Signed-in` | `identity_fn.*` |
| Pub/Sub | `pubsub.onMessagePublished` | `pubsub_fn.on_message_published` |
| Scheduled (cron) | `scheduler.onSchedule` | `scheduler_fn.on_schedule` |
| Eventarc | `eventarc.onCustomEventPublished` | `eventarc_fn.*` |

## Common Workflows

**Bootstrap a new function inside an existing project.**

1. `firebase init functions` (skip if `functions/` already exists).
2. Add the function to `src/index.ts` (Node) or `main.py` (Python).
3. Test in the emulator.
4. `firebase deploy --only functions:<name>`.

**Promote staging → prod.**

1. `firebase deploy --only functions:<name> -P staging`.
2. Verify with `firebase functions:log -P staging --only <name>`.
3. `firebase deploy --only functions:<name> -P prod`.

**Debug a deployed function.**

```bash
firebase functions:log --only webhook
gcloud logging read 'resource.type="cloud_function" AND resource.labels.function_name="webhook"' --limit=50 --format=json
```

## Important Notes

1. **Always v2 SDK for new functions** — `firebase-functions/v2/*` (Node) and `firebase_functions.*_fn` (Python). v1 is legacy.
2. **Set `region` explicitly** — defaults can shift; pin per function. Use `setGlobalOptions({ region: "..." })` to apply once.
3. **`enforceAppCheck: true` on callables** when the client is mobile/web — blocks tampered calls.
4. **Don't share state between invocations** beyond top-level cold-start init (Admin SDK, DB clients are fine).
5. **Scheduled functions cost $$ even when no work to do** — combine related crons rather than one per minute.
6. **First deploy enables `cloudbuild.googleapis.com`, `cloudfunctions.googleapis.com`, etc.** — the deploy fails the first time on a new project until those finish provisioning.

## Documentation

- [Cloud Functions for Firebase](https://firebase.google.com/docs/functions)
- [Get started (v2)](https://firebase.google.com/docs/functions/get-started)
- [HTTP triggers](https://firebase.google.com/docs/functions/http-events)
- [Callable functions](https://firebase.google.com/docs/functions/callable)
- [Firestore triggers](https://firebase.google.com/docs/functions/firestore-events)
- [Scheduled functions](https://firebase.google.com/docs/functions/schedule-functions)
- [Python (Preview)](https://firebase.google.com/docs/reference/functions/2nd-gen/python)
- [Function emulator](https://firebase.google.com/docs/emulator-suite/connect_functions)
- Companion skills: `use-firebase-cli`, `configure-firebase-app-hosting`.
