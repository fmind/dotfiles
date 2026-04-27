---
name: configure-firebase-rules
description: Guide for authoring and testing Firestore and Cloud Storage security rules — request/auth context, common patterns, helper functions, indexes, and emulator-based unit tests.
---

# Configure Firebase Security Rules

Firebase Security Rules guard Firestore and Cloud Storage access at the API edge — every read/write request from a client SDK is evaluated against the active ruleset before the database/storage layer is touched.

This skill covers authoring (`firestore.rules`, `storage.rules`), unit testing in the emulator, and deployment. The companion bundle `firebase/agent-skills` ships a dedicated `firebase-security-rules-auditor` skill (install via `install-firebase-skills`) that performs deeper vulnerability assessments.

## File Layout

```
firebase.json                  # references the rules + index files
firestore.rules                # Firestore security rules
firestore.indexes.json         # composite indexes
storage.rules                  # Cloud Storage security rules
```

`firebase.json` minimum:

```json
{
  "firestore": {
    "rules": "firestore.rules",
    "indexes": "firestore.indexes.json"
  },
  "storage": { "rules": "storage.rules" }
}
```

## Firestore Rules — Skeleton

```
rules_version = '2';

service cloud.firestore {
  match /databases/{database}/documents {

    // Helpers (DRY auth checks).
    function isSignedIn() { return request.auth != null; }
    function isOwner(uid) { return isSignedIn() && request.auth.uid == uid; }
    function hasRole(role) {
      return isSignedIn()
        && get(/databases/$(database)/documents/users/$(request.auth.uid)).data.role == role;
    }

    // Default deny.
    match /{document=**} { allow read, write: if false; }

    // Per-collection rules.
    match /users/{userId} {
      allow read: if isOwner(userId) || hasRole('admin');
      allow write: if isOwner(userId)
        && request.resource.data.keys().hasOnly(['name', 'email', 'photoURL']);
    }

    match /posts/{postId} {
      allow read: if true;                               // public reads
      allow create: if isSignedIn()
        && request.resource.data.authorId == request.auth.uid;
      allow update, delete: if isOwner(resource.data.authorId);
    }
  }
}
```

## Storage Rules — Skeleton

```
rules_version = '2';

service firebase.storage {
  match /b/{bucket}/o {

    function isSignedIn() { return request.auth != null; }
    function isOwner(uid) { return isSignedIn() && request.auth.uid == uid; }
    function isImage()    { return request.resource.contentType.matches('image/.*'); }
    function under(maxBytes) { return request.resource.size < maxBytes; }

    match /users/{userId}/avatar/{fileName} {
      allow read:  if true;
      allow write: if isOwner(userId) && isImage() && under(2 * 1024 * 1024);
    }

    match /{path=**} { allow read, write: if false; }     // default deny
  }
}
```

## Common Patterns

```
// Owner-only writes, public reads.
match /docs/{id} {
  allow read: if true;
  allow write: if isSignedIn()
    && (request.resource.data.ownerId == request.auth.uid
        || resource.data.ownerId == request.auth.uid);
}

// Append-only collection (no updates/deletes).
match /events/{id} {
  allow create: if isSignedIn();
  allow read: if isSignedIn();
  allow update, delete: if false;
}

// Subcollection inheriting parent permissions.
match /teams/{teamId} {
  allow read, write: if hasRole('admin');
  match /members/{memberId} {
    allow read: if request.auth.uid in resource.data.team.members;
  }
}

// Field-level constraints on writes.
match /profiles/{uid} {
  allow update: if isOwner(uid)
    && request.resource.data.diff(resource.data).changedKeys()
         .hasOnly(['displayName', 'photoURL']);
}
```

## Indexes (`firestore.indexes.json`)

Single-field indexes are automatic; composite indexes must be declared:

```json
{
  "indexes": [
    {
      "collectionGroup": "posts",
      "queryScope": "COLLECTION",
      "fields": [
        { "fieldPath": "authorId", "order": "ASCENDING" },
        { "fieldPath": "publishedAt", "order": "DESCENDING" }
      ]
    }
  ]
}
```

The Firestore SDK surfaces a console URL when a query lacks an index — paste the suggested index into this file and `firebase deploy --only firestore:indexes`.

## Unit Testing (emulator-based)

Install the test framework and write Jest/Vitest cases:

```bash
npm install --save-dev @firebase/rules-unit-testing
```

```typescript
// tests/firestore.rules.test.ts
import { initializeTestEnvironment, assertSucceeds, assertFails } from '@firebase/rules-unit-testing';
import { readFileSync } from 'node:fs';

const env = await initializeTestEnvironment({
  projectId: 'demo-rules',
  firestore: { rules: readFileSync('firestore.rules', 'utf8') },
});

const alice = env.authenticatedContext('alice');
await assertSucceeds(alice.firestore().doc('users/alice').set({ name: 'A' }));
await assertFails(   alice.firestore().doc('users/bob').set({ name: 'B' }));
```

Run against the emulator:

```bash
firebase emulators:exec --only firestore "npm test"
```

## Deploy

```bash
firebase deploy --only firestore:rules
firebase deploy --only storage:rules
firebase deploy --only firestore:indexes,firestore:rules
```

## Lint & Diff

```bash
# Validate syntax without deploying.
firebase deploy --only firestore:rules --dry-run

# Diff staging vs prod (manual but useful).
firebase use staging && firebase functions:config:get > /tmp/staging.json
firebase use prod    && firebase functions:config:get > /tmp/prod.json
diff /tmp/staging.json /tmp/prod.json
```

## Common Workflows

**Add a new collection with rules.**
1. Add a `match /<col>/{id} { allow ... }` block with explicit `allow` predicates.
2. Write unit tests covering authorized + unauthorized paths.
3. `firebase emulators:exec --only firestore "npm test"`.
4. `firebase deploy --only firestore:rules -P staging`, then prod after validation.

**Audit existing rules.**
- Look for `match /{document=**}` blocks granting reads/writes broadly.
- Look for `request.auth != null` as the only check (any signed-in user, including anonymous, satisfies it).
- Look for missing field-key constraints (`request.resource.data.keys().hasOnly([...])`) on writes.
- Run the `firebase-security-rules-auditor` skill (from `firebase/agent-skills`) for a structured pass.

## Important Notes

1. **Default-deny last** — always end with `match /{document=**} { allow read, write: if false; }`.
2. **`request.auth` is null for unauthenticated requests** — handle that branch explicitly. `request.auth != null` does NOT mean a "real" user; anonymous auth users count.
3. **Rules can read documents** via `get()` / `exists()` — but each costs a billable read and slows the rule. Cache role checks in custom claims when possible.
4. **`resource` = current state, `request.resource` = proposed state.** Constraint mismatches between them are the most common rule bugs.
5. **Test before deploying** — production rule mistakes lock real users out and cost a redeploy round-trip to fix.
6. **Indexes deploy independently** — a rules change that requires a new index will fail at runtime if `firestore:indexes` wasn't deployed first.

## Documentation

- [Firestore Security Rules](https://firebase.google.com/docs/firestore/security/get-started)
- [Storage Security Rules](https://firebase.google.com/docs/storage/security)
- [Rules language reference](https://firebase.google.com/docs/rules/rules-language)
- [Unit testing rules](https://firebase.google.com/docs/rules/unit-tests)
- [`@firebase/rules-unit-testing`](https://www.npmjs.com/package/@firebase/rules-unit-testing)
- [Firestore composite indexes](https://firebase.google.com/docs/firestore/query-data/index-overview)
- [Companion: `firebase-security-rules-auditor` skill](https://github.com/firebase/agent-skills)
