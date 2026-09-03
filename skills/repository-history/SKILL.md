---
name: repository-history
description: Reconstruct why tracked code exists from read-only Git history. Use to trace files, symbols, or lines to introducing commits or pull requests, rationale, reverts, renames, authorship, co-change, and issues.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/repository-history
  created: "2026-08-08"
  updated: "2026-09-03"
---

# Repository History

Recover why tracked code exists from read-only Git history, cited and labeled by confidence, before a risky edit; [systematic-debugging](../systematic-debugging/SKILL.md) owns reproducing a current failure and [diff-review](../diff-review/SKILL.md) owns defects in one change.

## Workflow

1. **Frame the question**: Name the repository, tracked path, line range or symbol, proposed change, and the specific uncertainty; keep the scope small enough that every cited commit can be inspected.
1. **Establish coverage**: Use `git` to record branch, `HEAD`, dirty state, available refs, and whether the clone is shallow; say which evidence Git cannot supply when the path is untracked, generated, vendored, or absent at `HEAD`. Preserve staged, unstaged, and untracked work.

   ```bash
   git status --short
   git rev-parse HEAD
   git rev-parse --is-shallow-repository
   ```

1. **Seed line provenance**: Treat movement and copy detection as clues, not guarantees; review a repository-owned `.git-blame-ignore-revs` before honoring it and disclose ignored revisions.

   ```bash
   git blame --line-porcelain -w -M -C -C -C -L <start>,<end> -- <path>
   ```

1. **Trace the timeline**: Use the smallest relevant view and name the exact revision range instead of treating every ref as one lineage.

   ```bash
   git log --follow --format=fuller -- <path>   # one file across renames
   git log -L <start>,<end>:<path>              # commits that shaped a line range
   git log -S '<literal>' -p -- <path>          # occurrence count of a string changed
   git log -G '<regex>' -p -- <path>            # a diff added or removed matching lines
   ```

1. **Inspect candidate commits**: Read subject, body, changed tests, schemas, migrations, configuration, and docs together; preserve reverts and behavior changes, and group mechanical edits only after verifying they are mechanical.

   ```bash
   git show --format=fuller --find-renames --find-copies --stat <sha>
   ```

1. **Resolve ancestry anomalies**: Check renames, splits, copies, bulk formatting, generated files, squashes, rebases, cherry-picks, backports, merge parents, and revert pairs; when line history stops at a rewrite, compare file history, pickaxe searches, and neighboring tests, and report the break rather than forcing one origin.
1. **Find coupling clues**: Count files that repeatedly changed with the target across behavior commits. Repeated co-change can suggest a test, schema, migration, or deployment constraint; one shared commit or a formatting sweep proves nothing.
1. **Add remote rationale only when needed**: Map an exact commit to its pull requests with `gh`, then read the PR body, linked issue, reviews, inline comments, and changed files; paginate, disclose truncation, and treat a title match or semantic search alone as low confidence.
1. **Extract decision atoms**: Separate facts, author-stated rationale, inference, contradiction, and unknowns; cite the commit, patch, test, issue, or review for every proposed constraint and check whether later changes superseded it.
1. **Return the history note**: Do not implement the change unless requested separately. Report:
   - **Bottom line**: the most likely explanation, its confidence, and whether it changes the proposed plan.
   - **Scope and identity**: `HEAD`, working-copy boundary, path, lines or symbol, revision range, shallow state, and remote evidence checked.
   - **Origin and timeline**: introducing evidence, later changes, fixes, reverts, and moves with short hashes and dates.
   - **Decision atoms and companion evidence**: each labeled fact, inference, contradiction, or unknown; co-changes stay correlation until code confirms the seam.
   - **Change risk**: what could break, which evidence may be stale, and the smallest current test or human answer that resolves the rest.
   - **Evidence sources**: local objects, remote discussion, runtime behavior, and human intent kept separate.

## Gotchas

- **Investigation is read-only**: Do not fetch, pull, checkout, bisect with mutations, reset, rebase, amend, revert, delete refs, rewrite history, or contact an author unless the user separately authorizes it.
- **Never print raw remote URLs**: identify the remote with `gh repo view --json nameWithOwner` and never echo a URL that carries a token.
- **Current blame is not original authorship**: a committer is not necessarily the designer, and a message can state intent without proving the constraint still holds; redact email addresses from returned evidence.
- **Confidence**: `High` needs explicit rationale that agrees with the patch, tests, and later history; `Medium` has agreeing lineage and co-change without stated rationale; `Low` rests on blame, one title match, a semantic search, sparse history, or an ancestry break; otherwise say `UNKNOWN`.
- **Stale rationale**: A revert describes a past decision; downgrade it when later architectural changes contradict the trade-off.
- **Shallow or incomplete clones**: Never infer that a missing commit or discussion does not exist; state the gap and lower confidence.

## Documentation

- [git blame](https://git-scm.com/docs/git-blame) · [git log](https://git-scm.com/docs/git-log) · [Pull requests associated with a commit](https://docs.github.com/en/rest/commits/commits#list-pull-requests-associated-with-a-commit)
- Adapted from [awesome-llm-apps commit-archaeologist](https://github.com/Shubhamsaboo/awesome-llm-apps/blob/779e9f9bcf87fa8cd95870a438b70b84e47d3173/agent_skills/commit-archaeologist/SKILL.md).
- Companion skills: [systematic-debugging](../systematic-debugging/SKILL.md) (reproduce a current failure), [diff-review](../diff-review/SKILL.md) (one change), [repository-review](../repository-review/SKILL.md) (cross-cutting audit), [technical-research](../technical-research/SKILL.md) (current external facts), [github-issues](../github-issues/SKILL.md) (remote issue state).
