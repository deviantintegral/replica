---
id: 4
version: 0.4.0
summary: "Asset Management - Object store integration, asset service with versioning, and image transformations"
created: 2025-12-19
depends_on: [0.3.0]
---

# Release 0.4.0: Asset Management

## Release Overview

This release adds complete binary asset management. After this release, users can upload, version, and transform media files. Assets are stored in pluggable object stores (S3, GCS, Azure, local filesystem) with full version history, enabling rollback and efficient sync.

## Prerequisites

- Release 0.3.0 (Content Management) completed

## Scope

### Included Tasks

| Task | Name | Description |
|------|------|-------------|
| 11 | Object Store Integration | Portable blob storage via gocloud.dev/blob |
| 12 | Asset Service | Asset CRUD, versioning, metadata management |
| 13 | Asset Transformations | Image variants via vipsgen (libvips) |

### Not Included

- API endpoints (0.5.0)
- Background job queue (0.6.0) - transformations run synchronously or on-demand
- Sync functionality (0.7.0)

## Architecture Additions

```
internal/
├── storage/
│   ├── blob.go            # gocloud.dev/blob abstraction
│   ├── s3.go              # S3-specific configuration
│   ├── gcs.go             # GCS-specific configuration
│   ├── azure.go           # Azure-specific configuration
│   ├── local.go           # Local filesystem
│   └── memory.go          # In-memory for testing
├── asset/
│   ├── service.go         # Asset business logic
│   ├── version.go         # Asset versioning
│   ├── metadata.go        # Metadata extraction
│   └── reference.go       # Content-to-asset references
└── transform/
    ├── service.go         # Transformation orchestration
    ├── image.go           # Image transforms with vipsgen
    ├── variant.go         # Variant definitions
    └── cache.go           # Transformed asset caching
```

## Task Details

### Task 11: Object Store Integration

**Objective**: Provide portable blob storage supporting multiple backends

**Supported Backends** (via gocloud.dev/blob):

| Backend | URL Scheme | Use Case |
|---------|------------|----------|
| AWS S3 | `s3://bucket-name` | Production |
| Google Cloud Storage | `gs://bucket-name` | Production |
| Azure Blob Storage | `azblob://container` | Production |
| Local Filesystem | `file:///path` | Development |
| In-Memory | `mem://` | Testing |

**Storage Interface**:

```go
type BlobStore interface {
    Upload(ctx context.Context, key string, data io.Reader, opts UploadOptions) error
    Download(ctx context.Context, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
    List(ctx context.Context, prefix string) ([]BlobInfo, error)
}

type UploadOptions struct {
    ContentType     string
    CacheControl    string
    Metadata        map[string]string
}
```

**Configuration**:
```yaml
storage:
  backend: s3://my-replica-assets
  # or: file:///var/lib/replica/assets
  # or: mem:// (for testing)

  s3:
    region: us-east-1
    # Credentials from environment or IAM role

  local:
    create_dirs: true
```

**Deliverables**:
- gocloud.dev/blob integration
- Configuration for each backend
- Signed URL generation for direct access
- Upload with content-type detection
- Semaphore-limited concurrent operations

**Acceptance Criteria**:
- [ ] All five backends work correctly
- [ ] Signed URLs enable direct client access
- [ ] Content-type is properly detected/set
- [ ] Large files stream without memory exhaustion
- [ ] Concurrent uploads are bounded by semaphore

### Task 12: Asset Service

**Objective**: Manage binary assets with full version history

**Asset Model**:

```go
type Asset struct {
    ID           uuid.UUID
    Filename     string
    ContentType  string
    Size         int64
    Checksum     string    // SHA-256
    StorageKey   string    // Object store key
    Metadata     AssetMetadata
    Versions     []AssetVersion
    RevisionPos  uuid.UUID // Content revision that added/changed this
    CreatedAt    time.Time
    UpdatedAt    time.Time
    CreatedBy    uuid.UUID
}

type AssetVersion struct {
    ID          uuid.UUID
    AssetID     uuid.UUID
    Version     int
    StorageKey  string
    Size        int64
    Checksum    string
    CreatedAt   time.Time
    CreatedBy   uuid.UUID
}

type AssetMetadata struct {
    Width       *int      // For images
    Height      *int      // For images
    Duration    *float64  // For audio/video
    ExifData    map[string]any
}
```

**Asset Operations**:

```go
type AssetService interface {
    // Upload
    Upload(ctx context.Context, filename string, data io.Reader) (*Asset, error)

    // Version management
    Get(ctx context.Context, id uuid.UUID) (*Asset, error)
    GetVersion(ctx context.Context, id uuid.UUID, version int) (*AssetVersion, error)
    ListVersions(ctx context.Context, id uuid.UUID) ([]AssetVersion, error)
    RestoreVersion(ctx context.Context, id uuid.UUID, version int) error

    // Download
    Download(ctx context.Context, id uuid.UUID) (io.ReadCloser, error)
    GetURL(ctx context.Context, id uuid.UUID, ttl time.Duration) (string, error)

    // Deletion
    Delete(ctx context.Context, id uuid.UUID, hard bool) error

    // Reference tracking
    GetReferences(ctx context.Context, id uuid.UUID) ([]ContentReference, error)
}
```

