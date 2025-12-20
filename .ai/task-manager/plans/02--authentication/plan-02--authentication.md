---
id: 2
version: 0.2.0
summary: "Data Models & Authentication - Core domain models, JWT/OAuth, RBAC, and rate limiting"
created: 2025-12-19
depends_on: [0.1.0]
---

# Release 0.2.0: Data Models & Authentication

## Release Overview

This release introduces the core domain models and complete authentication/authorization stack. After this release, the system has user management, secure API access with JWT tokens, role-based permissions, and rate limiting - all the security infrastructure needed for content management.

## Prerequisites

- Release 0.1.0 (Foundation Layer) completed

## Scope

### Included Tasks

| Task | Name | Description |
|------|------|-------------|
| 04 | Core Domain Models and Repository Layer | Content, revision, user, role entities with UUIDv7 |
| 05 | Authentication System | JWT/OAuth 2.0 with built-in IdP, RS256 signing |
| 06 | Role-Based Access Control | Custom roles, granular permissions, content-type scoping |
| 07 | Quota and Rate Limiting | Sliding window rate limits, storage quotas |

### Not Included

- Content service business logic
- Schema versioning
- Asset handling
- API endpoints (authentication infra only)
- Sync functionality

## Architecture Additions

```
internal/
├── domain/
│   ├── content.go         # Content item entity
│   ├── revision.go        # Revision entity with vector clocks
│   ├── contenttype.go     # Content type schema entity
│   ├── user.go            # User entity
│   ├── role.go            # Role and permission entities
│   └── instance.go        # Instance identity
├── repository/
│   ├── content.go         # Content repository interface
│   ├── user.go            # User repository interface
│   └── role.go            # Role repository interface
├── auth/
│   ├── jwt.go             # JWT generation and validation
│   ├── oauth.go           # OAuth 2.0 flows
│   ├── password.go        # argon2id password hashing
│   └── middleware.go      # Auth middleware
├── rbac/
│   ├── permission.go      # Permission definitions
│   ├── evaluator.go       # Permission evaluation
│   └── middleware.go      # RBAC middleware
└── ratelimit/
    ├── limiter.go         # Sliding window implementation
    └── quota.go           # Storage quota tracking
```

## Task Details

### Task 04: Core Domain Models and Repository Layer

**Objective**: Define all core entities with UUIDv7 identification and version tracking

**Domain Entities**:

```go
// Content item with revision tracking
type Content struct {
    ID          uuid.UUID
    ContentType string
    Revision    uuid.UUID
    Data        map[string]any
    CreatedAt   time.Time
    UpdatedAt   time.Time
    CreatedBy   uuid.UUID
    UpdatedBy   uuid.UUID
}

// Revision for version history
type Revision struct {
    ID          uuid.UUID
    ContentID   uuid.UUID
    ParentIDs   []uuid.UUID  // DAG structure for merges
    Data        map[string]any
    VectorClock map[string]uint64
    CreatedAt   time.Time
    CreatedBy   uuid.UUID
}

// User with argon2id password hash
type User struct {
    ID           uuid.UUID
    Username     string
    Email        string
    PasswordHash string
    Roles        []uuid.UUID
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// Instance identity
type Instance struct {
    ID        uuid.UUID  // Generated at first boot
    Name      string     // Admin-configured friendly name
    CreatedAt time.Time
}
```

**Deliverables**:
- All domain entity definitions
- Repository interfaces for each entity
- SQLite, MariaDB/MySQL, PostgreSQL implementations
- Database migrations for all tables
- UUIDv7 generation utility

**Acceptance Criteria**:
- [ ] All entities have proper UUIDv7 generation
- [ ] Repository CRUD operations work on all databases
- [ ] Migrations create proper indexes
- [ ] Vector clock structure supports distributed versioning

### Task 05: Authentication System

**Objective**: Implement JWT/OAuth 2.0 authentication with built-in identity provider

**Components**:

1. **Password Hashing**: argon2id with secure defaults
2. **JWT Generation**: RS256 signing with configurable key pair
3. **Token Management**: 1-hour access tokens, 30-day refresh tokens
4. **OAuth 2.0 Flows**:
   - Client credentials (service-to-service)
   - Authorization code (user authentication)

