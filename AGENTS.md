# Agent Guidelines

## Repository Overview

Ansible-based dotfiles repo that automates installation and configuration of developer tools on Linux/macOS.

- **`site.yml`** — Main playbook. Roles are listed alphabetically; commented-out roles are inactive. `nodejs` and `pipx` are loaded in `pre_tasks` as a dependency for other roles.
- **`roles/`** — One directory per tool. Each has `tasks/main.yml` and an optional `files/` dir for configs.
- **`justfile`** — Task runner: `just apply`, `just check`, `just install`, `just shell`, `just docker`.
- **`inventory.ini`** / **`ansible.cfg`** — Local-only inventory and config (auto python interpreter, no deprecation warnings).

## Role Conventions

- **Tags**: `admin` = requires root (`become: true`), `user` = user-level (no sudo).
- **Cross-platform**: Use `when: ansible_facts['distribution'] != 'MacOSX'` for Linux-only tasks.
- **Config files**: Symlinked from `roles/<name>/files/` to `~/` via `state: link, force: true`.
- **Naming**: Role name = tool name (lowercase, hyphenated). Tasks named by purpose: `package`, `config`, `directory`.

## Other Conventions

- **Conventional Commit**: Use Conventional Commits for commit messages.

## Running

```bash
just apply              # Run all active roles
just check              # Syntax check only
just sudo=true apply    # With sudo prompt (for admin-tagged tasks)
ansible-playbook -i inventory.ini site.yml --tags <role>  # Run a single role
```

## Adding a Role

1. Create `roles/<name>/tasks/main.yml` (and optional `files/`)
2. Add `- { role: <name>, tags: ["<name>"] }` to `site.yml` in alphabetical order
3. See `.agent/skills/ansible-role/SKILL.md` for detailed patterns and examples
