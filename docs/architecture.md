---
layout: default
title: Architecture
nav_order: 3
description: "System architecture, components, and data flow of SwiftLog"
---

# SwiftLog Architecture

This document describes the system architecture, components, and data flow of SwiftLog.

## System Overview

SwiftLog is a distributed log collection and analysis platform built with a microservices architecture.

```
┌─────────────┐
│     CLI     │ ──gRPC──► Ingestor ──► Loki (Logs)
└─────────────┘              │
                             ▼
                       PostgreSQL (Metadata)
                             ▲
          ┌──────────────────┼──────────────────┐
          │                  │                  │
    REST API          WebSocket Server    AI Worker
          │                  │                  │
          └──────────────────┴──────────────────┘
                             │
                         Frontend
```

## Components

### 1. CLI Tool (`cli/`)

**Technology:** Go 1.21+

**Purpose:** Command-line interface for running scripts and streaming their logs

**Key Features:**
- Wraps any command and captures stdout/stderr
- Real-time gRPC streaming to Ingestor
- Smart project/group auto-detection
- Configuration management
- Exit code preservation

**Communication:**
- gRPC client → Ingestor (port 50051)
- Metadata-based authentication

### 2. Ingestor Service (`backend/cmd/ingestor/`)

**Technology:** Go 1.21+ with gRPC

**Purpose:** Receives and processes log streams from CLI clients

**Key Features:**
- Bidirectional gRPC streaming
- Token-based authentication
- Real-time log writing to Loki
- Metadata storage in PostgreSQL
- Redis pub/sub for real-time notifications

**Data Flow:**
1. Authenticates CLI client via token
2. Creates/retrieves project, group, and run records
3. Streams logs to Loki with labels
4. Updates run status in PostgreSQL
5. Publishes events to Redis

**Ports:**
- gRPC: 50051

### 3. REST API Service (`backend/cmd/api/`)

**Technology:** Go 1.21+ with Gin framework

**Purpose:** HTTP API for web interface and integrations

**Key Features:**
- RESTful endpoints for CRUD operations
- Bearer token authentication
- Query logs from Loki
- Trigger AI analysis
- Health checks

**Endpoints:**
- `/api/v1/projects` - Project management
- `/api/v1/groups/:id` - Group details
- `/api/v1/runs/:id` - Run details and logs
- `/api/v1/runs/:id/analyze` - Trigger AI analysis
- `/health` - Health check

**Ports:**
- HTTP: 8080

### 4. WebSocket Service (`backend/cmd/websocket/`)

**Technology:** Go 1.21+ with gorilla/websocket

**Purpose:** Real-time log streaming to web browsers

**Key Features:**
- WebSocket connections for live log streaming
- Redis pub/sub subscription
- Connection hub management
- Token-based authentication via query parameter

**Data Flow:**
1. Client connects via WebSocket with auth token
2. Subscribes to Redis channel for specific run
3. Receives log events from Ingestor via Redis
4. Forwards logs to connected client

**Ports:**
- WebSocket: 8081

### 5. AI Worker Service (`backend/cmd/ai-worker/`)

**Technology:** Go 1.21+ with background job processing

**Purpose:** Asynchronous AI-powered log analysis

**Key Features:**
- Polls for analysis requests
- Fetches logs from Loki
- Calls OpenAI API for analysis
- Stores results in PostgreSQL
- Configurable with any OpenAI-compatible API

**Configuration:**
- `OPENAI_API_KEY` - API key
- `OPENAI_BASE_URL` - Custom endpoint (optional)
- `OPENAI_MODEL` - Model name (default: gpt-4o-mini)

### 6. Frontend (`frontend/`)

**Technology:** Next.js 14, TypeScript 5, Tailwind CSS

**Purpose:** Web user interface

**Key Features:**
- Project and group navigation
- Real-time log viewing
- AI analysis results display
- API token management (future)
- Responsive design

**Ports:**
- HTTP: 3000

### 7. Infrastructure

#### PostgreSQL 16
**Purpose:** Metadata storage

**Schema:**
- `users` - User accounts
- `api_tokens` - Authentication tokens (SHA-256 hashed)
- `projects` - Top-level containers
- `log_groups` - Organizational units
- `log_runs` - Script execution records

**Migrations:** Auto-applied on startup via `docker-entrypoint-initdb.d`

#### Grafana Loki 2.9
**Purpose:** Log storage and querying

**Features:**
- Efficient log storage with compression
- Label-based indexing
- LogQL query language
- Retention policies

**Labels:**
- `project` - Project name
- `group` - Log group name
- `run_id` - Unique run identifier
- `level` - stdout/stderr

#### Redis 7
**Purpose:** Pub/sub messaging

**Usage:**
- Real-time log event distribution
- Channel: `logs:<run_id>`
- Publisher: Ingestor
- Subscriber: WebSocket server

## Data Flow

### Log Ingestion Flow

```
1. User runs: swiftlog run --project myapp -- ./script.sh

2. CLI:
   - Spawns subprocess
   - Captures stdout/stderr
   - Connects to Ingestor via gRPC
   - Streams log lines

3. Ingestor:
   - Authenticates token
   - Creates/finds project → group → run
   - Writes logs to Loki with labels
   - Updates run status in PostgreSQL
   - Publishes to Redis channel

4. WebSocket (if client connected):
   - Receives from Redis
   - Forwards to browser

5. Frontend:
   - Displays logs in real-time
   - Shows run status and exit code
```

### Log Retrieval Flow

```
1. User opens run in web interface

2. Frontend:
   - Calls REST API: GET /api/v1/runs/:id/logs
   - Opens WebSocket connection for live updates

3. REST API:
   - Queries Loki using LogQL
   - Returns historical logs

4. WebSocket:
   - Subscribes to Redis channel
   - Streams new logs as they arrive

5. Frontend:
   - Displays historical logs
   - Appends new logs in real-time
```

