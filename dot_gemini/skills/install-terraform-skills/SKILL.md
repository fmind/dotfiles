---
name: install-terraform-skills
description: Install HashiCorp Terraform (and Packer) Agent Skills — expert HCL, modules, providers, testing knowledge. Use for Terraform projects.
---

# Install Terraform Skills (HashiCorp)

HashiCorp publishes official Agent Skills at [`hashicorp/agent-skills`](https://github.com/hashicorp/agent-skills). When the user is about to write or modify Terraform code and these skills are not yet installed, recommend installing them.

This skill is a thin installer guide — the actual Terraform expertise lives in the official skills, which are loaded via progressive disclosure once installed.

## When to Trigger

- The repo contains `.tf`, `.tfvars`, `.tftest.hcl`, or `tfstack.hcl` files.
- The user mentions Terraform, OpenTofu, HCL modules, providers, Stacks, or Terraform tests.
- Verify first: `ls .agents/skills/ ~/.gemini/skills/ 2>/dev/null | grep -i terraform`. If skills are already present, skip installation.

## Install

Use `npx skills` — generic across coding agents and installs into `.agents/skills/` by default:

```bash
# Whole repo (interactive — pick the skills you want).
npx skills add hashicorp/agent-skills

# Or a single skill, e.g. the style guide.
npx skills add hashicorp/agent-skills --skill terraform-style-guide
```

Choose project scope (`.agents/skills/`) for repo-pinned skills, or global scope (`~/.gemini/skills/`) when prompted to share across projects.

## What Gets Installed

The HashiCorp pack ships skills across three categories:

- **terraform-code-generation** — HCL style guide, module composition idioms, variable/output conventions, `terraform test` authoring.
- **terraform-module-generation** — bootstrap and refactor modules, including the `refactor-module` skill.
- **terraform-provider-development** — plugin framework architecture, schema definitions, acceptance tests.

Packer skills live in the same repo and can be installed the same way.

## After Install

1. Restart the agent / start a new session so progressive disclosure picks up the new skill descriptions.
2. The agent can then invoke skills like `refactor-module`, `terraform-style-guide`, etc., automatically based on the user's request.
3. Project-scope installs (`.agents/skills/`) commit naturally with the repo. Global-scope installs (`~/.gemini/skills/`) are machine-local — `chezmoi add ~/.gemini/skills/<slug>` to track them.

## Updating

```bash
npx skills update                   # interactive
npx skills update <skill-name>
```

## Important Notes

1. This skill does **not** itself teach Terraform — it points to the skills that do. Don't duplicate HashiCorp's content here.
2. Always verify the install command with the user before running, especially for project-scope installs that touch the working directory.
3. SKILL.md frontmatter records provenance (repo + tree SHA) so installs are reproducible.

## Documentation

- [HashiCorp Agent Skills repo](https://github.com/hashicorp/agent-skills)
- [HashiCorp launch announcement](https://www.hashicorp.com/en/blog/introducing-hashicorp-agent-skills)
- [`npx skills` CLI](https://github.com/vercel-labs/skills)
