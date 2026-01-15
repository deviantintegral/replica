---
id: 15
group: "testing"
dependencies: [5, 10, 11, 14]
status: "pending"
created: "2026-01-15"
skills:
  - github-actions
---
# CI Database Matrix Testing

## Objective
Extend the CI workflow to run integration tests against all three database backends (SQLite, MariaDB 11.8, PostgreSQL 18). This ensures compatibility across all supported databases.

## Skills Required
- github-actions: Matrix strategy and service containers

## Acceptance Criteria
- [ ] CI matrix runs tests against SQLite
- [ ] CI matrix runs tests against MariaDB 11.8
- [ ] CI matrix runs tests against PostgreSQL 18
- [ ] Integration tests are tagged and run separately
- [ ] All database tests must pass for PR merge

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- GitHub Actions matrix strategy
- Service containers for MariaDB and PostgreSQL
- Build tags for integration tests
- Separate unit and integration test jobs

## Input Dependencies
- Task 6: Base CI workflow
- Task 11: MariaDB driver
- Task 12: PostgreSQL driver
- Task 15: Test utilities

## Output Artifacts
- Updated `.github/workflows/ci.yml` with database matrix
- `internal/database/integration_test.go` - Integration tests (build tagged)

## Implementation Notes

<details>
<summary>Detailed Implementation Guide</summary>

### Updated CI Workflow

Update `.github/workflows/ci.yml` to add database matrix:

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
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest

  unit-test:
    name: Unit Tests
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - name: Install SQLite dependencies
        run: sudo apt-get update && sudo apt-get install -y libsqlite3-dev
      - name: Run unit tests
        run: go test -race -coverprofile=coverage.out ./...
      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: coverage.out
          fail_ci_if_error: false

  integration-test:
    name: Integration Tests (${{ matrix.database }})
    runs-on: ubuntu-24.04
    strategy:
      fail-fast: false
      matrix:
        database: [sqlite, mariadb, postgres]
        include:
          - database: sqlite
            db_driver: sqlite
            db_url: ":memory:"
          - database: mariadb
            db_driver: mariadb
            db_url: "replica:replica_password@tcp(127.0.0.1:3306)/replica"
          - database: postgres
            db_driver: postgres
            db_url: "postgres://replica:replica_password@127.0.0.1:5432/replica?sslmode=disable"

    services:
      mariadb:
        image: mariadb:11.8
        env:
          MARIADB_ROOT_PASSWORD: root_password
          MARIADB_DATABASE: replica
          MARIADB_USER: replica
          MARIADB_PASSWORD: replica_password
        ports:
          - 3306:3306
        options: >-
          --health-cmd="healthcheck.sh --connect --innodb_initialized"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=5
        # Only start for mariadb matrix entry
        if: matrix.database == 'mariadb'

      postgres:
        image: postgres:18
        env:
          POSTGRES_DB: replica
          POSTGRES_USER: replica
          POSTGRES_PASSWORD: replica_password
        ports:
          - 5432:5432
        options: >-
          --health-cmd="pg_isready -U replica -d replica"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=5
        # Only start for postgres matrix entry
        if: matrix.database == 'postgres'

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Install SQLite dependencies
        run: sudo apt-get update && sudo apt-get install -y libsqlite3-dev

      - name: Run integration tests
        env:
          TEST_DB_DRIVER: ${{ matrix.db_driver }}
          TEST_DB_URL: ${{ matrix.db_url }}
        run: go test -race -tags=integration ./...

  build:
    name: Build
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - name: Install SQLite dependencies
        run: sudo apt-get update && sudo apt-get install -y libsqlite3-dev
      - run: make build
      - run: ./bin/replica version

  docker:
    name: Docker Build
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - uses: docker/build-push-action@v6
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
      - run: npx renovate-config-validator --strict
```

### Integration Test File

Create `internal/database/integration_test.go`:

```go
//go:build integration

package database_test

import (
    "context"
    "os"
    "testing"
    "time"

    "github.com/rs/zerolog"

    "github.com/deviantintegral/replica/internal/database"
)

func TestIntegration_DatabaseConnection(t *testing.T) {
    driver := os.Getenv("TEST_DB_DRIVER")
    url := os.Getenv("TEST_DB_URL")

    if driver == "" || url == "" {
        t.Skip("TEST_DB_DRIVER and TEST_DB_URL must be set for integration tests")
    }

    logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

    cfg := database.Config{
        Driver: driver,
        URL:    url,
    }

    db, err := database.New(cfg, logger, database.Options{AutoMigrate: true})
    if err != nil {
        t.Fatalf("failed to create database: %v", err)
    }
    defer db.Close()

    // Test ping
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := db.Ping(ctx); err != nil {
        t.Errorf("ping failed: %v", err)
    }

    // Test transaction
    tx, err := db.Begin(ctx)
    if err != nil {
        t.Fatalf("failed to begin transaction: %v", err)
    }
    if err := tx.Rollback(); err != nil {
        t.Errorf("rollback failed: %v", err)
    }
}
```

### Running Integration Tests Locally

```bash
# SQLite (no setup needed)
TEST_DB_DRIVER=sqlite TEST_DB_URL=":memory:" go test -tags=integration ./...

# MariaDB
docker compose up -d mariadb
TEST_DB_DRIVER=mariadb TEST_DB_URL="replica:replica_password@tcp(127.0.0.1:3306)/replica" go test -tags=integration ./...

# PostgreSQL
docker compose up -d postgres
TEST_DB_DRIVER=postgres TEST_DB_URL="postgres://replica:replica_password@127.0.0.1:5432/replica?sslmode=disable" go test -tags=integration ./...
```

### Verification

1. Push changes to trigger CI
2. Verify matrix jobs run for all three databases
3. Confirm all database tests pass

</details>
