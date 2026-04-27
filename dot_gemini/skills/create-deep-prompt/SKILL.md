---
name: create-deep-prompt
description: "Turn a rough research idea into a structured deep research prompt with `deep-prompt`."
---

# Create Deep Prompt

Use `deep-prompt` (`~/.local/bin/deep-prompt`) when the user has a rough idea, not a finished deep research prompt.

## Usage

```bash
deep-prompt "Compare cloud GPU providers for ML training"

deep-prompt --colab "Research AI regulation trends"

deep-prompt --colab "Compare cloud GPU providers" | deep-research --max
```

## Workflow

1. Start with the user's rough research idea.
2. Run `deep-prompt` directly for a one-shot structured prompt.
3. Use `--colab` when the request is underspecified; it asks targeted questions, then prints the final prompt to stdout.
4. Pipe the result stdout or redirect it to a file.

## Notes

- Output is only the final prompt, so it composes cleanly with pipes.
- Requires `GOOGLE_API_KEY` in the environment.
- Pair with the `run-deep-research` skill to execute the prompt end-to-end.

## Documentation

- [Gemini API — get an API key](https://ai.google.dev/gemini-api/docs/api-key)
- [Gemini API — Deep Research models](https://ai.google.dev/gemini-api/docs/models)
- Local binary source: `~/.local/bin/deep-prompt`
