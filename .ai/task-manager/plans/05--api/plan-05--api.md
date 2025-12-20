---
id: 5
version: 0.5.0
summary: "API Layer - JSON:API server, bulk operations, CLI core, and OpenAPI documentation"
created: 2025-12-19
depends_on: [0.4.0]
---

# Release 0.5.0: API Layer

## Release Overview

This release delivers the complete REST API and CLI foundation. After this release, Replica is a fully functional single-instance CMS accessible via JSON:API endpoints and command-line interface. Users can perform all content and asset operations through the API with proper authentication, authorization, and documentation.

## Prerequisites

- Release 0.4.0 (Asset Management) completed

## Scope

### Included Tasks

| Task | Name | Description |
|------|------|-------------|
| 15 | JSON:API Server Core | Full JSON:API implementation with google/jsonapi |
| 16 | Bulk Operations API | Batch create/update/delete with per-item errors |
| 26 | CLI Core | Command structure, auth, output formatting |
| 28 | OpenAPI Documentation | Auto-generated Swagger/OpenAPI spec |

### Not Included

- Background processing (0.6.0)
- Sync functionality (0.7.0)
- CLI sync operations (0.8.0)

## Architecture Additions

```
internal/
├── api/
│   ├── server.go          # HTTP server setup
│   ├── router.go          # Route definitions
│   ├── middleware/
│   │   ├── auth.go        # JWT authentication
│   │   ├── rbac.go        # Permission checks
│   │   ├── ratelimit.go   # Rate limiting
│   │   ├── cors.go        # CORS handling
│   │   └── recovery.go    # Panic recovery
│   ├── handlers/
│   │   ├── content.go     # Content CRUD
│   │   ├── contenttype.go # Content type management
│   │   ├── asset.go       # Asset operations
│   │   ├── user.go        # User management
│   │   ├── role.go        # Role management
│   │   ├── auth.go        # Authentication endpoints
│   │   ├── bulk.go        # Bulk operations
│   │   └── health.go      # Health check
│   ├── jsonapi/
│   │   ├── marshal.go     # JSON:API serialization
│   │   ├── unmarshal.go   # JSON:API deserialization
│   │   ├── errors.go      # Error formatting
│   │   ├── pagination.go  # Cursor-based pagination
│   │   └── sparse.go      # Sparse fieldsets
│   └── openapi/
│       └── generator.go   # OpenAPI spec generation
└── cmd/
    └── replica/
        ├── main.go        # Entry point
        ├── root.go        # Root command
        ├── serve.go       # Server command
        ├── init.go        # Bootstrap command
        ├── login.go       # Authentication
        ├── config.go      # Configuration helpers
        └── completion.go  # Shell completions
```

## Task Details

### Task 15: JSON:API Server Core

**Objective**: Implement complete JSON:API 1.1 compliant REST API

**API Endpoints**:

```
# Authentication
POST   /v1/auth/token          # Get access token
POST   /v1/auth/refresh        # Refresh token
POST   /v1/auth/revoke         # Revoke token

# Content Types
GET    /v1/content-types                    # List content types
POST   /v1/content-types                    # Create content type
GET    /v1/content-types/{type}             # Get content type
PATCH  /v1/content-types/{type}             # Update content type
DELETE /v1/content-types/{type}             # Delete content type
GET    /v1/content-types/{type}/versions    # List schema versions

# Content
GET    /v1/content                          # List all content
POST   /v1/content                          # Create content
GET    /v1/content/{id}                     # Get content
PATCH  /v1/content/{id}                     # Update content
DELETE /v1/content/{id}                     # Delete content
GET    /v1/content/{id}/revisions           # List revisions
GET    /v1/content/{id}/revisions/{rev}     # Get revision
POST   /v1/content/{id}/restore/{rev}       # Restore revision
POST   /v1/content/{id}/workflow/{state}    # Transition state

# Assets
GET    /v1/assets                           # List assets
POST   /v1/assets                           # Upload asset
GET    /v1/assets/{id}                      # Get asset metadata
DELETE /v1/assets/{id}                      # Delete asset
GET    /v1/assets/{id}/download             # Download asset
GET    /v1/assets/{id}/variants/{variant}   # Get variant

# Users
GET    /v1/users                            # List users
POST   /v1/users                            # Create user
GET    /v1/users/{id}                       # Get user
PATCH  /v1/users/{id}                       # Update user
DELETE /v1/users/{id}                       # Delete user

# Roles
GET    /v1/roles                            # List roles
POST   /v1/roles                            # Create role
GET    /v1/roles/{id}                       # Get role
PATCH  /v1/roles/{id}                       # Update role
DELETE /v1/roles/{id}                       # Delete role

# Search
GET    /v1/search                           # Full-text search

# Health
GET    /health                              # Health check
```

