---
id: 1
summary: "Implement Replica 1.0 - a distributed Golang CMS with bidirectional content sync, schema versioning, and multi-instance federation"
created: 2025-12-18
---

# Plan: Replica 1.0 - Distributed Content Management System

## Original Work Order

> Define the scope and implementation for a 1.0 release of "Replica", a golang CMS focused on a distributed content model. The CMS allows for each instance to push or pull content from other Replica instances. This implies many features, such as having UUIDs for all objects, metadata over APIs to expose content type definitions, and versions not just for content but for the content schema as well. For example, a user should be able to pull content updates for some or all content from a production instance to a test instance. All schema updates should ship with upgrade and downgrade paths over the API, so consuming sites can migrate content to match different schemas. This should also enable Blue / Green deployments of equal production environments, as well as larger content distribution networks between independent sites.

## Plan Clarifications

| Question | Answer |
|----------|--------|
| Authentication model for inter-instance communication? | Dual support: API Keys for simple setups, OAuth 2.0/JWT for federated environments |
| Content conflict resolution strategy? | Version branching with manual or AI-assisted merge resolution |
| Content types for 1.0? | Full CMS: structured content, media, users, permissions. Binary files via pluggable object stores (S3 default, local for dev) |
| Schema migration approval? | Auto-migrate unless data loss possible, then require explicit approval |
| Sync operation model? | Bidirectional push/pull between peers |
| Sync granularity? | Query-based supporting full instance, content type, or individual item selection |
| Admin interface? | API-only with CLI tool for administration |
| Observability? | Basic logging, Prometheus/OpenTelemetry metrics, distributed tracing |
| Version sync depth? | Full history by default, configurable pruning by age or revision count |
| Database migrations? | golang-migrate for multi-database support |
| Large content sync? | Delta sync with chunked transfers and resume capability |
| Content relationship sync? | Automatically include referenced content (dependencies) in sync operations |
| User/permission model? | RBAC with custom roles and granular permissions |
| Binary asset versioning? | Full version history for all binary files, enabling rollback |
| Field type complexity? | Rich types: string, number, boolean, date, reference, media, plus rich text, JSON blob, geolocation, computed fields |
| Circular reference handling? | Detect cycles and include all items in the cycle in one sync batch |
| RBAC sync scope? | Role definitions sync between instances, but user assignments are instance-local |
| Rate limiting? | Storage and API quotas per instance/user with enforcement |
| Computed fields calculation? | On-write: calculated when content is saved, stored and indexable |
| Content deletion? | Configurable: soft delete by default, optional hard delete after retention period |
| Workflow states? | Custom configurable workflow states per content type (draft, review, published, etc.) |
| Instance identification? | UUID generated at first boot + admin-configured friendly name mapping to UUID |
| Full-text search? | Database-native search (PostgreSQL FTS, SQLite FTS5, MySQL FULLTEXT) |
| Scheduled publishing? | Content can be scheduled for future publication |
| External integrations? | Configurable webhooks triggered on content and sync events |
| Multilingual support? | Content items can have translations linked by UUID |
| Concurrent edit handling? | Optimistic locking with version checks; optional AI-assisted conflict resolution |
| Asset transformations? | Background variant generation with on-demand fallback for missing variants |
| API pagination? | Cursor-based pagination for efficient deep pagination |
| Field validation? | Expression language (CEL or similar) for complex validation rules |
| Configuration format? | YAML configuration files with environment variable overrides |
| Sync failure recovery? | Idempotent retry; partial syncs are consistent, dependencies written in order, co-dependencies atomic |
| Audit logging? | Immutable audit log of all content changes, access, and sync operations |
| New instance bootstrap? | Start empty; pull content from peers as needed |
| Migration tooling? | API only for data movement; no file-based import in 1.0 |
| API versioning? | URL path versioning (e.g., /v1/content) |
| Health checks? | Simple /health endpoint for container orchestration |
| TLS handling? | Terminate at reverse proxy/load balancer; Replica speaks HTTP internally |
| Deployment artifacts? | Static binary plus official Docker image |
| API documentation? | Auto-generated OpenAPI/Swagger specification from code |
| Content type management? | API only; content types created and modified exclusively through API |
| Graceful shutdown? | Interrupt syncs at next safe point; rely on idempotent retry for resume |
| MVP scope for 1.0? | All 32 architectural components are required for the core distributed CMS vision |
| Local user authentication? | CLI creates initial admin user; subsequent users created via API |
| Testing strategy? | Unit tests with code coverage, functional tests for CLI, mutation testing with gremlins |
| Instance bootstrap? | CLI init command creates admin user and optional starter content types |
| Starter content types? | Optional templates (Article, Page, Media) available via CLI init |
| Mutation testing framework? | gremlins for mutation testing |
| Extension model? | HTTP webhook-based; subscribers can intercept and modify content during CRUD operations |
| Webhook priority? | Subscribers specify advisory priority: EARLY, NEUTRAL, or LATE for call ordering |
| Webhook subscription management? | API only; subscriptions created and managed via REST endpoints |
| Webhook failure handling? | Configurable per-subscriber: 'required' (abort on fail) or 'optional' (skip on fail) |
| CLI authentication? | Login flow; CLI has 'login' command that stores JWT token locally |
| Peer instance configuration? | API registration; peers registered via API with URL, name, and credentials |
| Background job processing? | Database queue; jobs stored in database, any instance can pick up work |
| Content caching? | In-memory cache per instance; not shared between instances |

