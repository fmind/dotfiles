#!/usr/bin/env bash

set -euo pipefail

if (($# > 2)); then
  printf 'usage: %s [lazy-lock.json] [lazy-root]\n' "$0" >&2
  exit 2
fi

: "${HOME:?HOME must be set}"
config_home="${XDG_CONFIG_HOME:-${HOME}/.config}"
data_home="${XDG_DATA_HOME:-${HOME}/.local/share}"
lock_path="${1:-${config_home}/nvim/lazy-lock.json}"
lazy_root="${2:-${data_home}/nvim/lazy}"

if [[ ! -s ${lock_path} ]]; then
  printf '✗ Lazy lockfile is missing or empty: %s\n' "${lock_path}" >&2
  exit 1
fi
if [[ ! -d ${lazy_root} ]]; then
  printf '✗ Lazy plugin root is missing: %s\n' "${lazy_root}" >&2
  exit 1
fi

# Reject malformed or path-like names before they are joined to the plugin root.
if ! jq -e '
  type == "object" and
  length > 0 and
  all(to_entries[];
    (.key | test("^[A-Za-z0-9._-]+$")) and
    (.key != ".") and
    (.key != "..") and
    (.value | type == "object") and
    ((.value.commit? // null) | type == "string") and
    (.value.commit | test("^[0-9a-f]{40}([0-9a-f]{24})?$"))
  )
' "${lock_path}" >/dev/null; then
  printf '✗ Lazy lockfile has an invalid plugin name or commit: %s\n' "${lock_path}" >&2
  exit 1
fi

verify_checked_out_tree() {
  local repository=$1
  local plugin=$2
  local entry
  local entries=0
  local present=0
  local missing=0

  # Compare physical paths with HEAD rather than the index: staged deletions must
  # still fail, while modified files and plugin-generated dirt remain harmless.
  while IFS= read -r -d '' entry; do
    entries=$((entries + 1))
    if [[ -e "${repository}/${entry}" || -L "${repository}/${entry}" ]]; then
      present=$((present + 1))
      continue
    fi
    printf '✗ %s: missing tracked path %s\n' "${plugin}" "${entry}" >&2
    missing=1
  done < <(git -C "${repository}" ls-tree -r --name-only -z HEAD 2>/dev/null || true)

  if ((entries == 0)); then
    printf '✗ %s: HEAD has no tracked worktree entries\n' "${plugin}" >&2
    return 1
  fi
  if ((present == 0)); then
    printf '✗ %s: empty worktree at %s\n' "${plugin}" "${repository}" >&2
    return 1
  fi
  return "${missing}"
}

status=0
count=0
plugin_rows="$(jq -r 'to_entries | sort_by(.key)[] | [.key, .value.commit] | @tsv' "${lock_path}")"
while IFS=$'\t' read -r plugin expected_commit; do
  count=$((count + 1))
  repository="${lazy_root}/${plugin}"

  if [[ ! -d ${repository} ]] || [[ "$(git -C "${repository}" rev-parse --is-inside-work-tree 2>/dev/null || true)" != "true" ]]; then
    printf '✗ %s: missing Git repository at %s\n' "${plugin}" "${repository}" >&2
    status=1
    continue
  fi

  actual_commit="$(git -C "${repository}" rev-parse --verify HEAD 2>/dev/null || true)"
  if [[ ${actual_commit} != "${expected_commit}" ]]; then
    printf '✗ %s: locked commit %s, checkout is at %s\n' "${plugin}" "${expected_commit}" "${actual_commit:-an unreadable HEAD}" >&2
    status=1
  fi

  if ! verify_checked_out_tree "${repository}" "${plugin}"; then
    status=1
  fi
done <<<"${plugin_rows}"

if ((status != 0)); then
  exit "${status}"
fi

printf '✓ Verified %d Lazy plugin checkout(s) against %s.\n' "${count}" "${lock_path}"
