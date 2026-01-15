---
id: 5
group: "project-foundation"
dependencies: [4]
status: "pending"
created: "2026-01-14"
skills:
  - github-actions
  - ci-cd
---
# GitHub Actions CI Workflow

## Objective
Create GitHub Actions workflow for continuous integration: lint, test, build, and publish Docker images to ghcr.io.

## Skills Required
- **github-actions**: GitHub Actions workflow syntax and jobs
- **ci-cd**: CI/CD pipeline design and Docker registry publishing

## Acceptance Criteria
- [ ] CI runs on all PRs and pushes to main
- [ ] Lint job runs golangci-lint
- [ ] Test job runs with race detection
- [ ] Build job produces binary artifacts
- [ ] Docker image published to ghcr.io on releases
- [ ] Jobs run in parallel where possible
- [ ] Renovate config validation job runs `npx renovate-config-validator --strict`

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- GitHub Actions workflow in `.github/workflows/ci.yml`
- Publish to ghcr.io (GitHub Container Registry)
- Use matrix for future database testing (prepared but not yet active)

## Input Dependencies
- Task 4: Dockerfile (for Docker build/publish)

## Output Artifacts
- `.github/workflows/ci.yml`

## Implementation Notes

<details>
<summary>Detailed Implementation Instructions</summary>

1. **Create `.github/workflows/ci.yml`**:
   ```yaml
   name: CI

   on:
     push:
       branches: [main]
     pull_request:
       branches: [main]

   env:
     GO_VERSION: "1.25"
     REGISTRY: ghcr.io
     IMAGE_NAME: ${{ github.repository }}

   jobs:
     validate-renovate:
       name: Validate Renovate Config
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4

         - name: Validate Renovate config
           run: npx --yes --package renovate -- renovate-config-validator --strict

     lint:
       name: Lint
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4

         - uses: actions/setup-go@v5
           with:
             go-version: ${{ env.GO_VERSION }}

         - name: golangci-lint
           uses: golangci/golangci-lint-action@v6
           with:
             version: latest

     test:
       name: Test
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4

         - uses: actions/setup-go@v5
           with:
             go-version: ${{ env.GO_VERSION }}

         - name: Run tests
           run: make test

         - name: Upload coverage
           uses: codecov/codecov-action@v4
           with:
             files: coverage.out
             fail_ci_if_error: false

     build:
       name: Build
       runs-on: ubuntu-latest
       needs: [lint, test]
       steps:
         - uses: actions/checkout@v4

         - uses: actions/setup-go@v5
           with:
             go-version: ${{ env.GO_VERSION }}

         - name: Build binary
           run: make build

         - name: Verify binary
           run: ./replica version

     docker:
       name: Docker
       runs-on: ubuntu-latest
       needs: [build]
       if: github.event_name == 'push' && github.ref == 'refs/heads/main'
       permissions:
         contents: read
         packages: write
       steps:
         - uses: actions/checkout@v4

         - name: Log in to Container Registry
           uses: docker/login-action@v3
           with:
             registry: ${{ env.REGISTRY }}
             username: ${{ github.actor }}
             password: ${{ secrets.GITHUB_TOKEN }}

         - name: Extract metadata
           id: meta
           uses: docker/metadata-action@v5
           with:
             images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
             tags: |
               type=sha
               type=raw,value=latest

         - name: Build and push
           uses: docker/build-push-action@v6
           with:
             context: .
             push: true
             tags: ${{ steps.meta.outputs.tags }}
             labels: ${{ steps.meta.outputs.labels }}
             build-args: |
               VERSION=${{ github.sha }}
               COMMIT=${{ github.sha }}
               DATE=${{ github.event.head_commit.timestamp }}
   ```

2. **Update Makefile** for CI compatibility - update test target to output coverage:
   ```makefile
   ## test: Run tests with race detection and coverage
   test:
   	go test -race -coverprofile=coverage.out -covermode=atomic ./...
   ```

3. **Verify locally** (if `act` is installed):
   ```bash
   act -l  # List jobs
   ```

</details>