## Executive Summary

Replica is a distributed Content Management System built in Go that enables bidirectional content synchronization between independent instances. The system addresses the need for flexible content distribution across environments (production-to-staging, blue-green deployments) and organizations (content distribution networks between partner sites).

The architectural approach centers on treating content and schema as versioned, UUID-identified entities that can flow between instances while maintaining integrity and traceability. Each instance operates autonomously but can federate with peers through a standardized JSON:API interface. The version branching model for conflict resolution, combined with pluggable authentication (API Keys + OAuth 2.0/JWT), enables both simple internal deployments and complex multi-organization federations.

Key differentiators include schema version management with automatic upgrade/downgrade paths, query-based selective sync, and AI-assisted conflict resolution. The system supports multiple database backends (SQLite, MySQL/MariaDB, PostgreSQL) and delegates binary asset storage to pluggable object stores, optimizing for the performance characteristics of distributed storage systems.

## Context

### Current State vs Target State

| Current State | Target State | Why? |
|---------------|--------------|------|
| No codebase exists | Fully functional distributed CMS | Greenfield implementation of core vision |
| No content model | UUID-based content with full version history | Enable unique identification and history tracking across instances |
| No schema management | Versioned schemas with upgrade/downgrade migrations | Allow consuming sites to adapt content to different schema versions |
| No sync capability | Bidirectional delta sync with chunked transfers | Enable efficient content distribution between instances |
| No conflict handling | Version branching with manual/AI merge | Support concurrent modifications across distributed instances |
| No multi-database support | SQLite, MySQL/MariaDB, PostgreSQL backends | Support diverse deployment environments |
| No observability | Logging, metrics, distributed tracing | Enable operational visibility in distributed deployments |
| No CLI tooling | Full CLI for administration and sync operations | Provide operators with efficient management interface |

### Background

Distributed content management presents unique challenges that traditional CMS architectures don't address. Organizations need to:

1. **Replicate content across environments** - Pull production content to staging/test without manual export/import
2. **Deploy with zero downtime** - Blue/green deployments require synchronized content across parallel production environments
3. **Federate content across organizations** - Content distribution networks where independent sites share subsets of content

The JSON:API specification provides a standardized interface for content operations. Combined with UUID-based identification and vector clock versioning, this enables reliable content synchronization without central coordination.

The pluggable object store approach for binary assets acknowledges that file distribution has fundamentally different performance characteristics than metadata synchronization. S3-compatible storage provides the scalability needed for media-heavy deployments while local storage supports development workflows.

## Architectural Approach

```mermaid
graph TB
    subgraph "Replica Instance"
        API[JSON:API Server]
        CLI[CLI Tool]
        Auth[Auth Layer<br/>API Keys + JWT]
        RBAC[RBAC Service]
        QUOTA[Quota Service]

        subgraph "Core Services"
            CS[Content Service]
            SS[Schema Service]
            SY[Sync Service]
            VS[Version Service]
            AS[Asset Service]
            WF[Workflow Service]
            SR[Search Service]
        end

        subgraph "Background Jobs"
            SCH[Scheduler]
            WH[Webhook Dispatcher]
            ATR[Asset Transformer]
        end

        subgraph "Storage Layer"
            DB[(Database<br/>SQLite/MySQL/PG)]
            OBJ[Object Store<br/>S3/Local]
        end

        subgraph "Observability"
            LOG[Structured Logging]
            MET[Metrics Export]
            TRC[Distributed Tracing]
            AUD[Audit Log]
        end
    end

    CLI --> API
    API --> Auth
    Auth --> RBAC
    RBAC --> QUOTA
    QUOTA --> CS & SS & SY
    CS --> VS
    CS --> WF
    SS --> VS
    CS --> AS
    CS --> SR
    AS --> OBJ
    SY --> CS & SS
    CS & SS & SR --> DB
    AS --> DB
    SCH --> WF
    WH -.->|HTTP| EXT2[External Systems]
    CS -.->|Events| WH
    AS -.->|Queue| ATR
    ATR --> OBJ
    CS -.->|Events| AUD
    SY -.->|Events| AUD

    API -.->|Federation| EXT[Other Replica Instances]
```

### Core Data Model
**Objective**: Establish UUID-based entities with comprehensive version tracking that enables distributed synchronization

All entities use UUIDs as primary identifiers, enabling conflict-free creation across instances. The content model separates:

