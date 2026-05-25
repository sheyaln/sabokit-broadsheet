# GitHub Actions Workflows

Workflows for Broadsheet.

## Docker image build

Two workflows publish to **GitHub Container Registry** at `ghcr.io/<owner>/broadsheet`. Both authenticate with the workflow-provided `GITHUB_TOKEN` — no external secrets needed.

### `docker-release.yml` — auto on tag push

Triggers on push of any tag matching `v*.*` or the `latest` tag.

- Multi-arch: `linux/amd64` (ubuntu-latest) + `linux/arm64` (ubuntu-24.04-arm)
- Builds per-arch by digest, then merges into a single manifest list
- Tag policy: the literal tag name (`v1.2`, `latest`, etc.) — no derived `v1` / `v1.2` partial tags
- Per-arch and manifest-list build provenance attestations pushed to the registry
- OCI labels include `broadsheet.git.commit`, `broadsheet.git.tag`, `broadsheet.build.url`

Trigger:
```
git tag v1.2 && git push origin v1.2
# or
git tag -f latest && git push -f origin latest
```

### `docker-manual.yml` — workflow_dispatch

Manual trigger from the Actions UI. Inputs:
- `tag` — image tag (default `latest`)
- `push` — push to registry vs. build-only (default true)

Use for ad-hoc builds and testing.

## Required configuration

**Image visibility.** First push to `ghcr.io/<owner>/broadsheet` creates a private package linked to the repo owner. To make it public, go to the package settings on github.com after the first successful push.

**Permissions.** The workflows declare `permissions.packages: write` at the job level — no repo-wide setting needed.

## Pulling the image

```
docker pull ghcr.io/<owner>/broadsheet:latest
docker pull ghcr.io/<owner>/broadsheet:v1.2
docker run -p 8080:8080 ghcr.io/<owner>/broadsheet:latest
```

## Other workflows

- `go.yml` — Go unit tests on push to `dev`/`main` and PRs against either
- `claude.yml`, `claude-code-review.yml` — Claude integration hooks
