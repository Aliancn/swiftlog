---
layout: default
title: Development
nav_order: 8
description: "Development and contribution guide for SwiftLog"
---

# SwiftLog Development Guide

Guide for developers who want to contribute to SwiftLog or modify it for their needs.

## Prerequisites

### Required Tools

- **Go 1.21+** - Backend services and CLI
- **Node.js 20+** - Frontend development
- **Docker 24+** & **Docker Compose v2+** - Local infrastructure
- **Protocol Buffers** - gRPC code generation
- **Git** - Version control
- **Make** - Build automation

### Installing Protocol Buffers

**macOS:**
```bash
brew install protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

**Linux:**
```bash
# Ubuntu/Debian
sudo apt install -y protobuf-compiler
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

**Windows:**
```powershell
# Using chocolatey
choco install protoc
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

---

## Project Structure

```
swiftlog/
├── backend/                    # Go backend services
│   ├── cmd/                   # Service entry points
│   │   ├── ingestor/         # gRPC log ingestor
│   │   ├── api/              # REST API server
│   │   ├── websocket/        # WebSocket server
│   │   └── ai-worker/        # AI analysis worker
│   ├── internal/             # Internal packages
│   │   ├── auth/             # Authentication
│   │   ├── database/         # Database connection
│   │   ├── models/           # Data models
│   │   ├── repository/       # Data access layer
│   │   ├── loki/             # Loki client
│   │   ├── ingestor/         # Ingestor logic
│   │   ├── websocket/        # WebSocket hub
│   │   └── ai/               # AI analyzer
│   ├── migrations/           # SQL migrations
│   ├── proto/                # Protobuf definitions
│   └── go.mod                # Go dependencies
├── cli/                       # CLI tool
│   ├── cmd/                  # CLI commands
│   ├── internal/             # CLI internals
│   └── proto/                # Generated protobuf code
├── frontend/                  # Next.js frontend
│   ├── src/                  # Source code
│   │   ├── app/             # Next.js 14 App Router
│   │   ├── components/      # React components
│   │   └── lib/             # Utilities
│   └── package.json         # Node dependencies
├── docs/                      # Documentation
├── tests/                     # Integration tests
├── docker-compose.yaml        # Production deployment
├── docker-compose.dev.yaml    # Development deployment
├── Makefile                   # Build automation
└── .env.example              # Environment template
```

---

## Getting Started

### 1. Clone Repository

```bash
git clone https://github.com/aliancn/swiftlog.git
cd swiftlog
```

### 2. Setup Environment

```bash
# Copy environment template
cp .env.example .env

# Edit for development
nano .env
```

**Minimal development configuration:**
```bash
ENVIRONMENT=development
LOG_LEVEL=debug

POSTGRES_PASSWORD=devpassword
JWT_SECRET=dev-jwt-secret-at-least-32-chars-long
ENCRYPTION_KEY=dev-encryption-key-32-chars-abc
```

### 3. Start Infrastructure

Start only the infrastructure services (PostgreSQL, Loki, Redis):

```bash
make dev-up
```

This starts:
- PostgreSQL on port 5432
- Loki on port 3100
- Redis on port 6379

---

## Backend Development

### Running Backend Services Locally

Open multiple terminal windows and run each service:

**Terminal 1 - Ingestor:**
```bash
cd backend/cmd/ingestor
go run main.go
```

**Terminal 2 - API:**
```bash
cd backend/cmd/api
go run main.go
```

**Terminal 3 - WebSocket:**
```bash
cd backend/cmd/websocket
go run main.go
```

**Terminal 4 - AI Worker:**
```bash
cd backend/cmd/ai-worker
go run main.go
```

### Backend Structure

#### Adding a New Endpoint

1. **Define the route** in `backend/cmd/api/main.go`:
```go
api := r.Group("/api/v1")
api.Use(authMiddleware)
{
    api.GET("/your-endpoint", handlers.YourHandler)
}
```

2. **Create the handler** in `backend/internal/handlers/`:
```go
package handlers

