---
id: 3
version: 0.3.0
summary: "Content Management - Content service, schema versioning, field validation, search, caching, and multilingual support"
created: 2025-12-19
depends_on: [0.2.0]
---

# Release 0.3.0: Content Management

## Release Overview

This release delivers the complete content management layer. After this release, users can define content types with semantic versioning, create and manage content with CEL-based field validation, search content using database-native full-text search, and manage multilingual content with translation links.

## Prerequisites

- Release 0.2.0 (Data Models & Authentication) completed

## Scope

### Included Tasks

| Task | Name | Description |
|------|------|-------------|
| 08 | Content Service Core | CRUD operations, revision management, workflow states |
| 09 | Schema Versioning | Semantic versioning for content types, migrations, version selection |
| 10 | Field Validation with CEL | Expression-based validation, built-in validators |
| 14 | Full-Text Search | Database-native FTS for PostgreSQL, MySQL, SQLite |
| 25 | Content Caching | LRU in-memory cache with invalidation |
| 29 | Multilingual Support | Translation links, language fallback chains |

### Not Included

- Asset management (0.4.0)
- API endpoints (0.5.0)
- Background processing (0.6.0)
- Sync functionality (0.7.0)

## Architecture Additions

```
internal/
├── content/
│   ├── service.go         # Content business logic
│   ├── revision.go        # Revision management
│   ├── workflow.go        # Workflow state machine
│   └── deletion.go        # Soft/hard delete handling
├── schema/
│   ├── version.go         # Semantic versioning
│   ├── migration.go       # Schema migrations
│   ├── transform.go       # Content transformation between versions
│   └── field_types.go     # Field type definitions
├── validation/
│   ├── cel.go             # CEL expression evaluation
│   ├── builtin.go         # Built-in validators
│   └── sanitize.go        # HTML/Markdown sanitization
├── search/
│   ├── service.go         # Search abstraction
│   ├── postgres.go        # PostgreSQL FTS
│   ├── mysql.go           # MySQL FULLTEXT
│   └── sqlite.go          # SQLite FTS5
├── cache/
│   ├── lru.go             # LRU cache implementation
│   └── invalidation.go    # Cache invalidation logic
└── i18n/
    ├── translation.go     # Translation link management
    └── fallback.go        # Language fallback chains
```

## Task Details

### Task 08: Content Service Core

**Objective**: Implement complete content lifecycle management

**Content Operations**:

```go
type ContentService interface {
    // CRUD
    Create(ctx context.Context, content *Content) error
    Get(ctx context.Context, id uuid.UUID) (*Content, error)
    Update(ctx context.Context, content *Content) error
    Delete(ctx context.Context, id uuid.UUID, hard bool) error

    // Queries
    List(ctx context.Context, query ContentQuery) (*ContentPage, error)
    GetByType(ctx context.Context, contentType string, query ContentQuery) (*ContentPage, error)

    // Revisions
    GetRevision(ctx context.Context, revisionID uuid.UUID) (*Revision, error)
    ListRevisions(ctx context.Context, contentID uuid.UUID) ([]*Revision, error)
    RestoreRevision(ctx context.Context, contentID, revisionID uuid.UUID) error

    // Workflow
    TransitionState(ctx context.Context, id uuid.UUID, newState string) error
    GetWorkflowStates(ctx context.Context, contentType string) ([]string, error)
}
```

**Workflow States**:
- Configurable per content type
- Default states: `draft`, `review`, `published`, `archived`
- State transition rules enforced

**Deletion Modes**:
- Soft delete: Content marked deleted, retained for sync
- Hard delete: Permanent removal after retention period
- Cascade options: block, cascade, nullify references

**Deliverables**:
- Content CRUD with revision tracking
- Optimistic concurrency control (version checks)
- Workflow state machine
- Soft/hard delete with retention
- Reference integrity enforcement

**Acceptance Criteria**:
- [ ] Content CRUD operations work correctly
- [ ] Every update creates a new revision
- [ ] Optimistic locking rejects stale updates
- [ ] Workflow transitions enforce valid paths
- [ ] Soft delete preserves content for sync

### Task 09: Schema Versioning

**Objective**: Implement semantic versioning for content type schemas

**Version Model**:

```go
type SchemaVersion struct {
    ContentType string
    Version     semver.Version  // MAJOR.MINOR.PATCH
    Fields      []FieldDefinition
    Migration   *Migration
    Changelog   string
    CreatedAt   time.Time
}

type Migration struct {
    FieldsAdded    []FieldDefinition
    FieldsRemoved  []string
    FieldsChanged  []FieldChange
    Transformations []Transformation
}
```

**Version Bump Rules**:
- **MAJOR**: Breaking changes (field removal, type changes)
- **MINOR**: Additive changes (new optional fields)
- **PATCH**: Metadata only (descriptions, relaxed validation)

**Version Selection**:
- Header: `Accept: application/vnd.api+json; schema-version=1.2`
- Query: `?schemaVersion=1.2`
- Per-item in bulk: `meta.schemaVersion`

**Deliverables**:
- Schema version storage and retrieval
- Automatic version bump detection
- Content transformation between versions
- Version discovery API data
- Minimum version enforcement

**Acceptance Criteria**:
- [ ] Schema versions follow semantic versioning
- [ ] Content transforms correctly between versions
- [ ] Version discovery returns all available versions
- [ ] Minimum version blocks old version requests
- [ ] Breaking changes require MAJOR bump

### Task 10: Field Validation with CEL

