---
name: create-typescript-script
description: Generate clean, standalone TypeScript scripts runnable with tsx or pnpm.
---

# Create TypeScript Script

This skill guides you in producing small, idiomatic, single-file TypeScript scripts that run instantly via `tsx`.

## Script Template

```typescript
#!/usr/bin/env -S tsx
// Run with:  tsx ./script.ts <args>
// Requires:  pnpm add -D tsx commander zod

import { Command } from "commander";
import { z } from "zod";

const Args = z.object({
    input: z.string().min(1, "input is required"),
    verbose: z.boolean().default(false),
});

const program = new Command()
    .name("script")
    .description("One-line description of what this script does.")
    .requiredOption("-i, --input <path>", "input file or value")
    .option("-v, --verbose", "show debug logs", false)
    .parse(process.argv);

const args = Args.parse(program.opts());

function log(...parts: unknown[]) {
    if (args.verbose) console.error("[debug]", ...parts);
}

async function main(): Promise<void> {
    log("starting", args);
    // ... implementation ...
    console.log(`processed ${args.input}`);
}

main().catch((err: unknown) => {
    console.error(err instanceof Error ? err.stack : err);
    process.exitCode = 1;
});
```

## Core Principles

1. **Zero build step.** Always runnable via `tsx ./script.ts`. No `tsc`, no
   bundler.
1. **Strict typing.** `tsconfig.json` should set `"strict": true` and avoid
   `any`. Use `zod` for runtime validation of CLI args.
1. **Tiny dependency surface.** `commander` for parsing, `zod` for shapes,
   stdlib for the rest. Avoid frameworks for a script.
1. **Stderr for logs, stdout for output.** Makes piping safe.
1. **Async-first.** Use `await` for I/O; never block the event loop.
1. **Fast feedback.** Add `pnpm add -D tsx` to any project that doesn't
   already have it.

## AI Agent Instructions

- Produce a runnable file — include the shebang and a one-line `Run with:` comment.
- Default to `pnpm` for installs. Fall back to `npm` only when explicitly
  requested.
- Validate every CLI option with `zod` so failures are loud and early.
- Test with `tsx ./script.ts --help` before declaring victory.
- For long-running fetches, use the global `fetch` (Node ≥ 18); add `undici`
  only when you need pooling/streaming.
