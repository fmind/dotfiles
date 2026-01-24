---
name: Ansible Role Creation
description: Guide for creating Ansible roles in the dotfiles repository
---

# Ansible Role Creation Skill

This skill documents how to create Ansible roles for the dotfiles repository, following established patterns and conventions.

## Repository Overview

The dotfiles repository uses Ansible to automate the installation and configuration of developer tools and environments. It contains:
- **`site.yml`** - Main playbook that orchestrates all roles
- **`roles/`** - Individual role directories for each tool/package
- **`inventory.ini`** - Local host inventory
- **`ansible.cfg`** - Ansible configuration
- **`justfile`** - Task runner for common operations

## Role Directory Structure

Every role follows this standard structure:

```
roles/<role-name>/
  ├── tasks/
  │   └── main.yml       # Main task definitions (required)
  └── files/             # Configuration files to copy/link (optional)
      └── <config-file>
```

**Key Points:**
- Role name should match the tool/package name (lowercase, hyphenated if needed)
- `tasks/main.yml` is the entry point and is required
- `files/` directory is optional, used for configuration files

## Writing Tasks (`tasks/main.yml`)

### Basic Structure

All task files start with `---` and contain a list of tasks:

```yaml
---
- name: <descriptive-name>
  <module>:
    <parameters>
  <optional-directives>
```

### Common Patterns

#### Pattern 1: Simple Package Installation

For basic CLI tools (example from `jules` role):

```yaml
---
- name: package
  community.general.npm:
    name: "@google/jules"
    state: latest
    global: true
  tags: user
```

#### Pattern 2: System Package with Configuration

For system packages requiring config files (example from `git` role):

```yaml
---
- name: package
  ansible.builtin.package:
    name: git
  become: true
  tags: admin

- name: storage
  ansible.builtin.package:
    name: git-lfs
  become: true
  tags: admin

- name: config
  ansible.builtin.copy:
    src: "{{ role_path }}/files/gitconfig"
    dest: ~/.gitconfig
    force: false
  tags: user
```

#### Pattern 3: Cross-Platform Package Installation

For packages that differ between Linux and macOS (example from `python` role):

```yaml
---
- name: package
  ansible.builtin.package:
    name: "{{ 'python3-dev' if ansible_facts['distribution'] != 'MacOSX' else 'python' }}"
  become: true
  tags: admin

- name: manager
  ansible.builtin.package:
    name: python3-pip
  when: ansible_facts['distribution'] != 'MacOSX'
  become: true
  tags: admin

- name: hosting
  ansible.builtin.file:
    src: "{{ role_path }}/files/pypirc"
    dest: ~/.pypirc
    state: link
    force: true
  tags: user
```

#### Pattern 4: Complex Setup with Post-Installation

For tools requiring setup steps (example from `gemini` role):

```yaml
---
- name: package
  community.general.npm:
    name: "@google/gemini-cli"
    state: latest
    global: true
  tags: user

- name: directory
  ansible.builtin.file:
    path: ~/.gemini
    state: directory
    mode: "0755"
  tags: user

- name: config
  ansible.builtin.file:
    src: "{{ role_path }}/files/settings.json"
    dest: ~/.gemini/settings.json
    state: link
    force: true
  tags: user

- name: extensions
  ansible.builtin.shell: |
    gemini extensions install --consent --auto-update {{ item }}
  loop:
    - https://github.com/example/extension
  register: install_result
  changed_when: "'Successfully installed' in install_result.stdout"
  failed_when:
    - install_result.rc != 0
    - "'is already installed' not in install_result.stderr"
  tags: user
```

## Tag Strategy

Use tags to categorize tasks by privilege requirements:

### `admin` Tag
- For tasks requiring root/sudo privileges
- Typically system package installations
- Always used with `become: true`

```yaml
- name: package
  ansible.builtin.package:
    name: some-package
  become: true
  tags: admin
```

### `user` Tag
- For user-level installations and configurations
- No root privileges required
- Most configuration file operations

```yaml
- name: config
  ansible.builtin.copy:
    src: "{{ role_path }}/files/config"
    dest: ~/.config/tool/config
  tags: user
```

## Privilege Escalation (`become`)

Use `become: true` for operations requiring elevated privileges:

**When to use:**
- Installing system packages via `apt`, `yum`, `package`, etc.
- Modifying system directories
- Installing packages to system-wide locations

**When NOT to use:**
- User-level package installations (npm, pip with `--user`, pipx, etc.)
- Creating files in user's home directory
- Linking configuration files to `~/`

