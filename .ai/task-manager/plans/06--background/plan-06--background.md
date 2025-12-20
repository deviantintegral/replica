---
id: 6
version: 0.6.0
summary: "Background Processing - Job queue, webhooks, scheduled publishing, and audit logging"
created: 2025-12-19
depends_on: [0.5.0]
---

# Release 0.6.0: Background Processing

## Release Overview

This release adds asynchronous processing capabilities. After this release, Replica can dispatch webhooks to external systems, schedule content for future publication, run background asset transformations, and maintain an immutable audit log of all operations.

## Prerequisites

- Release 0.5.0 (API Layer) completed

## Scope

### Included Tasks

| Task | Name | Description |
|------|------|-------------|
| 20 | Background Job Queue | Database-backed job queue with retry logic |
| 21 | Webhook System | HTTP-based extension points with HMAC signing |
| 22 | Scheduled Publishing | Workflow transitions at specified times |
| 24 | Audit Logging | Immutable audit trail with webhook streaming |

### Not Included

- Sync functionality (0.7.0)
- Observability stack (0.8.0)

## Architecture Additions

```
internal/
├── jobs/
│   ├── queue.go           # Database job queue
│   ├── worker.go          # Job worker with semaphores
│   ├── scheduler.go       # Cron-like scheduling
│   └── types/
│       ├── webhook.go     # Webhook delivery job
│       ├── publish.go     # Scheduled publish job
│       ├── transform.go   # Asset transform job
│       ├── prune.go       # Content pruning job
│       └── audit.go       # Audit streaming job
├── webhook/
│   ├── service.go         # Webhook orchestration
│   ├── subscriber.go      # Subscriber management
│   ├── dispatcher.go      # Event dispatching
│   ├── signer.go          # HMAC-SHA256 signing
│   └── retry.go           # Retry with backoff
├── scheduler/
│   ├── service.go         # Scheduled operations
│   └── transition.go      # Workflow transitions
└── audit/
    ├── log.go             # Audit record storage
    ├── events.go          # Event types
    └── stream.go          # Webhook streaming
```

## Task Details

### Task 20: Background Job Queue

**Objective**: Implement reliable database-backed job queue

**Job Model**:

```go
type Job struct {
    ID          uuid.UUID
    Type        string          // webhook, publish, transform, prune, audit
    Payload     json.RawMessage
    Priority    int             // 0=low, 5=normal, 10=high
    Status      JobStatus       // pending, running, completed, failed, dead
    Attempts    int
    MaxAttempts int
    LastError   *string
    RunAt       time.Time       // Scheduled execution time
    StartedAt   *time.Time
    CompletedAt *time.Time
    CreatedAt   time.Time
}

type JobStatus string
const (
    JobPending   JobStatus = "pending"
    JobRunning   JobStatus = "running"
    JobCompleted JobStatus = "completed"
    JobFailed    JobStatus = "failed"
    JobDead      JobStatus = "dead"  // Max retries exceeded
)
```

**Queue Interface**:

```go
type JobQueue interface {
    Enqueue(ctx context.Context, job *Job) error
    Dequeue(ctx context.Context, types []string) (*Job, error)
    Complete(ctx context.Context, jobID uuid.UUID) error
    Fail(ctx context.Context, jobID uuid.UUID, err error) error
    Retry(ctx context.Context, jobID uuid.UUID, runAt time.Time) error
    List(ctx context.Context, filter JobFilter) ([]*Job, error)
    Cleanup(ctx context.Context, olderThan time.Duration) error
}
```

**Worker Design**:
```go
// Worker spawns goroutine per job, uses semaphores for external I/O
type Worker struct {
    queue     JobQueue
    handlers  map[string]JobHandler
    semaphore chan struct{}  // Limits concurrent external calls
}

func (w *Worker) Run(ctx context.Context) error {
    for {
        job, err := w.queue.Dequeue(ctx, w.types)
        if err != nil {
            return err
        }

        // Spawn goroutine per job
        go func(j *Job) {
            // Acquire semaphore for external I/O jobs
            if j.RequiresExternalIO() {
                w.semaphore <- struct{}{}
                defer func() { <-w.semaphore }()
            }

            w.execute(ctx, j)
        }(job)
    }
}
```

**Retry Configuration**:
```yaml
jobs:
  workers: 10  # Goroutine concurrency
  external_io_limit: 5  # Semaphore for webhooks, S3, etc.

  retry:
    webhook:
      max_attempts: 5
      backoff: exponential
      base_delay: 1s
      max_delay: 5m
    transform:
      max_attempts: 3
      backoff: exponential
      base_delay: 5s
    prune:
      max_attempts: 3
      backoff: linear
      base_delay: 1m
```

