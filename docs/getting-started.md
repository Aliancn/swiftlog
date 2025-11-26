---
layout: default
title: Getting Started
nav_order: 2
description: "Quick start guide for SwiftLog - get up and running in minutes"
---

# Getting Started with SwiftLog

This guide will help you get SwiftLog up and running in minutes.

## Prerequisites

Before you begin, ensure you have:

- **Docker** 24+ & **Docker Compose** v2+
- **Go** 1.21+ (only for building CLI tool)
- **Git**

## Quick Start

### 1. Clone and Setup

```bash
git clone <repository-url>
cd swiftlog

# Copy environment template
cp .env.example .env
```

### 2. Configure Environment Variables

Edit `.env` and set your values:

**Required (Production):**
```bash
# Database password
POSTGRES_PASSWORD=your-secure-password

# JWT secret (at least 32 characters)
JWT_SECRET=$(openssl rand -base64 32)

# API key encryption key
ENCRYPTION_KEY=$(openssl rand -base64 32)

# CORS allowed origins
CORS_ORIGINS=https://your-domain.com
```

**Optional (AI Features):**
```bash
# OpenAI API key for AI analysis
OPENAI_API_KEY=sk-your-openai-key-here
```

See [Configuration Guide](configuration) for all available options.

### 3. Start All Services

**Using Makefile (Recommended):**
```bash
make start
```

**Or using Docker Compose directly:**
```bash
docker compose up -d
```

This will start:
- PostgreSQL database with auto-migrated schema
- Grafana Loki for log storage
- Redis for pub/sub
- Ingestor (gRPC server)
- REST API server
- WebSocket server
- AI Worker
- Frontend web application

### 4. Verify Services

```bash
# Check all services are running
docker compose ps

# Test API health
curl http://localhost:8080/health

# View logs
docker compose logs -f
```

### 5. Build and Configure CLI

```bash
# Build CLI tool
make cli

# Configure CLI
./cli/swiftlog config set --token YOUR_API_TOKEN --server localhost:50051
```

**Note:** You'll need to create an API token first. See [Configuration Guide](configuration#creating-api-tokens) for details.

### 6. Run Your First Command

```bash
cd cli
./swiftlog run --project myapp --group tests -- echo "Hello, SwiftLog!"
```

## Access Points

Once started, access SwiftLog at:

| Service | URL | Description |
|---------|-----|-------------|
| **Frontend** | http://localhost:3000 | Web interface |
| **REST API** | http://localhost:8080 | HTTP API |
| **WebSocket** | ws://localhost:8081 | Real-time streaming |
| **gRPC Ingestor** | localhost:50051 | CLI connection |

## Basic Usage Examples

### Run a Shell Script

```bash
swiftlog run --project webapp --group build -- ./build.sh
```

### Run a Python Script

```bash
swiftlog run --project ml --group training -- python train_model.py
```

### Run with Auto-Detection

```bash
# In a git repository, project name is auto-detected
swiftlog run -- npm test
```

### Complex Command with Pipes

```bash
swiftlog run --project data -- bash -c "cat data.csv | grep ERROR | wc -l"
```

## What's Next?

- **Learn more about CLI**: [CLI Guide](cli-guide)
- **Explore the API**: [API Reference](api-reference)
- **Deploy to production**: [Deployment Guide](deployment)
- **Understand the architecture**: [Architecture](architecture)
- **Configure advanced options**: [Configuration Guide](configuration)
## Common Issues

### Services won't start

```bash
# Check Docker is running
docker info

# View logs for errors
docker compose logs

# Restart everything
docker compose down && docker compose up -d
```

### CLI can't connect

```bash
# Verify Ingestor is running
docker compose ps ingestor

# Check Ingestor logs
docker compose logs ingestor

# Test connectivity
telnet localhost 50051
```

### Missing API token

You need to create an API token first. For testing:

```bash
docker compose exec postgres psql -U swiftlog -d swiftlog -c \
  "INSERT INTO api_tokens (user_id, token_hash, name)
   SELECT id, encode(sha256('test-token'::bytea), 'hex'), 'CLI Test Token'
   FROM users LIMIT 1
   RETURNING id;"
```

Then configure CLI:
```bash
./swiftlog config set --token test-token --server localhost:50051
```

For production, use the web interface or API to create tokens securely.

## Development Mode

For local development with hot-reload:

```bash
# Start only infrastructure (postgres, loki, redis)
make dev-up

# Then run services locally in separate terminals:
cd backend/cmd/ingestor && go run main.go
cd backend/cmd/api && go run main.go
cd backend/cmd/websocket && go run main.go
cd backend/cmd/ai-worker && go run main.go
cd frontend && npm run dev

# When done:
make dev-down
```

See [Development Guide](development) for more details.

## Need Help?

- **Documentation Index**: [docs/index](index)
- **API Documentation**: [api-reference](api-reference)
- **CLI Documentation**: [cli-guide](cli-guide)
- **GitHub Issues**: https://github.com/aliancn/swiftlog/issues
