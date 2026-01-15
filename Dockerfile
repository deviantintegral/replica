# syntax=docker/dockerfile:1

# Build stage: Compile the Go binary with CGO support for SQLite
FROM golang:1.25-alpine AS builder

# Install build dependencies for CGO and SQLite
RUN apk add --no-cache \
    gcc \
    musl-dev \
    sqlite-dev

# Set working directory
WORKDIR /src

# Copy go module files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments for version info
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# Build the binary with CGO enabled and version info
ENV CGO_ENABLED=1
RUN go build -mod=readonly \
    -ldflags "-X github.com/deviantintegral/replica/internal/version.Version=${VERSION} \
              -X github.com/deviantintegral/replica/internal/version.Commit=${COMMIT} \
              -X github.com/deviantintegral/replica/internal/version.BuildDate=${BUILD_DATE}" \
    -o /replica ./cmd/replica

# Runtime stage: Minimal Alpine image
FROM alpine:3.23

# Install runtime dependencies
RUN apk add --no-cache \
    sqlite-libs \
    ca-certificates

# Create non-root user for security
RUN addgroup -g 1000 replica && \
    adduser -u 1000 -G replica -s /bin/sh -D replica

# Create data directory for SQLite database
RUN mkdir -p /data && chown replica:replica /data

# Copy binary from builder
COPY --from=builder /replica /usr/local/bin/replica

# Switch to non-root user
USER replica

# Set working directory
WORKDIR /data

# Expose HTTP port
EXPOSE 8080

# Default command
ENTRYPOINT ["replica"]
CMD ["serve"]
