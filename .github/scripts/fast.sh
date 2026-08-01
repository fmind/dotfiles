#!/usr/bin/env bash
set -euo pipefail

usage() {
  printf 'Usage: %s [--list] [--files-from <path>]\n' "${0##*/}"
}

mode=run
files_from=
while (($# > 0)); do
  case "$1" in
  --files-from)
    if (($# < 2)); then
      usage >&2
      exit 2
    fi
    files_from=$2
    shift 2
    ;;
  --list)
    mode=list
    shift
    ;;
  --help | -h)
    usage
    exit 0
    ;;
  *)
    printf 'fast: unknown argument %s\n' "$1" >&2
    usage >&2
    exit 2
    ;;
  esac
done

repo_root=$(git rev-parse --show-toplevel)
changes_file=$(mktemp "${TMPDIR:-/tmp}/dot-fast-changes.XXXXXX")
tasks_file=$(mktemp "${TMPDIR:-/tmp}/dot-fast-tasks.XXXXXX")
cleanup() {
  rm -f "${changes_file}" "${tasks_file}"
}
trap cleanup EXIT

uncertain=false
if [[ -n ${files_from} ]]; then
  if [[ ! -f ${files_from} ]]; then
    printf 'fast: files list does not exist: %s\n' "${files_from}" >&2
    exit 2
  fi
  cp "${files_from}" "${changes_file}"
else
  base_ref=${FAST_BASE_REF:-origin/main}
  if git -C "${repo_root}" rev-parse --verify --quiet "${base_ref}^{commit}" >/dev/null; then
    base_commit=$(git -C "${repo_root}" merge-base HEAD "${base_ref}")
    git -C "${repo_root}" diff --name-only --diff-filter=ACMR "${base_commit}"...HEAD >>"${changes_file}"
  else
    uncertain=true
  fi
  git -C "${repo_root}" diff --name-only --diff-filter=ACMR HEAD >>"${changes_file}"
  git -C "${repo_root}" ls-files --others --exclude-standard >>"${changes_file}"
fi
sort -u -o "${changes_file}" "${changes_file}"

add_task() {
  printf '%s\n' "$1" >>"${tasks_file}"
}

broad=${uncertain}
while IFS= read -r path; do
  [[ -z ${path} ]] && continue
  case "${path}" in
  mise.toml | mise.lock | lefthook.yml | trivy.yaml)
    broad=true
    ;;
  dot/*.go | dot/go.mod | dot/go.sum | dot/mise.toml)
    add_task fast:go
    ;;
  *.sh | *.sh.tmpl)
    add_task fast:shell
    ;;
  *.tmpl | .chezmoi*)
    add_task check:chezmoi
    ;;
  .github/workflows/*.yml | .github/workflows/*.yaml)
    add_task check:actions
    add_task check:dprint
    ;;
  skills/*)
    add_task check:dprint
    add_task check:skills
    ;;
  *.py | ruff.toml)
    add_task check:dprint
    add_task check:python
    ;;
  *.lua | .stylua.toml)
    add_task check:dprint
    add_task check:lua
    ;;
  *.md | *.json | *.toml | *.yaml | *.yml)
    add_task check:dprint
    ;;
  *)
    broad=true
    ;;
  esac
done <"${changes_file}"

if [[ ${broad} == true ]]; then
  printf 'check\n' >"${tasks_file}"
elif [[ ! -s ${tasks_file} ]]; then
  add_task check:dprint
  add_task fast:go
fi
sort -u -o "${tasks_file}" "${tasks_file}"

if [[ ${mode} == list ]]; then
  cat "${tasks_file}"
  exit 0
fi

selected=$(paste -sd, "${tasks_file}")
printf 'fast: selected %s\n' "${selected}"
SECONDS=0
count=0
while IFS= read -r task; do
  mise -C "${repo_root}" run "${task}"
  count=$((count + 1))
done <"${tasks_file}"
printf 'fast: completed %d task(s) in %ds; run mise run all before completion\n' "${count}" "${SECONDS}"
