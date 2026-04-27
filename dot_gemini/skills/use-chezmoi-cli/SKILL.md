---
name: use-chezmoi-cli
description: Guide for managing dotfiles with chezmoi — source vs target state, templates, encryption, externals, and the apply/diff/edit workflow.
---

# Use chezmoi CLI

`chezmoi` manages dotfiles across machines from a single source directory (typically `~/.local/share/chezmoi`). The agent must distinguish between **source state** (what chezmoi tracks, with mangled filenames like `dot_zshrc`) and **target state** (what lands in `$HOME`, e.g. `~/.zshrc`).

## Source ↔ Target Naming

| Source filename            | Target filename       | Meaning                               |
|----------------------------|-----------------------|---------------------------------------|
| `dot_zshrc`                | `~/.zshrc`            | Files starting with `.`               |
| `dot_config/git/config`    | `~/.config/git/config`| Nested directories                    |
| `private_dot_ssh/config`   | `~/.ssh/config`       | `private_` → `chmod 600`              |
| `executable_bin/foo`       | `~/bin/foo`           | `executable_` → `+x`                  |
| `symlink_dot_vimrc`        | `~/.vimrc` (symlink)  | Symlink whose target is the file body |
| `dot_zshrc.tmpl`           | `~/.zshrc`            | `.tmpl` → rendered via Go templates   |
| `encrypted_private_dot_*`  | (decrypted on apply)  | age/gpg-encrypted content             |
| `run_once_install.sh.tmpl` | (executed once)       | Run-once script                       |
| `run_onchange_*.sh`        | (re-runs when changed)| Run-on-change script                  |

## Core Commands

```bash
# Inspect.
chezmoi status                       # changed files (like git status)
chezmoi diff                         # show pending changes
chezmoi diff ~/.zshrc                # diff a single target
chezmoi managed                      # list everything chezmoi tracks
chezmoi unmanaged                    # files in $HOME that aren't tracked

# Edit (always edit the source, not the target).
chezmoi edit ~/.zshrc                # opens dot_zshrc in $EDITOR
chezmoi edit --apply ~/.zshrc        # edit + apply on save
chezmoi cd                           # drop into the source directory

# Apply changes to $HOME.
chezmoi apply                        # apply everything
chezmoi apply --dry-run --verbose    # preview without writing
chezmoi apply ~/.zshrc               # apply just one target

# Pull updates from the remote dotfiles repo.
chezmoi update                       # = git pull + apply
chezmoi git pull                     # pull only, no apply

# Add new files to chezmoi.
chezmoi add ~/.config/foo/bar.toml
chezmoi add --template ~/.gitconfig  # add as .tmpl
chezmoi add --encrypt ~/.ssh/id_ed25519
```

## Templates

Source files ending in `.tmpl` are rendered with Go templates and chezmoi data. Useful variables: `.chezmoi.os`, `.chezmoi.arch`, `.chezmoi.hostname`, `.chezmoi.username`, plus anything declared in `~/.config/chezmoi/chezmoi.toml`.

```gotemplate
{{- if eq .chezmoi.os "darwin" }}
export HOMEBREW_PREFIX="/opt/homebrew"
{{- else if eq .chezmoi.os "linux" }}
export HOMEBREW_PREFIX="/home/linuxbrew/.linuxbrew"
{{- end }}
```

Render or debug:

```bash
chezmoi execute-template < path.tmpl   # render a template stdin → stdout
chezmoi data                           # dump the data context as JSON
chezmoi cat ~/.zshrc                   # show the rendered target
```

## Encryption

```bash
# Configure age once (preferred over GPG for new setups).
chezmoi generate age-key
# then in chezmoi.toml:
#   encryption = "age"
#   [age]
#     identity = "~/.config/chezmoi/key.txt"
#     recipient = "age1..."

chezmoi add --encrypt ~/.aws/credentials
chezmoi decrypt encrypted_private_dot_aws/credentials
```

## Externals (`.chezmoiexternal.toml`)

Pull in third-party files / archives without vendoring:

```toml
[".oh-my-zsh"]
    type = "archive"
    url = "https://github.com/ohmyzsh/ohmyzsh/archive/master.tar.gz"
    refreshPeriod = "168h"
    stripComponents = 1
```

## Ignoring Files (`.chezmoiignore`)

Same syntax as `.gitignore`, but applies to the *target* state — i.e. things in the source dir you don't want chezmoi to manage on this host. It is itself a template, so it can be host-conditional.

## Run Scripts

- `run_once_install-deps.sh` — runs once per machine (tracked by SHA).
- `run_onchange_install-packages.sh.tmpl` — re-runs whenever the rendered content changes (great for package lists).
- Scripts run in alphabetical order; prefix with `before_` / `after_` to control ordering relative to file application.

## Common Workflows

**Add a new dotfile.**
1. `chezmoi add ~/.config/foo/bar.toml`
2. Optionally rename to `.tmpl` and template it.
3. `chezmoi diff` to confirm.
4. `chezmoi cd` → `git add . && git commit -m "..."`.

**Modify an existing dotfile.**
1. `chezmoi edit ~/.zshrc` (NOT direct edit of `~/.zshrc`).
2. `chezmoi diff`.
3. `chezmoi apply`.

**Sync a new machine.**
```bash
chezmoi init --apply <github-user>
```

## Important Notes

1. **Never edit target files directly** — they get clobbered on `apply`. Always edit the source via `chezmoi edit` or by editing under `~/.local/share/chezmoi/`.
2. `chezmoi apply` is the only step that writes to `$HOME`. Always `diff` first if uncertain.
3. Templates are evaluated lazily — bugs surface at apply time. Use `chezmoi execute-template` to debug.
4. When committing changes, do it from inside the source dir (`chezmoi cd`); chezmoi does not auto-commit.
5. The chezmoi source dir for this user is `~/.local/share/chezmoi`.

## Documentation

- [chezmoi user guide](https://www.chezmoi.io/user-guide/command-overview/)
- [Templates reference](https://www.chezmoi.io/reference/templates/)
- [Source state attributes](https://www.chezmoi.io/reference/source-state-attributes/)