### AI Analysis Flow

```
1. User clicks "Analyze" in web interface

2. Frontend:
   - Calls REST API: POST /api/v1/runs/:id/analyze

3. REST API:
   - Creates analysis request in PostgreSQL
   - Returns immediately (async)

4. AI Worker:
   - Polls for pending requests
   - Fetches logs from Loki
   - Calls OpenAI API
   - Saves report to PostgreSQL

5. Frontend (polling or WebSocket):
   - Detects analysis completion
   - Displays AI-generated report
```

## Project Structure

```
swiftlog/
├── backend/
│   ├── cmd/                      # Service entry points
│   │   ├── ingestor/            # gRPC log ingestor
│   │   ├── api/                 # REST API server
│   │   ├── websocket/           # WebSocket server
│   │   └── ai-worker/           # AI analysis worker
│   ├── internal/                # Internal packages
│   │   ├── auth/                # Token authentication
│   │   ├── database/            # Database connection
│   │   ├── models/              # Data models
│   │   ├── repository/          # Data access layer
│   │   ├── loki/                # Loki client
│   │   ├── ingestor/            # Ingestor service logic
│   │   ├── websocket/           # WebSocket hub
│   │   └── ai/                  # AI analyzer
│   ├── migrations/              # SQL migrations
│   ├── proto/                   # Protobuf definitions
│   └── Dockerfile.*             # Service containers
├── cli/
│   ├── cmd/                     # CLI commands (run, config)
│   ├── internal/
│   │   ├── config/              # Config management
│   │   └── client/              # gRPC client
│   └── proto/                   # Protobuf (generated)
├── frontend/
│   └── src/                     # Next.js application
├── docker-compose.yaml          # All-in-one deployment
├── .env.example                 # Environment template
└── docs/                        # Documentation
```

## Database Schema

### Core Tables

#### users
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### api_tokens
```sql
CREATE TABLE api_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,  -- SHA-256
    name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP
);
```

#### projects
```sql
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
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
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, name)
);
```

#### log_runs
```sql
CREATE TABLE log_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID REFERENCES log_groups(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL,  -- 'running', 'success', 'failed'
    exit_code INT,
    started_at TIMESTAMP NOT NULL,
    ended_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Security Considerations

### Authentication

1. **API Tokens**
   - SHA-256 hashed in database
   - Transmitted via Bearer token (HTTP) or metadata (gRPC)
   - User-specific, revocable

2. **Password Storage**
   - bcrypt hashed
   - Minimum strength requirements

### Network Security

1. **Service Isolation**
   - Infrastructure services (PostgreSQL, Loki, Redis) not exposed to host
   - Only application services exposed on necessary ports
   - Internal Docker network for inter-service communication

2. **CORS Configuration**
   - Configurable allowed origins
   - Production deployment requires explicit CORS_ORIGINS

3. **TLS/SSL**
   - Recommended for production with reverse proxy
   - CLI supports TLS connections

### Data Protection

1. **SQL Injection Prevention**
   - Parameterized queries throughout
   - Input validation

2. **Sensitive Data**
   - API keys encrypted at rest (ENCRYPTION_KEY)
   - JWT secrets for session management
   - Environment variable configuration

## Performance Characteristics

### Targets

- **CLI Overhead**: <5% for scripts running >10 seconds
- **Real-Time Latency**: <2 seconds from log generation to browser
- **API Response Time**: <200ms (p95)
- **Concurrent Streams**: 1000+ simultaneous gRPC connections

### Optimization Strategies

1. **Connection Pooling**
   - PostgreSQL: Configurable pool size
   - Loki: HTTP keep-alive
   - Redis: Connection reuse

2. **Buffering**
   - CLI: Line-buffered streaming
   - Ingestor: Batch writes to Loki
   - WebSocket: Message batching

3. **Caching**
   - Project/group lookups cached in-memory
   - API token validation cached (5 min)

4. **Indexing**
   - Database indexes on frequently queried fields
   - Loki labels for efficient log queries

## Scalability

### Horizontal Scaling

- **Ingestor**: Multiple replicas behind load balancer
- **API**: Stateless, scales horizontally
- **WebSocket**: Sticky sessions required
- **AI Worker**: Scale based on analysis queue depth

### Vertical Scaling

- **PostgreSQL**: Increase memory and CPU for more connections
- **Loki**: Increase storage for more logs
- **Redis**: Increase memory for more concurrent streams

## Monitoring and Observability

### Health Checks

- `/health` endpoints on API and WebSocket
- Container health checks in Docker Compose
- Database connection verification

### Logging

- Structured JSON logging
- Configurable log levels (debug, info, warn, error)
- Service-specific log streams

### Metrics (Future)

- Prometheus integration planned
- Key metrics: request rates, latency, error rates
- Grafana dashboards

## Technology Stack Summary

| Component | Technology | Version |
|-----------|-----------|---------|
| **Backend Services** | Go | 1.21+ |
| **Frontend** | Next.js | 14 |
| **Database** | PostgreSQL | 16 |
| **Log Storage** | Grafana Loki | 2.9 |
| **Pub/Sub** | Redis | 7 |
| **HTTP Framework** | Gin | Latest |
| **WebSocket** | gorilla/websocket | Latest |
| **RPC** | gRPC | Latest |
| **AI** | OpenAI API | gpt-4o-mini |
| **Container** | Docker | 24+ |
| **Orchestration** | Docker Compose | v2+ |

## Related Documentation

- [Getting Started](getting-started)
- [CLI Guide](cli-guide)
- [API Reference](api-reference)
- [Deployment Guide](deployment)
- [Configuration](configuration)