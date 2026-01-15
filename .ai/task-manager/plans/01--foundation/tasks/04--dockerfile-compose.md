---
id: 4
group: "build-system"
dependencies: [3]
status: "completed"
created: "2026-01-15"
skills:
  - docker
---
# Dockerfile and Docker Compose

## Objective
Create Dockerfile for production container image and docker-compose.yml for local development with all three database backends. This enables containerized deployment and consistent development environments.

## Skills Required
- docker: Dockerfile best practices, multi-stage builds, docker-compose

## Acceptance Criteria
- [ ] Dockerfile builds a minimal Alpine-based image
- [ ] Docker image runs and responds to `replica version`
- [ ] docker-compose.yml starts SQLite, MariaDB 11.8, and PostgreSQL 18
- [ ] Health checks are configured for database services
- [ ] Image size is optimized where reasonable (no hard limit)

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- Alpine base image (no libvips until 0.4.0)
- Multi-stage build for smaller image
- CGO support in build stage for go-sqlite3
- MariaDB 11.8 and PostgreSQL 18 for development

## Input Dependencies
- Task 4: Makefile with build targets

## Output Artifacts
- `Dockerfile` - Production container image
- `docker-compose.yml` - Local development environment
- `.dockerignore` - Build context exclusions

## Implementation Notes

<details>
<summary>Detailed Implementation Guide</summary>

### Dockerfile

Create `Dockerfile`:

```dockerfile
# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies for CGO (go-sqlite3)
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled for SQLite
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=1 go build \
    -ldflags "-X github.com/deviantintegral/replica/internal/version.Version=${VERSION} \
              -X github.com/deviantintegral/replica/internal/version.Commit=${COMMIT} \
              -X github.com/deviantintegral/replica/internal/version.BuildDate=${BUILD_DATE}" \
    -o /replica ./cmd/replica

# Runtime stage
FROM alpine:3.23

# Install runtime dependencies for SQLite
RUN apk add --no-cache sqlite-libs ca-certificates

# Create non-root user
RUN addgroup -S replica && adduser -S replica -G replica

WORKDIR /app

# Copy binary from builder
COPY --from=builder /replica .

# Use non-root user
USER replica

# Default port (will be configurable)
EXPOSE 8080

ENTRYPOINT ["./replica"]
```

### Docker Compose

Create `docker-compose.yml`:

```yaml
services:
  # SQLite doesn't need a service - it uses a file
  # But we can include replica with SQLite config
  replica-sqlite:
    build: .
    volumes:
      - ./data:/app/data
    environment:
      - REPLICA_DATABASE_DRIVER=sqlite
      - REPLICA_DATABASE_URL=/app/data/replica.db
    ports:
      - "8080:8080"
    profiles:
      - sqlite

  mariadb:
    image: mariadb:11.8
    environment:
      MARIADB_ROOT_PASSWORD: replica_root
      MARIADB_DATABASE: replica
      MARIADB_USER: replica
      MARIADB_PASSWORD: replica_password
    ports:
      - "3306:3306"
    volumes:
      - mariadb_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "healthcheck.sh", "--connect", "--innodb_initialized"]
      interval: 10s
      timeout: 5s
      retries: 5

  postgres:
    image: postgres:18
    environment:
      POSTGRES_DB: replica
      POSTGRES_USER: replica
      POSTGRES_PASSWORD: replica_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U replica -d replica"]
      interval: 10s
      timeout: 5s
      retries: 5

  replica-mariadb:
    build: .
    depends_on:
      mariadb:
        condition: service_healthy
    environment:
      - REPLICA_DATABASE_DRIVER=mariadb
      - REPLICA_DATABASE_URL=replica:replica_password@tcp(mariadb:3306)/replica
    ports:
      - "8081:8080"
    profiles:
      - mariadb

  replica-postgres:
    build: .
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      - REPLICA_DATABASE_DRIVER=postgres
      - REPLICA_DATABASE_URL=postgres://replica:replica_password@postgres:5432/replica?sslmode=disable
    ports:
      - "8082:8080"
    profiles:
      - postgres

volumes:
  mariadb_data:
  postgres_data:
```

### Dockerignore

Create `.dockerignore`:

```
# Git
.git
.gitignore

# Build artifacts
bin/
*.out
coverage.*

# IDE
.idea/
.vscode/
*.swp
*.swo

# Documentation
docs/
*.md
!README.md

# CI/CD
.github/

# Task manager
.ai/

# Test data
testdata/
*_test.go

# Local development
docker-compose.yml
.env*
data/
```

### Verification

```bash
# Build image
docker build -t replica:dev .

# Run version command
docker run --rm replica:dev version

# Start databases for development
docker compose up -d mariadb postgres

# Check health
docker compose ps
```

</details>
