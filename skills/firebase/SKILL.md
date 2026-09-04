---
name: firebase
description: "Set up and ship Firebase backends with the Firebase CLI: emulators, Firestore rules, Auth, Functions, Hosting or App Hosting. Use when an app needs Firebase."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/firebase
  created: "2026-09-02"
  updated: "2026-09-03"
---

# Firebase Standard

Firebase is the default backend and hosting for TypeScript web apps: Auth, Firestore, Cloud Functions, Hosting for static bundles, App Hosting for SSR. The `firebase` CLI comes from mise; the upstream skills carry the product detail, so this skill fixes the workflow, the local loop, and the deploy authority.

## 1. Setup

1. **Authenticate once**: `firebase login` (or `firebase login --no-localhost` on a headless machine); `firebase projects:create <project-id>` or `firebase use <project-id>` in the repo.
1. **Initialize features**: `firebase init <feature>` for `firestore`, `hosting`, `functions`, or `apphosting`; answers land in `firebase.json`, `.firebaserc`, `firestore.rules`, and `firestore.indexes.json`, all committed.
1. **Angular integration**: `ng add @angular/fire` in the app per [angular](../angular/SKILL.md); keep the Firebase config in `src/environments/` and never commit service-account keys.

## 2. Local Loop

```bash
firebase emulators:start                      # Auth, Firestore, Functions, Hosting on localhost with the Emulator UI
firebase emulators:exec "mise run test"       # run the suite against emulators, then stop them
```

- Point the app at the emulators in development (`connectAuthEmulator`, `connectFirestoreEmulator`) behind an environment flag, never in production builds.
- Rules are tested with `@firebase/rules-unit-testing` against the emulator; run them from `mise run test` so hooks and CI cover them.

## 3. Deploy

```bash
firebase deploy --only firestore:rules,firestore:indexes   # rules first, reviewed as code
firebase deploy --only hosting                            # static bundle from mise run build
firebase hosting:channel:deploy preview-<branch> --expires 7d   # preview URL for review
firebase deploy --only functions                          # Cloud Functions (2nd gen, Node LTS)
firebase deploy --only apphosting                         # SSR backends (Blaze plan)
```

- Expose deployment as an explicit `deploy` mise task per [mise](../mise/SKILL.md); it never runs from a hook because it mutates production and can cost money.
- CI deploys authenticate with Workload Identity Federation (`google-github-actions/auth`) and `firebase deploy --project <id>`; no `FIREBASE_TOKEN` secrets — see [github-actions](../github-actions/SKILL.md).
- Audit the rules with the upstream security-rules skill before every rules deploy; its score and findings go in the PR.

## Gotchas

- **Rules are the security boundary**: client SDKs talk to Firestore directly, so a permissive rule is a public database; deny by default, validate types and ownership, and test the rules.
- **Two hosting products**: Hosting serves static files and rewrites; App Hosting builds and runs SSR frameworks and needs the Blaze plan; do not mix them for one app.
- **Emulator data is ephemeral**: export with `firebase emulators:export ./emulator-data` when a seed must persist, and gitignore it.
- **Functions have their own `package.json`**: keep them on the Node LTS the runtime supports and lint them with the same Biome config as the app when they are plain TypeScript.
- **Cloud Run remains the default for services**: Firebase covers app backends; standalone APIs and agents ship per [cloud-run](../cloud-run/SKILL.md).

## Official Skills

Upstream: `firebase/agent-skills`. List the current release, then install what the task needs at project scope after reviewing the snapshot (see [agent-skills](../agent-skills/SKILL.md)):

```bash
skills add firebase/agent-skills --list
skills add firebase/agent-skills --skill <name> -y
```

## Documentation

- [Firebase CLI](https://firebase.google.com/docs/cli) · [Emulator Suite](https://firebase.google.com/docs/emulator-suite) · [Firebase agent skills](https://firebase.google.com/docs/ai-assistance/agent-skills) · [firebase/agent-skills](https://github.com/firebase/agent-skills)
- Companion skills: [angular](../angular/SKILL.md), [typescript-stack](../typescript-stack/SKILL.md), [genkit](../genkit/SKILL.md) (AI features on Firebase), [secure](../secure/SKILL.md).