**Deliverables**:
- Database-backed job storage
- Priority-based dequeuing
- Configurable retry per job type
- Exponential/linear backoff
- Dead letter handling
- Job status API endpoints

**Acceptance Criteria**:
- [ ] Jobs persist across restarts
- [ ] Priority ordering is respected
- [ ] Retries follow configured policy
- [ ] Dead jobs are tracked for inspection
- [ ] Concurrent execution is bounded

### Task 21: Webhook System

**Objective**: Enable HTTP-based extensions with modification capability

**Webhook Events**:

| Event | Trigger | Modification Allowed |
|-------|---------|---------------------|
| `content.creating` | Before create | Yes |
| `content.created` | After create | No (notification) |
| `content.updating` | Before update | Yes |
| `content.updated` | After update | No (notification) |
| `content.deleting` | Before delete | Yes (can abort) |
| `content.deleted` | After delete | No (notification) |
| `workflow.transitioning` | Before transition | Yes |
| `workflow.transitioned` | After transition | No (notification) |
| `schema.updating` | Before schema change | Yes |
| `schema.updated` | After schema change | No (notification) |

**Subscriber Model**:

```go
type WebhookSubscriber struct {
    ID           uuid.UUID
    URL          string
    Events       []string          // Event types to receive
    ContentTypes []string          // Filter by content type (or all)
    Priority     WebhookPriority   // EARLY, NEUTRAL, LATE
    FailureMode  FailureMode       // required, optional
    Secret       string            // For HMAC signing
    SchemaVersion string           // e.g., "1.2" or "1.*"
    Timeout      time.Duration
    Enabled      bool
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type WebhookPriority string
const (
    PriorityEarly   WebhookPriority = "EARLY"
    PriorityNeutral WebhookPriority = "NEUTRAL"
    PriorityLate    WebhookPriority = "LATE"
)

type FailureMode string
const (
    FailureRequired FailureMode = "required"  // Abort operation on failure
    FailureOptional FailureMode = "optional"  // Skip and continue
)
```

**Webhook Payload**:

```json
{
  "id": "event-uuid",
  "type": "content.creating",
  "timestamp": "2025-12-19T10:00:00Z",
  "data": {
    "type": "content",
    "id": "content-uuid",
    "attributes": { ... }
  },
  "meta": {
    "schemaVersion": "1.2.0",
    "contentType": "article",
    "actor": "user-uuid"
  }
}
```

**HMAC Signature**:
```
X-Replica-Signature: sha256=base64(HMAC-SHA256(payload, secret))
X-Replica-Timestamp: 1703066400
```

**Modification Response**:
```json
{
  "data": {
    "type": "content",
    "id": "content-uuid",
    "attributes": {
      "title": "Modified by webhook",
      "enrichedField": "Added by subscriber"
    }
  }
}
```

**API Endpoints**:
```
GET    /v1/webhooks                 # List subscribers
POST   /v1/webhooks                 # Create subscriber
GET    /v1/webhooks/{id}            # Get subscriber
PATCH  /v1/webhooks/{id}            # Update subscriber
DELETE /v1/webhooks/{id}            # Delete subscriber
GET    /v1/webhooks/{id}/deliveries # List delivery history
POST   /v1/webhooks/{id}/test       # Send test event
```

**Deliverables**:
- Subscriber CRUD via API
- Event dispatching with priority ordering
- HMAC-SHA256 signature generation
- Content modification handling
- Retry with exponential backoff
- Delivery history tracking
- Schema version transformation for payloads

**Acceptance Criteria**:
- [ ] Webhooks fire on content operations
- [ ] Priority ordering (EARLY → NEUTRAL → LATE) works
- [ ] HMAC signatures are valid
- [ ] Modifications are applied to content
- [ ] Required subscribers abort on failure
- [ ] Optional subscribers are skipped on failure
- [ ] Retries use exponential backoff

### Task 22: Scheduled Publishing

**Objective**: Enable content to transition workflow states at specified times

**Scheduling Model**:

```go
type ScheduledTransition struct {
    ID          uuid.UUID
    ContentID   uuid.UUID
    FromState   string
    ToState     string
    ScheduledAt time.Time
    ExecutedAt  *time.Time
    Status      ScheduleStatus
    Error       *string
    CreatedBy   uuid.UUID
    CreatedAt   time.Time
}

type ScheduleStatus string
const (
    SchedulePending   ScheduleStatus = "pending"
    ScheduleExecuted  ScheduleStatus = "executed"
    ScheduleCancelled ScheduleStatus = "cancelled"
    ScheduleFailed    ScheduleStatus = "failed"
)
```