- **Content Items**: The actual content data with UUID, content type reference, and field values
- **Content Types**: Schema definitions describing fields, validation rules, and relationships
- **Revisions**: Immutable snapshots of content state with parent revision references forming a DAG
- **Schema Versions**: Versioned content type definitions with migration paths
- **Assets**: Binary files with full version history, stored in object stores with metadata in the database

**Supported Field Types:**
- **Basic**: String, number (integer/float), boolean, date/datetime
- **References**: Links to other content items (with referential integrity)
- **Media**: References to versioned binary assets
- **Rich**: Rich text (HTML/Markdown with sanitization), JSON blob (schema-validated)
- **Specialized**: Geolocation (lat/lng with optional address), computed fields (calculated on-write, stored and indexable)

Vector clocks accompany revisions to enable proper ordering and conflict detection without centralized coordination. Each instance maintains its logical clock, incremented on local modifications.

### Schema Version Management
**Objective**: Enable content type evolution with bidirectional migration support

Content types evolve over time. Replica treats schema changes as first-class versioned entities:

- Each content type maintains a version history
- Schema versions include structured migration definitions (not raw SQL)
- Migrations specify field additions, removals, transformations, and validation changes
- Downgrade paths enable syncing content to instances running older schema versions
- The system auto-migrates additive changes (new optional fields) but requires approval for potentially destructive changes (field removal, type changes)

Migration definitions use a declarative format describing the transformation, enabling the system to generate appropriate SQL for each database backend.

### Content Type Management
**Objective**: Provide API-driven content type definition and evolution

Content types are managed exclusively through the API:

- **API Operations**: Create, update, delete content types via REST endpoints
- **No Config Files**: Content types not defined in YAML; schema is data, not configuration
- **Sync as Content**: Content type definitions sync between instances like any other content
- **Schema Validation**: API validates content type changes; prevents invalid field configurations
- **Migration Generation**: Schema changes automatically generate migration definitions

This approach enables content type management through the same sync mechanisms as content itself.

### Authentication and Authorization
**Objective**: Support both simple internal deployments and complex federated environments

Dual authentication strategy:

- **API Keys**: Shared secrets for trusted internal networks, simple setup for development and single-organization deployments
- **OAuth 2.0 / JWT**: Token-based authentication with expiration for federated environments, supporting instance-to-instance authorization with scoped permissions

Authorization operates at multiple levels:
- Instance-level: Which remote instances can connect
- Content type-level: Which content types can be synced
- Operation-level: Push, pull, or both

### Role-Based Access Control (RBAC)
**Objective**: Provide granular permission management for local content operations

The RBAC system manages user access to content and operations:

- **Custom Roles**: Define roles with specific permission sets (beyond basic admin/editor/viewer)
- **Granular Permissions**: Permissions scoped to content types, operations (create/read/update/delete), and field-level access
- **Role Definitions Sync**: Role templates can sync between instances for consistency
- **Instance-Local Assignments**: User-to-role assignments remain instance-specific (not synced)

This separation allows organizations to maintain consistent role definitions across environments while managing users independently per instance.

### Quota and Rate Limiting
**Objective**: Protect system resources and enforce usage limits

The quota system enforces limits at multiple levels:

- **API Rate Limits**: Per-client request limits to prevent abuse
- **Storage Quotas**: Limits on content items, revisions, and asset storage per user or instance
- **Sync Quotas**: Limits on sync operation frequency and data volume
- **Enforcement**: Requests exceeding quotas receive appropriate HTTP status codes with retry guidance

### Sync Engine
**Objective**: Enable efficient bidirectional content synchronization with conflict detection

```mermaid
sequenceDiagram
    participant A as Instance A
    participant B as Instance B

    Note over A,B: Pull Operation
    A->>B: Request sync (query filter, last sync vector)
    B->>B: Identify delta (changed since vector)
    B->>A: Stream chunks (content + revisions + assets refs)
    A->>A: Detect conflicts (divergent branches)
    A->>A: Apply non-conflicting changes
    A-->>A: Queue conflicts for resolution

    Note over A,B: Push Operation
    A->>A: Identify local changes for push
    A->>B: Propose changes (content + revisions)
    B->>B: Validate against schema
    B->>B: Check authorization
    B->>A: Accept/reject response
```

Sync operations support:
- **Query-based selection**: Filter by content type, individual UUIDs, tags, date ranges, or custom queries
- **Delta sync**: Only transfer content modified since last synchronization
- **Chunked transfer**: Large sync operations broken into resumable chunks
- **Bidirectional flow**: Both push and pull supported between any connected instances
- **Dependency resolution**: Automatically include referenced content items in sync operations
- **Circular reference handling**: Detect reference cycles and include all items in the cycle as a single atomic batch

The sync protocol transmits content metadata and revision history. Binary assets are referenced by their object store URLs; the receiving instance can fetch assets from the origin's object store or trigger replication to its own store.

