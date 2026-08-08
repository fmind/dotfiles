---
name: containerize
description: Build minimal, non-root OCI images — `ko` for Go or a distroless multi-stage Dockerfile — then scan, sign, and SBOM them. Use when containerizing or packaging an app for deployment.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/containerize
  created: 2026-07-04
  updated: 2026-08-02
---

# Containerize an Application

Build a small, non-root, reproducible OCI image and verify it before it ships. Pairs with [k8s-local](../k8s-local/SKILL.md) for the local dev loop and [security-scan](../security-scan/SKILL.md) for image scanning.

## Choose an Approach

1. **Go → `ko` (default, no Dockerfile)**: builds a minimal, shell-less, multi-arch, reproducible image straight from a package path (base defaults to `cgr.dev/chainguard/static`, override with `KO_DEFAULTBASEIMAGE`). Pin it per project (`go get -tool github.com/google/ko`, then `go tool ko`) so builds stay reproducible even where a global toolchain already provides `ko`.
   ```bash
   export KO_DOCKER_REPO=registry.localhost:5050/<slug>   # or a real registry
   go tool ko build ./cmd/<slug> --bare --platform=linux/amd64,linux/arm64
   ```
1. **Python (or any other language) → multi-stage Dockerfile** on a distroless or minimal base (optimized with `uv`). Copy and customize the image digests and the `<slug>` console-script entry point:
   - [Dockerfile](references/Dockerfile)
   - [.dockerignore](references/.dockerignore)

   ```bash
   # Build locally for current platform
   docker build -t <registry>/<slug>:<tag> .

   # Build multi-platform using Buildx (recommended for multi-arch registries)
   docker buildx build --platform linux/amd64,linux/arm64 -t <registry>/<slug>:<tag> --push .
   ```

## Verify Before Ship

1. **Scan** the built image (fail on HIGH/CRITICAL — see [security-scan](../security-scan/SKILL.md)):
   ```bash
   trivy image <registry>/<slug>:<tag>
   ```
1. **Sign** keyless with Sigstore (OIDC, no long-lived keys):
   ```bash
   cosign sign --yes <registry>/<slug>@<digest> # --yes is mandatory: cosign prompts by default and hangs any CI job or agent session
   ```
1. **SBOM** for provenance:
   ```bash
   trivy image --format cyclonedx -o sbom.json <registry>/<slug>:<tag>
   ```

## Mise Task Integration

Integrate image building and checking into the project's canonical task vocabulary (`mise.toml`):

```toml
[tasks."build:image"]
description = "Build OCI image"
run = "go tool ko build ./cmd/<slug> --bare" # For Go
# run = "docker build -t <registry>/<slug>:<tag> ." # For Python/Docker

[tasks."check:image"]
description = "Scan container image for vulnerabilities"
run = "trivy image <registry>/<slug>:<tag>"

[tasks."check:dockerfile"]
description = "Lint Dockerfile (hadolint)"
run = "hadolint Dockerfile"
```

## Local Dev Loop

Push to the shared k3d registry and deploy per [k8s-local](../k8s-local/SKILL.md):

```bash
docker push registry.localhost:5050/<slug>:<tag>   # ko pushes automatically
```

## Gotchas

- **Non-root + minimal**: the `ko` base (`chainguard/static`) has no shell or package manager — debug with `kubectl debug` ephemeral containers, not by adding a shell. `python:*-slim` (the [Dockerfile](references/Dockerfile) reference) is non-root but still Debian-based with a shell and `apt-get`; the template removes the unused runtime `pip`, refresh its exact release digest during upgrades, and swap in a shell-less final stage (e.g. `gcr.io/distroless/python3-debian13` or a Chainguard Python image) only once the copied `.venv` is verified compatible.
- **Use `.dockerignore`**: Always exclude development artifacts, local virtual environments, and secrets to speed up builds and prevent credential leakage (see [.dockerignore](references/.dockerignore)).
- **Static binaries**: `CGO_ENABLED=0` for `distroless/static`; use `distroless/base` only when cgo is required.
- **Digests over tags in manifests**: reference images by digest in Kubernetes so deploys are immutable.

## Documentation

- [ko](https://ko.build) · [distroless](https://github.com/GoogleContainerTools/distroless) · [Chainguard Images](https://images.chainguard.dev) · [cosign](https://docs.sigstore.dev/cosign/)
- Companion skills:
  - [k8s-local](../k8s-local/SKILL.md) — run the image locally.
  - [security-scan](../security-scan/SKILL.md) — scan deps and images.
  - [go-stack](../go-stack/SKILL.md) — the Go build it packages.
  - [python-stack](../python-stack/SKILL.md) — the Python app it packages.
