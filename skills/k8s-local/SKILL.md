---
name: k8s-local
description: Create and manage local Kubernetes clusters (k3d or kind) and deploy to them with kubectl, helm, helmfile, and skaffold. Use for local k8s cluster setup, dev loops, and debugging.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/k8s-local
  created: 2026-06-23
  updated: 2026-07-09
---

# Local Kubernetes Cluster Management

Bootstrap and manage lightweight local Kubernetes clusters with **k3d** (default, Docker-based K3s) or **kind**, plus the local dev-tool ecosystem (kubectl, helm, helmfile, skaffold, k9s).

## Global vs Project-Scoped Setup

When designing project local environments, choose the appropriate isolation scope:

1. **Project-Scoped Setup (Recommended for important/complex projects)**: Important projects must remain independent, self-contained, and reproducible across hosts. Do not rely on the global `k3d-local` cluster or default configs. Instead, declare all dependencies (like `k3d`, `kubectl`, `helm`, `flux2`) under the `[tools]` section of the project's mise.toml, store cluster engine configurations within the repository, and isolate active workloads by generating temporary, process-scoped `KUBECONFIG` variables. Author a project-scoped skill at `.agents/skills/local-cluster/SKILL.md` to guide agents to use your local task vocabulary instead of global `dot cluster` commands.
1. **Global Setup (Optional fallback for small/simple projects)**: Simple projects or quick experiments without custom orchestration needs can rely on the global cluster configuration managed here. Use the unified `dot cluster` CLI or manually bootstrap using the global templates.

## Templates & Resources

- **k3d Configuration**: Deployed at `~/.config/k3d/local.yaml` (source: `dot_config/k3d/local.yaml`).
- **kind Configuration**: [kind-config.yaml](resources/kind-config.yaml)
- **Local Cluster CLI**: Deployed at `~/.local/bin/dot` (source: `dot/`).
- **Ingress Template**: [ingress-template.yaml](resources/ingress-template.yaml)

## Core Principles

- **Cloud Agnostic**: Local clusters must mirror production API interfaces. Avoid vendor-specific or platform-specific configurations unless mapping local routes.
- **Engine Choice (Hybrid Pattern)**: Match the engine to the workflow. Use `k3d` for interactive developer loops (fast startup, built-in Traefik ingress, and registry integration). Use `kind` for automated/disposable integration tests to validate vanilla, upstream-conformant Kubernetes API and security compliance.
- **Local Registry Dev Loop**: Push locally built Docker images to the local registry at `registry.localhost:5050` or load them directly into the cluster engine nodes. (The registry container is named `registry.localhost`, so the same `registry.localhost:5050` reference resolves for both host push and in-cluster pull via the k3s containerd mirror.)
- **Declarative Operations**: Manage all workloads using declarative configurations (Helm, Helmfile, Skaffold, Kustomize) rather than raw imperatives.

## AI Agent Instructions

- **Docker Healthcheck**: Always run `docker info` before launching, stopping, or configuring clusters.
- **Context Verification**: Verify that the active context matches the intended local cluster before running `kubectl` commands. Use `kubectx` or `kubectl config current-context` to ensure the context is `k3d-local`.
- **Local Images**: Build local images, tag them for the local registry (`registry.localhost:5050/image:tag`), and push them, or load them directly into the cluster engine. Set `imagePullPolicy: IfNotPresent` or `Never` in deployment specs if using sideloaded or locally tagged images.

## Workflow

1. **Verify Docker Daemon**: Ensure that the Docker runtime is active and running:
   ```bash
   docker info
   ```
1. **Choose and Create Cluster**:
   - **Using k3d (Recommended)**: Use the `dot` CLI to idempotently ensure the cluster is configured and running:
     ```bash
     dot cluster start
     ```
     Or manually create it using the configuration file:
     ```bash
     k3d cluster create --config ~/.config/k3d/local.yaml
     ```
   - **Using kind**:
     ```bash
     kind create cluster --config resources/kind-config.yaml
     ```
1. **Context & Namespace Switching**:
   - List and switch contexts easily:
     ```bash
     kubectx
     kubectx k3d-local # or kind-local-cluster
     ```
   - List and switch active namespaces:
     ```bash
     kubens
     kubens default
     ```
1. **Validate and Lint Manifests**:
   - Validate YAML structures:
     ```bash
     kubeconform -strict <file-or-directory>
     ```
   - Lint configurations for best practices and security:
     ```bash
     kube-linter lint <file-or-directory>
     ```
1. **Publish or Load Local Images**:
   - **Using Local Registry (Recommended for k3d)**: Tag and push the image:
     ```bash
     docker tag <image>:<tag> registry.localhost:5050/<image>:<tag>
     docker push registry.localhost:5050/<image>:<tag>
     ```
   - **Sideloading directly (kind or k3d fallback)**:
     - **For kind**:
       ```bash
       kind load docker-image <image>:<tag> --name local-cluster
       ```
     - **For k3d**:
       ```bash
       k3d image load <image>:<tag> -c local
       ```
1. **Declarative Deployments**:
   - **Via Helm**:
     ```bash
     helm upgrade --install <release-name> <chart-path> --namespace <namespace> --create-namespace
     ```
   - **Via Helmfile**:
     ```bash
     helmfile apply
     ```
   - **Via Skaffold (Development inner-loop)**:
     ```bash
     skaffold dev
     ```
   - **Via Kustomize**:
     ```bash
     kubectl apply -k <kustomization-directory>
     ```
1. **Observability & Debugging**:
   - Launch `k9s` to monitor and manage the cluster interactively:
     ```bash
     k9s
     ```
   - Tail logs from multiple pods matching a query:
     ```bash
     stern <pod-name-query> -n <namespace>
     ```
1. **Local-Remote Integration**:
   - Use `mirrord` to run a local process inside the context of the cluster:
     ```bash
     mirrord exec --target deploy/<deployment-name> -- <local-command>
     ```
1. **Teardown & Cleanup**:
   - **For k3d**:
     ```bash
     k3d cluster stop local  # Pause to save resources
     k3d cluster start local # Resume cluster
     k3d cluster delete local
     ```
   - **For kind**:
     ```bash
     kind delete cluster --name local-cluster
     ```

## Gotchas

1. **Port Conflict**: Host ports `8080` (HTTP Ingress), `8443` (HTTPS Ingress), `6443` (API Server), and `5050` (Local Registry) must not be occupied by other local services.
1. **Image Pull Policies**: Kubernetes defaults to pulling images if the tag is `latest`. Sideloaded or locally published images must use a specific tag, and the manifest must set `imagePullPolicy: IfNotPresent` or `Never`.
1. **Ingress Controllers**: `k3d` has Traefik enabled by default. `kind` requires applying an ingress controller manually (e.g. Nginx Ingress Controller).
