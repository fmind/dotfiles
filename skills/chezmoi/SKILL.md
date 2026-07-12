---
name: chezmoi
description: Canonical chezmoi dotfiles setup — the source-attribute naming grammar, Go templates, age-encrypted secrets, and the edit-source then apply/diff workflow. Use when adding, editing, or debugging a managed dotfile.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/chezmoi
  created: 2026-07-12
  updated: 2026-07-12
---

# Chezmoi Dotfiles Standard

Canonical setup for **chezmoi**, the source-of-truth manager for dotfiles across machines. The source tree (`~/.local/share/chezmoi`) is the only thing you edit; `chezmoi apply` renders it into the home directory. File **names** encode target path, permissions, encryption, and rendering, so the grammar below is the core of the tool. Pins the toolchain via the [mise skill](../mise/SKILL.md) and formats configs via the [dprint skill](../dprint/SKILL.md).

## 1. Source-Attribute Naming Grammar

A source name is built from optional prefixes + name + suffixes; chezmoi strips them to compute the target. Combine in this order (type → attributes → `dot_` → name → suffix):

`[create_|modify_|remove_|symlink_]` `[encrypted_]` `[private_]` `[readonly_]` `[executable_]` `[dot_]` `<name>` `[.tmpl]` `[.age]`

| Prefix / suffix                                    | Effect on the target                                                         |
| -------------------------------------------------- | ---------------------------------------------------------------------------- |
| `dot_foo`                                          | `~/.foo` — never write a literal leading dot in a source path.               |
| `private_`                                         | mode `0600`. `executable_` → `0755`. `readonly_` → drop write bits.          |
| `encrypted_` + `.age`                              | decrypted on apply (see §3). Order is `encrypted_private_dot_foo`.           |
| `symlink_<name>.tmpl`                              | a symlink whose **target path is the rendered file content** (verbatim).     |
| `modify_<name>`                                    | script/template that rewrites the _existing_ target (partial ownership, §2). |
| `create_` / `remove_`                              | write only if absent / delete the target.                                    |
| `<name>.tmpl`                                      | Go-template rendered with chezmoi data (§2).                                 |
| `run_[once_\|onchange_][before_\|after_]<name>.sh` | a hook script run during `apply` (§4).                                       |

Run scripts key off content: `run_once_after_*` runs once per unique content hash (bootstrap/install); `run_onchange_after_*` runs whenever the script body changes (regenerate derived state).

## 2. Templates

`*.tmpl` files render with Go `text/template` + [sprig](https://masterminds.github.io/sprig/) plus chezmoi data. Branch on the host with `.chezmoi.os` / `.chezmoi.arch`; reach the home dir with `.chezmoi.homeDir`; read `[data]` keys from the config (§5).

- **Literal delimiters**: to emit a literal `{{ ... }}` (e.g. another tool's template), wrap it in backticks inside a template action — `{{`{{ .Destination }}`}}`.
- **`modify_` templates**: a `modify_` file whose body starts with the `# chezmoi:modify-template` marker is rendered as a template with the **current target piped in on `.chezmoi.stdin`**, so it can merge managed keys into a file another tool also writes (parse with `fromToml`/`fromJson`, merge, re-emit). A `modify_` file is _already_ a script — do **not** add a `.tmpl` suffix.
- **`symlink_` templates**: the rendered content is the link destination, e.g. a one-line `{{ .chezmoi.homeDir }}/.agents/skills`.

## 3. Secrets (age)

Encrypt with **age** (`encryption = "age"` in the config). Name a secret `encrypted_private_dot_<name>.age`; chezmoi decrypts it to a `0600` target on apply using the identity key, and encrypts new secrets to the configured recipient:

```bash
chezmoi add --encrypt ~/.config/<tool>/secret   # imports + encrypts into the source
chezmoi edit ~/.config/<tool>/secret            # edits the plaintext, re-encrypts on save
```

Never commit or apply a decrypted copy — a `*.age` blob is the only committable form. A leaked secret is compromised even after removal; rotate it (see the [security-scan skill](../security-scan/SKILL.md)). Gate machines without the age key by ignoring key-dependent files in `.chezmoiignore`.

## 4. Workflow (edit source → apply)

1. **Edit the source**, never the deployed copy under `~/.config`, `~/.claude`, etc. — an apply would overwrite it.
2. **Start managing** an existing file: `chezmoi add <target>` (adds with inferred attributes; `--encrypt` for secrets, `--template` to templatize).
3. **Preview**: `chezmoi diff` (add `--force` in automation to skip prompts).
4. **Apply**: `chezmoi apply --force` — `--force` is mandatory in scripts/hooks so a changed target never blocks on a prompt. Add `--dry-run` to preview without writing.
5. **Pull target edits back**: `chezmoi re-add` folds manual changes to a managed file back into the source (e.g. a regenerated lockfile).
6. **Diagnose**: `chezmoi doctor` (config, encryption, template health); `chezmoi managed` / `chezmoi unmanaged` list coverage; `chezmoi cd` opens a shell in the source root.

In this repo the tasks are wrapped by [mise](../mise/SKILL.md): `mr a` (`chezmoi apply --force`), `mr d` (diff), `mr x` (`chezmoi doctor` + `mise doctor`), and `dot chezmoi clean` scans for orphaned once-managed files in `$HOME`.

## 5. Config Seed & Ignore

- **`.chezmoi.toml.tmpl`** seeds each machine's `~/.config/chezmoi/chezmoi.toml` on `chezmoi init`. Prompt once for per-host data with `promptStringOnce . "key" "question" "default"`, and set `encryption`, the `[age]` identity/recipient, and `[edit] apply = true` (so `chezmoi edit` applies on save). It is itself a template, so escape literal merge-tool delimiters as in §2.
- **`.chezmoiignore`** (templated, gitignore syntax) keeps repo-only files out of `apply` — the `dot` CLI source, `skills/`, `AGENTS.md`, CI configs — and skips host- or secret-conditional files (e.g. a `.desktop` file off Linux, secret files when the age key is absent). Patterns match **target** paths; later patterns win, and a leading `!` re-includes.

## Gotchas

- **Edit source, not target**: changes to `~/.config/...` are erased on the next apply. Always change `~/.local/share/chezmoi/...`.
- **Attribute order is fixed**: `encrypted_` before `private_` before `dot_`; wrong order yields a literally-named file, not the effect you wanted.
- **`modify_` ≠ `.tmpl`**: `modify_` files are executed and template themselves via the marker; adding `.tmpl` is wrong.
- **`--force` in automation**: interactive apply prompts on a diverged target; scripts and hooks must pass `--force`.
- **Templates fail closed**: a template error aborts the whole apply — run `chezmoi execute-template < file` or `chezmoi apply --dry-run` to debug before committing.

## Documentation

- [chezmoi reference](https://www.chezmoi.io/reference/) · [source-state attributes](https://www.chezmoi.io/reference/source-state-attributes/) · [templating](https://www.chezmoi.io/user-guide/templating/) · [encryption](https://www.chezmoi.io/user-guide/encryption/age/)
- Companion skills:
  - [mise](../mise/SKILL.md) — pins `chezmoi` and wraps `apply`/`diff` as tasks.
  - [dprint](../dprint/SKILL.md) — formats the JSON/TOML/YAML configs chezmoi deploys.
  - [agent-project](../agent-project/SKILL.md) / [agent-skills](../agent-skills/SKILL.md) — the AGENTS.md + skills layer chezmoi symlinks into each agent CLI.
  - [security-scan](../security-scan/SKILL.md) — secret scanning around the age-encrypted files.