**JSON:API Features**:
- Compound documents with `include` parameter
- Sparse fieldsets with `fields[type]` parameter
- Cursor-based pagination with `page[cursor]` and `page[size]`
- Sorting with `sort` parameter
- Filtering with `filter[field]` parameter
- Schema version selection via header or query

**Response Format**:
```json
{
  "data": {
    "type": "content",
    "id": "01234567-89ab-cdef-0123-456789abcdef",
    "attributes": {
      "title": "Hello World",
      "body": "Content here..."
    },
    "relationships": {
      "author": {
        "data": { "type": "user", "id": "..." }
      }
    },
    "meta": {
      "schemaVersion": "1.2.0",
      "revision": "..."
    }
  },
  "included": [...],
  "links": {
    "self": "/v1/content/...",
    "next": "/v1/content?page[cursor]=..."
  }
}
```

**Deliverables**:
- Complete JSON:API endpoint implementation
- google/jsonapi integration for marshaling
- Cursor-based pagination
- Sparse fieldsets
- Compound documents
- Error responses per JSON:API spec
- Schema version header handling

**Acceptance Criteria**:
- [ ] All endpoints respond with valid JSON:API
- [ ] Pagination works correctly with cursors
- [ ] Include parameter fetches related resources
- [ ] Sparse fieldsets reduce payload size
- [ ] Authentication middleware protects endpoints
- [ ] RBAC blocks unauthorized operations

### Task 16: Bulk Operations API

**Objective**: Enable efficient batch operations with detailed per-item feedback

**Bulk Endpoints**:
```
POST   /v1/bulk/content        # Bulk create/update/delete
POST   /v1/bulk/assets         # Bulk asset operations
```

**Request Format**:
```json
{
  "operations": [
    {
      "op": "create",
      "data": {
        "type": "content",
        "attributes": { ... },
        "meta": { "schemaVersion": "1.2" }
      }
    },
    {
      "op": "update",
      "ref": { "type": "content", "id": "..." },
      "data": { ... }
    },
    {
      "op": "delete",
      "ref": { "type": "content", "id": "..." }
    }
  ],
  "atomic": false
}
```

**Response Format**:
```json
{
  "results": [
    {
      "status": 201,
      "data": { "type": "content", "id": "..." }
    },
    {
      "status": 200,
      "data": { "type": "content", "id": "..." }
    },
    {
      "status": 404,
      "errors": [{ "status": "404", "title": "Not Found" }]
    }
  ],
  "meta": {
    "succeeded": 2,
    "failed": 1
  }
}
```

**Atomic Mode**:
- `atomic: true`: All-or-nothing semantics (transaction)
- `atomic: false`: Partial success allowed (default)

**Deliverables**:
- Bulk endpoint implementation
- Per-item status and error reporting
- Atomic transaction support
- Schema version per-item override
- Webhook firing per-item (when webhooks available)

**Acceptance Criteria**:
- [ ] Bulk create/update/delete works
- [ ] Per-item errors are reported correctly
- [ ] Atomic mode rolls back on failure
- [ ] Schema versions can be specified per-item
- [ ] Large batches (1000+ items) complete successfully

### Task 26: CLI Core

**Objective**: Provide command-line interface foundation with authentication

