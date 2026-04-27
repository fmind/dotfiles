---
name: use-google-workspace-cli
description: Guide for using the Google Workspace CLI (gws) to interact with Google Workspace APIs.
---

# Use Google Workspace CLI (gws)

This skill documents how to use the Google Workspace CLI (`gws`) to interact with all Google Workspace APIs (Drive, Gmail, Calendar, Sheets, Docs, Chat, Admin, etc.).

## How it Works

`gws` does not ship with a static list of commands. Instead, it reads Google's [Discovery Service](https://developers.google.com/discovery) at runtime to dynamically build its command surface. When Google adds an API endpoint or method, `gws` picks it up automatically.

For AI agents, `gws` is particularly powerful because every response is structured JSON. There is zero boilerplate, and it allows agents to interact with Google Workspace without custom tooling.

## Usage Pattern

Invoke the CLI directly using `gws`:

```bash
gws <service> <resource> [sub-resource] <method> [flags]
```

## Exploring API Schemas

Since the CLI is built dynamically, you can introspect any method's request and response schema using the `schema` command. This will detail the exact parameters and payload required:

```bash
# Discover parameters and payload format for drive.files.list
gws schema drive.files.list

# Discover parameters for gmail messages
gws schema gmail.users.messages.list
```

## Examples

### Google Drive

```bash
# List the 10 most recent files
gws drive files list --params '{"pageSize": 10}'

# Stream paginated results as NDJSON and extract file names
gws drive files list --params '{"pageSize": 100}' --page-all | jq -r '.files[].name'

# Get metadata of a specific file
gws drive files get --params '{"fileId": "abc123"}'
```

### Google Sheets

```bash
# Create a spreadsheet
gws sheets spreadsheets create --json '{"properties": {"title": "Q1 Budget"}}'

# Get spreadsheet details
gws sheets spreadsheets get --params '{"spreadsheetId": "YOUR_SPREADSHEET_ID"}'
```

### Gmail

```bash
# List recent emails for the authenticated user
gws gmail users messages list --params '{"userId": "me", "maxResults": 5}'
```

### Google Chat

```bash
# Send a Chat message (using dry-run to test)
gws chat spaces messages create \
  --params '{"parent": "spaces/xyz"}' \
  --json '{"text": "Deploy complete."}' \
  --dry-run
```

## Key Flags

- `--params <JSON>`: Passes URL/Query parameters as JSON.
- `--json <JSON>`: Passes the Request body as JSON (for POST/PATCH/PUT methods).
- `--format <FMT>`: Output format. Options are `json` (default), `table`, `yaml`, `csv`.
- `--page-all`: Auto-paginate to fetch multiple pages of results (outputs as JSON lines / NDJSON).
- `--page-limit <N>`: Sets the maximum number of pages to fetch (default: 10).
- `--output <PATH>`: Saves binary responses directly to a file path.
- `--dry-run`: Previews requests without executing them.
- `--help`: Shows dynamically generated help for any resource or method.

## Important Notes

1. **JSON Strings:** Always ensure that arguments passed to `--params` or `--json` are strictly valid JSON strings enclosed in single quotes.
2. **Error Codes:**
   - `Exit Code 1`: API Error (Google returned an error).
   - `Exit Code 2`: Auth Error (Credentials missing or invalid).
3. **Authentication:** If you encounter an Auth Error, the user may need to authenticate using `gws auth login` or set up an OAuth client via `gws auth setup`. Check status with `gws auth status`.

## Documentation

- [GitHub: googleworkspace/cli](https://github.com/googleworkspace/cli)
- [Google APIs Discovery Service](https://developers.google.com/discovery)
- [Google Workspace Developers home](https://developers.google.com/workspace)
- [Configure Google Workspace MCP servers](https://developers.google.com/workspace/guides/configure-mcp-servers)
