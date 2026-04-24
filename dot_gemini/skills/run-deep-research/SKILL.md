---
name: run-deep-research
description: "Run Gemini Deep Research with deep-research."
---

# Run Deep Research

Use `deep-research` (`~/.local/bin/deep-research`) to execute a prepared deep
research prompt and print the final report.

## Usage

```bash
deep-research "Compare cloud GPU providers for ML training"

deep-prompt --colab "Research AI regulation trends" | deep-research

deep-prompt "Analyze EV battery competition" | deep-research --max > report.md
```

## Flags

| Flag | Purpose |
|------|---------|
| `--max` | Use `deep-research-max` for broader or more expensive runs |

## Workflow

1. If requested by the user, run `deep-research` with the final prompt from an argument or stdin.
1. Watch stderr for thought summaries and progress while the report is being prepared.
1. Use stdout or redirect it to a file if the user wants to save the report.

## Notes

- The final report is written to stdout, so piping and redirection stay simple.
- Requires `GOOGLE_API_KEY` in the environment.