**Command Structure**:
```
replica
├── version              # Show version
├── init                 # Initialize new instance
├── serve                # Start API server
├── login                # Authenticate and store token
├── logout               # Remove stored token
├── config               # Configuration helpers
│   ├── show             # Show current config
│   └── validate         # Validate config file
├── completion           # Generate shell completions
│   ├── bash
│   ├── zsh
│   └── fish
└── (future commands in 0.8.0)
```

**Authentication Flow**:
```bash
# Login with username/password
$ replica login
Username: admin
Password: ********
Successfully logged in. Token stored in ~/.config/replica/token

# Token stored in XDG-compliant location
$ cat ~/.config/replica/token
eyJhbGciOiJSUzI1NiIs...
```

**Output Formatting**:
```bash
# Human-readable by default
$ replica config show
Instance: production-us-east
Database: postgresql://...
Server: 0.0.0.0:8080

# JSON for scripting
$ replica config show --json
{"instance":"production-us-east","database":"..."}
```

**Deliverables**:
- cobra-based CLI structure
- Login/logout with token storage
- XDG-compliant config directory
- Shell completion generation
- Human-readable and JSON output modes
- Instance init command

**Acceptance Criteria**:
- [ ] `replica init` creates admin user
- [ ] `replica serve` starts API server
- [ ] `replica login` stores JWT token
- [ ] `replica completion` generates working completions
- [ ] `--json` flag produces parseable output
- [ ] Token auto-refreshes when near expiration

### Task 28: OpenAPI Documentation

**Objective**: Auto-generate OpenAPI/Swagger specification from code

**OpenAPI Generation**:
```go
// Annotations on handlers generate OpenAPI spec
// @Summary Get content by ID
// @Description Returns a single content item
// @Tags content
// @Accept json
// @Produce json
// @Param id path string true "Content ID"
// @Param schemaVersion query string false "Schema version"
// @Success 200 {object} jsonapi.Document
// @Failure 404 {object} jsonapi.Errors
// @Router /v1/content/{id} [get]
func (h *ContentHandler) Get(w http.ResponseWriter, r *http.Request) {
    // ...
}
```

**Endpoints**:
```
GET    /v1/openapi.json        # OpenAPI 3.0 spec
GET    /v1/docs                # Swagger UI (optional)
```

**Configuration**:
```yaml
api:
  openapi:
    enabled: true
    swagger_ui: false  # Disable in production
```

**Deliverables**:
- swaggo/swag integration for spec generation
- OpenAPI 3.0 JSON endpoint
- Optional Swagger UI for development
- Schema validation against spec

**Acceptance Criteria**:
- [ ] OpenAPI spec generates from code
- [ ] Spec validates as valid OpenAPI 3.0
- [ ] All endpoints are documented
- [ ] Request/response schemas are accurate
- [ ] Swagger UI works in development mode

## Configuration Additions

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
  shutdown_timeout: 30s

api:
  version: v1
  pagination:
    default_size: 100
    max_size: 1000
  cors:
    enabled: true
    allowed_origins: ["*"]
    allowed_methods: ["GET", "POST", "PATCH", "DELETE"]
  openapi:
    enabled: true
    swagger_ui: false
```

## Dependencies

| Dependency | From Release |
|------------|--------------|
| Content service | 0.3.0 |
| Asset service | 0.4.0 |
| Authentication | 0.2.0 |
| RBAC | 0.2.0 |
| Rate limiting | 0.2.0 |

## Success Criteria for 0.5.0

1. **JSON:API Compliance**: All endpoints follow JSON:API 1.1 spec
2. **Authentication**: JWT tokens protect all endpoints
3. **Authorization**: RBAC blocks unauthorized operations
4. **Bulk Operations**: Batch operations with per-item feedback
5. **CLI**: Basic commands work with proper auth flow
6. **Documentation**: OpenAPI spec accurately describes API

## Release Checklist

- [ ] All tasks completed and tested
- [ ] JSON:API compliance validated
- [ ] All endpoints documented in OpenAPI
- [ ] CLI works across Linux/macOS/Windows
- [ ] Shell completions tested
- [ ] Rate limiting verified under load
- [ ] CI pipeline green
- [ ] Release notes drafted
- [ ] Git tag created: v0.5.0
