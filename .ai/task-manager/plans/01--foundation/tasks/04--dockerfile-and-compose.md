---
id: 4
group: "project-foundation"
dependencies: [3]
status: "pending"
created: "2026-01-14"
skills:
  - docker
---
# Dockerfile and Docker Compose

## Objective
Create a multi-stage Dockerfile for building the Replica container image and docker-compose.yml for local development with all three database backends.

## Skills Required
- **docker**: Dockerfile multi-stage builds, Docker Compose configuration

## Acceptance Criteria
- [ ] `make docker` builds container image
- [ ] Multi-stage build produces minimal Alpine-based image
- [ ] docker-compose.yml provides SQLite, MariaDB 11.8, PostgreSQL 18 for local dev
- [ ] Container runs and `replica version` works
- [ ] Image size optimized (no libvips until 0.4.0)

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- Base image: Alpine (latest)
- Multi-stage build: builder stage with Go, final stage with minimal runtime
- CGO enabled for go-sqlite3
- Docker Compose with MariaDB 11.8 and PostgreSQL 18 services

## Input Dependencies
- Task 3: Makefile (provides `docker` target structure)

## Output Artifacts
- `Dockerfile`
- `docker-compose.yml`
- Updated `Makefile` with `docker` target

## Implementation Notes

<details>
<summary>Detailed Implementation Instructions</summary>

1. **Create Dockerfile**:
   ```dockerfile
   # Build stage
   FROM golang:1.25-alpine AS builder

   # Install build dependencies for CGO (required for go-sqlite3)
   RUN apk add --no-cache gcc musl-dev

   WORKDIR /app

   # Copy go mod files first for caching
   COPY go.mod go.sum ./
   RUN go mod download

   # Copy source code
   COPY . .

   # Build arguments for version info
   ARG VERSION=dev
   ARG COMMIT=none
   ARG DATE=unknown

   # Build static binary with CGO enabled for SQLite
   RUN CGO_ENABLED=1 go build \
       -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
       -o replica ./cmd/replica

   # Runtime stage
   FROM alpine:latest

   # Install runtime dependencies for SQLite
   RUN apk add --no-cache ca-certificates

   WORKDIR /app

   # Copy binary from builder
   COPY --from=builder /app/replica .

   # Create non-root user
   RUN adduser -D -H replica
   USER replica

   ENTRYPOINT ["./replica"]
   ```

2. **Create docker-compose.yml**:
   ```yaml
   version: "3.8"

   services:
     # MariaDB for development/testing
     mariadb:
       image: mariadb:11.8
       environment:
         MYSQL_ROOT_PASSWORD: replica_root
         MYSQL_DATABASE: replica
         MYSQL_USER: replica
         MYSQL_PASSWORD: replica_pass
       ports:
         - "3306:3306"
       volumes:
         - mariadb_data:/var/lib/mysql
       healthcheck:
         test: ["CMD", "healthcheck.sh", "--connect", "--innodb_initialized"]
         interval: 10s
         timeout: 5s
         retries: 5

     # PostgreSQL for development/testing
     postgres:
       image: postgres:18
       environment:
         POSTGRES_DB: replica
         POSTGRES_USER: replica
         POSTGRES_PASSWORD: replica_pass
       ports:
         - "5432:5432"
       volumes:
         - postgres_data:/var/lib/postgresql/data
       healthcheck:
         test: ["CMD-SHELL", "pg_isready -U replica -d replica"]
         interval: 10s
         timeout: 5s
         retries: 5

   volumes:
     mariadb_data:
     postgres_data:
   ```

3. **Update Makefile** - add docker target:
   ```makefile
   ## docker: Build Docker image
   docker:
   	docker build \
   		--build-arg VERSION=$(VERSION) \
   		--build-arg COMMIT=$(COMMIT) \
   		--build-arg DATE=$(DATE) \
   		-t replica:$(VERSION) \
   		-t replica:latest \
   		.
   ```

4. **Verify**:
   ```bash
   make docker
   docker run --rm replica:latest version
   docker-compose up -d mariadb postgres
   docker-compose ps
   docker-compose down
   ```

</details>
