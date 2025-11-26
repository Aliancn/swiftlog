---
layout: default
title: API Reference
nav_order: 5
description: "Complete REST API and WebSocket API reference for SwiftLog"
---

# SwiftLog API Reference

Complete reference for the SwiftLog REST API and WebSocket API.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

All API endpoints require authentication using Bearer tokens.

### Header Format

```http
Authorization: Bearer YOUR_API_TOKEN
```

### Creating API Tokens

**For testing (via database):**
```bash
docker compose exec postgres psql -U swiftlog -d swiftlog -c \
  "INSERT INTO api_tokens (user_id, token_hash, name)
   SELECT id, encode(sha256('your-token'::bytea), 'hex'), 'My Token'
   FROM users LIMIT 1
   RETURNING id;"
```

**For production:** Use the web interface (future) or API endpoint to create tokens securely.

## REST API Endpoints

### Health Check

#### GET /health

Check API service health.

**Parameters:** None

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2025-11-26T12:00:00Z"
}
```

**Status Codes:**
- `200 OK` - Service is healthy

---

### Projects

#### GET /api/v1/projects

List all projects.

**Parameters:**
- `limit` (query, optional): Maximum number of results (default: 50)
- `offset` (query, optional): Number of results to skip (default: 0)

**Response:**
```json
{
  "projects": [
    {
      "id": "uuid",
      "name": "myapp",
      "description": "My application logs",
      "created_at": "2025-11-26T10:00:00Z",
      "updated_at": "2025-11-26T10:00:00Z"
    }
  ],
  "total": 1,
  "limit": 50,
  "offset": 0
}
```

**Status Codes:**
- `200 OK` - Success
- `401 Unauthorized` - Invalid or missing token

---

#### GET /api/v1/projects/:id

Get a single project by ID.

**Parameters:**
- `id` (path): Project UUID

**Response:**
```json
{
  "id": "uuid",
  "name": "myapp",
  "description": "My application logs",
  "created_at": "2025-11-26T10:00:00Z",
  "updated_at": "2025-11-26T10:00:00Z"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Project not found
- `401 Unauthorized` - Invalid token

---

#### POST /api/v1/projects

Create a new project.

**Request Body:**
```json
{
  "name": "myapp",
  "description": "My application logs"
}
```

**Response:**
```json
{
  "id": "uuid",
  "name": "myapp",
  "description": "My application logs",
  "created_at": "2025-11-26T10:00:00Z",
  "updated_at": "2025-11-26T10:00:00Z"
}
```

**Status Codes:**
- `201 Created` - Success
- `400 Bad Request` - Invalid input
- `409 Conflict` - Project name already exists
- `401 Unauthorized` - Invalid token

---

#### GET /api/v1/projects/:id/groups

List groups within a project.

**Parameters:**
- `id` (path): Project UUID
- `limit` (query, optional): Maximum number of results (default: 50)
- `offset` (query, optional): Number of results to skip (default: 0)

**Response:**
```json
{
  "groups": [
    {
      "id": "uuid",
      "project_id": "uuid",
      "name": "build",
      "description": "Build logs",
      "created_at": "2025-11-26T10:00:00Z"
    }
  ],
  "total": 1,
  "limit": 50,
  "offset": 0
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Project not found
- `401 Unauthorized` - Invalid token

---

### Groups

#### GET /api/v1/groups/:id

Get a single group by ID.

**Parameters:**
- `id` (path): Group UUID

**Response:**
```json
{
  "id": "uuid",
  "project_id": "uuid",
  "name": "build",
  "description": "Build logs",
  "created_at": "2025-11-26T10:00:00Z"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Group not found
- `401 Unauthorized` - Invalid token

---

#### GET /api/v1/groups/:id/runs

List runs within a group.

**Parameters:**
- `id` (path): Group UUID
- `limit` (query, optional): Maximum number of results (default: 50)
- `offset` (query, optional): Number of results to skip (default: 0)
- `status` (query, optional): Filter by status (running, success, failed)

**Response:**
```json
{
  "runs": [
    {
      "id": "uuid",
      "group_id": "uuid",
      "status": "success",
      "exit_code": 0,
      "started_at": "2025-11-26T10:00:00Z",
      "ended_at": "2025-11-26T10:05:00Z",
      "created_at": "2025-11-26T10:00:00Z"
    }
  ],
  "total": 1,
  "limit": 50,
  "offset": 0
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Group not found
- `401 Unauthorized` - Invalid token

---

### Runs

#### GET /api/v1/runs/:id

Get a single run by ID.

**Parameters:**
- `id` (path): Run UUID

**Response:**
```json
{
  "id": "uuid",
  "group_id": "uuid",
  "status": "success",
  "exit_code": 0,
  "started_at": "2025-11-26T10:00:00Z",
  "ended_at": "2025-11-26T10:05:00Z",
  "created_at": "2025-11-26T10:00:00Z"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Run not found
- `401 Unauthorized` - Invalid token

---

#### GET /api/v1/runs/:id/logs

Get logs for a specific run.

**Parameters:**
- `id` (path): Run UUID
- `limit` (query, optional): Maximum number of log lines (default: 1000)
- `start` (query, optional): Start time (RFC3339)
- `end` (query, optional): End time (RFC3339)

**Response:**
```json
{
  "run_id": "uuid",
  "logs": [
    {
      "timestamp": "2025-11-26T10:00:01Z",
      "level": "stdout",
      "content": "Starting process..."
    },
    {
      "timestamp": "2025-11-26T10:00:02Z",
      "level": "stdout",
      "content": "Process completed successfully"
    }
  ],
  "total": 2
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Run not found
- `401 Unauthorized` - Invalid token
- `503 Service Unavailable` - Loki unavailable

---

#### POST /api/v1/runs/:id/analyze

Trigger AI analysis for a run.

**Parameters:**
- `id` (path): Run UUID

**Request Body:** (optional)
```json
{
  "prompt": "Focus on errors and performance issues"
}
```

**Response:**
```json
{
  "run_id": "uuid",
  "status": "pending",
  "created_at": "2025-11-26T10:00:00Z"
}
```

**Status Codes:**
- `202 Accepted` - Analysis queued
- `404 Not Found` - Run not found
- `409 Conflict` - Analysis already in progress
- `401 Unauthorized` - Invalid token

---

#### GET /api/v1/runs/:id/analysis

Get AI analysis results for a run.

**Parameters:**
- `id` (path): Run UUID

**Response:**
```json
{
  "run_id": "uuid",
  "status": "completed",
  "report": "## Analysis Summary\n\nThe script executed successfully...",
  "created_at": "2025-11-26T10:00:00Z",
  "completed_at": "2025-11-26T10:00:30Z"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Run or analysis not found
- `401 Unauthorized` - Invalid token

---

## WebSocket API

### Connection

Connect to the WebSocket server for real-time log streaming.

**URL:**
```
ws://localhost:8081/ws/runs/:run_id?token=YOUR_TOKEN
```

**Parameters:**
- `run_id` (path): Run UUID to subscribe to
- `token` (query): API authentication token

### Message Format

**Server → Client (Log Events):**
```json
{
  "type": "log",
  "run_id": "uuid",
  "timestamp": "2025-11-26T10:00:00Z",
  "level": "stdout",
  "content": "log line content"
}
```

**Server → Client (Status Updates):**
```json
{
  "type": "status",
  "run_id": "uuid",
  "status": "completed",
  "exit_code": 0
}
```

**Server → Client (Error):**
```json
{
  "type": "error",
  "message": "Error description"
}
```

### Example (JavaScript)

```javascript
const runId = 'your-run-uuid';
const token = 'your-api-token';
const ws = new WebSocket(`ws://localhost:8081/ws/runs/${runId}?token=${token}`);

ws.onopen = () => {
  console.log('Connected to log stream');
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);

  switch (data.type) {
    case 'log':
      console.log(`[${data.level}] ${data.content}`);
      break;
    case 'status':
      console.log(`Run ${data.status} with exit code ${data.exit_code}`);
      break;
    case 'error':
      console.error('Error:', data.message);
      break;
  }
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('Connection closed');
};
```

### Connection Lifecycle

1. **Connect**: Client connects with auth token
2. **Subscribe**: Server subscribes to Redis channel for the run
3. **Stream**: Server forwards logs from Redis to client
4. **Disconnect**: Connection closes when client disconnects or run completes

---

## Error Responses

All API errors follow this format:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {}
}
```

### Common Error Codes

| Status Code | Error Code | Description |
|-------------|-----------|-------------|
| 400 | `BAD_REQUEST` | Invalid request parameters |
| 401 | `UNAUTHORIZED` | Missing or invalid authentication token |
| 403 | `FORBIDDEN` | Insufficient permissions |
| 404 | `NOT_FOUND` | Resource not found |
| 409 | `CONFLICT` | Resource already exists or conflict |
| 429 | `RATE_LIMIT` | Too many requests |
| 500 | `INTERNAL_ERROR` | Server error |
| 503 | `SERVICE_UNAVAILABLE` | Service temporarily unavailable |

---

## Rate Limiting

Currently, there is no rate limiting implemented. For production deployments, consider adding rate limiting at the reverse proxy level (e.g., nginx).

---

## CORS Configuration

Configure allowed origins via the `CORS_ORIGINS` environment variable:

```bash
# Single origin
CORS_ORIGINS=https://logs.example.com

# Multiple origins
CORS_ORIGINS=https://logs.example.com,https://app.example.com

# Development
CORS_ORIGINS=http://localhost:3000
```

---

## API Versioning

The API uses URL-based versioning (`/api/v1/`). Breaking changes will be introduced in new versions (e.g., `/api/v2/`) while maintaining backward compatibility with previous versions.

---

## Examples

### Complete Workflow Example

```bash
# 1. Run a script with CLI
swiftlog run --project myapp --group build -- ./build.sh

# 2. List projects
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/projects

# 3. Get project groups
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/projects/{project_id}/groups

# 4. Get runs in a group
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/groups/{group_id}/runs

# 5. Get logs for a run
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/runs/{run_id}/logs

# 6. Trigger AI analysis
curl -X POST -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/runs/{run_id}/analyze

# 7. Get analysis results
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/runs/{run_id}/analysis
```

### Create and Track a New Project

```bash
# Create project
curl -X POST -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"newproject","description":"My new project"}' \
  http://localhost:8080/api/v1/projects

# Run a command in the new project
swiftlog run --project newproject --group tests -- npm test

# Watch logs in real-time (JavaScript)
# See WebSocket example above
```

---

## Related Documentation

- [Getting Started](getting-started)
- [CLI Guide](cli-guide)
- [Architecture](architecture)
- [Configuration](configuration)