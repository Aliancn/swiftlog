# SwiftLog - Script Log Collection and Analysis Platform

SwiftLog is a lightweight, high-performance platform for collecting, storing, and analyzing logs from script executions. Never lose track of your script runs again!

---

## 📚 文档导航

- 📖 **[文档索引](DOCS_INDEX.md)** - 所有文档的快速导航
- 🔒 **[安全更新](SECURITY_UPDATES.md)** - 安全修复和生产配置（部署前必读！）
- 📚 **[API文档](API_CONSISTENCY_REPORT.md)** - 完整的API接口文档
- 📝 **[贡献指南](CONTRIBUTING.md)** - 如何参与项目开发
- 📝 **[变更日志](CHANGELOG.md)** - 版本历史和变更记录

**⚠️ 生产部署提示**: 部署前请务必阅读 [SECURITY_UPDATES.md](SECURITY_UPDATES.md#生产环境配置清单) 配置安全相关环境变量！

---

## 🌟 Features

- **Zero-Intrusion Collection**: Wrap any command with `swiftlog run` or pipe output directly
- **Real-Time Streaming**: Watch script output live in the web interface
- **Accurate State Tracking**: Captures exact exit codes and execution status
- **Structured Storage**: PostgreSQL metadata + Grafana Loki for log lines
- **AI-Powered Analysis**: Automatic report generation using OpenAI
- **RESTful API**: Full-featured API for integrations
- **WebSocket Support**: Real-time log streaming to browsers

## 🏗️ Architecture

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

### Services

- **CLI** (`cli/`): Command-line interface for running and piping logs
- **Ingestor** (`backend/cmd/ingestor/`): gRPC service for log ingestion
- **API** (`backend/cmd/api/`): REST API server (port 8080)
- **WebSocket** (`backend/cmd/websocket/`): Real-time log streaming (port 8081)
- **AI Worker** (`backend/cmd/ai-worker/`): Background job processor for AI analysis
- **Frontend** (`frontend/`): Next.js 14 web interface

## 🚀 Quick Start (Production)

### Prerequisites

- **Docker** 24+ & **Docker Compose** v2+
- **Go** 1.21+ (only for building CLI tool)
- **Git**

### 1️⃣ Clone and Setup

```bash
git clone <repository-url>
cd swiftlog

# Copy environment template
cp .env.example .env

# Edit .env and set your values
nano .env
```

**必需环境变量（生产环境）:**
```bash
# 数据库密码
POSTGRES_PASSWORD=your-secure-password

# AI API密钥（可选，如不使用AI分析可不设置）
OPENAI_API_KEY=sk-your-openai-key-here

# JWT密钥（生产环境必须设置强密钥，至少32字符）
JWT_SECRET=$(openssl rand -base64 32)

# API密钥加密密钥（生产环境必需）
ENCRYPTION_KEY=$(openssl rand -base64 32)

# CORS允许的域名（生产环境必须设置）
CORS_ORIGINS=https://your-domain.com
```

**可选环境变量:**
```bash
# JWT过期时间（推荐1-2小时）
JWT_EXPIRATION=2h

# 自定义OpenAI兼容API端点
OPENAI_BASE_URL=https://api.openai.com/v1

# AI模型名称
OPENAI_MODEL=gpt-4o-mini

# 应用环境
ENVIRONMENT=production
LOG_LEVEL=info
```

**🔒 安全提示**:
- 生产环境必须设置强`JWT_SECRET`和`ENCRYPTION_KEY`（使用`openssl rand -base64 32`生成）
- 更多安全配置详见 [SECURITY_UPDATES.md](SECURITY_UPDATES.md#生产环境配置清单)

### 2️⃣ Start All Services (One Command!)

**Option A: Using Makefile (Recommended)**
```bash
make start
```

**Option B: Using startup script**
```bash
./start.sh
```

**Option C: Using docker compose directly**
```bash
docker compose up -d
```

That's it! 🎉 All services will start automatically:
- ✅ PostgreSQL with auto-migrated schema
- ✅ Grafana Loki for log storage
- ✅ Redis for pub/sub
- ✅ Ingestor (gRPC server)
- ✅ REST API server
- ✅ WebSocket server
- ✅ AI Worker
- ✅ Frontend web app

### 3️⃣ Verify Services

```bash
# Check all services are running
docker compose ps

# View logs
docker compose logs -f

# Test API health
curl http://localhost:8080/health
```

### 4️⃣ Build and Configure CLI

```bash
# Build CLI tool
make cli

# Or manually
cd cli
go build -o swiftlog

# Install globally (optional)
sudo cp swiftlog /usr/local/bin/

# Configure CLI
./swiftlog config set --token YOUR_API_TOKEN --server localhost:50051
```

### 5️⃣ Run Your First Command

```bash
swiftlog run --project myapp --group tests -- echo "Hello, SwiftLog!"
```

## 🎯 Access Points

Once started, access SwiftLog at:

| Service | URL | Exposed | Description |
|---------|-----|---------|-------------|
| **Frontend** | http://localhost:3000 | ✅ | Web interface |
| **REST API** | http://localhost:8080 | ✅ | HTTP API |
| **WebSocket** | ws://localhost:8081 | ✅ | Real-time streaming |
| **gRPC Ingestor** | localhost:50051 | ✅ | CLI connection |
| **PostgreSQL** | N/A (internal) | ❌ | Database (Docker network only) |
| **Loki** | N/A (internal) | ❌ | Log storage (Docker network only) |
| **Redis** | N/A (internal) | ❌ | Pub/sub (Docker network only) |

**Note**: Infrastructure services (PostgreSQL, Loki, Redis) are only accessible within the Docker network for security. To access them for development/debugging, uncomment the port mappings in `docker compose.yaml`.

## 📝 CLI Usage

### Run a Command

```bash
swiftlog run --project <project> --group <group> -- <command> [args...]
```

**Examples:**
```bash
# Run a build script
swiftlog run --project webapp --group build -- ./build.sh

# Run Python script
swiftlog run --project ml --group training -- python train_model.py

# Run with default project/group
swiftlog run -- npm test

# Complex command with pipes
swiftlog run --project data -- bash -c "cat data.csv | grep ERROR | wc -l"
```

### Configuration

```bash
# Set API token and server
swiftlog config set --token YOUR_TOKEN --server localhost:50051

# View current configuration
swiftlog config get

# Show config file path
swiftlog config path
```

## 🤖 Using Custom OpenAI-Compatible APIs

SwiftLog supports any OpenAI-compatible API endpoint. This allows you to use:

- **Azure OpenAI Service**
- **LocalAI** (self-hosted)
- **Ollama** (with OpenAI compatibility layer)
- **Other OpenAI-compatible providers**

### Configuration

Set the `OPENAI_BASE_URL` environment variable in your `.env` file:

```bash
# Example: Azure OpenAI
OPENAI_BASE_URL=https://your-resource.openai.azure.com/openai/deployments/your-deployment
OPENAI_API_KEY=your-azure-api-key

# Example: LocalAI
OPENAI_BASE_URL=http://localhost:8080/v1
OPENAI_API_KEY=not-needed-for-localai

# Example: Custom endpoint
OPENAI_BASE_URL=https://your-custom-endpoint.com/v1
OPENAI_API_KEY=your-api-key
```

**Note**: The base URL should end with `/v1` (without `/chat/completions`). SwiftLog will automatically append the correct endpoint path.

### Supported Models

Any model that supports the OpenAI Chat Completions API format will work. Set the model name in `.env`:

```bash
OPENAI_MODEL=gpt-4o-mini          # OpenAI
OPENAI_MODEL=gpt-35-turbo         # Azure OpenAI
OPENAI_MODEL=llama2               # LocalAI/Ollama
```

## 🛠️ Development Mode

For local development with hot-reload:

```bash
# Start only infrastructure (postgres, loki, redis)
make dev-up

# Then run services locally:
cd backend/cmd/ingestor && go run main.go  # Terminal 1
cd backend/cmd/api && go run main.go       # Terminal 2
cd backend/cmd/websocket && go run main.go # Terminal 3
cd backend/cmd/ai-worker && go run main.go # Terminal 4
cd frontend && npm run dev                  # Terminal 5

# When done:
make dev-down
```

## 📚 API Overview

### Authentication

All API endpoints require an `Authorization` header:
```
Authorization: Bearer YOUR_API_TOKEN
```

### REST Endpoints (Summary)

**Base URL:** `http://localhost:8080/api/v1`

#### Projects
- `GET /projects` - List all projects
- `GET /projects/:id` - Get project details
- `POST /projects` - Create new project
- `GET /projects/:id/groups` - List groups in project

#### Groups
- `GET /groups/:id` - Get group details

#### Runs
- `GET /groups/:id/runs?limit=50&offset=0` - List runs in a group
- `GET /runs/:id` - Get run details
- `GET /runs/:id/logs` - Get run logs from Loki
- `POST /runs/:id/analyze` - Trigger AI analysis

### WebSocket API

Connect to real-time log streaming:

```javascript
const ws = new WebSocket('ws://localhost:8081/ws/runs/:run_id?token=YOUR_TOKEN');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log(data);
  // {
  //   "type": "log",
  //   "run_id": "uuid",
  //   "timestamp": "2025-11-18T10:00:00Z",
  //   "level": "stdout",
  //   "content": "log line content"
  // }
};
```

**For complete API documentation with examples, see [docs/API.md](docs/API.md)**

## 📊 Database Schema

### Core Tables

- **users**: Authenticated users
- **api_tokens**: API authentication tokens (SHA-256 hashed)
- **projects**: Top-level log containers
- **log_groups**: Organizational units within projects
- **log_runs**: Individual script executions with metadata

### Auto-Migration

Database schema is automatically created on first startup via PostgreSQL's `docker-entrypoint-initdb.d`. All migrations are idempotent and can be safely re-run.

## 🔧 Makefile Commands

```bash
make help         # Show all available commands
make start        # Start all services
make stop         # Stop all services
make restart      # Restart all services
make logs         # View logs from all services
make build        # Rebuild all Docker images
make clean        # Remove all containers and volumes
make cli          # Build CLI tool
make dev-up       # Start infrastructure only
make dev-down     # Stop infrastructure
```

## 🐳 Docker Compose Commands

```bash
# Start all services
docker compose up -d

# Stop all services
docker compose down

# View logs
docker compose logs -f

# View logs for specific service
docker compose logs -f api

# Restart a service
docker compose restart ingestor

# Rebuild and restart
docker compose up -d --build

# Scale AI workers
docker compose up -d --scale ai-worker=3

# Remove everything including volumes
docker compose down -v
```

## 🔐 Security

- **Token Authentication**: SHA-256 hashed API tokens
- **gRPC Security**: Metadata-based authentication
- **HTTP Security**: Bearer token authentication
- **SQL Injection Prevention**: Parameterized queries
- **CORS**: Configured for localhost (update for production)
- **Network Isolation**: Infrastructure services (PostgreSQL, Loki, Redis) are not exposed to the host, only accessible within Docker network

**Production Checklist:**
- [ ] Change `POSTGRES_PASSWORD` in `.env`
- [ ] Set strong `JWT_SECRET` in `.env`
- [ ] Add your `OPENAI_API_KEY` in `.env`
- [ ] Update CORS settings in API and WebSocket services
- [ ] Use HTTPS/TLS for all external connections
- [ ] Set up rate limiting (e.g., nginx reverse proxy)
- [ ] Configure firewall rules
- [ ] Set up monitoring and alerting

## 📈 Performance Targets

- **CLI Overhead**: <5% for scripts running >10 seconds
- **Real-Time Latency**: <2 seconds from log generation to browser display
- **API Response Time**: <200ms (p95)
- **Concurrent Streams**: 1000+ simultaneous gRPC connections

## 🗂️ Project Structure

```
swiftlog/
├── backend/
│   ├── cmd/                    # Service entry points
│   │   ├── ingestor/          # gRPC log ingestor
│   │   ├── api/               # REST API server
│   │   ├── websocket/         # WebSocket server
│   │   └── ai-worker/         # AI analysis worker
│   ├── internal/              # Internal packages
│   │   ├── auth/              # Authentication
│   │   ├── database/          # DB connection
│   │   ├── models/            # Data models
│   │   ├── repository/        # Data access
│   │   ├── loki/              # Loki client
│   │   ├── ingestor/          # Ingestor logic
│   │   ├── websocket/         # WebSocket hub
│   │   └── ai/                # AI analyzer
│   ├── migrations/            # SQL migrations (auto-run)
│   ├── proto/                 # Protobuf definitions
│   ├── Dockerfile.ingestor    # Ingestor container
│   ├── Dockerfile.api         # API container
│   ├── Dockerfile.websocket   # WebSocket container
│   └── Dockerfile.ai-worker   # AI Worker container
├── cli/
│   ├── cmd/                   # CLI commands
│   ├── internal/
│   │   ├── config/            # Config management
│   │   └── client/            # gRPC client
│   └── proto/                 # Protobuf (generated)
├── frontend/
│   ├── src/                   # Next.js application
│   └── Dockerfile             # Frontend container
├── docker compose.yaml        # All-in-one deployment
├── Makefile                   # Build automation
├── start.sh                   # Startup script
├── .env.example               # Environment template
└── README.md                  # This file
```

## 🧪 Testing

### Unit Tests

```bash
# Backend tests
cd backend
go test ./...

# Frontend tests
cd frontend
npm test
```

### Integration Tests

SwiftLog includes a comprehensive test suite for testing different logging scenarios:

```bash
# Run all integration tests
cd tests
./run_all_tests.sh

# Or run individual tests
./cli/swiftlog run --project test-project --group 01_simple_test -- bash tests/01_simple_test.sh
./cli/swiftlog run --project test-project --group 02_stderr_test -- bash tests/02_stderr_test.sh
./cli/swiftlog run --project test-project --group 03_long_logs -- bash tests/03_long_logs.sh
./cli/swiftlog run --project test-project --group 04_multiline_output -- bash tests/04_multiline_output.sh
```

**Test Suite Features:**
- Simple stdout logging
- Mixed stdout/stderr output
- High-volume logs (100+ entries)
- Multiline output (JSON, SQL, stack traces)

See [tests/README.md](tests/README.md) for detailed test documentation.

## 🐛 Troubleshooting

### Services won't start

```bash
# Check Docker is running
docker info

# Check logs for errors
docker compose logs

# Restart everything
docker compose down
docker compose up -d
```

### Database connection errors

```bash
# Check PostgreSQL is healthy
docker compose ps postgres

# View PostgreSQL logs
docker compose logs postgres

# Access PostgreSQL for debugging (from within Docker network)
docker compose exec postgres psql -U swiftlog -d swiftlog

# Reset database (WARNING: deletes all data)
docker compose down -v
docker compose up -d
```

### Need to access internal services?

By default, infrastructure services are not exposed to the host for security. To access them during development:

**Edit `docker compose.yaml` and uncomment the port mappings:**

```yaml
postgres:
  ports:
    - "5432:5432"  # Uncomment this line

loki:
  ports:
    - "3100:3100"  # Uncomment this line

redis:
  ports:
    - "6379:6379"  # Uncomment this line
```

Then restart:
```bash
docker compose down
docker compose up -d
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

### Missing OpenAI API key

```bash
# Edit .env file
nano .env

# Add your key
OPENAI_API_KEY=sk-your-key-here

# Restart AI Worker
docker compose restart ai-worker
```

## 📦 Technology Stack

### Backend
- **Go 1.21+**: CLI, gRPC, REST API, Workers
- **PostgreSQL 16**: Metadata storage
- **Grafana Loki 2.9**: Log storage
- **Redis 7**: Pub/sub messaging
- **gRPC**: Log streaming protocol
- **Gin**: HTTP framework
- **gorilla/websocket**: WebSocket server

### Frontend
- **Next.js 14**: React framework with App Router
- **TypeScript 5**: Type safety
- **Tailwind CSS**: Styling

### AI
- **OpenAI API**: gpt-4o-mini for log analysis

## 🚢 Deployment & CI/CD

SwiftLog includes production-ready deployment automation and CI/CD pipelines.

### GitHub Actions Workflows

- **Release**: Automatically builds CLI binaries and Docker images on version tags
- **Deploy**: Automated deployment to production/staging servers

### Quick Deployment

**Using the deployment script:**
```bash
./deploy.sh -h your-server.com -u deploy -v v0.1.0
```

**Using GitHub Actions:**
```bash
# Create and push a version tag
git tag v0.1.0
git push origin v0.1.0
```

This will automatically:
- Build CLI binaries for Linux, macOS (Intel/ARM), Windows
- Create a GitHub Release with all binaries
- Build and push Docker images to GitHub Container Registry
- Deploy to your configured server (if secrets are set)

### CLI Binary Downloads

Pre-built CLI binaries are available for download from the [Releases page](https://github.com/aliancn/swiftlog/releases):

- **Linux**: `swiftlog-linux-amd64`, `swiftlog-linux-arm64`
- **macOS**: `swiftlog-darwin-amd64`, `swiftlog-darwin-arm64`
- **Windows**: `swiftlog-windows-amd64.exe`

**Installation:**
```bash
# Download for your platform
wget https://github.com/aliancn/swiftlog/releases/latest/download/swiftlog-linux-amd64

# Make executable and install
chmod +x swiftlog-linux-amd64
sudo mv swiftlog-linux-amd64 /usr/local/bin/swiftlog

# Verify
swiftlog --version
```

### Docker Images

Pre-built Docker images are available on GitHub Container Registry:

```bash
ghcr.io/aliancn/swiftlog/api:v0.1.0
ghcr.io/aliancn/swiftlog/ingestor:v0.1.0
ghcr.io/aliancn/swiftlog/websocket:v0.1.0
ghcr.io/aliancn/swiftlog/ai-worker:v0.1.0
ghcr.io/aliancn/swiftlog/frontend:v0.1.0
```

See **[Deployment Guide](docs/DEPLOYMENT.md)** for detailed instructions on:
- Production deployment strategies
- GitHub Actions configuration
- **Environment variable management and injection**
- Server requirements and setup
- SSL/TLS and reverse proxy configuration
- Monitoring and maintenance
- Backup and restore procedures

### Environment Variables

SwiftLog supports automatic environment variable injection during deployment:

**Configure via GitHub Secrets:**
```bash
# Add to GitHub Settings → Secrets
POSTGRES_PASSWORD    # Database password
JWT_SECRET          # JWT signing key
ADMIN_PASSWORD      # Admin password
# Note: OpenAI API key must be configured manually on server
```

**Deploy with auto-configuration:**
```bash
git tag v0.1.0
git push origin v0.1.0
# Variables automatically injected from GitHub Secrets
```

See **[Environment Variables Guide](ENV_OVERRIDE_GUIDE.md)** for complete documentation.

## 📚 Documentation

Comprehensive documentation is available:

- **[Deployment Guide](docs/DEPLOYMENT.md)** - Production deployment and CI/CD
- **[Environment Variables](docs/ENVIRONMENT_VARIABLES.md)** - Complete environment configuration guide
- **[Environment Override Guide](ENV_OVERRIDE_GUIDE.md)** - Auto-inject variables during deployment (中文)
- **[OpenAI Configuration](docs/OPENAI_CONFIGURATION.md)** - How to configure OpenAI API for AI features
- **[Quick Start Guide](QUICKSTART.md)** - Get started in 5 minutes
- **[API Documentation](docs/API.md)** - Complete REST API reference
- **[CLI Documentation](cli/README.md)** - Command-line tool usage
- **[Frontend Documentation](frontend/README.md)** - Web interface development
- **[Architecture](docs/ARCHITECTURE.md)** - System design and data flow
- **[Test Suite](tests/README.md)** - Integration testing guide
- **[Contributing Guide](CONTRIBUTING.md)** - How to contribute
- **[Deployment Summary](DEPLOYMENT_SUMMARY.md)** - Deployment guide (中文)
- **[CHANGELOG](CHANGELOG.md)** - Version history and release notes

## 🤝 Contributing

Contributions welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:

- Code style and standards
- Development workflow
- Testing requirements
- Pull request process

## 📄 License

[Your License Here]

## 🙋 Support

- **Documentation**: See links above
- **Issues**: [GitHub Issues](https://github.com/your-repo/swiftlog/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-repo/swiftlog/discussions)
- **Spec Docs**: `/specs/001-script-log-platform/`

## ⭐ Star History

If you find SwiftLog useful, please consider giving it a star! ⭐

---

**Built with ❤️ using Go, TypeScript, and Docker**
