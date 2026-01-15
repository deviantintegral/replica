---
id: 6
group: "project-foundation"
dependencies: [5]
status: "pending"
created: "2026-01-14"
skills:
  - github-actions
  - release-automation
---
# Release Please and GoReleaser Configuration

## Objective
Configure release-please for automated changelog/releases and goreleaser for cross-platform binary distribution with CGO support.

## Skills Required
- **github-actions**: GitHub Actions for release automation
- **release-automation**: release-please and goreleaser configuration

## Acceptance Criteria
- [ ] release-please creates release PRs from conventional commits
- [ ] goreleaser produces binaries for Linux/macOS/Windows (amd64/arm64)
- [ ] Cross-compilation with CGO works via goreleaser-cross
- [ ] Release workflow triggers on version tags
- [ ] Binaries uploaded as release assets

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- release-please for changelog and version management
- goreleaser-cross Docker image for CGO cross-compilation
- Targets: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64

## Input Dependencies
- Task 5: GitHub Actions CI (release workflow builds on CI)

## Output Artifacts
- `.github/workflows/release.yml`
- `.goreleaser.yml`
- `release-please-config.json`
- `.release-please-manifest.json`

## Implementation Notes

<details>
<summary>Detailed Implementation Instructions</summary>

1. **Create release-please-config.json**:
   ```json
   {
     "$schema": "https://raw.githubusercontent.com/googleapis/release-please/main/schemas/config.json",
     "release-type": "go",
     "packages": {
       ".": {
         "changelog-path": "CHANGELOG.md",
         "release-type": "go"
       }
     },
     "bump-minor-pre-major": true,
     "bump-patch-for-minor-pre-major": true
   }
   ```

2. **Create .release-please-manifest.json**:
   ```json
   {
     ".": "0.0.0"
   }
   ```

3. **Create .goreleaser.yml**:
   ```yaml
   version: 2

   project_name: replica

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
       ignore:
         - goos: windows
           goarch: arm64
       ldflags:
         - -s -w
         - -X main.version={{.Version}}
         - -X main.commit={{.Commit}}
         - -X main.date={{.Date}}

   archives:
     - id: default
       format: tar.gz
       name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
       format_overrides:
         - goos: windows
           format: zip
       files:
         - LICENSE
         - README.md
         - replica.yaml.example

   checksum:
     name_template: "checksums.txt"

   changelog:
     use: github-native

   release:
     github:
       owner: deviantintegral
       name: replica
     draft: false
     prerelease: auto
   ```

4. **Create .github/workflows/release.yml**:
   ```yaml
   name: Release

   on:
     push:
       branches: [main]

   permissions:
     contents: write
     pull-requests: write
     packages: write

   jobs:
     release-please:
       name: Release Please
       runs-on: ubuntu-latest
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
       name: GoReleaser
       runs-on: ubuntu-latest
       needs: release-please
       if: ${{ needs.release-please.outputs.release_created }}
       env:
         REGISTRY: ghcr.io
         IMAGE_NAME: ${{ github.repository }}
       steps:
         - uses: actions/checkout@v4
           with:
             fetch-depth: 0

         - name: Set up Go
           uses: actions/setup-go@v5
           with:
             go-version: "1.25"

         - name: Run GoReleaser
           uses: goreleaser/goreleaser-action@v6
           with:
             distribution: goreleaser-cross
             version: latest
             args: release --clean
           env:
             GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

         - name: Log in to Container Registry
           uses: docker/login-action@v3
           with:
             registry: ${{ env.REGISTRY }}
             username: ${{ github.actor }}
             password: ${{ secrets.GITHUB_TOKEN }}

         - name: Build and push release Docker image
           uses: docker/build-push-action@v6
           with:
             context: .
             push: true
             tags: |
               ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ needs.release-please.outputs.tag_name }}
               ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:latest
             build-args: |
               VERSION=${{ needs.release-please.outputs.tag_name }}
               COMMIT=${{ github.sha }}
               DATE=${{ github.event.head_commit.timestamp }}
   ```

5. **Verify configuration**:
   ```bash
   # Check goreleaser config
   goreleaser check
   ```

</details>
