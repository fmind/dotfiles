# dotfiles

Configuration of my favorite languages, editors, shells, and tools.

# Requirements

- python
- pipx

# Installation

```bash
# with pyinvoke
inv install
# on Linux system (no sudo required)
ansible-playbook site.yml
# on Linux system (sudo required)
ansible-playbook -K site.yml
# on MacOS system
ansible-playbook site.yml --become-user=$USER
```

**For Mac OSX**:
- To enable the unarchive module: `brew install gnu-tar`

# Configuration

```bash
# on Linux system
chsh -s /usr/bin/fish
# on MacOS system
chsh -s /bin/zsh
```