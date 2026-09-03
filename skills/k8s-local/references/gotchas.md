# Local Kubernetes Gotchas

Longer explanations behind the one-line gotchas in [SKILL.md](SKILL.md).

## Registry push name

The cluster pulls `registry.localhost:5050/...` because k3d writes that mirror into containerd, but the Docker daemon rejects a push to the same name with `server gave HTTP response to HTTPS client`: the registry speaks plain HTTP and `registry.localhost` is not covered by the daemon's default insecure ranges. Pushing to `localhost:5050/...` reaches the same registry container and works, but any image a manifest must pull has to be named `registry.localhost:5050/...`. To use one name for both, add `{"insecure-registries": ["registry.localhost:5050"]}` to `/etc/docker/daemon.json` and restart the daemon.

## DiskPressure

The kubelet reserves 10% of the filesystem by default. A local cluster shares the workstation disk, so an ordinarily full machine taints the node `NoSchedule` and even Traefik stays `Pending` while the node still reports `Ready`. The k3d config overrides `eviction-hard` with an absolute reserve; check `kubectl get node -o jsonpath='{.spec.taints}'` first whenever pods will not schedule.

## Port conflicts

Host ports `8042` (HTTP ingress), `8043` (HTTPS ingress), `6443` (API server), and `5050` (local registry) must be free. The ingress pair deliberately avoids `8080` and `8443`: those are the first ports a local web server or another project's cluster claims, and k3d only discovers the clash after it has partly created the cluster, then rolls the whole thing back.