### Conflict Resolution
**Objective**: Handle concurrent modifications across distributed instances gracefully

When the same content is modified on multiple instances before sync:

1. **Detection**: Vector clock comparison identifies divergent revision branches
2. **Branching**: Both versions are preserved as branches of the revision DAG
3. **Resolution options**:
   - Manual: Present both versions to user for merge decision
   - AI-assisted: Invoke configured AI service to propose merge
   - Automated rules: Apply configured merge strategies per content type

Resolved conflicts create a new revision with both branches as parents, preserving full history.

### Object Store Integration
**Objective**: Handle binary assets efficiently with pluggable storage backends and full version history

Binary files (media, documents) are stored in object stores rather than the database:

- **S3-compatible**: Default production backend supporting AWS S3, MinIO, and compatible services
- **Local filesystem**: Development and testing backend

**Asset Versioning:**
- Each asset maintains full version history similar to content items
- Asset versions are immutable; updates create new versions
- Version history enables rollback to previous asset states
- Pruning policies can limit retained versions by count or age

The content database stores object store references (bucket, key, version, metadata). Asset sync between instances can operate in multiple modes:
- Reference-only: Receiving instance accesses assets from origin's store
- Copy-on-sync: Assets replicated to receiving instance's store
- Lazy-copy: Assets copied on first local access

### CLI Tool
**Objective**: Provide operators with efficient command-line administration

The CLI tool provides administrative operations:
- Instance configuration and peer registration
- Sync operations (push/pull with query filters)
- Schema management (list versions, preview migrations)
- Conflict resolution queue management
- Content pruning operations
- Health and status checks

**Authentication:**
- **Login Command**: `replica login` authenticates with username/password and stores JWT token locally
- **Token Storage**: JWT stored in user's config directory (~/.config/replica or platform equivalent)
- **Token Refresh**: Automatic refresh when token approaches expiration
- **Logout**: `replica logout` removes stored credentials

Commands output structured data (JSON) for scripting integration while providing human-readable formatting for interactive use.

### Observability Stack
**Objective**: Enable operational visibility across distributed deployments

Three-pillar observability:

- **Structured Logging**: JSON-formatted logs with correlation IDs that span cross-instance operations
- **Metrics**: Prometheus-compatible metrics exposing sync operations, latency, error rates, queue depths
- **Distributed Tracing**: OpenTelemetry integration for tracing requests across instance boundaries

Correlation IDs propagate through sync operations, enabling operators to trace a content item's journey across multiple instances.

### Database Abstraction
**Objective**: Support multiple database backends for diverse deployment needs

Using golang-migrate with database-agnostic migration definitions:

- **SQLite**: Default for development and small deployments
- **MySQL/MariaDB**: Production deployments with existing MySQL infrastructure
- **PostgreSQL**: Production deployments preferring PostgreSQL

The data access layer abstracts database-specific SQL, with migrations generating appropriate DDL for each backend.

### Workflow and Publishing
**Objective**: Support configurable content lifecycle states with scheduled publishing

Content items move through configurable workflow states:

- **Custom States**: Each content type defines its own workflow states (e.g., draft → review → published → archived)
- **State Transitions**: Configurable rules for valid state transitions and required permissions
- **Scheduled Publishing**: Content can be scheduled to transition states at specific times (e.g., publish at midnight)
- **Sync Filtering**: Sync operations can filter by workflow state (e.g., only sync published content to production)

A background scheduler processes pending state transitions, with distributed locking to prevent duplicate execution across instances.

### Content Deletion
**Objective**: Handle content removal with configurable retention and sync propagation

Deletion operates in two modes:

- **Soft Delete**: Content marked as deleted but retained; deletion syncs as a tombstone marker
- **Hard Delete**: Permanent removal after configurable retention period
- **Cascade Behavior**: Configurable handling of referenced content (block, cascade, or nullify references)

Tombstones ensure deletions propagate correctly across instances, with configurable retention before permanent removal.

### Full-Text Search
**Objective**: Enable content discovery through database-native full-text search

Search leverages each database's native capabilities:

- **PostgreSQL**: Full-text search with tsvector/tsquery, GIN indexes
- **SQLite**: FTS5 virtual tables with porter stemmer
- **MySQL/MariaDB**: FULLTEXT indexes with natural language mode

The search abstraction provides a unified query interface while using database-specific implementations for optimal performance. Indexed fields are configurable per content type.

### Instance Identity
**Objective**: Uniquely identify instances for federation and sync tracking

Each instance maintains:

- **Instance UUID**: Generated at first boot, immutable, used for vector clocks and sync tracking
- **Instance Name**: Admin-configured friendly identifier (e.g., "production-us-east")
- **Name-to-UUID Mapping**: Remote instances reference by name; resolved to UUID for operations

This allows operators to use meaningful names while maintaining stable UUIDs for distributed system coordination.

### Peer Registration
**Objective**: Enable federation between Replica instances

