---
name: release
description: Cut or verify a versioned release — bump semver, generate the changelog with git-cliff, tag and publish on GitHub, or reconcile an already-published tag and assets.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/release
  created: 2026-07-04
  updated: 2026-08-30
---

# Release Process

Turn the Conventional Commits since the last tag into a versioned release: a bumped `CHANGELOG.md`, package manifest updates, an annotated git tag, and a GitHub release. Uses **git-cliff** (config: `dot_config/git-cliff/cliff.toml`, deployed to `~/.config/git-cliff/cliff.toml`).

## Authority Boundaries

- Preparing or reviewing a release does not authorize a commit, tag, push, GitHub release, asset upload, package or image publication, deployment, or announcement. Confirm each requested external mutation before performing it.
- Verification is read-only apart from downloading public assets into a disposable directory. It never creates or edits a release, moves a tag, replaces an asset, rebuilds an artifact, or resigns it.
- If a repository workflow owns release creation, let the pushed tag trigger that workflow and verify its result; do not race it with `gh release create`.

## Publication Preconditions

- Clean working tree on the default branch (`main`), synced with `origin`.
- History follows [conventional-commit](../conventional-commit/SKILL.md) — git-cliff groups commits by `type` and skips `chore(release)`/`chore(deps)`.
- The repository's full gate passes on the release tree, normally `mise run all`. Before running it, inspect the full gate's task definition and working-tree state. If it invokes whole-tree write-formatters or the tree contains unrelated changes, use an isolated temporary worktree; never reformat unrelated work.
- The proposed tag is absent locally and remotely. Stop if either copy exists; never move a published tag.

## Workflow

1. **Compute the next version** from the commit types since the last tag (`feat` → minor, `fix`/others → patch, `!`/`BREAKING CHANGE` → major):

   ```bash
   # Resolves the local/global cliff.toml configuration
   git-cliff --bumped-version

   # Or explicitly reference the global config if a local one is not present:
   git-cliff --config ~/.config/git-cliff/cliff.toml --bumped-version
   ```