**Configuration**:
```yaml
auth:
  jwt:
    access_token_ttl: 1h
    refresh_token_ttl: 720h  # 30 days
    private_key_path: /etc/replica/jwt.key
    public_key_path: /etc/replica/jwt.pub
  oauth:
    authorization_code:
      enabled: true
    client_credentials:
      enabled: true
```

**Deliverables**:
- Password hashing with argon2id
- JWT generation and validation
- Refresh token rotation
- OAuth client registration
- Authentication middleware
- Token revocation support

**Acceptance Criteria**:
- [ ] Users can authenticate with username/password
- [ ] JWT tokens are properly signed with RS256
- [ ] Refresh tokens enable long-lived sessions
- [ ] OAuth client credentials flow works
- [ ] Expired/revoked tokens are rejected

### Task 06: Role-Based Access Control

**Objective**: Implement granular permission system with custom roles

**Permission Model**:

```go
type Permission struct {
    Resource    string   // "content", "content:article", "user", "role"
    Actions     []string // "create", "read", "update", "delete"
    Constraints map[string]any // Optional field-level restrictions
}

type Role struct {
    ID          uuid.UUID
    Name        string
    Description string
    Permissions []Permission
    IsSystem    bool  // Built-in roles cannot be deleted
}
```

**Built-in Roles**:
- `admin`: Full system access
- `editor`: Content CRUD, no user/role management
- `viewer`: Read-only access

**Deliverables**:
- Permission definition system
- Role CRUD operations
- Permission evaluation engine
- RBAC middleware for request authorization
- Content-type scoped permissions

**Acceptance Criteria**:
- [ ] Custom roles can be created with specific permissions
- [ ] Permissions can be scoped to content types
- [ ] RBAC middleware correctly blocks unauthorized access
- [ ] Built-in roles cannot be deleted
- [ ] Role changes take effect immediately

### Task 07: Quota and Rate Limiting

**Objective**: Protect system resources with rate limits and storage quotas

**Rate Limiting**:
- Sliding window algorithm
- Per-client (authenticated user or IP)
- Configurable limits per endpoint category

**Quota Types**:
- API requests per minute/hour
- Storage (bytes) per user/instance
- Content items per content type

**Configuration**:
```yaml
quotas:
  rate_limits:
    default:
      requests_per_minute: 60
      requests_per_hour: 1000
    sync:
      requests_per_minute: 10
  storage:
    max_per_user: 10GB
    max_per_instance: 100GB
```

**Deliverables**:
- Sliding window rate limiter
- Storage quota tracking
- Quota enforcement middleware
- HTTP 429 responses with Retry-After header
- Quota usage API endpoint

**Acceptance Criteria**:
- [ ] Rate limits enforce sliding window correctly
- [ ] Storage quotas are tracked and enforced
- [ ] 429 responses include proper Retry-After
- [ ] Quota usage is queryable via API
- [ ] Different limits can apply to different users/roles

## Database Migrations

This release adds the following tables:
- `instances` - Instance identity
- `users` - User accounts
- `roles` - Role definitions
- `role_permissions` - Role-permission mappings
- `user_roles` - User-role assignments
- `oauth_clients` - OAuth client registrations
- `refresh_tokens` - Active refresh tokens
- `rate_limit_windows` - Rate limit tracking
- `quota_usage` - Quota consumption tracking
- `content_types` - Content type schema definitions
- `contents` - Content items
- `revisions` - Content revision history

## Dependencies

| Dependency | From Release |
|------------|--------------|
| Database abstraction | 0.1.0 |
| Configuration system | 0.1.0 |
| Testing infrastructure | 0.1.0 |

## Success Criteria for 0.2.0

1. **Users**: Can create users, authenticate, receive JWT tokens
2. **Roles**: Custom roles with granular permissions work correctly
3. **Rate Limiting**: Excessive requests are throttled with proper 429 responses
4. **Quotas**: Storage quotas are tracked and enforced
5. **Security**: Passwords are properly hashed, tokens properly signed

## Release Checklist

- [ ] All tasks completed and tested
- [ ] Security review of authentication implementation
- [ ] Password hashing parameters reviewed
- [ ] JWT key generation documented
- [ ] CI pipeline green
- [ ] Release notes drafted
- [ ] Git tag created: v0.2.0