Peer instances are registered via API:

- **Registration Endpoint**: POST to register new peer with URL, friendly name, and credentials
- **Credential Types**: API key or OAuth client credentials for the remote instance
- **Health Verification**: Optional connectivity check during registration
- **Sync Permissions**: Configure which content types can sync with this peer, and in which direction
- **Peer Management**: List, update, disable, or remove peers via API

Peers are stored in the database and available to all instances in a multi-instance deployment.

### Background Job Queue
**Objective**: Process asynchronous tasks reliably across multiple instances

Jobs are queued in the database for distributed processing:

- **Database-Backed Queue**: Jobs stored as database records; survives instance restarts
- **Worker Polling**: Instances poll for available jobs with distributed locking
- **Job Types**: Scheduled publishing, webhook delivery, asset transformation, pruning
- **Retry Logic**: Failed jobs retry with exponential backoff; dead-letter after max attempts
- **Distributed Locking**: Database-level locks prevent duplicate job execution

This approach enables horizontal scaling without external dependencies like Redis.

### Content Caching
**Objective**: Improve read performance for frequently accessed content

In-memory caching per instance:

- **Cache Scope**: Per-instance cache; not shared between instances
- **Cache Strategy**: LRU eviction with configurable size limits
- **Invalidation**: Cache invalidated on local writes and incoming sync
- **Cache Bypass**: API parameter to bypass cache for fresh data
- **Metrics**: Cache hit/miss rates exposed via metrics endpoint

Caching is optional and can be disabled for memory-constrained deployments.

### Webhook Extension System
**Objective**: Enable external integrations and content modification through HTTP-based extension points

The webhook system serves as the primary extension model, allowing external services to intercept and modify content during CRUD operations:

**Extension Model:**
- **Interceptor Pattern**: Webhooks are called during content operations, not just after
- **Data Modification**: Subscribers can inspect and modify content data in their response
- **Synchronous Flow**: CRUD operations wait for webhook responses before completing
- **Timeout Handling**: Configurable timeouts; operations proceed if subscriber is unresponsive

**Event Types:**
- **Content Events**: `content.creating`, `content.created`, `content.updating`, `content.updated`, `content.deleting`, `content.deleted`
- **Workflow Events**: `workflow.transitioning`, `workflow.transitioned`
- **Sync Events**: `sync.started`, `sync.completed`, `sync.conflict_detected`
- **Schema Events**: `schema.updating`, `schema.updated`

**Subscriber Priority:**
Subscribers specify an advisory priority when registering, influencing call order:
- **EARLY**: Called first; typically for validation, enrichment, or transformation
- **NEUTRAL**: Called in middle; default priority for general processing
- **LATE**: Called last; typically for logging, analytics, or final modifications

Priority is advisory; the system orders subscribers by priority but doesn't guarantee strict ordering within the same priority level.

**Webhook Configuration:**
- **API Management**: Subscriptions created, updated, deleted via REST endpoints
- **Multiple Endpoints**: Multiple subscribers per event type
- **Selective Subscription**: Subscribe to specific content types or all content
- **Failure Mode**: Per-subscriber setting - `required` (abort operation on failure) or `optional` (skip and continue)
- **Payload Signing**: HMAC signatures for webhook authenticity verification
- **Retry Logic**: Failed deliveries retry with exponential backoff (for notification-only events)

```mermaid
sequenceDiagram
    participant C as Client
    participant R as Replica
    participant E as EARLY Subscriber
    participant N as NEUTRAL Subscriber
    participant L as LATE Subscriber

    C->>R: Create Content
    R->>E: content.creating (with data)
    E->>R: Modified data (or unchanged)
    R->>N: content.creating (with data)
    N->>R: Modified data (or unchanged)
    R->>L: content.creating (with data)
    L->>R: Modified data (or unchanged)
    R->>R: Persist final content
    R-->>E: content.created (notification)
    R-->>N: content.created (notification)
    R-->>L: content.created (notification)
    R->>C: Response
```

**Response Contract:**
- Subscribers return modified content in response body to alter the operation
- Return unchanged content or empty response to pass through
- Return error status to abort the operation (for `.creating`, `.updating`, `.deleting` events)
- Post-operation events (`.created`, `.updated`, `.deleted`) are notification-only and fire asynchronously

### Multilingual Content
**Objective**: Support content translation through linked content items

Translation model:

- **Translation Links**: Content items link to translations via UUID reference
- **Language Codes**: Standard language identifiers (e.g., en-US, fr-FR)
- **Fallback Chain**: Configurable language fallback order per instance
- **Sync Behavior**: Translations treated as distinct content items; sync includes all linked translations by default

This approach keeps the content model simple while supporting complex multilingual requirements.

### Optimistic Concurrency Control
**Objective**: Handle concurrent edits gracefully without blocking locks

Edit conflicts are handled through version-based optimistic locking:

- **Version Tracking**: Each content revision has a version identifier
- **Conflict Detection**: Save operations fail if the base version doesn't match current
- **Resolution Options**: Manual resolution or AI-assisted automatic merge
- **Retry Guidance**: Conflict responses include current state for client retry

This approach maximizes concurrency while ensuring data integrity.

### Asset Transformations
**Objective**: Provide derived asset variants (thumbnails, resizes) efficiently

Asset transformation model:

- **Variant Definitions**: Configurable transformation rules per content type (dimensions, format, quality)
- **Background Generation**: Variants generated asynchronously after upload
- **On-Demand Fallback**: If a requested variant doesn't exist, generate synchronously and cache
- **Sync Behavior**: Original assets sync; variants regenerated on destination or synced optionally

Transformations use standard image processing; video transcoding is out of scope for 1.0.

### Cursor-Based Pagination
**Objective**: Enable efficient pagination through large result sets

API pagination uses opaque cursors:

- **Cursor Tokens**: Encode position in result set without exposing internals
- **Stable Results**: Consistent ordering even as content changes
- **Deep Pagination**: Efficient retrieval regardless of offset depth
- **JSON:API Compliance**: Links section includes next/prev cursors per specification

### Field Validation
**Objective**: Enable expressive validation rules for content fields

Validation uses an expression language (CEL - Common Expression Language):

- **Built-in Validators**: Standard rules (required, min/max, regex, enum) as shortcuts
- **Custom Expressions**: CEL expressions for complex cross-field validation
- **Error Messages**: Custom error messages with field references
- **Schema Integration**: Validation rules stored as part of content type schema versions

This allows validation like `size(this.tags) <= 10 && this.publishDate > this.createdAt` without custom code.

### Configuration Management
**Objective**: Provide flexible configuration for diverse deployment environments

Configuration uses YAML files with environment variable support:

- **Primary Config**: YAML file for instance configuration, peer definitions, content types
- **Environment Overrides**: Environment variables override YAML values for deployment-specific settings
- **Secrets Handling**: Sensitive values (API keys, database credentials) via environment variables
- **Hot Reload**: Configuration changes applied without restart where possible

### Sync Reliability
**Objective**: Ensure sync operations are reliable and recoverable

Sync operations are designed for idempotent retry:

- **Idempotent Operations**: Applying the same sync payload multiple times produces identical results
- **Dependency Ordering**: Referenced content written before referencing content to avoid dangling references
- **Co-dependency Atomicity**: Circular reference groups written atomically (all-or-nothing)
- **Resume on Retry**: Failed syncs can be safely retried; already-synced content is recognized and skipped

This design ensures partial syncs leave the system in a consistent state.

### Audit Logging
**Objective**: Provide immutable audit trail for compliance and security

The audit log captures:

- **Content Operations**: Create, update, delete, publish, unpublish with actor and timestamp
- **Access Events**: Content reads by authenticated users (configurable)
- **Sync Operations**: Push/pull operations with source/destination and content summary
- **Permission Changes**: Role and permission modifications
- **Immutability**: Audit records cannot be modified or deleted; retention configurable

Audit logs are stored separately from operational data and can be exported for compliance systems.

### API Versioning
**Objective**: Enable backwards-compatible API evolution

API versioning uses URL path prefixes:

- **Version Format**: `/v1/`, `/v2/`, etc. in URL path
- **Compatibility Policy**: Minor versions add features; major versions may break compatibility
- **Deprecation Process**: Old versions supported for defined period after new version release
- **Version Negotiation**: Sync protocol includes version handshake for cross-instance compatibility

### Health Endpoints
**Objective**: Enable container orchestration and monitoring integration

Health check endpoint:

- **Endpoint**: `GET /health` returns instance health status
- **Response**: JSON with status (healthy/unhealthy) and optional component details
- **Use Case**: Kubernetes liveness probes, load balancer health checks
- **Lightweight**: Minimal processing to avoid impacting system under load

### Transport Security
**Objective**: Secure inter-instance and client communication

TLS is handled at the infrastructure layer:

- **Proxy Termination**: TLS terminated at reverse proxy (nginx, Traefik, cloud load balancer)
- **Internal HTTP**: Replica speaks plain HTTP; assumes trusted network or sidecar proxy
- **Header Trust**: Configurable trusted proxy headers for client IP and protocol detection
- **Recommendation**: Deploy behind TLS-terminating proxy in all non-development environments

### Deployment Model
**Objective**: Provide flexible deployment options for various environments

Deployment artifacts:

- **Static Binary**: Single self-contained executable for Linux, macOS, Windows
- **Docker Image**: Official multi-arch image with minimal base (distroless or Alpine)
- **Configuration**: Environment variables and mounted config files
- **Database**: External database connection; no embedded database in container

The binary includes all assets; no external file dependencies beyond configuration.

### API Documentation
**Objective**: Provide accurate, up-to-date API documentation

Documentation generation:

- **OpenAPI Specification**: Auto-generated from code annotations/structs
- **Swagger UI**: Optional embedded UI for API exploration (disabled by default in production)
- **Versioned Docs**: Each API version has corresponding OpenAPI spec
- **Schema Validation**: Request/response validation against generated schemas

### Graceful Shutdown
**Objective**: Handle process termination without data corruption

Shutdown behavior:

- **Signal Handling**: Respond to SIGTERM/SIGINT for controlled shutdown
- **Sync Interruption**: Active sync operations interrupted at next safe checkpoint
- **Request Draining**: Stop accepting new requests; complete in-flight API calls
- **Background Job Completion**: Allow short-running background jobs to finish
- **Idempotent Resume**: Interrupted syncs resume correctly on restart due to idempotent design

The system prioritizes data consistency over completion; interrupted operations are designed for safe retry.

### Instance Bootstrap
**Objective**: Provide streamlined first-run experience for new instances

Bootstrap workflow via CLI:

- **Init Command**: `replica init` creates the initial admin user with secure credentials
- **Admin Credentials**: Generated password displayed once; can be changed via API
- **Starter Templates**: Optional content type templates (Article, Page, Media) can be created during init
- **Instance Identity**: UUID generated and instance name configured during bootstrap
- **Database Setup**: Migrations run automatically on first boot

This approach enables quick local development setup while supporting production deployments.

### Local User Management
**Objective**: Manage users within a single instance

User management model:

- **Initial Admin**: Created via CLI init command; has full system access
- **User CRUD via API**: Subsequent users created, updated, deleted through REST endpoints
- **Password Authentication**: Secure password hashing (bcrypt/argon2) for local authentication
- **API Key Generation**: Users can generate personal API keys for programmatic access
- **Session Management**: JWT tokens issued on login with configurable expiration

User accounts are instance-local; only role definitions sync between instances (per RBAC sync scope clarification).

### Testing Strategy
**Objective**: Ensure code quality and correctness across the distributed system

Testing approach:

- **Unit Tests**: Comprehensive unit tests for all business logic with high code coverage targets
- **Functional CLI Tests**: End-to-end tests exercising CLI commands against test databases
- **Mutation Testing**: gremlins framework to identify untested code paths and weak assertions
- **Code Coverage**: Coverage reports integrated into CI; coverage gates for PRs
- **Database Testing**: Tests run against SQLite by default; CI matrix includes MySQL and PostgreSQL

Test infrastructure:

- **Docker Compose**: Multi-database test environment for local development
- **CI Matrix**: GitHub Actions runs tests against all three database backends
- **MinIO**: Local S3-compatible storage for object store tests

## Risk Considerations and Mitigation Strategies

<details>
<summary>Technical Risks</summary>

- **Conflict resolution complexity**: Concurrent modifications create complex merge scenarios
    - **Mitigation**: Version branching preserves both versions; resolution is explicit and auditable. Start with manual resolution, add AI-assisted as enhancement.

- **Schema migration data loss**: Downgrades may lose data from fields not present in older schema
    - **Mitigation**: Require explicit approval for destructive migrations. Preview mode shows impact before execution. Store removed field data in metadata for potential recovery.

- **Large sync performance**: Full history sync of large content sets could overwhelm resources
    - **Mitigation**: Chunked transfers with resume capability. Configurable history depth limits. Delta sync minimizes transferred data.

- **Webhook extension latency**: Synchronous webhook calls during CRUD operations add latency
    - **Mitigation**: Configurable timeouts per subscriber. Circuit breaker pattern for unresponsive subscribers. Option to mark subscribers as non-blocking (notification-only).
</details>

<details>
<summary>Implementation Risks</summary>

- **Multi-database SQL compatibility**: Subtle differences between database engines
    - **Mitigation**: Comprehensive integration tests against all supported databases. Use golang-migrate's database-agnostic features. Avoid database-specific SQL where possible.

- **Object store abstraction leakiness**: S3 and local filesystem have different semantics
    - **Mitigation**: Define minimal interface covering required operations. Integration tests against both backends. Document behavioral differences.

- **Distributed system debugging**: Issues spanning multiple instances are hard to diagnose
    - **Mitigation**: Correlation ID propagation from day one. Comprehensive distributed tracing. Structured logging with consistent schema.
</details>

<details>
<summary>Operational Risks</summary>

- **Security of inter-instance communication**: Misconfigured instances could leak content
    - **Mitigation**: Mutual authentication required. Explicit allowlisting of peer instances. Content-type level authorization. Audit logging of all sync operations.

- **Version drift across instances**: Instances running different Replica versions
    - **Mitigation**: API versioning with compatibility matrix. Clear minimum version requirements for federation. Version negotiation in sync protocol.
</details>

## Success Criteria

### Primary Success Criteria