## File Operations

### Copying Files

Use `ansible.builtin.copy` to copy files:

```yaml
- name: config
  ansible.builtin.copy:
    src: "{{ role_path }}/files/gitconfig"
    dest: ~/.gitconfig
    force: false  # Don't overwrite if exists
  tags: user
```

### Symlinking Files

Use `ansible.builtin.file` with `state: link` for symlinks:

```yaml
- name: config
  ansible.builtin.file:
    src: "{{ role_path }}/files/pypirc"
    dest: ~/.pypirc
    state: link
    force: true  # Update symlink if exists
  tags: user
```

## Integration with `site.yml`

After creating the role, add it to `site.yml`:

1. **Find the alphabetically correct position** in the roles list
2. **Add the role entry** with tag:

```yaml
roles:
  # ... other roles ...
  - { role: your-role-name, tags: ["your-role-name"] }
  # ... other roles ...
```

3. **Comment out initially** (optional) for testing:

```yaml
# - { role: your-role-name, tags: ["your-role-name"] }
```

4. **Uncomment to enable** when ready to use

### Pre-tasks Example

If your role needs to be loaded in `pre_tasks` (like `nodejs` for npm dependencies):

```yaml
pre_tasks:
  - name: ensure your-tool
    include_role:
      name: your-tool
    tags: your-tool
```

## Common Modules

### Package Management
- `ansible.builtin.package` - Generic package manager (auto-detects apt/yum/brew)
- `community.general.npm` - Node.js packages
- `community.general.pipx` - Python isolated environments

### File Operations
- `ansible.builtin.file` - Create directories, symlinks
- `ansible.builtin.copy` - Copy files
- `ansible.builtin.template` - Template files (with variables)

### Execution
- `ansible.builtin.shell` - Run shell commands
- `ansible.builtin.command` - Run commands (safer, no shell expansion)

## Testing Your Role

### Syntax Check

```bash
just check
# or
ansible-playbook -i inventory.ini site.yml --syntax-check
```

### Run Specific Role

```bash
ansible-playbook -i inventory.ini site.yml --tags your-role-name
```

### Run with Specific Privilege Mode

```bash
# With sudo prompt (Linux)
just sudo=true apply

# Without sudo (user-level only)
just apply
```

## Complete Example: New CLI Tool

Creating a role for a new CLI tool called `newtool`:

**1. Create directory structure:**
```bash
mkdir -p roles/newtool/tasks
mkdir -p roles/newtool/files
```

**2. Create `roles/newtool/tasks/main.yml`:**
```yaml
---
- name: package
  community.general.npm:
    name: newtool
    state: latest
    global: true
  tags: user

- name: config
  ansible.builtin.file:
    src: "{{ role_path }}/files/newtool.conf"
    dest: ~/.newtool/config
    state: link
    force: true
  tags: user
```

**3. Add config file `roles/newtool/files/newtool.conf`:**
```
# Newtool configuration
setting = value
```

**4. Add to `site.yml`:**
```yaml
roles:
  # ... other roles in alphabetical order ...
  - { role: newtool, tags: ["newtool"] }
  # ... other roles ...
```

**5. Test:**
```bash
just check
ansible-playbook -i inventory.ini site.yml --tags newtool
```

## Best Practices

1. **Name tasks descriptively** - Use meaningful names like "package", "config", "directory"
2. **Use appropriate tags** - `admin` for system changes, `user` for user-level
3. **Handle idempotency** - Tasks should be safe to run multiple times
4. **Use `when` for conditionals** - Platform-specific logic
5. **Register and check results** - For shell commands, register output and handle failures
6. **Keep it simple** - One role = one tool/package
7. **Follow existing patterns** - Look at similar roles for guidance

## Quick Reference

| Task                   | Module                    | Common Options                     | Requires `become` |
| ---------------------- | ------------------------- | ---------------------------------- | ----------------- |
| Install system package | `ansible.builtin.package` | `name`, `state`                    | Yes               |
| Install npm package    | `community.general.npm`   | `name`, `global`, `state`          | No                |
| Install pipx package   | `community.general.pipx`  | `name`                             | No                |
| Copy file              | `ansible.builtin.copy`    | `src`, `dest`, `force`             | Usually no        |
| Create symlink         | `ansible.builtin.file`    | `src`, `dest`, `state: link`       | Usually no        |
| Create directory       | `ansible.builtin.file`    | `path`, `state: directory`, `mode` | Usually no        |
| Run command            | `ansible.builtin.shell`   | Command string, `register`         | Depends           |
