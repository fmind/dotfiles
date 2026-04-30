---
name: use-docker-cli
description: Guide for the docker CLI — image build (multi-stage, buildx), run, exec, compose, registries, prune, and cleanup hygiene.
---

# Use Docker CLI

`docker` is the canonical container CLI. The agent should prefer `docker buildx` for builds (multi-platform + cache), use multi-stage Dockerfiles to keep images small, and clean up dangling images/containers when iterating.

## One-time Setup

```bash
# Install via mise / system package manager (Docker Desktop on macOS).
mise use -g docker@latest      # or platform-native install

# Sanity check.
docker version
docker info

# (Optional) authenticate to GCP Artifact Registry.
gcloud auth configure-docker us-central1-docker.pkg.dev
```

## Build

```bash
# Standard build.
docker build -t my-app:dev .

# Multi-platform with buildx (default builder is `default`).
docker buildx build --platform linux/amd64,linux/arm64 -t my-app:dev .

# Push during build (registry must be reachable).
docker buildx build --push -t us-central1-docker.pkg.dev/$PROJ/repo/app:1.0 .

# Cache to a registry between CI runs.
docker buildx build \
  --cache-from=type=registry,ref=$REG/app:cache \
  --cache-to=type=registry,ref=$REG/app:cache,mode=max \
  -t $REG/app:1.0 --push .
```

## Multi-stage Dockerfile (canonical pattern)

```dockerfile
# Stage 1 — build.
FROM node:22-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

# Stage 2 — runtime (minimal).
FROM gcr.io/distroless/nodejs22 AS runtime
WORKDIR /app
COPY --from=build /app/dist ./dist
COPY --from=build /app/node_modules ./node_modules
USER nonroot:nonroot
CMD ["dist/server.js"]
```

## Run / Exec

```bash
# Run with port mapping + env.
docker run --rm -it -p 8080:8080 -e NODE_ENV=development my-app:dev

# Detached + named.
docker run -d --name api -p 8080:8080 my-app:dev
docker logs -f api
docker exec -it api sh

# Mount a host directory (read-write).
docker run --rm -v "$PWD":/work -w /work python:3.13 bash

# One-shot tool (without persisting).
docker run --rm -it ghcr.io/charmbracelet/glow:latest README.md
```

## Inspect

```bash
docker ps                     # running
docker ps -a                  # including stopped
docker images
docker inspect api            # JSON detail
docker stats                  # live CPU/mem (Ctrl+C to exit)
docker logs --tail 100 api
```

## Volumes & Networks

```bash
docker volume ls
docker volume create pgdata
docker run -v pgdata:/var/lib/postgresql/data postgres:16

docker network ls
docker network create dev
docker run --network=dev --name=db -d postgres:16
docker run --network=dev -e DB_HOST=db my-app
```

## Compose (multi-service dev)

```yaml
# compose.yaml
services:
  db:
    image: postgres:16
    environment: { POSTGRES_PASSWORD: dev }
    volumes: [pgdata:/var/lib/postgresql/data]
  api:
    build: .
    ports: ["8080:8080"]
    depends_on: [db]
    environment: { DATABASE_URL: postgres://postgres:dev@db/postgres }
volumes:
  pgdata:
```

```bash
docker compose up -d
docker compose logs -f api
docker compose down                 # stop + remove containers
docker compose down -v              # also remove volumes (destructive)
```

## Registry

```bash
# GAR (Google Artifact Registry).
docker tag my-app:dev us-central1-docker.pkg.dev/$PROJ/repo/app:1.0
docker push us-central1-docker.pkg.dev/$PROJ/repo/app:1.0

# Generic / Docker Hub.
docker login                        # interactive
docker login ghcr.io -u $USER --password-stdin <<< "$GH_TOKEN"
docker push ghcr.io/me/my-app:1.0
```

## Cleanup (do this often when iterating)

```bash
# Remove dangling images, stopped containers, unused networks.
docker system prune

# Aggressive — also unused images and build cache.
docker system prune -a --volumes

# Targeted.
docker container prune
docker image prune -a
docker buildx prune
docker volume prune
```

## Common Workflows

**Build + run a single-service Python app.**

```bash
docker build -t app .
docker run --rm -p 8080:8080 app
```

**Iterate on a Compose stack.**

```bash
docker compose up -d
docker compose logs -f api db
# Edit code; rebuild only the changed service:
docker compose build api && docker compose up -d api
```

**Smaller images.**

- Start from `distroless` or `alpine` for runtime stages.
- `.dockerignore` to keep build context tight (mirror `.gitignore` plus `.git/`, `node_modules/`).
- Multi-stage so build tools never reach the runtime image.

## Important Notes

1. **`docker run`'s `--rm`** removes the container on exit — use it for one-shot commands; omit for long-running services.
2. **Compose `down -v`** deletes volumes — confirm before running it against shared dev DBs.
3. **Buildx is the modern builder** — supports multi-platform, advanced caching, BuildKit features. Prefer it for everything.
4. **Image tags are mutable** — pin by digest (`@sha256:...`) for reproducibility.
5. **`.dockerignore` matters** — large build context (e.g. `node_modules`) makes every build slow.
6. **Don't run as root in production images** — use `USER nonroot:nonroot` or a numeric UID/GID.

## Documentation

- [Docker CLI reference](https://docs.docker.com/reference/cli/docker/)
- [Buildx (multi-platform builds)](https://docs.docker.com/build/building/multi-platform/)
- [Compose v2](https://docs.docker.com/compose/)
- [Multi-stage builds](https://docs.docker.com/build/building/multi-stage/)
- [`.dockerignore`](https://docs.docker.com/build/concepts/context/#dockerignore-files)
- [Distroless images](https://github.com/GoogleContainerTools/distroless)
- [GAR auth](https://docs.cloud.google.com/artifact-registry/docs/docker/authentication)
