---
name: install-firebase-skills
description: Install Firebase's official Agent Skills bundle so the agent gains expert knowledge for Firestore, Auth, Hosting, App Hosting, Genkit, AI Logic, and security rules.
---

# Install Firebase Skills

Firebase publishes the official [`firebase/agent-skills`](https://github.com/firebase/agent-skills) bundle. It's the canonical source for Firebase-specific guidance; this skill explains when and how to install it.

## When to Trigger

- The repo contains `firebase.json`, `firestore.rules`, `apphosting.yaml`, `genkit.config.ts`, or imports from `firebase/*` / `firebase-admin`.
- The user mentions Firebase, Firestore, Firebase Auth, App Hosting, Genkit, Firebase AI Logic, or security rules.
- Verify first: `ls .agents/skills/ ~/.gemini/skills/ 2>/dev/null | grep -i firebase`. If installed, skip.

## Install

```bash
# Interactive: pick the Firebase skills you want.
npx skills add firebase/agent-skills
```

`npx skills` writes to `.agents/skills/` for project scope (default — repo-pinned, commits with the codebase) or to `~/.gemini/skills/` for global scope when prompted.

## What Gets Installed

13 skills at the time of writing (verify with `npx skills add firebase/agent-skills --list`):

**Core**
- `firebase-basics` — initial setup and platform workflows.

**Auth & Hosting**
- `firebase-auth-basics` — sign-in / authentication.
- `firebase-hosting-basics` — static sites and SPAs.
- `firebase-app-hosting-basics` — Next.js, Angular, etc., on App Hosting.

**Data**
- `firebase-firestore-standard` — Firestore essentials.
- `firebase-firestore-enterprise-native-mode` — Firestore Enterprise Native Mode.
- `firebase-data-connect-basics` — Data Connect (PostgreSQL + GraphQL).
- `firebase-security-rules-auditor` — vulnerability assessment for `firestore.rules`.

**AI**
- `firebase-ai-logic-basics` — Firebase AI Logic (Gemini API for web/mobile).
- `developing-genkit-js` — Genkit in Node.js / TypeScript.
- `developing-genkit-go` — Genkit in Go.
- `developing-genkit-dart` — Genkit for Dart / Flutter.
- `developing-genkit-python` — Genkit in Python (Preview).

## Related: `firebase` CLI

The skills assume the Firebase CLI (installed via mise) is logged in:

```bash
firebase login
firebase projects:list
firebase use <project-id>
firebase init                     # interactive feature picker
firebase emulators:start          # local Auth/Firestore/Functions/etc.
firebase deploy --only hosting,functions
```

## After Install

1. Restart the agent so the new skills are picked up by progressive disclosure.
2. Project-scope installs (`.agents/skills/`) commit naturally with the repo. Global-scope installs are machine-local — `chezmoi add ~/.gemini/skills/<slug>` to track them.
3. The `firebase-security-rules-auditor` skill can be invoked explicitly any time `firestore.rules` is touched.

## Important Notes

1. This skill installs the bundle — it does **not** duplicate Firebase docs. Defer Firebase-specific guidance to the installed skills.
2. Firebase also appears in `google/skills` as `firebase-basics`; the dedicated `firebase/agent-skills` bundle is more comprehensive — prefer it.
3. The Genkit skills assume a Genkit project layout; don't apply them to plain Cloud Functions code.

## Documentation

- [Firebase agent skills docs](https://firebase.google.com/docs/ai-assistance/agent-skills)
- [`firebase/agent-skills` repo](https://github.com/firebase/agent-skills)
- [Firebase blog announcement (Feb 2026)](https://firebase.blog/posts/2026/02/ai-agent-skills-for-firebase/)
- [Firebase CLI reference](https://firebase.google.com/docs/cli)