func YourHandler(c *gin.Context) {
    // Your logic here
    c.JSON(200, gin.H{"message": "success"})
}
```

3. **Add repository methods** in `backend/internal/repository/` if needed.

#### Adding a New Service

1. Create directory: `backend/cmd/your-service/`
2. Create `main.go` with service logic
3. Add Dockerfile: `backend/Dockerfile.your-service`
4. Update `docker-compose.yaml`

### Running Tests

```bash
# Backend tests
cd backend
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/auth/...
```

### Code Style

**Go:**
- Follow standard Go conventions
- Use `gofmt` for formatting
- Use `golint` for linting

```bash
# Format code
gofmt -w .

# Run linter
golint ./...

# Run vet
go vet ./...
```

---

## Frontend Development

### Running Frontend Locally

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

Frontend will be available at `http://localhost:3000`.

### Frontend Structure

```
frontend/src/
├── app/                    # Next.js 14 App Router
│   ├── layout.tsx         # Root layout
│   ├── page.tsx           # Home page
│   ├── projects/          # Projects pages
│   └── runs/              # Runs pages
├── components/            # React components
│   ├── LogViewer.tsx     # Log display component
│   ├── ProjectList.tsx   # Project list
│   └── ...
└── lib/                   # Utilities
    ├── api.ts            # API client
    └── utils.ts          # Helper functions
```

### Adding a New Page

1. Create file in `frontend/src/app/your-page/page.tsx`:
```tsx
export default function YourPage() {
  return <div>Your content</div>;
}
```

2. Add navigation link in layout or components.

### Adding a New Component

1. Create file in `frontend/src/components/YourComponent.tsx`:
```tsx
export function YourComponent() {
  return <div>Your component</div>;
}
```

2. Import and use in pages:
```tsx
import { YourComponent } from '@/components/YourComponent';
```

### Frontend Tests

```bash
# Run tests
npm test

# Run with coverage
npm test -- --coverage

# Type checking
npm run type-check
```

### Code Style

**TypeScript/React:**
- Use TypeScript for type safety
- Follow ESLint rules
- Use Prettier for formatting

```bash
# Lint code
npm run lint

# Format code
npm run format
```

---

## CLI Development

### Building CLI

```bash
cd cli

# Build
go build -o swiftlog

# Install locally
make cli
```

### Testing CLI

```bash
# Configure CLI
./swiftlog config set --token test-token --server localhost:50051

# Test run command
./swiftlog run --project test --group dev -- echo "Hello"
```

### CLI Structure

```
cli/
├── cmd/
│   ├── root.go          # Root command
│   ├── run.go           # Run command
│   ├── config.go        # Config command
│   └── version.go       # Version command
├── internal/
│   ├── config/          # Config management
│   │   └── config.go
│   └── client/          # gRPC client
│       └── client.go
└── proto/               # Generated protobuf
```

---

## Protocol Buffers

### Regenerating Protobuf Code

When you modify `.proto` files:

**Backend:**
```bash
cd backend
protoc --go_out=. --go-grpc_out=. proto/ingestor.proto
```

**CLI:**
```bash
cd cli
protoc --go_out=. --go-grpc_out=. proto/ingestor.proto
```

### Proto File Location

- Source: `backend/proto/ingestor.proto`
- Generated (backend): `backend/proto/*.pb.go`
- Generated (CLI): `cli/proto/*.pb.go`

---

## Database Migrations

### Adding a New Migration

1. Create file in `backend/migrations/`:
```sql
-- backend/migrations/006_add_your_table.sql

CREATE TABLE IF NOT EXISTS your_table (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

2. Update `backend/migrations/init.sql` to include your migration.

### Running Migrations

Migrations are automatically run on startup via Docker's `docker-entrypoint-initdb.d`.

For manual migration:
```bash
docker compose exec postgres psql -U swiftlog -d swiftlog -f /docker-entrypoint-initdb.d/006_add_your_table.sql
```

---

## Docker Development

### Building Images

```bash
# Build all images
make build

# Build specific service
docker compose build api