**Sync Optimization**:
- Each asset tracks `RevisionPos` (like CouchDB's `revpos`)
- During sync, only transfer assets where `RevisionPos > lastSyncedRevision`
- Avoids re-transferring unchanged assets

**Deliverables**:
- Asset upload with metadata extraction
- Version history with rollback
- SHA-256 checksum for integrity
- Content-to-asset reference tracking
- Soft/hard delete with retention
- Revision position tracking for sync optimization

**Acceptance Criteria**:
- [ ] Assets upload and store correctly
- [ ] Every upload creates a new version
- [ ] Previous versions are retrievable
- [ ] Rollback restores previous version as current
- [ ] Checksums verify integrity
- [ ] References prevent orphaned assets

### Task 13: Asset Transformations

**Objective**: Generate image variants (thumbnails, resizes) efficiently

**Transformation Framework**:

```go
type TransformService interface {
    GetVariant(ctx context.Context, assetID uuid.UUID, variant string) (*Asset, error)
    GenerateVariant(ctx context.Context, assetID uuid.UUID, variant string) (*Asset, error)
    ListVariants(ctx context.Context, assetID uuid.UUID) ([]VariantInfo, error)
}

type VariantDefinition struct {
    Name        string
    Width       *int
    Height      *int
    Fit         string  // "cover", "contain", "fill", "inside", "outside"
    Format      string  // "jpeg", "webp", "png", "avif"
    Quality     int     // 1-100
}
```

**Default Variants**:
```yaml
transforms:
  variants:
    thumbnail:
      width: 150
      height: 150
      fit: cover
      format: webp
      quality: 80
    medium:
      width: 800
      height: 600
      fit: inside
      format: webp
      quality: 85
    large:
      width: 1920
      height: 1080
      fit: inside
      format: jpeg
      quality: 90
```

**Processing Strategy**:
- **Background**: Generate variants after upload (when background jobs available in 0.6.0)
- **On-Demand**: If variant doesn't exist, generate synchronously and cache
- **Semaphore**: Limit concurrent transforms to prevent resource exhaustion

**vipsgen Integration**:
```go
// Using cshum/vipsgen for high-performance image processing
import "github.com/cshum/vipsgen/vips"

func (s *TransformService) resize(input io.Reader, opts VariantDefinition) (io.Reader, error) {
    img, err := vips.NewImageFromReader(input)
    if err != nil {
        return nil, err
    }
    defer img.Close()

    // Apply transformations
    if opts.Width != nil || opts.Height != nil {
        img.Resize(opts.Width, opts.Height, opts.Fit)
    }

    // Export to format
    return img.Export(opts.Format, opts.Quality)
}
```

**Deliverables**:
- vipsgen/libvips integration
- Configurable variant definitions
- On-demand generation with caching
- Variant storage in object store
- Semaphore-limited concurrent transforms

**Acceptance Criteria**:
- [ ] Image variants generate correctly
- [ ] WebP/AVIF formats work
- [ ] On-demand generation caches results
- [ ] Large images don't cause OOM
- [ ] Concurrent transforms are bounded
- [ ] Variants sync optionally (flag to regenerate vs sync)

## Docker Image Updates

The Docker image must include libvips:

```dockerfile
FROM alpine:3.20

# Install libvips for image processing
RUN apk add --no-cache \
    vips \
    vips-dev \
    ca-certificates

COPY replica /usr/local/bin/replica

ENTRYPOINT ["replica"]
```

## Database Migrations

This release adds:
- `assets` - Asset metadata and current version
- `asset_versions` - Asset version history
- `asset_variants` - Generated variant references
- `content_assets` - Content-to-asset references

## Dependencies

| Dependency | From Release |
|------------|--------------|
| Content service | 0.3.0 |
| Configuration system | 0.1.0 |
| Database abstraction | 0.1.0 |

## Success Criteria for 0.4.0

1. **Object Store**: All backends work correctly (S3, GCS, Azure, local, memory)
2. **Asset Upload**: Files upload with metadata extraction
3. **Versioning**: Full version history with rollback capability
4. **Transforms**: Image variants generate correctly
5. **Performance**: Large files stream efficiently, transforms are bounded

## Release Checklist

- [ ] All tasks completed and tested
- [ ] Docker image includes libvips and works
- [ ] All five object store backends tested
- [ ] Large file upload tested (>1GB)
- [ ] Transform performance benchmarked
- [ ] Memory usage profiled during transforms
- [ ] CI pipeline green
- [ ] Release notes drafted
- [ ] Git tag created: v0.4.0
