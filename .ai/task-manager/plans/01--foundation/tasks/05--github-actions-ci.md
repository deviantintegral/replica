---
id: 5
group: "ci-cd"
dependencies: [4]
status: "pending"
created: "2026-01-15"
skills:
  - github-actions
---
# GitHub Actions CI Workflow

## Objective
Create GitHub Actions workflow for continuous integration that lints, tests, and builds the project on every PR and push to main. This ensures code quality and catches issues early.

## Skills Required
- github-actions: GitHub Actions workflow syntax and best practices

## Acceptance Criteria
- [ ] CI workflow runs on PR and push to main
- [ ] Workflow runs lint, test, and build jobs
- [ ] Docker image builds successfully in CI
- [ ] Job failures block PR merge
- [ ] Renovate config validation job passes
- [ ] Workflow caches Go modules for faster builds

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- GitHub Actions workflow syntax
- Go module caching
- Docker build in CI
- Renovate config validation with `npx renovate-config-validator --strict`

## Input Dependencies
- Task 5: Dockerfile for building image in CI

## Output Artifacts
- `.github/workflows/ci.yml` - CI workflow definition

## Implementation Notes

<details>
<summary>Detailed Implementation Guide</summary>

### CI Workflow

Create `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:
  contents: read

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest

  test:
    name: Test
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Install SQLite dependencies
        run: sudo apt-get update && sudo apt-get install -y libsqlite3-dev

      - name: Run tests
        run: go test -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: coverage.out
          fail_ci_if_error: false

  build:
    name: Build
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Install SQLite dependencies
        run: sudo apt-get update && sudo apt-get install -y libsqlite3-dev

      - name: Build binary
        run: make build

      - name: Verify binary
        run: ./bin/replica version

  docker:
    name: Docker Build
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Build Docker image
        uses: docker/build-push-action@v6
        with:
          context: .
          push: false
          tags: replica:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  validate-renovate:
    name: Validate Renovate Config
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4

      - name: Validate renovate.json
        run: npx renovate-config-validator --strict
```

### Key Features

1. **Separate Jobs**: Each concern (lint, test, build, docker) is a separate job for parallel execution
2. **Go Module Caching**: `actions/setup-go@v5` automatically caches Go modules
3. **Docker Layer Caching**: Uses GitHub Actions cache for Docker builds
4. **Coverage Upload**: Optional codecov integration
5. **Renovate Validation**: Ensures config is always valid

### Pull Request Requirements

In GitHub repository settings, configure branch protection:
- Require status checks to pass before merging
- Required checks: lint, test, build, docker, validate-renovate

### Verification

After pushing the workflow:
1. Create a PR to trigger the workflow
2. Verify all jobs pass
3. Check that Docker image builds successfully
4. Confirm Renovate config validation passes

</details>