1. **Bidirectional Sync**: Content created on Instance A can be pushed to Instance B, modified, and pulled back to Instance A with full revision history intact
2. **Schema Evolution**: A content type schema can be upgraded on one instance, content synced to another instance running the old schema with automatic downgrade migration applied
3. **Conflict Handling**: When the same content is modified on two instances, sync detects the conflict, preserves both versions as branches, and enables resolution
4. **Multi-Database**: The same Replica binary runs correctly against SQLite, MySQL, and PostgreSQL with identical behavior
5. **Blue/Green Deployment**: Two production instances can maintain synchronized content enabling traffic switching without data inconsistency
6. **Delta Efficiency**: Syncing 10 changed items out of 10,000 total transfers only the 10 changed items plus metadata overhead

### Secondary Success Criteria

1. **Query Flexibility**: Sync operations can filter by content type, UUID list, date range, and custom field values
2. **Observability**: Operators can trace a sync operation across instance boundaries using correlation IDs
3. **CLI Completeness**: All administrative operations achievable via CLI with scriptable JSON output
4. **RBAC Enforcement**: Users with editor role cannot access content types they lack permission for; role definitions sync but assignments remain local
5. **Asset Versioning**: Binary assets support rollback to previous versions; asset history syncs with content history
6. **Quota Enforcement**: Requests exceeding quotas are rejected with clear error messages; quota usage is trackable via API
7. **Dependency Sync**: Syncing a content item automatically includes all referenced items, handling circular references correctly
8. **Full-Text Search**: Content is searchable using database-native full-text capabilities across all three database backends
9. **Scheduled Publishing**: Content scheduled for future publication transitions automatically at the specified time
10. **Webhook Extensions**: External subscribers can intercept content CRUD operations and modify data; priority ordering (EARLY/NEUTRAL/LATE) is respected
11. **Translations**: Content items can be linked as translations; fetching content can include or follow translation links
12. **Audit Trail**: All content modifications, sync operations, and permission changes are recorded in immutable audit log
13. **Sync Idempotency**: Retrying a failed sync operation produces consistent results; partial syncs don't create invalid states

## Resource Requirements

### Development Skills

- **Go expertise**: Strong Go experience including goroutines, channels, and context handling
- **Database knowledge**: SQL proficiency across SQLite, MySQL, PostgreSQL; understanding of migration strategies
- **Distributed systems**: Understanding of vector clocks, conflict resolution, and eventual consistency
- **API design**: JSON:API specification implementation experience
- **Security**: Authentication protocols (OAuth 2.0, JWT), secure inter-service communication

### Technical Infrastructure

- **Development**: Go 1.21+, Docker for database testing, MinIO for S3 testing
- **CI/CD**: GitHub Actions for testing against all database backends
- **Deployment**: Multi-arch Docker builds, goreleaser for binary releases
- **Testing**: gremlins for mutation testing, standard Go test framework with coverage
- **Dependencies**:
  - golang-migrate for database migrations
  - Standard library net/http for API server
  - OpenTelemetry Go SDK for observability
  - AWS SDK for S3 integration
  - google/cel-go for validation expressions
  - swaggo/swag or similar for OpenAPI generation
  - bcrypt or argon2 for password hashing

### External Services

- **Object Storage**: S3-compatible service for production binary asset storage
- **Databases**: Access to test instances of MySQL and PostgreSQL for integration testing

## Integration Strategy

Replica operates as a standalone service exposing JSON:API endpoints. Integration patterns:

- **Headless CMS**: Frontend applications consume content via JSON:API
- **CI/CD Pipelines**: CLI tool integrates with deployment pipelines for sync operations
- **Monitoring**: Prometheus scrapes metrics endpoint; traces export to configured collector
- **Object Storage**: Standard S3 API for asset storage; no proprietary integrations

## Notes

- The 1.0 scope explicitly excludes a web admin UI; API and CLI provide complete functionality
- AI-assisted conflict resolution is architectural provision; the interface is defined but specific AI integration can be enhanced post-1.0
- Content pruning (version history limits) is configurable but not required; instances can retain full history if storage permits
- The system assumes instances can reach each other's APIs; network topology and firewall configuration is deployment-specific
- All 32 architectural components confirmed as required for 1.0 MVP scope

### Change Log

- **2025-12-18**: Initial plan creation with 38 clarifications
- **2025-12-18**: Plan refinement - added 7 new clarifications (MVP scope, local auth, testing, bootstrap, starter types, mutation testing), added Instance Bootstrap, Local User Management, and Testing Strategy architectural sections
- **2025-12-18**: Extended Webhook System to full extension model - webhooks can now intercept and modify content during CRUD operations with EARLY/NEUTRAL/LATE priority ordering; added sequence diagram for extension flow
- **2025-12-18**: Added webhook subscription management (API only), configurable failure handling (required/optional), and CLI login flow authentication
- **2025-12-18**: Added Peer Registration, Background Job Queue, and Content Caching architectural sections
- **2025-12-19**: Plan refinement review completed - no additional clarifications needed; plan confirmed ready for task generation with 53 clarifications and 38 architectural components