1. **Update package manifests** (if the project doesn't use dynamic/VCS-based versioning) to match the computed version (e.g., `vX.Y.Z` or `X.Y.Z`):
   - **Python**: Bump `version` in `pyproject.toml` (unless using dynamic versioning via `hatch-vcs` or similar).
   - **Node.js**: Run `npm version --no-git-tag-version X.Y.Z` or update `package.json`.
   - **Go / OpenTofu**: Versioned via git tags (no file changes needed).
1. **Generate the changelog** for that version (pass `--config` to be explicit):
   ```bash
   git-cliff --config ~/.config/git-cliff/cliff.toml --bump -o CHANGELOG.md
   ```
1. **Commit** the changelog and manifest changes with a release commit (excluded from the changelog by design):
   ```bash
   git add CHANGELOG.md
   # plus the manifest you bumped in step 2, if any (Python: pyproject.toml · Node: package.json · Go/OpenTofu: none)
   git commit -m "chore(release): vX.Y.Z"
   ```
1. **Record, tag, and push the exact release commit** after the authorized release commit:
   ```bash
   tag=vX.Y.Z
   release_sha=$(git rev-parse HEAD)
   git tag -a "$tag" -m "$tag" "$release_sha"
   git push --atomic origin main "refs/tags/$tag"
   ```
1. **Publish** the GitHub release using only the latest section as notes (write to a temp file to stay shell-agnostic):
   ```bash
   release_tmp=$(mktemp -d)
   git-cliff --config ~/.config/git-cliff/cliff.toml --latest --strip all > "$release_tmp/release-notes.md"
   gh release create "$tag" --verify-tag --title "$tag" --notes-file "$release_tmp/release-notes.md"
   ```
1. Print the release URL and the resolved version.

## Verify a Published Release

Use this phase after publication, including when another workflow or person created the release.

1. **Resolve the expected identity:** Record the repository, tag, and exact release commit from the approved candidate rather than from the latest branch state.

   ```bash
   repo=$(gh repo view --json nameWithOwner --jq .nameWithOwner)
   tag=vX.Y.Z
   release_sha=$(git rev-parse HEAD) # Run from the approved candidate checkout.
   ```

1. **Reconcile the remote tag:** An annotated tag emits its tag-object line and a peeled `^{}` line. Compare the peeled commit to `release_sha`; the first line is not the release commit.

   ```bash
   git ls-remote --exit-code --tags origin "refs/tags/$tag" "refs/tags/$tag^{}"
   ```

1. **Inspect GitHub's release state:** Require the expected tag, final draft/prerelease state, publication time, asset inventory, immutability policy, and URL. `targetCommitish` may only name `main`; it does not replace the peeled-tag check.

   ```bash
   gh release view "$tag" -R "$repo" --json tagName,name,isDraft,isPrerelease,isImmutable,publishedAt,targetCommitish,assets,url
   ```

1. **Prove exact-head automation:** List runs at `release_sha` and require every expected CI/CD workflow to be present, completed, and successful. A green latest-branch run, or merely some successful runs, is insufficient.

   ```bash
   gh run list -R "$repo" --commit "$release_sha" --limit 100 --json workflowName,attempt,headSha,status,conclusion,event,url
   ```

1. **Verify published assets:** Download the expected inventory into a new temporary directory, verify the repository's checksum manifest, and then verify integrity and provenance where the release publishes attestations.

   ```bash
   release_dir=$(mktemp -d)
   gh release download "$tag" -R "$repo" --dir "$release_dir"
   gh release verify "$tag" -R "$repo"
   gh release verify-asset "$tag" "$release_dir/checksums.txt" -R "$repo"
   (cd "$release_dir" && shasum -a 256 -c checksums.txt)
   asset=project_vX.Y.Z_linux_amd64.tar.gz
   signer_workflow="$repo/.github/workflows/release.yml"
   gh release verify-asset "$tag" "$release_dir/$asset" -R "$repo"
   gh attestation verify "$release_dir/$asset" --repo "$repo" --signer-workflow "$signer_workflow" --source-ref "refs/tags/$tag" --source-digest "$release_sha"
   ```

   Run `gh release verify` and `verify-asset` only for immutable releases with release attestations; generated source archives cannot be verified as uploaded assets. Verify the checksum manifest itself before trusting its entries. Run `gh attestation verify` when separate build provenance exists, with the expected signer workflow. Missing expected assets, checksums, attestations, or signer identity is a failed proof, not permission to regenerate or replace them.

1. **Verify the delivered boundary:** After integrity and provenance pass, exercise the packaged binary or installation contract and confirm its version. Use [containerize](../containerize/SKILL.md) for digest-bound OCI, Cosign, and SBOM verification, and [production-readiness](../production-readiness/SKILL.md) for a deployment or operational go/no-go.
1. **Report a release receipt:** Include the expected commit, remote tag object and peeled commit, workflow names and URLs, release URL and state, expected versus downloaded assets, checksum and attestation results, packaged version, and any package, image, deployment, or anonymous-public boundary that remains unproved.

## Gotchas

- **Semver source of truth**: let `git-cliff --bumped-version` decide from commits; only override for a deliberate bump (e.g. first stable `v1.0.0`).
- **Tag prefix**: keep the `v` prefix consistent — git-cliff, `gh`, and Go module tags all expect `vX.Y.Z`.
- **Pre-1.0**: git-cliff applies the same rules below `v1.0.0` as above it — a `feat` bumps the **minor** (`v0.1.0` → `v0.2.0`) and a breaking change jumps straight to **`v1.0.0`**. Verify before tagging: `git-cliff --bumped-version`. To keep a 0.x line breaking-change-tolerant instead, set `features_always_bump_minor = false` and `breaking_always_bump_major = false` under `[bump]` in `cliff.toml` — neither key is set in the global config, which holds only `initial_tag`.
- **First Release (No Tags)**: if the repository has no tags, `git-cliff` defaults to the `initial_tag` (configured as `v0.1.0` in `cliff.toml` to match the `tag_pattern`).
- **Idempotency**: if the tag already exists, stop — never move a published tag.
- **Remote truth**: a local tag, local green gate, draft release, or latest-branch workflow is not proof of the published release. Reconcile the remote peeled tag, exact-head workflows, release state, and public artifacts independently.
- **Config Resolution**: run without `--config` and git-cliff searches `cliff.toml`, `.cliff.toml`, then `.config/cliff.toml` at the repository root, and automatically falls back to `~/.config/git-cliff/cliff.toml` — so `--config` is only needed to force a non-default file.

## Documentation

- [git-cliff Documentation](https://git-cliff.org)
- [GitHub release integrity](https://docs.github.com/en/code-security/how-tos/secure-your-supply-chain/secure-your-dependencies/verify-release-integrity)
- [GitHub CLI release manual](https://cli.github.com/manual/gh_release)
- Companion skills:
  - [conventional-commit](../conventional-commit/SKILL.md) — the commit grammar git-cliff parses.
  - [github-pull-request](../github-pull-request/SKILL.md) — merge work before releasing.
