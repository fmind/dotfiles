---
name: use-clasp-cli
description: Guide for using clasp (Command Line Apps Script Projects) to develop, push, pull, and deploy Google Apps Script projects locally, including MCP mode for coding agents.
---

# Use clasp CLI

`clasp` is Google's CLI for managing [Apps Script](https://developers.google.com/apps-script) projects from the local filesystem instead of the web editor. It also has an **MCP mode** that exposes Apps Script operations as tools to coding agents.

## One-time Setup

```bash
# Enable the Apps Script API for your account (one-time, browser):
#   https://script.google.com/home/usersettings

# Login (opens browser).
clasp login
clasp login --status

# For headless / CI use:
clasp login --no-localhost
clasp login --creds creds.json     # use a service-account / OAuth client
```

## Project Lifecycle

```bash
# Create a new standalone script.
clasp create --type standalone --title "My Script"

# Or bind a script to an existing Doc/Sheet/Form.
clasp create --type sheets --title "Sheet Helper" --parentId <drive-file-id>

# Clone an existing script by its scriptId.
clasp clone <scriptId>

# Push local files to Apps Script (respects .claspignore).
clasp push
clasp push --watch                 # auto-push on file change
clasp push --force                 # overwrite manifest changes

# Pull remote changes back.
clasp pull
```

## Run, Logs, Open

```bash
clasp open                         # open the script in the editor
clasp open --webapp                # open the latest deployed web app
clasp logs                         # tail Stackdriver logs
clasp logs --watch
clasp run myFunction               # execute a function remotely (Executable API)
clasp run myFunction --params '[1,2,3]'
```

## Versions & Deployments

clasp v3 renamed the deployment commands; the old verbs (`deploy`, `undeploy`, `deployments`) are no longer valid except where the README documents them as aliases (`redeploy` is still an alias for `create-deployment -i <id>`).

```bash
# Pin a version (immutable snapshot).
clasp create-version "v1: initial release"
clasp list-versions

# Create / update / list / delete deployments.
clasp create-deployment --description "prod" --versionNumber 3
clasp list-deployments
clasp update-deployment <deploymentId> --versionNumber 4   # or: clasp redeploy <deploymentId>
clasp delete-deployment <deploymentId>
clasp delete-deployment --all
```

## Project Layout

```text
.clasp.json          # scriptId, rootDir, parentId
.claspignore         # files to skip on `clasp push`
appsscript.json      # Apps Script manifest (scopes, runtime, time zone)
src/
  Code.js            # → Code.gs on the server
  ui/Sidebar.html    # HTML files preserve their extension
```

The `.clasp.json` minimum:

```json
{
  "scriptId": "1abc...",
  "rootDir": "src"
}
```

`.claspignore` example:

```text
**/node_modules/**
**/.git/**
**/*.test.ts
README.md
```

## TypeScript

clasp transpiles `.ts` → `.gs` automatically when the project has a `tsconfig.json`. Add type definitions:

```bash
npm install --save-dev @types/google-apps-script
```

Reference `appsscript.json` to enable advanced services (Drive, Sheets, Gmail, etc.) before calling them in code.

## MCP Mode (Coding Agents)

clasp can run as an MCP server so coding agents can drive it without spawning a shell process per call:

```bash
# Start the MCP server (STDIO transport).
clasp mcp
```

Register it in Gemini CLI's MCP config (`~/.gemini/settings.json` for global, `.gemini/settings.json` for project scope):

```json
{
  "mcpServers": {
    "clasp": { "command": "clasp", "args": ["mcp"] }
  }
}
```

Once registered, the agent gets typed tools like `clasp_push`, `clasp_run`, `clasp_logs`, etc., and skips per-call auth prompts.

## Common Workflows

**Bootstrap a new Apps Script project locally.**

1. `mkdir my-script && cd my-script`
2. `clasp create --type standalone --rootDir ./src --title "My Script"`
3. Add `appsscript.json`, write code under `src/`.
4. `clasp push` → `clasp open` to verify in the web editor.

**Promote dev → prod.**

1. `clasp push`
2. `clasp create-version "release notes"`
3. `clasp create-deployment --description "prod" --versionNumber <n>`

**Read recent errors.**

```bash
clasp logs --json | jq 'select(.severity=="ERROR")'
```

## Important Notes

1. The Apps Script API must be enabled at <https://script.google.com/home/usersettings> before `clasp login` works.
2. `clasp push` is destructive on the server side — it overwrites the remote with the local snapshot. Always `clasp pull` first if anyone might have edited the script in the web UI.
3. `appsscript.json` controls runtime (V8 vs Rhino), OAuth scopes, advanced services. Editing it requires `clasp push --force`.
4. `clasp run` requires an OAuth client with the Executable API scope and the function deployed as an API executable.
5. Bound scripts (Sheets/Docs/Forms) need `--parentId`; otherwise the script is created as standalone.

## Documentation

- [clasp on GitHub](https://github.com/google/clasp)
- [Apps Script reference](https://developers.google.com/apps-script/reference)
- [Manifest (`appsscript.json`)](https://developers.google.com/apps-script/concepts/manifests)
