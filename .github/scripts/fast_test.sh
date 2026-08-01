#!/usr/bin/env bash
set -euo pipefail

repo_root=$(git rev-parse --show-toplevel)
selector="${repo_root}/.github/scripts/fast.sh"
fixture=$(mktemp "${TMPDIR:-/tmp}/dot-fast-fixture.XXXXXX")
expected_file=$(mktemp "${TMPDIR:-/tmp}/dot-fast-expected.XXXXXX")
actual_file=$(mktemp "${TMPDIR:-/tmp}/dot-fast-actual.XXXXXX")
cleanup() {
  rm -f "${fixture}" "${expected_file}" "${actual_file}"
}
trap cleanup EXIT

assert_selection() {
  name=$1
  expected=$2
  shift 2
  printf '%s\n' "$@" >"${fixture}"
  printf '%s\n' "${expected}" >"${expected_file}"
  "${selector}" --list --files-from "${fixture}" >"${actual_file}"
  if ! diff -u "${expected_file}" "${actual_file}"; then
    printf 'FAIL: %s\n' "${name}" >&2
    exit 1
  fi
}

assert_selection go fast:go dot/verify.go
assert_selection shell fast:shell .github/scripts/trust-mise.sh
assert_selection template check:chezmoi dot_config/mise/config.toml.tmpl
assert_selection workflow $'check:actions\ncheck:dprint' .github/workflows/ci.yml
assert_selection skill $'check:dprint\ncheck:skills' skills/release/SKILL.md
assert_selection unknown check unknown.extension
assert_selection uncertain check dot/verify.go unknown.extension
assert_selection empty $'check:dprint\nfast:go' ''

printf 'PASS: deterministic fast-gate selection\n'
