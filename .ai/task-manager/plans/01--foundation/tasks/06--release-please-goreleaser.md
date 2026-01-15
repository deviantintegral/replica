---
id: 6
group: "ci-cd"
dependencies: [5]
status: "completed"
created: "2026-01-15"
skills:
  - github-actions
  - devops
---
# Release Please and GoReleaser Configuration

## Objective
Configure automated release management with release-please for changelog generation and GoReleaser for cross-platform binary distribution. This enables professional release workflows with minimal manual intervention.

## Skills Required
- github-actions: Release automation workflows
- devops: Release tooling and cross-compilation

## Acceptance Criteria
- [ ] release-please workflow creates release PRs from conventional commits
- [ ] GoReleaser configuration builds for Linux/macOS/Windows (amd64, arm64)
- [ ] Release workflow publishes binaries to GitHub Releases
- [ ] Docker images are pushed to ghcr.io on release
- [ ] CGO cross-compilation works via goreleaser-cross

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- release-please for changelog and version management
- GoReleaser with goreleaser-cross Docker image for CGO
- GitHub Container Registry (ghcr.io) for Docker images
- Cross-compilation for Linux, macOS, Windows (amd64, arm64)

## Input Dependencies
- Task 6: CI workflow foundation

## Output Artifacts
- `.github/workflows/release.yml` - Release automation workflow
- `.goreleaser.yml` - GoReleaser configuration
- `.release-please-manifest.json` - Version manifest
- `release-please-config.json` - Release-please configuration

## Implementation Notes

<details>
<summary>Detailed Implementation Guide</summary>

### Release Please Configuration

Create `release-please-config.json`:

```json
{
  "$schema": "https://raw.githubusercontent.com/googleapis/release-please/main/schemas/config.json",
  "packages": {
    ".": {
      "release-type": "go",
      "bump-minor-pre-major": true,
      "bump-patch-for-minor-pre-major": true,
      "changelog-sections": [
        {"type": "feat", "section": "Features"},
        {"type": "fix", "section": "Bug Fixes"},
        {"type": "perf", "section": "Performance Improvements"},
        {"type": "docs", "section": "Documentation"},
        {"type": "chore", "section": "Miscellaneous"}
      ]
    }
  }
}
```

Create `.release-please-manifest.json`:

```json
{
  ".": "0.0.0"
}
```

### GoReleaser Configuration

Create `.goreleaser.yml`:

```yaml
version: 2

before:
  hooks:
    - go mod tidy

builds:
  - id: replica
    main: ./cmd/replica
    binary: replica
    env:
      - CGO_ENABLED=1
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X github.com/deviantintegral/replica/internal/version.Version={{.Version}}
      - -X github.com/deviantintegral/replica/internal/version.Commit={{.Commit}}
      - -X github.com/deviantintegral/replica/internal/version.BuildDate={{.Date}}

archives:
  - format: tar.gz
    name_template: >-
      {{ .ProjectName }}_
      {{- .Version }}_
      {{- .Os }}_
      {{- .Arch }}
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'

changelog:
  use: github-native

dockers:
  - image_templates:
      - "ghcr.io/deviantintegral/replica:{{ .Version }}"
      - "ghcr.io/deviantintegral/replica:latest"
    dockerfile: Dockerfile
    build_flag_templates:
      - "--build-arg=VERSION={{.Version}}"
      - "--build-arg=COMMIT={{.Commit}}"
      - "--build-arg=BUILD_DATE={{.Date}}"

release:
  github:
    owner: deviantintegral
    name: replica
```

### Release Workflow

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    branches: [main]

permissions:
  contents: write
  packages: write
  pull-requests: write

jobs:
  release-please:
    runs-on: ubuntu-24.04
    outputs:
      release_created: ${{ steps.release.outputs.release_created }}
      tag_name: ${{ steps.release.outputs.tag_name }}
    steps:
      - uses: googleapis/release-please-action@v4
        id: release
        with:
          config-file: release-please-config.json
          manifest-file: .release-please-manifest.json

  goreleaser:
    needs: release-please
    if: needs.release-please.outputs.release_created
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Login to GitHub Container Registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser-cross
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Renovate Configuration

Add to `renovate.json` custom managers (from Task 01):

```json
{
  "customType": "regex",
  "fileMatch": ["^\\.goreleaser\\.ya?ml$"],
  "matchStrings": [
    "image:\\s*['\"]?(?<depName>goreleaser/goreleaser-cross):(?<currentValue>[^\\s'\"]+)"
  ],
  "datasourceTemplate": "docker"
}
```

### Verification

1. Merge a PR with conventional commits (e.g., `feat: add feature`)
2. release-please should create a release PR
3. Merging the release PR triggers GoReleaser
4. Check GitHub Releases for binaries
5. Check ghcr.io for Docker image

</details>
