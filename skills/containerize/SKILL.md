---
name: containerize
description: Build minimal, non-root OCI images with ko for Go or a distroless multi-stage Dockerfile, then scan, sign, and SBOM them. Use when containerizing an app.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/containerize
  created: "2026-07-04"
  updated: "2026-09-03"
---

# Containerize an Application

Build a small, non-root, reproducible OCI image and verify it before it ships; [cloud-run](../cloud-run/SKILL.md) deploys it, [trivy](../trivy/SKILL.md) scans it, and [cosign](../cosign/SKILL.md) signs it.

## Workflow

1. **Go: `ko` (default, no Dockerfile)**: builds a shell-less, multi-arch, reproducible image from a package path on `cgr.dev/chainguard/static`. Pin it per project (`go get -tool github.com/google/ko`, then `go tool ko`).

   ```bash
   export KO_DOCKER_REPO=<registry>/<slug>
   go tool ko build ./cmd/<slug> --bare --platform=linux/amd64,linux/arm64
   ```

1. **Other languages: multi-stage Dockerfile**: copy [Dockerfile](references/Dockerfile) and [.dockerignore](references/.dockerignore), then set the image digests and the `<slug>` entry point.

   ```bash
   docker build -t <registry>/<slug>:<tag> .
   docker buildx build --platform linux/amd64,linux/arm64 -t <registry>/<slug>:<tag> --push .
   ```

1. **Wire the tasks** into `mise.toml`: `build:image` builds, `check:image` scans a local tarball so no registry push or digest is needed before the image ships.

   ```toml
   [env]
   KO_DOCKER_REPO = "<registry>/<slug>" # ko names images from it even when not pushing

   [tasks."build:image"]
   description = "Build the OCI image (ko)"
   run = "go tool ko build ./cmd/<slug> --bare" # or: docker build -t <registry>/<slug>:<tag> .

   [tasks."check:image"]
   description = "Scan the OCI image for vulnerabilities (trivy)"
   run = [
     "mkdir -p tmp",
     "go tool ko build ./cmd/<slug> --bare --push=false --tarball tmp/image.tar",
     "trivy --config trivy.yaml image --input tmp/image.tar",
   ]
   ```

1. **Verify the pushed digest**: use the registry digest for every scan, signature, SBOM, and deployment reference; scan per [trivy](../trivy/SKILL.md), then sign, verify, and attest the SBOM per [cosign](../cosign/SKILL.md).
1. **Run locally** the way Cloud Run will: `docker run --rm -p 8080:8080 -e PORT=8080 <registry>/<slug>@<digest>`; a project with a local cluster pushes to its registry per [k8s-local](../k8s-local/SKILL.md).

## Gotchas

- **No shell in the image**: `cgr.dev/chainguard/static` has no shell or package manager; debug with logs or an ephemeral sidecar, never by adding a shell.
- **cgo**: build with `CGO_ENABLED=0`; when cgo is unavoidable, set `KO_DEFAULTBASEIMAGE=cgr.dev/chainguard/glibc-dynamic` and stay in the Chainguard family.
- **Python base**: `python:*-slim` in the [Dockerfile](references/Dockerfile) is non-root but Debian-based with a shell; the template uninstalls `pip`, and [upgrade-tools](../upgrade-tools/SKILL.md) refreshes its digest.
- **`.dockerignore`**: exclude development artifacts, virtual environments, and secrets ([.dockerignore](references/.dockerignore)).
- **Digests over tags**: reference images by digest in Cloud Run services and Kubernetes manifests so deploys are immutable.
- **Pushes and signatures are registry writes**: a local packaging request does not authorize them.

## Documentation

- [ko](https://ko.build) · [Chainguard Images](https://images.chainguard.dev) · [distroless](https://github.com/GoogleContainerTools/distroless)
- Companion skills: [cloud-run](../cloud-run/SKILL.md) (deploys), [trivy](../trivy/SKILL.md) and [cosign](../cosign/SKILL.md) (scan, sign, attest), [github-actions](../github-actions/SKILL.md) (CD job), [secure](../secure/SKILL.md), [go-stack](../go-stack/SKILL.md), [python-stack](../python-stack/SKILL.md).
