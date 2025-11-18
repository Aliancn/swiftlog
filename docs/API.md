# SwiftLog API Documentation

Complete REST API reference for SwiftLog platform.

## Table of Contents

- [Overview](#overview)
- [Authentication](#authentication)
- [Base URL](#base-url)
- [Response Format](#response-format)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Endpoints](#endpoints)
  - [Health Check](#health-check)
  - [Authentication](#authentication-endpoints)
  - [Projects](#projects)
  - [Groups](#groups)
  - [Runs](#runs)
  - [Logs](#logs)
  - [AI Analysis](#ai-analysis)
- [WebSocket API](#websocket-api)
- [Examples](#examples)

## Overview

The SwiftLog REST API provides programmatic access to log data, projects, groups, and runs. All API endpoints return JSON responses.

**API Version**: v1
**Protocol**: HTTP/HTTPS
**Format**: JSON

## Authentication

All API requests (except health check) require authentication using Bearer tokens.

### Obtaining a Token

**For Testing:**
```bash
# Insert test token into database
docker compose exec postgres psql -U swiftlog -d swiftlog -c \
  "INSERT INTO api_tokens (user_id, token_hash, name)
   SELECT id, encode(sha256('test-token'::bytea), 'hex'), 'Test Token'
   FROM users LIMIT 1
   RETURNING id, name;"
```

**For Production:**
Use the web interface or create via API (requires existing token).

### Using Tokens

Include the token in the `Authorization` header:

```http
Authorization: Bearer YOUR_API_TOKEN
```

**Example:**
```bash
curl -H "Authorization: Bearer test-token" \
  http://localhost:8080/api/v1/projects
```

## Base URL

**Development:** `http://localhost:8080/api/v1`
**Production:** `https://your-domain.com/api/v1`

## Response Format

### Success Response

```json
{
  "data": {
    "id": "uuid",
    "name": "example"
  }
}
```

Or for lists:

```json
{
  "data": [...],
  "total": 100,
  "limit": 50,
  "offset": 0
}
```

### Error Response

```json
{
  "error": "Error message describing what went wrong"
}
```

**HTTP Status Codes:**
- `200 OK` - Request succeeded
- `201 Created` - Resource created
- `400 Bad Request` - Invalid request parameters
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - Not authorized to access resource
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

## Error Handling

All errors return a JSON object with an `error` field and appropriate HTTP status code.

**Example:**
```json
{
  "error": "Project not found"
}
```

## Rate Limiting

**Current Status:** Not implemented

**Future Plans:**
- 1000 requests per hour per token
- Headers: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`

## Endpoints

### Health Check

#### GET /health

Check API server health (no authentication required).

**Request:**
```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "status": "healthy"
}
```

---

### Authentication Endpoints

#### POST /api/v1/auth/register

Register a new user.

**Request:**
```json
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "secure_password123"
}
```

**Response:**
```json
{
  "id": "uuid",
  "username": "john_doe",
  "email": "john@example.com"
}
```

#### POST /api/v1/auth/login

Authenticate and receive JWT token.

**Request:**
```json
{
  "email": "john@example.com",
  "password": "secure_password123"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "username": "john_doe",
    "email": "john@example.com"
  }
}
```

#### POST /api/v1/auth/tokens

Create a new API token (requires authentication).

**Request:**
```json
{
  "name": "CI/CD Token"
}
```

**Response:**
```json
{
  "token": "generated-api-token-here",
  "name": "CI/CD Token"
}
```

**Note:** The raw token is only shown once. Store it securely.

---

### Projects

#### GET /api/v1/projects

List all projects for the authenticated user.

**Query Parameters:**
- `limit` (optional): Number of results (default: 50, max: 100)
- `offset` (optional): Pagination offset (default: 0)

**Request:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:8080/api/v1/projects?limit=10&offset=0"
```

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "my-project",
      "user_id": "uuid",
      "created_at": "2025-11-18T10:00:00Z",
      "updated_at": "2025-11-18T10:00:00Z"
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0
}
```

#### GET /api/v1/projects/:id

Get a specific project by ID.

**Request:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/projects/{project-id}
```

**Response:**
```json
{
  "id": "uuid",
  "name": "my-project",
  "user_id": "uuid",
  "created_at": "2025-11-18T10:00:00Z",
  "updated_at": "2025-11-18T10:00:00Z"
}
```

#### POST /api/v1/projects

Create a new project.

**Request:**
```json
{
  "name": "new-project"
}
```

**Response:**
```json
{
  "id": "uuid",
  "name": "new-project",
  "user_id": "uuid",
  "created_at": "2025-11-18T10:00:00Z",
  "updated_at": "2025-11-18T10:00:00Z"
}
```

#### GET /api/v1/projects/:id/groups

List all groups in a project.

**Request:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/projects/{project-id}/groups
```

**Response:**
```json
[
  {
    "id": "uuid",
    "project_id": "uuid",
    "name": "build",
    "created_at": "2025-11-18T10:00:00Z",
    "updated_at": "2025-11-18T10:00:00Z"
  },
  {
    "id": "uuid",
    "project_id": "uuid",
    "name": "tests",
    "created_at": "2025-11-18T10:00:00Z",
    "updated_at": "2025-11-18T10:00:00Z"
  }
]
```

---

### Groups

#### GET /api/v1/groups/:id

Get a specific group by ID.

**Request:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/groups/{group-id}
```

**Response:**
```json
{
  "id": "uuid",
  "project_id": "uuid",
  "name": "build",
  "created_at": "2025-11-18T10:00:00Z",
  "updated_at": "2025-11-18T10:00:00Z"
}
```

---

### Runs

#### GET /api/v1/groups/:id/runs

List all runs in a group.

**Query Parameters:**
- `limit` (optional): Number of results (default: 50, max: 100)
- `offset` (optional): Pagination offset (default: 0)

**Request:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:8080/api/v1/groups/{group-id}/runs?limit=20&offset=0"
```

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "group_id": "uuid",
      "start_time": "2025-11-18T10:00:00Z",
      "end_time": "2025-11-18T10:05:00Z",
      "exit_code": 0,
      "status": "success",
      "ai_status": "pending",
      "ai_report": null,
      "created_at": "2025-11-18T10:00:00Z"
    }
  ],
  "total": 1,
  "limit": 20,
  "offset": 0
}
```

**Fields:**
- `status`: `running`, `success`, `failed`
- `ai_status`: `pending`, `processing`, `completed`, `failed`
- `exit_code`: null if still running
- `end_time`: null if still running
- `ai_report`: markdown text, null if not generated

#### GET /api/v1/runs/:id

Get details of a specific run.

**Request:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/runs/{run-id}
```

**Response:**
```json
{
  "id": "uuid",
  "group_id": "uuid",
  "start_time": "2025-11-18T10:00:00Z",
  "end_time": "2025-11-18T10:05:00Z",
  "exit_code": 0,
  "status": "success",
  "ai_status": "completed",
  "ai_report": "# Analysis Report\n\nThe script executed successfully...",
  "created_at": "2025-11-18T10:00:00Z"
}
```

---

### Logs

#### GET /api/v1/runs/:id/logs

Retrieve log entries for a run from Loki.

**Query Parameters:**
- `limit` (optional): Max entries to return (default: 1000, max: 10000)
- `start` (optional): Start time (RFC3339)
- `end` (optional): End time (RFC3339)

**Request:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:8080/api/v1/runs/{run-id}/logs?limit=100"
```

**Response:**
```json
[
  {
    "timestamp": "2025-11-18T10:00:01Z",
    "level": "STDOUT",
    "content": "Starting process..."
  },
  {
    "timestamp": "2025-11-18T10:00:02Z",
    "level": "STDOUT",
    "content": "Processing item 1"
  },
  {
    "timestamp": "2025-11-18T10:00:03Z",
    "level": "STDERR",
    "content": "Warning: Low memory"
  }
]
```

**Log Levels:**
- `STDOUT` - Standard output
- `STDERR` - Standard error

---

### AI Analysis

#### POST /api/v1/runs/:id/analyze

Trigger AI analysis for a run.

**Request:**
```bash
curl -X POST \
  -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/runs/{run-id}/analyze
```

**Response:**
```json
{
  "message": "AI analysis queued",
  "run_id": "uuid"
}
```

**Process:**
1. Request is queued in Redis
2. AI Worker picks up the job
3. Fetches logs from Loki
4. Sends to OpenAI for analysis
5. Saves report to database
6. Updates run status

**Check Status:**

Poll the run endpoint to check AI analysis progress:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/runs/{run-id}
```

When `ai_status` is `completed`, the `ai_report` field will contain the markdown analysis.

---

## WebSocket API

Real-time log streaming via WebSocket.

### Connect to Run Stream

**Endpoint:** `ws://localhost:8081/ws/runs/:run_id?token=YOUR_TOKEN`

**JavaScript Example:**
```javascript
const runId = 'your-run-id';
const token = 'your-api-token';
const ws = new WebSocket(`ws://localhost:8081/ws/runs/${runId}?token=${token}`);

ws.onopen = () => {
  console.log('Connected to log stream');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log(message);
  // {
  //   "type": "log",
  //   "run_id": "uuid",
  //   "timestamp": "2025-11-18T10:00:00Z",
  //   "level": "STDOUT",
  //   "content": "log line content"
  // }
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('Disconnected from log stream');
};
```

**Message Types:**

1. **Log Message**
```json
{
  "type": "log",
  "run_id": "uuid",
  "timestamp": "2025-11-18T10:00:00Z",
  "level": "STDOUT",
  "content": "log line"
}
```

2. **Status Update**
```json
{
  "type": "status",
  "run_id": "uuid",
  "status": "completed",
  "exit_code": 0
}
```

3. **Error Message**
```json
{
  "type": "error",
  "run_id": "uuid",
  "error": "Connection lost"
}
```

---

## Examples

### Complete Workflow Example

```bash
# 1. Create a project
PROJECT_ID=$(curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"my-app"}' \
  http://localhost:8080/api/v1/projects | jq -r '.id')

echo "Created project: $PROJECT_ID"

# 2. Run a command (CLI creates group and run automatically)
./swiftlog run --project my-app --group build -- ./build.sh

# 3. List runs in the group
GROUP_ID=$(curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/projects/$PROJECT_ID/groups" | jq -r '.[0].id')

RUN_ID=$(curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/groups/$GROUP_ID/runs" | jq -r '.data[0].id')

echo "Latest run: $RUN_ID"

# 4. Get logs
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/runs/$RUN_ID/logs?limit=100"

# 5. Trigger AI analysis
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/runs/$RUN_ID/analyze"

# 6. Check analysis status
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/runs/$RUN_ID" | jq '.ai_status'

# 7. Get analysis report (when completed)
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/runs/$RUN_ID" | jq -r '.ai_report'
```

### Python Client Example

```python
import requests

class SwiftLogClient:
    def __init__(self, base_url, token):
        self.base_url = base_url
        self.headers = {"Authorization": f"Bearer {token}"}

    def create_project(self, name):
        response = requests.post(
            f"{self.base_url}/projects",
            json={"name": name},
            headers=self.headers
        )
        return response.json()

    def list_projects(self):
        response = requests.get(
            f"{self.base_url}/projects",
            headers=self.headers
        )
        return response.json()["data"]

    def get_run(self, run_id):
        response = requests.get(
            f"{self.base_url}/runs/{run_id}",
            headers=self.headers
        )
        return response.json()

    def get_logs(self, run_id, limit=1000):
        response = requests.get(
            f"{self.base_url}/runs/{run_id}/logs",
            params={"limit": limit},
            headers=self.headers
        )
        return response.json()

    def trigger_analysis(self, run_id):
        response = requests.post(
            f"{self.base_url}/runs/{run_id}/analyze",
            headers=self.headers
        )
        return response.json()

# Usage
client = SwiftLogClient("http://localhost:8080/api/v1", "your-token")

# Create project
project = client.create_project("my-new-project")
print(f"Created project: {project['id']}")

# List projects
projects = client.list_projects()
for p in projects:
    print(f"- {p['name']}")

# Get run logs
logs = client.get_logs("run-id-here", limit=100)
for log in logs:
    print(f"[{log['timestamp']}] {log['level']}: {log['content']}")
```

### Node.js Client Example

```javascript
const axios = require('axios');

class SwiftLogClient {
  constructor(baseURL, token) {
    this.client = axios.create({
      baseURL,
      headers: { 'Authorization': `Bearer ${token}` }
    });
  }

  async createProject(name) {
    const response = await this.client.post('/projects', { name });
    return response.data;
  }

  async listProjects() {
    const response = await this.client.get('/projects');
    return response.data.data;
  }

  async getRun(runId) {
    const response = await this.client.get(`/runs/${runId}`);
    return response.data;
  }

  async getLogs(runId, limit = 1000) {
    const response = await this.client.get(`/runs/${runId}/logs`, {
      params: { limit }
    });
    return response.data;
  }

  async triggerAnalysis(runId) {
    const response = await this.client.post(`/runs/${runId}/analyze`);
    return response.data;
  }
}

// Usage
const client = new SwiftLogClient('http://localhost:8080/api/v1', 'your-token');

(async () => {
  // Create project
  const project = await client.createProject('my-new-project');
  console.log(`Created project: ${project.id}`);

  // List projects
  const projects = await client.listProjects();
  projects.forEach(p => console.log(`- ${p.name}`));

  // Get run logs
  const logs = await client.getLogs('run-id-here', 100);
  logs.forEach(log => {
    console.log(`[${log.timestamp}] ${log.level}: ${log.content}`);
  });
})();
```

---

## Testing the API

### Using curl

```bash
# Set your token
export TOKEN="your-api-token"

# Test health endpoint
curl http://localhost:8080/health

# List projects
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/projects

# Create a project
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"test-project"}' \
  http://localhost:8080/api/v1/projects
```

### Using HTTPie

```bash
# Install httpie
pip install httpie

# Test endpoints
http localhost:8080/health

http localhost:8080/api/v1/projects \
  Authorization:"Bearer $TOKEN"

http POST localhost:8080/api/v1/projects \
  Authorization:"Bearer $TOKEN" \
  name="test-project"
```

### Using Postman

1. Import the SwiftLog API collection (coming soon)
2. Set environment variable `TOKEN` to your API token
3. Set environment variable `BASE_URL` to `http://localhost:8080/api/v1`
4. Run requests

---

## API Versioning

**Current Version:** v1

The API version is included in the URL path: `/api/v1/...`

Future versions will be released as `/api/v2/...` with backwards compatibility maintained for v1.

---

## Support

- **Documentation**: [Main README](../README.md)
- **CLI Documentation**: [cli/README.md](../cli/README.md)
- **Issues**: GitHub Issues
- **Discussions**: GitHub Discussions

---

## Changelog

### v1.0.0 (2025-11-18)
- Initial API release
- Projects, Groups, Runs endpoints
- Log retrieval from Loki
- AI analysis trigger
- WebSocket streaming
- Token authentication
