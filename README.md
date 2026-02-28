# dotfiles

[![CI](https://github.com/fmind/dotfiles/actions/workflows/ci.yml/badge.svg)](https://github.com/fmind/dotfiles/actions/workflows/ci.yml)
[![CD](https://github.com/fmind/dotfiles/actions/workflows/cd.yml/badge.svg)](https://github.com/fmind/dotfiles/actions/workflows/cd.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE.txt)

Automated configuration for my personal development environment.

## Requirements

- [Docker](https://www.docker.com)
- [Just](https://just.systems)
- [Python](https://www.python.org)

## Usage

```bash
# Install project
just install

# Check configuration
just check

# Apply configuration
just apply

# Apply configuration with sudo
just apply sudo=true

# Configure the default shell
just shell
```

## Docker Image

A pre-built image is published to the GitHub Container Registry on every push to `main`:

```bash
docker pull ghcr.io/fmind/dotfiles:latest
docker run --rm -it ghcr.io/fmind/dotfiles:latest
```