# Build with no cache
docker compose build --no-cache
```

### Viewing Logs

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f api

# Last 100 lines
docker compose logs --tail=100 api
```

### Accessing Services

```bash
# PostgreSQL shell
docker compose exec postgres psql -U swiftlog -d swiftlog

# Redis CLI
docker compose exec redis redis-cli

# Check Loki
curl http://localhost:3100/ready
```

---

## Testing

### Integration Tests

Run the included test suite:

```bash
cd tests
./run_all_tests.sh
```

Individual tests:
```bash
./cli/swiftlog run --project test --group simple -- bash tests/01_simple_test.sh
```

### Writing New Tests

1. Create test script in `tests/`:
```bash
#!/bin/bash
echo "Running test..."
exit 0
```

2. Make executable:
```bash
chmod +x tests/05_your_test.sh
```

3. Add to `tests/run_all_tests.sh`.

---

## Debugging

### Debug Backend Services

Use Delve debugger:

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug API service
cd backend/cmd/api
dlv debug

# Run with breakpoints
(dlv) break main.main
(dlv) continue
```

### Debug with VS Code

Create `.vscode/launch.json`:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug API",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/backend/cmd/api",
      "env": {
        "LOG_LEVEL": "debug"
      }
    }
  ]
}
```

### View Real-Time Logs

```bash
# Backend service logs
docker compose logs -f api

# Database queries (if logging enabled)
docker compose logs -f postgres

# All logs
docker compose logs -f
```

---

## Git Workflow

### Branch Strategy

- `main` - Stable production branch
- `develop` - Development branch
- `feature/*` - Feature branches
- `bugfix/*` - Bug fix branches

### Commit Messages

Follow conventional commits:

```
feat: add new API endpoint for user settings
fix: resolve database connection leak
docs: update CLI documentation
refactor: simplify authentication logic
test: add integration tests for API
```

### Pull Request Process

1. Create feature branch:
```bash
git checkout -b feature/your-feature
```

2. Make changes and commit:
```bash
git add .
git commit -m "feat: add your feature"
```

3. Push and create PR:
```bash
git push origin feature/your-feature
```

4. Wait for review and CI checks.

---

## Makefile Commands

```bash
make help           # Show all commands
make dev-up         # Start infrastructure only
make dev-down       # Stop infrastructure
make start          # Start all services
make stop           # Stop all services
make restart        # Restart all services
make build          # Build all Docker images
make cli            # Build CLI tool
make clean          # Remove containers and volumes
make logs           # View all logs
make test           # Run tests
```

---

## Common Development Tasks

### Adding a New Feature

1. Create feature branch
2. Add backend endpoint/logic
3. Update frontend UI
4. Add tests
5. Update documentation
6. Create PR

### Fixing a Bug

1. Reproduce the bug
2. Create bugfix branch
3. Fix the issue
4. Add test to prevent regression
5. Create PR

### Updating Dependencies

**Backend:**
```bash
cd backend
go get -u ./...
go mod tidy
```

**Frontend:**
```bash
cd frontend
npm update
npm audit fix
```

**CLI:**
```bash
cd cli
go get -u ./...
go mod tidy
```

---

## CI/CD

SwiftLog uses GitHub Actions for CI/CD. See `.github/workflows/`:

- `release.yml` - Build and release
- `deploy.yml` - Deploy to servers
- `test.yml` - Run tests (future)

### Local CI Testing

Use [act](https://github.com/nektos/act) to test workflows locally:

```bash
# Install act
brew install act

# Run workflow
act push
```

---

## Code Review Guidelines

When reviewing PRs:

1. **Functionality** - Does it work as intended?
2. **Tests** - Are there tests?
3. **Documentation** - Is it documented?
4. **Code Quality** - Is it clean and maintainable?
5. **Security** - Are there security concerns?
6. **Performance** - Are there performance implications?

---

## Related Documentation

- [Getting Started](getting-started)
- [Architecture](architecture)
- [API Reference](api-reference)
- [Configuration](configuration)
- [Deployment Guide](deployment)