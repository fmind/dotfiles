---
name: goreleaser
description: "Release Go binaries with GoReleaser: minimal .goreleaser.yaml, snapshot builds, tag-driven GitHub Actions release, cosign signing, ko images. Use when releasing a Go CLI."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/goreleaser
  created: "2026-09-03"
  updated: "2026-09-03"
---

# GoReleaser

Build, archive, checksum, sign, and publish Go binaries from one `.goreleaser.yaml` when a `v*` tag lands. The version, changelog, and tag come from [release](../release/SKILL.md) (git-cliff); images come from `ko` per [containerize](../containerize/SKILL.md); signatures follow [cosign](../cosign/SKILL.md).

## Workflow

1. **Config**: `goreleaser init` writes a starter; trim it to the minimum and keep GoReleaser out of changelog duty.
   ```yaml
   version: 2
   builds:
     - main: ./cmd/<slug>
       env: [CGO_ENABLED=0]
       goos: [linux, darwin]
       goarch: [amd64, arm64]
       ldflags: [-s -w -X main.version={{.Version}}]
   archives:
     - formats: [tar.gz]
   checksum:
     name_template: checksums.txt
   changelog:
     disable: true # git-cliff owns CHANGELOG.md and the release notes
   ```
1. **Validate**: `goreleaser check` after every edit; `goreleaser healthcheck` confirms `git`, `cosign`, and `ko` are on the runner.
1. **Local dry run**: build every target without a tag and without publishing.
   ```bash
   goreleaser release --snapshot --clean
   goreleaser build --single-target --snapshot --clean -o dist/<slug> # this machine only
   ```
1. **CI**: a `release.yml` workflow on `push: tags: ['v*']` with `permissions: contents: write` (plus `id-token: write` when signing), `actions/checkout` with `fetch-depth: 0`, `jdx/mise-action` for the toolchain, then the release with git-cliff notes; this workflow owns the GitHub release, so skip the `gh release create` step of the release skill.
   ```bash
   git-cliff --latest --strip all > "$RUNNER_TEMP/notes.md"
   goreleaser release --clean --release-notes "$RUNNER_TEMP/notes.md" # GITHUB_TOKEN from secrets.GITHUB_TOKEN
   ```
1. **Sign and ship images (optional)**: add `signs:` with `cmd: cosign`, `args: [sign-blob, --output-certificate=${certificate}, --output-signature=${signature}, --yes, "${artifact}"]`, `artifacts: checksum`, and `kos:` (`repositories`, `platforms`) for a multi-arch image; verify with `cosign verify-blob` per [cosign](../cosign/SKILL.md).
1. **Homebrew (optional)**: `homebrew_casks:` publishes a cask to a tap repository; it needs a token with write access to that repository, not `GITHUB_TOKEN`.
1. **Verify**: `gh release view <tag>` lists the archives, `checksums.txt`, and signatures; finish with the verification steps of [release](../release/SKILL.md).

## Gotchas

- **Tag first**: `goreleaser release` refuses a dirty tree or a missing tag; `--snapshot` is the only tag-less path.
- **Shallow clones**: without `fetch-depth: 0`, tag and version detection fail.
- **Deprecated keys**: `brews` became `homebrew_casks`, `archives.format` became `formats`, `kos.repository` became `repositories`; `goreleaser check` prints them.
- **Cross-compilation**: keep `CGO_ENABLED=0`; a cgo dependency needs a per-platform toolchain and usually a separate build.

## Documentation

- [GoReleaser](https://goreleaser.com) · [GitHub Actions](https://goreleaser.com/ci/actions/) · [Signing](https://goreleaser.com/customization/sign/) · [ko](https://goreleaser.com/customization/ko/) · [Homebrew casks](https://goreleaser.com/customization/homebrew_casks/)
- Companion skills: [release](../release/SKILL.md) (version and notes), [github-actions](../github-actions/SKILL.md) (workflow shape), [cosign](../cosign/SKILL.md) (signatures), [containerize](../containerize/SKILL.md) (`ko` images), [go-stack](../go-stack/SKILL.md) (project layout).
