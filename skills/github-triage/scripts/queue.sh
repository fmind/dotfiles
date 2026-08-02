#!/usr/bin/env bash

set -euo pipefail

readonly LEASE_LABEL="status/in-progress"
readonly HUMAN_LABEL="needs-human"
readonly EPIC_LABEL="kind/epic"
readonly DEFAULT_TTL_SECONDS=7200

usage() {
  cat <<'EOF'
Usage:
  queue.sh setup OWNER/REPO
  queue.sh runnable OWNER/REPO [--input FILE]
  queue.sh stale OWNER/REPO [--input FILE] [--now UNIX_SECONDS] [--ttl SECONDS]
EOF
}

die() {
  printf 'queue: %s\n' "$*" >&2
  exit 1
}

require_repo() {
  [[ ${1:-} =~ ^[^/[:space:]]+/[^/[:space:]]+$ ]] || die "repository must be OWNER/REPO"
}

setup_labels() {
  local repo=$1

  gh label create "${LEASE_LABEL}" --repo "${repo}" --color 0E8A16 --description \
    "Cooperative lease: active implementation is in progress" --force
  gh label create "${HUMAN_LABEL}" --repo "${repo}" --color B60205 --description \
    "Blocked on a human decision, credential, approval, or spend" --force
  gh label create "${EPIC_LABEL}" --repo "${repo}" --color 5319E7 --description \
    "Tracker issue grouping dependent work; not directly runnable" --force
  # The runnable filter requires exactly one priority/p* and one effort/* label,
  # so the fixed sets are provisioned here or a fresh repository's queue stays
  # permanently empty. Only area/* is repository-specific and created at triage.
  gh label create "priority/p0" --repo "${repo}" --color D93F0B --description \
    "Now: the active frontier and anything blocking it" --force
  gh label create "priority/p1" --repo "${repo}" --color FBCA04 --description \
    "Next: near-term once the p0 frontier clears" --force
  gh label create "priority/p2" --repo "${repo}" --color C2E0C6 --description \
    "Later: lower urgency within its track" --force
  gh label create "effort/s" --repo "${repo}" --color C5DEF5 --description \
    "Small: one focused, bounded change" --force
  gh label create "effort/m" --repo "${repo}" --color 84B6EB --description \
    "Medium: several coordinated changes in one area" --force
  gh label create "effort/l" --repo "${repo}" --color 3C6EB4 --description \
    "Large: cross-cutting change; consider splitting first" --force
}

issue_json() {
  local repo=$1
  local input=$2

  if [[ -n ${input} ]]; then
    jq -e 'if type == "array" then . else error("fixture must be a JSON array") end' "${input}"
    return
  fi

  gh issue list --repo "${repo}" --state open --limit 100 \
    --search "-label:${LEASE_LABEL} -label:${HUMAN_LABEL} -label:${EPIC_LABEL} -is:blocked" \
    --json number,title,state,url,labels,blockedBy,comments
}

runnable_issues() {
  jq --arg lease "${LEASE_LABEL}" --arg human "${HUMAN_LABEL}" --arg epic "${EPIC_LABEL}" '
    def names: [.labels[].name];
    def matching($prefix): [names[] | select(startswith($prefix))];
    def blockers: if (.blockedBy | type) == "array" then .blockedBy else (.blockedBy.nodes // []) end;
    def rank($prefix; $ordered):
      matching($prefix)[0] as $value
      | ($ordered | index($value) // length);
    map(select(
      (.state | ascii_upcase) == "OPEN"
      and (blockers | length) == 0
      and (names | index($lease) | not)
      and (names | index($human) | not)
      and (names | index($epic) | not)
      and (matching("priority/") | length) == 1
      and (matching("effort/") | length) == 1
      and (matching("area/") | length) >= 1
    ))
    | sort_by(rank("priority/"; ["priority/p0", "priority/p1", "priority/p2"]), rank("effort/"; ["effort/s", "effort/m", "effort/l"]), .number)
    | map({number, title, url, priority: matching("priority/")[0], effort: matching("effort/")[0], areas: matching("area/")})
  '
}

stale_leases() {
  local now=$1
  local ttl=$2

  jq --arg lease "${LEASE_LABEL}" --argjson now "${now}" --argjson ttl "${ttl}" '
    def names: [.labels[].name];
    def lease_comments: [(.comments // [])[] | select(.body | test("^\\*\\*(Claim|Lease update)\\*\\*"))];
    map(select((names | index($lease)) != null))
    | map(
        . as $issue
        | (lease_comments | sort_by(.createdAt) | last) as $heartbeat
        | if $heartbeat == null then
            {number, title, url, reason: "missing lease comment", lastLeaseAt: null}
          elif (($heartbeat.createdAt | fromdateiso8601) + $ttl) <= $now then
            {number, title, url, reason: "lease heartbeat expired", lastLeaseAt: $heartbeat.createdAt}
          else empty end
      )
    | sort_by(.number)
  '
}

main() {
  (($# >= 2)) || {
    usage >&2
    exit 2
  }

  local command=$1
  local repo=$2
  local input=""
  local now
  local ttl=${DEFAULT_TTL_SECONDS}
  shift 2
  require_repo "${repo}"

  while (($# > 0)); do
    case $1 in
    --input)
      (($# >= 2)) || die "--input requires a file"
      input=$2
      shift 2
      ;;
    --now)
      (($# >= 2)) || die "--now requires UNIX seconds"
      now=$2
      shift 2
      ;;
    --ttl)
      (($# >= 2)) || die "--ttl requires seconds"
      ttl=$2
      shift 2
      ;;
    *) die "unknown argument: $1" ;;
    esac
  done

  case ${command} in
  setup)
    [[ -z ${input} ]] || die "setup does not accept --input"
    setup_labels "${repo}"
    ;;
  runnable)
    issue_json "${repo}" "${input}" | runnable_issues
    ;;
  stale)
    now=${now:-$(date -u +%s)}
    [[ ${now} =~ ^[0-9]+$ && ${ttl} =~ ^[1-9][0-9]*$ ]] || die "--now and --ttl must be positive integers"
    issue_json "${repo}" "${input}" | stale_leases "${now}" "${ttl}"
    ;;
  *)
    usage >&2
    exit 2
    ;;
  esac
}

main "$@"