**Objective**: Provide expressive field validation using Common Expression Language

**Validation Types**:

```go
// Built-in validators (shortcuts)
type BuiltinValidation struct {
    Required  bool
    MinLength *int
    MaxLength *int
    Min       *float64
    Max       *float64
    Pattern   *string  // regex
    Enum      []any
}

// CEL expressions for complex validation
type CELValidation struct {
    Expression string   // CEL expression
    Message    string   // Error message
}
```

**Example Expressions**:
```cel
// Cross-field validation
this.endDate > this.startDate

// Array constraints
size(this.tags) <= 10 && size(this.tags) >= 1

// Conditional validation
this.status == "published" ? has(this.publishedAt) : true

// String format
this.email.matches("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$")
```

**Sanitization**:
- HTML: bluemonday policy-based sanitization
- Markdown: goldmark parsing with safe rendering

**Deliverables**:
- CEL expression evaluation engine
- Built-in validators for common cases
- HTML sanitization with bluemonday
- Markdown sanitization with goldmark
- Validation error messages with field paths

**Acceptance Criteria**:
- [ ] Built-in validators work correctly
- [ ] CEL expressions evaluate properly
- [ ] Cross-field validation works
- [ ] HTML is sanitized before storage
- [ ] Markdown is safely rendered to HTML

### Task 14: Full-Text Search

**Objective**: Implement database-native full-text search across all backends

**Search Abstraction**:

```go
type SearchService interface {
    Index(ctx context.Context, content *Content) error
    Remove(ctx context.Context, contentID uuid.UUID) error
    Search(ctx context.Context, query SearchQuery) (*SearchResults, error)
}

type SearchQuery struct {
    Query       string
    ContentType *string
    Fields      []string  // Limit search to specific fields
    Cursor      *string
    Limit       int
}
```

**Database Implementations**:
- **PostgreSQL**: tsvector/tsquery with GIN indexes
- **MySQL**: FULLTEXT indexes with natural language mode
- **SQLite**: FTS5 virtual tables with porter stemmer

**Deliverables**:
- Search abstraction interface
- PostgreSQL FTS implementation
- MySQL FULLTEXT implementation
- SQLite FTS5 implementation
- Automatic index updates on content changes

**Acceptance Criteria**:
- [ ] Search works identically across all databases
- [ ] Results are ranked by relevance
- [ ] Partial word matching works (stemming)
- [ ] Indexes update automatically on content changes
- [ ] Search performance is acceptable (<100ms for 100k items)

### Task 25: Content Caching

**Objective**: Improve read performance with LRU in-memory cache

**Cache Design**:

```go
type ContentCache interface {
    Get(ctx context.Context, id uuid.UUID) (*Content, bool)
    Set(ctx context.Context, content *Content)
    Invalidate(ctx context.Context, id uuid.UUID)
    InvalidateType(ctx context.Context, contentType string)
    Stats() CacheStats
}
```

**Configuration**:
```yaml
cache:
  enabled: true
  max_entries: 10000
  max_size_mb: 100
  ttl: 5m
```

**Invalidation Triggers**:
- Local content updates
- Local content deletes
- Incoming sync operations

**Deliverables**:
- LRU cache implementation
- TTL support
- Size-based eviction
- Cache bypass parameter
- Metrics for hit/miss rates

**Acceptance Criteria**:
- [ ] Cache hit improves response time significantly
- [ ] Cache invalidates on writes
- [ ] Cache respects size limits
- [ ] Bypass parameter works
- [ ] Metrics are exposed correctly

### Task 29: Multilingual Support

**Objective**: Support content translations through linked items

**Translation Model**:

```go
type Translation struct {
    ContentID   uuid.UUID  // Original content
    Language    string     // e.g., "en-US", "fr-FR"
    TranslatedID uuid.UUID // Translated content item
}
```

**Fallback Chain**:
```yaml
i18n:
  default_language: en-US
  fallback_chains:
    fr-CA: [fr-FR, en-US]
    de-AT: [de-DE, en-US]
```

**Deliverables**:
- Translation link management
- Language fallback resolution
- Fetch with translations included
- Sync includes linked translations

**Acceptance Criteria**:
- [ ] Translations link correctly to source content
- [ ] Fallback chains resolve properly
- [ ] API can return content with all translations
- [ ] Missing translations fall back correctly
- [ ] Translations sync as dependencies

## Database Migrations

This release adds/modifies:
- `content_types` - Add version tracking columns
- `schema_versions` - Schema version history
- `contents` - Add workflow state, deleted_at
- `content_translations` - Translation links
- `search_indexes` - FTS configuration per database

## Dependencies

| Dependency | From Release |
|------------|--------------|
| Core domain models | 0.2.0 |
| RBAC middleware | 0.2.0 |
| Rate limiting | 0.2.0 |

## Success Criteria for 0.3.0

1. **Content CRUD**: Create, read, update, delete with full revision history
2. **Schema Versions**: Content types evolve with semantic versioning
3. **Validation**: CEL expressions validate complex rules correctly
4. **Search**: Full-text search works across all database backends
5. **Caching**: Read performance improved with proper invalidation
6. **Translations**: Multilingual content with fallback chains

## Release Checklist

- [ ] All tasks completed and tested
- [ ] CEL security review (no unsafe operations)
- [ ] Search performance benchmarked
- [ ] Cache hit rates verified
- [ ] Translation fallback tested
- [ ] CI pipeline green
- [ ] Release notes drafted
- [ ] Git tag created: v0.3.0
