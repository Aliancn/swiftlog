# SwiftLog Architecture

Comprehensive architecture documentation for the SwiftLog platform.

## Table of Contents

- [Overview](#overview)
- [System Design](#system-design)
- [Components](#components)
- [Data Flow](#data-flow)
- [Technology Stack](#technology-stack)
- [Database Schema](#database-schema)
- [API Design](#api-design)
- [Security](#security)
- [Scalability](#scalability)
- [Future Enhancements](#future-enhancements)

## Overview

SwiftLog is a distributed log collection and analysis platform designed for:

- **Zero-intrusion**: Wrap any command without modifying code
- **Real-time**: Stream logs as they're generated
- **Accurate**: Capture exact exit codes and timing
- **Intelligent**: AI-powered log analysis
- **Scalable**: Handle thousands of concurrent streams

### Design Principles

1. **Separation of Concerns**: Each service has a single responsibility
2. **Stateless Services**: All services are horizontally scalable
3. **Event-Driven**: Redis pub/sub for real-time updates
4. **Dual Storage**: PostgreSQL for metadata, Loki for logs
5. **API-First**: All functionality exposed via REST API

## System Design

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Users                                 │
│  ┌─────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │     CLI     │    │   Browser    │    │   Scripts    │  │
│  └──────┬──────┘    └──────┬───────┘    └──────┬───────┘  │
└─────────┼──────────────────┼───────────────────┼──────────┘
          │                  │                   │
          │ gRPC             │ HTTP/WS          │ HTTP
          │                  │                   │
┌─────────▼──────────────────▼───────────────────▼──────────┐
│                      Services Layer                        │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐          │
│  │  Ingestor  │  │  REST API  │  │ WebSocket  │          │
│  │   :50051   │  │   :8080    │  │   :8081    │          │
│  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘          │
│        │               │               │                   │
│        │               │               │                   │
│  ┌─────▼───────────────▼───────────────▼──────┐           │
│  │            Message Queue (Redis)             │           │
│  │              Pub/Sub Channel                 │           │
│  └────────────────────┬─────────────────────────┘           │
│                       │                                     │
│                       │                                     │
│                 ┌─────▼──────┐                             │
│                 │ AI Worker  │                             │
│                 │ (Worker)   │                             │
│                 └─────┬──────┘                             │
│                       │                                     │
└───────────────────────┼─────────────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────────────┐
│                    Storage Layer                            │
│  ┌────────────────┐  ┌────────────────┐  ┌──────────────┐ │
│  │   PostgreSQL   │  │  Grafana Loki  │  │    Redis     │ │
│  │   (Metadata)   │  │     (Logs)     │  │ (Pub/Sub)    │ │
│  └────────────────┘  └────────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Component Communication

```
┌─────────┐  gRPC   ┌──────────┐  Insert  ┌────────────┐
│   CLI   ├────────►│ Ingestor ├─────────►│ PostgreSQL │
└─────────┘         └────┬─────┘          └────────────┘
                         │
                         │ Push Logs
                         ▼
                    ┌─────────┐
                    │  Loki   │
                    └────┬────┘
                         │
                         │ Publish Event
                         ▼
                    ┌─────────┐
                    │  Redis  │
                    │ Pub/Sub │
                    └────┬────┘
                         │
               ┌─────────┴─────────┐
               │                   │
               ▼                   ▼
          ┌─────────┐        ┌──────────┐
          │   API   │        │WebSocket │
          │         │        │ Server   │
          └─────────┘        └────┬─────┘
                                  │
                                  ▼
                             ┌─────────┐
                             │ Browser │
                             └─────────┘
```

## Components

### 1. CLI (`cli/`)

**Purpose**: Command-line tool for wrapping scripts and streaming logs.

**Technology**: Go 1.21+

**Key Responsibilities**:
- Parse command-line arguments
- Execute wrapped commands
- Capture stdout/stderr
- Stream logs via gRPC
- Report exit codes

**Communication**:
- **Outbound**: gRPC to Ingestor (`:50051`)

**Data Flow**:
```
Command Output → Pipe → Line Buffer → Label → gRPC Stream → Ingestor
```

**Configuration**:
- Stored in `~/.swiftlog/config.yaml`
- Contains: server address, API token

### 2. Ingestor (`backend/cmd/ingestor/`)

**Purpose**: gRPC server for receiving log streams.

**Technology**: Go 1.21+, gRPC

**Key Responsibilities**:
- Accept gRPC connections
- Authenticate requests
- Create/update run records
- Store logs in Loki
- Update run metadata
- Publish events to Redis

**Communication**:
- **Inbound**: gRPC from CLI clients
- **Outbound**:
  - HTTP to Loki (`:3100`)
  - SQL to PostgreSQL (`:5432`)
  - Pub/Sub to Redis (`:6379`)

**Concurrency**:
- One goroutine per active stream
- Connection pooling for database
- Batch writes to Loki (configurable)

**State Management**:
- Stateless (all state in database)
- In-memory buffer for active streams

### 3. REST API (`backend/cmd/api/`)

**Purpose**: HTTP API for web interface and integrations.

**Technology**: Go 1.21+, Gin framework

**Key Responsibilities**:
- User authentication (JWT)
- CRUD operations for projects/groups/runs
- Query logs from Loki
- Trigger AI analysis
- Serve API documentation

**Endpoints**:
- `/health` - Health check
- `/api/v1/auth/*` - Authentication
- `/api/v1/projects/*` - Project management
- `/api/v1/groups/*` - Group management
- `/api/v1/runs/*` - Run management
- `/api/v1/runs/:id/logs` - Log retrieval
- `/api/v1/runs/:id/analyze` - Trigger AI

**Communication**:
- **Inbound**: HTTP from browsers/scripts
- **Outbound**:
  - SQL to PostgreSQL
  - HTTP to Loki
  - Pub/Sub to Redis (for AI jobs)

**Authentication**:
- Bearer token authentication
- SHA-256 hashed tokens in database
- JWT for session management

### 4. WebSocket Server (`backend/cmd/websocket/`)

**Purpose**: Real-time log streaming to browsers.

**Technology**: Go 1.21+, gorilla/websocket

**Key Responsibilities**:
- Manage WebSocket connections
- Subscribe to Redis pub/sub channels
- Forward log events to connected clients
- Handle connection lifecycle

**Communication**:
- **Inbound**: WebSocket from browsers
- **Inbound**: Redis pub/sub (log events)
- **Outbound**: WebSocket to browsers

**Connection Management**:
```go
Hub {
    clients: map[*Client]bool
    register: chan *Client
    unregister: chan *Client
    broadcast: chan Message
}
```

**Message Format**:
```json
{
  "type": "log",
  "run_id": "uuid",
  "timestamp": "2025-11-18T10:00:00Z",
  "level": "STDOUT",
  "content": "log line"
}
```

### 5. AI Worker (`backend/cmd/ai-worker/`)

**Purpose**: Background processor for AI analysis.

**Technology**: Go 1.21+, OpenAI API

**Key Responsibilities**:
- Poll Redis for analysis jobs
- Fetch logs from Loki
- Send to OpenAI API
- Store analysis results
- Update run status

**Communication**:
- **Inbound**: Redis queue
- **Outbound**:
  - HTTP to Loki
  - HTTP to OpenAI API
  - SQL to PostgreSQL

**Process Flow**:
1. Pop job from Redis queue (`swiftlog:ai-jobs`)
2. Fetch run metadata from PostgreSQL
3. Query logs from Loki
4. Construct prompt for OpenAI
5. Send request to OpenAI
6. Parse markdown response
7. Update run record with report
8. Update status to `completed` or `failed`

**OpenAI Integration**:
- Model: `gpt-4o-mini` (configurable)
- Max tokens: 2000
- Temperature: 0.3 (focused responses)
- Custom base URL support (Azure, LocalAI, etc.)

### 6. Frontend (`frontend/`)

**Purpose**: Web interface for viewing and analyzing logs.

**Technology**: Next.js 14, TypeScript, Tailwind CSS

**Key Responsibilities**:
- User interface
- Project/group/run navigation
- Log display
- AI report rendering
- Real-time updates via WebSocket

**Pages**:
- `/` - Project list
- `/projects/:id` - Project detail + groups
- `/projects/:id/groups/:groupId` - Group detail + runs
- `/runs/:id` - Run detail + logs + AI report

**State Management**:
- React hooks (useState, useEffect)
- No global state library (future: React Query)

**API Client**:
- Centralized in `src/lib/api.ts`
- Type-safe with TypeScript interfaces

## Data Flow

### Log Ingestion Flow

```
1. CLI starts command
   ├── Creates run record (via Ingestor)
   ├── Start time recorded
   └── Status: "running"

2. Command outputs logs
   ├── CLI captures line-by-line
   ├── Labels: [STDOUT] or [STDERR]
   ├── Buffers in memory
   └── Streams via gRPC

3. Ingestor receives log stream
   ├── Authenticates request
   ├── Batches log lines
   ├── Pushes to Loki
   ├── Publishes to Redis pub/sub
   └── Updates run metadata periodically

4. WebSocket server
   ├── Subscribes to Redis channel
   ├── Forwards to connected browsers
   └── Maintains connection state

5. Command exits
   ├── CLI sends final metadata
   ├── End time recorded
   ├── Exit code stored
   └── Status: "success" or "failed"
```

### AI Analysis Flow

```
1. User clicks "Generate AI Report"
   └── Frontend → API: POST /runs/:id/analyze

2. API enqueues job
   ├── Validates run exists
   ├── Checks if already analyzed
   ├── Pushes job to Redis queue
   └── Returns immediately (async)

3. AI Worker processes job
   ├── Pops from Redis queue
   ├── Fetches run metadata
   ├── Queries logs from Loki (limit 10,000 lines)
   ├── Constructs analysis prompt
   └── Sends to OpenAI API

4. OpenAI responds
   ├── Returns markdown analysis
   ├── Contains: summary, errors, warnings, recommendations
   └── Typically 500-2000 tokens

5. Worker saves result
   ├── Updates run.ai_report field
   ├── Sets ai_status = "completed"
   └── Commits to PostgreSQL

6. Frontend polls or subscribes
   ├── Polls /runs/:id every 2 seconds
   ├── Or WebSocket real-time update
   └── Displays markdown report
```

### Query Flow

```
User Request
    │
    ▼
Frontend (Browser)
    │
    ├── GET /api/v1/projects
    │   └─► API → PostgreSQL → Returns projects
    │
    ├── GET /api/v1/groups/:id/runs
    │   └─► API → PostgreSQL → Returns run list
    │
    └── GET /api/v1/runs/:id/logs
        └─► API → Loki → Returns log entries
```

## Technology Stack

### Backend

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| Language | Go | 1.21+ | All backend services |
| API Framework | Gin | 1.9+ | HTTP routing |
| gRPC | grpc-go | 1.60+ | Log streaming |
| WebSocket | gorilla/websocket | 1.5+ | Real-time updates |
| Database Driver | pgx | 5.5+ | PostgreSQL connection |
| Log Storage | Grafana Loki | 2.9+ | Time-series logs |
| Cache/Queue | Redis | 7+ | Pub/sub and jobs |

### Frontend

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| Framework | Next.js | 14+ | React framework |
| Language | TypeScript | 5+ | Type safety |
| Styling | Tailwind CSS | 3+ | Utility CSS |
| Markdown | react-markdown | 9+ | AI report rendering |

### Infrastructure

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| Database | PostgreSQL | 16+ | Metadata storage |
| Log Storage | Loki | 2.9+ | Log time-series |
| Cache/Queue | Redis | 7+ | Pub/sub + jobs |
| Container | Docker | 24+ | Deployment |
| Orchestration | Docker Compose | v2+ | Service management |

## Database Schema

### Entity-Relationship Diagram

```
┌─────────────┐       ┌──────────────┐
│    users    │       │  api_tokens  │
├─────────────┤       ├──────────────┤
│ id (PK)     │◄──────┤ user_id (FK) │
│ username    │       │ token_hash   │
│ email       │       │ name         │
│ password    │       │ created_at   │
│ created_at  │       └──────────────┘
└──────┬──────┘
       │
       │ 1:N
       │
┌──────▼──────┐
│  projects   │
├─────────────┤       ┌──────────────┐
│ id (PK)     │       │  log_groups  │
│ name        │       ├──────────────┤
│ user_id(FK) │◄──────┤ project_id(FK)│
│ created_at  │  1:N  │ name         │
│ updated_at  │       │ created_at   │
└─────────────┘       └──────┬───────┘
                             │
                             │ 1:N
                             │
                      ┌──────▼───────┐
                      │  log_runs    │
                      ├──────────────┤
                      │ id (PK)      │
                      │ group_id(FK) │
                      │ start_time   │
                      │ end_time     │
                      │ exit_code    │
                      │ status       │
                      │ ai_status    │
                      │ ai_report    │
                      │ created_at   │
                      └──────────────┘
```

### Key Tables

#### users
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### api_tokens
```sql
CREATE TABLE api_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) UNIQUE NOT NULL,  -- SHA-256 hash
    name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### projects
```sql
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### log_groups
```sql
CREATE TABLE log_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, name)
);
```

#### log_runs
```sql
CREATE TABLE log_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID REFERENCES log_groups(id) ON DELETE CASCADE,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    exit_code INTEGER,
    status VARCHAR(50) DEFAULT 'running',
    ai_status VARCHAR(50) DEFAULT 'pending',
    ai_report TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Indexes

```sql
CREATE INDEX idx_log_runs_group_id ON log_runs(group_id);
CREATE INDEX idx_log_runs_start_time ON log_runs(start_time DESC);
CREATE INDEX idx_log_runs_status ON log_runs(status);
CREATE INDEX idx_log_runs_ai_status ON log_runs(ai_status);
CREATE INDEX idx_api_tokens_hash ON api_tokens(token_hash);
```

## API Design

### RESTful Principles

- **Resource-based URLs**: `/projects/:id`, `/runs/:id`
- **HTTP methods**: GET, POST, PUT, DELETE
- **JSON format**: All requests/responses
- **HTTP status codes**: Standard 2xx, 4xx, 5xx
- **Pagination**: `limit` and `offset` query params

### Authentication

**Bearer Token Format**:
```
Authorization: Bearer <token>
```

**Token Validation**:
1. Extract token from header
2. Hash with SHA-256
3. Query database: `SELECT user_id FROM api_tokens WHERE token_hash = ?`
4. Attach `user_id` to request context

### Error Responses

```json
{
  "error": "Resource not found"
}
```

**Status Codes**:
- `400` - Bad request (invalid input)
- `401` - Unauthorized (missing/invalid token)
- `403` - Forbidden (not owner)
- `404` - Not found
- `500` - Internal server error

## Security

### Threat Model

**Threats**:
1. Unauthorized access to logs
2. Token theft
3. SQL injection
4. XSS attacks
5. DoS attacks

### Mitigations

#### Authentication
- API tokens hashed with SHA-256
- JWT for web sessions
- Token-per-client recommendation

#### Authorization
- Row-level security (user_id checks)
- Project/group ownership validation
- No public endpoints (except health)

#### Input Validation
- Parameterized SQL queries (prevent injection)
- UUID validation for IDs
- String length limits

#### Network Security
- Internal services not exposed (PostgreSQL, Loki, Redis)
- CORS configured for known origins
- HTTPS required in production

#### Rate Limiting
- Future: nginx reverse proxy with rate limits
- Per-token: 1000 req/hour

## Scalability

### Current Limits

- **Concurrent streams**: 1000+
- **Logs per run**: 10,000 (AI analysis limit)
- **Storage**: Unlimited (Loki retention policy)
- **API throughput**: 5000 req/sec (single instance)

### Horizontal Scaling

All services are stateless and can be scaled:

```bash
docker compose up -d --scale ingestor=3
docker compose up -d --scale api=5
docker compose up -d --scale websocket=2
docker compose up -d --scale ai-worker=3
```

### Bottlenecks

1. **PostgreSQL**: Use read replicas for queries
2. **Loki**: Distribute with S3 backend
3. **Redis**: Use Redis Cluster
4. **Network**: Add load balancer (nginx/HAProxy)

### Performance Targets

| Metric | Target | Current |
|--------|--------|---------|
| CLI overhead | <5% | ~2% |
| Log latency | <2s | <1s |
| API p95 | <200ms | ~50ms |
| Concurrent streams | 1000+ | ✓ |

## Future Enhancements

### Short Term

- [ ] React Query for frontend caching
- [ ] Pagination for log viewer
- [ ] Log search/filtering
- [ ] Export logs (CSV, JSON)
- [ ] Email notifications
- [ ] Webhooks for run completion

### Medium Term

- [ ] Multi-user organizations
- [ ] Role-based access control (RBAC)
- [ ] Audit logging
- [ ] Metrics dashboard (Prometheus/Grafana)
- [ ] Advanced AI features (anomaly detection)
- [ ] CLI plugins (custom output parsers)

### Long Term

- [ ] Kubernetes deployment
- [ ] Multi-region support
- [ ] Log retention policies
- [ ] Cost analytics
- [ ] Integration marketplace (Slack, PagerDuty, etc.)
- [ ] Custom AI models (fine-tuning)

## References

- **Main Documentation**: [README.md](../README.md)
- **API Documentation**: [API.md](./API.md)
- **CLI Documentation**: [cli/README.md](../cli/README.md)
- **Frontend Documentation**: [frontend/README.md](../frontend/README.md)

---

**Last Updated**: 2025-11-19
