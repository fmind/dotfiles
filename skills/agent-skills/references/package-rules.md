# Package Rules

Packaging constraints that `gh skill publish --dry-run` and the catalog tests enforce; the workflow lives in [agent-skills](SKILL.md).

## Catalog portability

- The first-party suite installs as one catalog and may link between packages under `skills/` and `.agents/skills/`.
- Catalog containment does not prove that one copied folder is independently distributable under the [file-reference rule](https://agentskills.io/specification#file-references).
- Before standalone publication, replace sibling links with skill-name routing or bundle the dependency, copy only that folder to a disposable directory, and validate it there.

## Direct disclosure

- Keep progressive-disclosure resources one level deep (`references/`, `templates/`, `scripts/`).
- Every resource other than host-owned `agents/openai.yaml` must be named directly by the root `SKILL.md` or by a local path in that OpenAI metadata; a nested reference does not make another file discoverable.
- Nested instructional Markdown is validated for missing or unsafe package links; output templates under `templates/` may point at destinations that exist only after materialization but still reject unsafe local schemes.
- Test helpers live in a disposable directory; document required tools and side effects; keep executable code minimal and paths relative.

## `agents/openai.yaml`

- Add it only when Codex needs interface metadata or dependencies.
- Use the supported `interface`, `dependencies.tools`, and `policy` fields, mention `$skill-name` in `interface.default_prompt`, and link the file from `SKILL.md`.
