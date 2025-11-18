# SwiftLog - Quick Start Guide

Get SwiftLog up and running in **5 minutes**! 🚀

## Step 1: Prerequisites ✅

Ensure you have:
- Docker 24+ & Docker Compose v2+
- Go 1.21+ (for CLI tool)

```bash
docker --version
docker compose --version
go version
```

## Step 2: Clone & Configure ⚙️

```bash
# Clone repository
git clone <repository-url>
cd swiftlog

# Set up environment
cp .env.example .env
```

**Edit `.env` and set these required values:**
```bash
POSTGRES_PASSWORD=your-secure-password
OPENAI_API_KEY=sk-your-openai-api-key
JWT_SECRET=your-random-secret-key

# Optional: Use custom OpenAI-compatible endpoint
# OPENAI_BASE_URL=https://api.openai.com/v1
```

## Step 3: Start Everything 🚀

**One command to rule them all:**

```bash
make start
```

Or without Make:
```bash
docker compose up -d
```

**That's it!** All services are now running:
- ✅ Database (PostgreSQL) with auto-migrated schema
- ✅ Log storage (Loki)
- ✅ Message queue (Redis)
- ✅ Backend services (Ingestor, API, WebSocket, AI Worker)
- ✅ Frontend web app

## Step 4: Build CLI Tool 🔧

```bash
make cli
```

Or manually:
```bash
cd cli
go build -o swiftlog
```

## Step 5: Test It Out 🎯

```bash
# Configure CLI (you'll need to create a token first via API/database)
./cli/swiftlog config set --token YOUR_TOKEN --server localhost:50051

# Run your first command
./cli/swiftlog run --project demo --group test -- echo "Hello, SwiftLog!"
```

## Access Points 🌐

Once started, these ports are exposed:

| Service | URL | Purpose |
|---------|-----|---------|
| **Web UI** | http://localhost:3000 | Browse logs |
| **REST API** | http://localhost:8080 | HTTP API |
| **WebSocket** | ws://localhost:8081 | Real-time logs |
| **gRPC** | localhost:50051 | CLI connection |

**Note**: Infrastructure services (PostgreSQL, Loki, Redis) are only accessible within Docker network for security.

## Verify Everything Works ✅

```bash
# Check service status
docker compose ps

# View logs
docker compose logs -f

# Test API
curl http://localhost:8080/health
```

## Common Commands 📝

```bash
# View all available commands
make help

# Stop services
make stop

# Restart services
make restart

# View logs
make logs

# Clean everything (WARNING: removes data)
make clean
```

## Need Help? 🆘

- **Full docs**: See [README.md](README.md)
- **Troubleshooting**: Check [README.md#troubleshooting](README.md#troubleshooting)
- **Issues**: GitHub Issues

---

**You're all set!** 🎉 Start logging your scripts with SwiftLog!
