# dotfiles

Automated configuration for my personal development environment.

## Requirements

- [Just](https://just.systems)
- [Python](https://www.python.org)

## Usage

```bash
# Install project
just install

# Apply configuration
just apply

# Apply configuration with sudo
just apply sudo=true

# Configure the default shell
just shell
```

## Docker

Run the configuration in an isolated container:

```bash
just docker
```
