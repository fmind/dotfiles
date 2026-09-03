---
name: release
description: "Cut or verify a versioned release: bump semver, generate the changelog with git-cliff, tag, and publish on GitHub. Use when releasing or reconciling a published tag."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/release
  created: "2026-07-04"
  updated: "2026-09-03"
---

# Release

Turn the Conventional Commits since the last tag into a versioned release (`CHANGELOG.md` from git-cliff, manifest bumps, an annotated tag, a GitHub release), then prove what was published; [conventional-commit](../conventional-commit/SKILL.md) owns the commit grammar git-cliff parses.

## Workflow

1. **Check the preconditions**:
   - Clean working tree on `main`, synced with `origin`.
   - The proposed tag is absent locally and remotely; stop if either copy exists and never move a published tag.
   - A repository workflow that owns release creation runs from the pushed tag; verify its result instead of publishing a second release from the CLI.
1. **Gate**: Run the full gate (`mise run all`); if the tree carries unrelated changes and the gate write-formats, run it in a temporary `git worktree` or fall back to `mise run check` and `mise run test` (see [mise](../mise/SKILL.md)).
1. **Compute the next version** from the commit types since the last tag: `feat` → minor, `fix` and others → patch, `!` or `BREAKING CHANGE` → major:

   ```bash
   git-cliff --bumped-version
   ```

1. **Bump manifests** that are not VCS-versioned: Python `version` in `pyproject.toml` (unless `hatch-vcs` or similar), Node `npm version --no-git-tag-version X.Y.Z`; Go and OpenTofu need no file change.
1. **Generate the changelog** for that version:

   ```bash
   git-cliff --bump -o CHANGELOG.md
   ```

1. **Commit the release**; the `chore(release)` subject is excluded from the changelog by design:

   ```bash
   git add CHANGELOG.md   # plus the manifest bumped above, if any
   git commit -m "chore(release): vX.Y.Z"
   ```

1. **Tag and push the exact release commit** atomically:

   ```bash
   tag=vX.Y.Z
   release_sha=$(git rev-parse HEAD)
   git tag -a "$tag" -m "$tag" "$release_sha"
   git push --atomic origin main "refs/tags/$tag"
   ```

1. **Publish** with the latest changelog section as notes, written to a temporary file to stay shell-agnostic:

   ```bash
   release_tmp=$(mktemp -d)
   git-cliff --latest --strip all > "$release_tmp/release-notes.md"
   gh release create "$tag" --verify-tag --title "$tag" --notes-file "$release_tmp/release-notes.md"
   ```

1. **Report** the release URL and the resolved version, then run the verification below.

## Verify

Use after publication, including when a workflow or another person created the release. Verification only reads and downloads public assets into a disposable directory; it never edits a release, moves a tag, or replaces an asset.

1. **Resolve the expected identity** from the approved candidate, not from the latest branch state:

   ```bash
   repo=$(gh repo view --json nameWithOwner --jq .nameWithOwner)
   tag=vX.Y.Z
   release_sha=$(git rev-parse HEAD)   # from the approved candidate checkout
   ```

1. **Reconcile the remote tag**: an annotated tag lists its tag object and a peeled `^{}` line; compare the peeled commit to `release_sha`, the first line is not the release commit:

   ```bash
   git ls-remote --exit-code --tags origin "refs/tags/$tag" "refs/tags/$tag^{}"
   ```

1. **Inspect the release state**: tag, draft and prerelease flags, publication time, assets, immutability, URL; `targetCommitish` may only name `main` and does not replace the peeled-tag check:

   ```bash
   gh release view "$tag" -R "$repo" --json tagName,name,isDraft,isPrerelease,isImmutable,publishedAt,targetCommitish,assets,url
   ```

1. **Prove exact-head automation**: every expected workflow at `release_sha` is present, completed, and successful; a green latest-branch run is not proof:

   ```bash
   gh run list -R "$repo" --commit "$release_sha" --limit 100 --json workflowName,attempt,headSha,status,conclusion,event,url
   ```

1. **Verify published assets** (checksums, release attestations, build provenance) per [verify-assets](references/verify-assets.md); anything missing is a failed proof, not permission to regenerate it.
1. **Verify the delivered boundary**: run the packaged binary or installation contract and confirm its version; [containerize](../containerize/SKILL.md) covers digest-bound OCI, Cosign, and SBOM checks.
1. **Report a release receipt** ending with the highest proven rung of the [proof ladder](../production-readiness/SKILL.md):
   - Expected commit, remote tag object and peeled commit, workflow names and URLs.
   - Release URL and state, expected versus downloaded assets, checksum and attestation results, packaged version.

## Gotchas

- **Semver source of truth**: let `git-cliff --bumped-version` decide; override only for a deliberate bump such as the first stable `v1.0.0`.
- **Tag prefix**: git-cliff, `gh`, and Go module tags all expect `vX.Y.Z`; keep the `v`.
- **Pre-1.0**: git-cliff applies the same rules below `v1.0.0`, so a `feat` bumps the minor and a breaking change jumps to `v1.0.0`.
- **Tolerant 0.x line**: set `features_always_bump_minor = false` and `breaking_always_bump_major = false` under `[bump]` in `cliff.toml`.
- **First release**: with no tag yet, git-cliff starts from the configured `initial_tag` (`v0.1.0` in the global config).
- **Remote truth**: a local tag, a green local gate, a draft release, or a latest-branch run proves nothing about the published release; reconcile the remote tag, exact-head workflows, state, and assets.
- **Config resolution**: git-cliff reads `cliff.toml` from the repository root or falls back to `~/.config/git-cliff/cliff.toml`; `--config` only forces another file.

## Documentation

- [git-cliff](https://git-cliff.org) · [gh release manual](https://cli.github.com/manual/gh_release)
- [GitHub release integrity](https://docs.github.com/en/code-security/how-tos/secure-your-supply-chain/secure-your-dependencies/verify-release-integrity)
- Companion skills: [conventional-commit](../conventional-commit/SKILL.md) (commit grammar), [github-pull-request](../github-pull-request/SKILL.md) (merge first), [mise](../mise/SKILL.md) (the gate).
