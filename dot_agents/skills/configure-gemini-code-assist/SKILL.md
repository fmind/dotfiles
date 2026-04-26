---
name: configure-gemini-code-assist
description: Guide for installing and customizing Gemini Code Assist on a GitHub repository, including the .gemini/config.yaml schema and styleguide.md.
---

# Configure Gemini Code Assist for GitHub

This skill documents how to enable Gemini Code Assist on a GitHub repository — the GitHub App that auto-summarizes pull requests and posts in-depth code reviews — and how to tune its behavior with `.gemini/config.yaml` and `.gemini/styleguide.md`.

## How it Works

Gemini Code Assist is delivered as a GitHub App. Once installed on a repo, every opened pull request triggers an automatic review (configurable). Reviewers can prompt the agent in PR comments using the `/gemini` tag (e.g. `/gemini summary`, `/gemini review`, free-form questions).

For security, the agent always **excludes `.github/workflows/`** files from reviews to prevent insecure CI changes from being suggested or approved.

## One-time Setup (Consumer / Free Tier)

1. Open the [Gemini Code Assist GitHub App](https://github.com/apps/gemini-code-assist).
2. Click **Install**, choose the personal account or organization.
3. Pick **All repositories** or a curated list, then **Complete setup** in the admin console.
4. Accept the terms — the app starts reviewing new PRs immediately.

## One-time Setup (Enterprise / Google Cloud)

Required for organizations that want IAM-controlled access and group configurations:

1. Grant the operator IAM roles `roles/serviceusage.serviceUsageAdmin` and `roles/geminicodeassistmanagement.scmConnectionAdmin`.
2. In the Google Cloud Console, open **Gemini Code Assist → Agents & Tools** and create a **Developer Connect** connection to GitHub (created in `us-east1` — existing Developer Connect connections are not reused).
3. Authenticate via the GitHub OAuth flow and select the repositories to link.
4. Optionally create a **Group Configuration** to share settings across repos.

## Repository Configuration

Add a `.gemini/` folder at the repository root:

```
.gemini/
├── config.yaml      # behavior tuning
└── styleguide.md    # repo-specific review rules (free-form Markdown)
```

Repository-level `config.yaml` overrides any group-level settings.

## `config.yaml` Schema

```yaml
# .gemini/config.yaml — defaults shown below.

have_fun: false                         # Adds a poem to the initial PR summary.

ignore_patterns: []                     # Glob patterns excluded from review.
                                        # Example: ["dist/**", "**/*.lock", "vendor/**"]

memory_config:
  disabled: false                       # Disable persistent cross-repo memory
                                        # for THIS repo only (group configs).

code_review:
  disable: false                        # Master switch — true silences the agent.
  comment_severity_threshold: MEDIUM    # LOW | MEDIUM | HIGH | CRITICAL
                                        # Comments below this severity are dropped.
  max_review_comments: -1               # -1 = unlimited.
  pull_request_opened:
    code_review: true                   # Post a full review when a PR opens.
    summary: false                      # Post the PR summary on open.
    help: false                         # Post the /gemini help message on open.
    include_drafts: true                # Run on draft PRs as well.
```

### Severity Threshold Cheatsheet

| Threshold | Drops |
|-----------|------|
| `LOW`      | Nothing — every comment is posted |
| `MEDIUM`   | Nit-picks |
| `HIGH`     | + minor refactorings |
| `CRITICAL` | Only blocking issues (security, correctness) |

## `styleguide.md`

Free-form Markdown that the agent reads on every review. Use it to encode review rules that are specific to this repo — naming conventions, banned APIs, performance rules, security gates, language-specific idioms, etc.

```markdown
# Project Code Review Rules

- Prefer `structlog` over the stdlib `logging` module.
- All public functions must have type hints and a one-line docstring.
- Never call `requests`; use the project's `httpx` wrapper in `src/http.py`.
- Mark every TODO with an owner and a tracking issue.
```

Keep it tight (a few hundred lines max) — the agent reads the whole file as context for every PR.

## Useful PR Comment Triggers

| Comment | Effect |
|---------|--------|
| `/gemini` | Show the help menu |
| `/gemini summary` | Re-generate the PR summary |
| `/gemini review` | Re-run the full code review |
| `/gemini <question>` | Ask a free-form question about the PR |

## Recommended Starting Configuration

Reasonable defaults for a polyglot project that wants signal over noise:

```yaml
have_fun: false
ignore_patterns:
  - "**/*.lock"
  - "**/*.snap"
  - "dist/**"
  - "vendor/**"
  - "**/generated/**"
code_review:
  comment_severity_threshold: HIGH
  max_review_comments: 20
  pull_request_opened:
    code_review: true
    summary: true
    help: false
    include_drafts: false
```

## Important Notes

1. **Workflows are always skipped:** `.github/workflows/` is hard-excluded — you cannot opt back in.
2. **Override hierarchy:** repo `config.yaml` > group config > built-in defaults.
3. **Apply on push:** changes to `.gemini/config.yaml` and `styleguide.md` take effect on the next PR event; no redeploy needed.
4. **Privacy:** for the consumer tier, code is processed under the consumer Gemini Terms; use the enterprise tier for IAM-controlled handling.

## Documentation

- [Review repository code (overview)](https://developers.google.com/gemini-code-assist/docs/review-repo-code)
- [Customize Gemini Code Assist behavior](https://developers.google.com/gemini-code-assist/docs/customize-repo-review)
- [Set up Gemini Code Assist on GitHub](https://developers.google.com/gemini-code-assist/docs/set-up-code-assist-github)
- [Code review style guide reference](https://developers.google.com/gemini-code-assist/docs/code-review-style-guide)
- [GitHub App install page](https://github.com/apps/gemini-code-assist)
