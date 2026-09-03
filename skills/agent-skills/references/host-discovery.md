# Host Discovery

How each host finds the persona, global skills, and workspace skills, and the read-only command that lists what it loaded. Global paths follow the dotfiles layout where every host path links back to `~/.agents/AGENTS.md` and `~/.agents/skills`.

| Host        | Persona                                              | Global skills                                                       | Workspace skills                                        | Read-only listing                           |
| ----------- | ---------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------- | ------------------------------------------- |
| Antigravity | `~/.gemini/GEMINI.md`                                | `~/.gemini/config/skills` (link to `~/.agents/skills`)              | `.agents/skills`                                        | `/skills` inside the session (unverified)   |
| Claude Code | `~/.claude/CLAUDE.md`                                | `~/.claude/skills` (link to `~/.agents/skills`)                     | `.claude/skills` (link to `../.agents/skills`)          | `/skills` inside the session                |
| Codex       | `~/.codex/AGENTS.md`                                 | `~/.agents/skills`                                                  | `.agents/skills`                                        | `codex debug prompt-input`                  |
| Copilot     | `~/.copilot/copilot-instructions.md`                 | `~/.copilot/skills` or `~/.agents/skills`                           | `.github/skills`, `.agents/skills`, or `.claude/skills` | `copilot skill list`                        |
| Grok        | `~/.grok/AGENTS.md`                                  | `~/.grok/skills` (link to `~/.agents/skills`)                       | `.agents/skills`                                        | `grok inspect`                              |
| OpenCode    | `instructions` in `~/.config/opencode/opencode.json` | `~/.config/opencode/skills`, `~/.agents/skills`, `~/.claude/skills` | `.opencode/skills` or `.agents/skills`                  | `opencode debug skill`                      |

## Reading the output

- `codex debug prompt-input` renders the model-visible prompt as JSON: check that every expected skill name and full description appear and keep any shortening warning; record the CLI version and model, because the metadata budget depends on them.
- `opencode debug skill` prints every loaded skill with its `location`; a skill at the wrong path is simply absent.
- `copilot skill list` groups skills by source (project, personal, plugin); `--json` gives machine-readable output.
- `grok inspect` lists project instructions, permissions, and every skill with its scope (`project` or `user`); `--json` is available.
- Claude Code and Antigravity expose `/skills` in the interactive session only; explicit invocation (`/<skill-name>`) is the fallback proof in Claude.
- A listed skill proves inclusion in the prompt, not instruction following; validate behavior in a disposable, instrumented run per [agent-evaluation](../agent-evaluation/SKILL.md).
