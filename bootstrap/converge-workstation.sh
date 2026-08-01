#!/usr/bin/env bash
set -euo pipefail

run_phase() {
  local phase="$1"
  shift

  echo "Convergence phase: ${phase}"
  if ! "$@"; then
    echo "Workstation convergence failed at phase: ${phase}" >&2
    return 1
  fi
}

run_phase repository-tools mise install -y
run_phase dotfiles mise run apply
run_phase global-trust mise trust -y "${HOME}/.config/mise/config.toml"
run_phase global-tools mise -C "${HOME}" install -y
run_phase completions mise run completions
# `dot verify` consumes the shared capability registry, so convergence checks
# executable behavior rather than accepting a path-visible but broken shim.
run_phase capabilities mise run verify

echo "Workstation convergence complete."