**API Endpoints**:
```
POST   /v1/content/{id}/schedule           # Schedule transition
GET    /v1/content/{id}/schedule           # Get scheduled transitions
DELETE /v1/content/{id}/schedule/{schedId} # Cancel scheduled transition
GET    /v1/schedules                       # List all scheduled transitions
```

**Request Example**:
```json
{
  "toState": "published",
  "scheduledAt": "2025-12-25T00:00:00Z"
}
```

**Scheduler Job**:
- Runs every minute
- Queries pending transitions where `scheduledAt <= now()`
- Executes transitions via content service
- Creates jobs for each transition

**Deliverables**:
- Schedule creation and cancellation API
- Background scheduler job
- Transition execution with audit trail
- Schedule status tracking

**Acceptance Criteria**:
- [ ] Content can be scheduled for future publication
- [ ] Scheduled transitions execute at correct time
- [ ] Cancelled schedules don't execute
- [ ] Failed transitions are logged with error
- [ ] Schedule history is queryable

### Task 24: Audit Logging

**Objective**: Maintain immutable audit trail with real-time streaming

**Audit Event Model**:

```go
type AuditEvent struct {
    ID          uuid.UUID
    Type        AuditEventType
    Actor       uuid.UUID         // User who performed action
    ActorType   string            // "user", "system", "sync"
    Resource    string            // "content", "user", "role", etc.
    ResourceID  uuid.UUID
    Action      string            // "create", "update", "delete", etc.
    Changes     json.RawMessage   // Field-level changes
    Metadata    json.RawMessage   // Additional context
    IPAddress   string
    UserAgent   string
    Timestamp   time.Time
}

type AuditEventType string
const (
    AuditContentCreate  AuditEventType = "content.create"
    AuditContentUpdate  AuditEventType = "content.update"
    AuditContentDelete  AuditEventType = "content.delete"
    AuditContentPublish AuditEventType = "content.publish"
    AuditUserCreate     AuditEventType = "user.create"
    AuditUserLogin      AuditEventType = "user.login"
    AuditRoleChange     AuditEventType = "role.change"
    AuditSyncPush       AuditEventType = "sync.push"
    AuditSyncPull       AuditEventType = "sync.pull"
)
```

**Immutability**:
- Audit records are append-only
- No UPDATE or DELETE operations allowed
- Retention period configurable (default: forever)

**Streaming to External Systems**:
```go
type AuditStreamConfig struct {
    WebhookURL  string
    BatchSize   int           // Events per batch
    FlushPeriod time.Duration // Max time before flush
    Secret      string        // HMAC signing
}
```

**API Endpoints**:
```
GET    /v1/audit                    # Query audit log
GET    /v1/audit/{id}               # Get specific event
GET    /v1/audit/stream             # Get stream config
POST   /v1/audit/stream             # Configure stream
DELETE /v1/audit/stream             # Disable streaming
```

**Query Parameters**:
```
GET /v1/audit?type=content.update&actor=uuid&resource=content&from=2025-12-01&to=2025-12-31
```

**Deliverables**:
- Audit event recording for all operations
- Immutable storage with retention
- Query API with filters
- Webhook streaming with batching
- HMAC-signed stream payloads

**Acceptance Criteria**:
- [ ] All content operations are logged
- [ ] Audit records cannot be modified
- [ ] Query API filters correctly
- [ ] Streaming batches events efficiently
- [ ] Stream survives webhook failures (queued)

## Database Migrations

This release adds:
- `jobs` - Background job queue
- `webhook_subscribers` - Webhook registrations
- `webhook_deliveries` - Delivery history
- `scheduled_transitions` - Scheduled workflow changes
- `audit_events` - Immutable audit log
- `audit_stream_config` - Stream configuration

## Dependencies

| Dependency | From Release |
|------------|--------------|
| Content service | 0.3.0 |
| JSON:API server | 0.5.0 |
| Database abstraction | 0.1.0 |

## Success Criteria for 0.6.0

1. **Job Queue**: Jobs persist and process reliably
2. **Webhooks**: External systems receive events with modification capability
3. **Scheduling**: Content publishes at scheduled times
4. **Audit**: All operations are logged immutably
5. **Reliability**: Failed jobs retry correctly; streaming survives failures

## Release Checklist

- [ ] All tasks completed and tested
- [ ] Webhook HMAC verified with external client
- [ ] Scheduled publishing tested across timezone
- [ ] Audit log immutability verified
- [ ] Job retry behavior tested
- [ ] CI pipeline green
- [ ] Release notes drafted
- [ ] Git tag created: v0.6.0
