# SwiftLog

**A lightweight, high-performance platform for collecting, storing, and analyzing logs from script executions.**

Never lose track of your script runs again! SwiftLog captures every execution, streams logs in real-time, and provides AI-powered analysis—all with zero code changes.

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![Node Version](https://img.shields.io/badge/node-20+-green.svg)](https://nodejs.org)

---

## ✨ Why SwiftLog?

### Zero-Intrusion Collection
Wrap any command or pipe output directly—no code changes required:
```bash
swiftlog run -- ./your-script.sh
# or
./your-script.sh | swiftlog pipe
```

### Real-Time Streaming
Watch script output live in a beautiful web interface with <2s latency from execution to browser.

### Accurate State Tracking
Captures exact exit codes, execution duration, and status. Know exactly when and why your scripts fail.

### AI-Powered Analysis
Automatic log analysis using OpenAI. Get instant insights, error summaries, and recommendations.

### Full-Featured API
RESTful API and WebSocket support for seamless integrations with your existing tools.

---

## 🚀 Quick Start

### 1. Start SwiftLog (One Command)

```bash
git clone https://github.com/aliancn/swiftlog.git
cd swiftlog
cp .env.example .env  # Edit with your settings
make start
```

**That's it!** All services start automatically:
- ✅ PostgreSQL with auto-migrated schema
- ✅ Grafana Loki for log storage
- ✅ Redis for pub/sub
- ✅ Backend services (Ingestor, API, WebSocket, AI Worker)
- ✅ Frontend web application

### 2. Build CLI Tool

```bash
make cli
./cli/swiftlog config set --token YOUR_TOKEN --server localhost:50051
```

### 3. Run Your First Script

```bash
cd cli
./swiftlog run --project myapp --group tests -- echo "Hello, SwiftLog!"
```

### 4. View Results

Open http://localhost:3000 to see your logs in real-time!

---

## 💡 Usage Examples

### Basic Script Execution

```bash
# Run a build script
swiftlog run --project webapp --group build -- ./build.sh

# Run Python script
swiftlog run --project ml --group training -- python train_model.py

# Auto-detect project from git repo
swiftlog run -- npm test
```

### Complex Commands

```bash
# Commands with pipes
swiftlog run -- bash -c "cat data.csv | grep ERROR | wc -l"

# Long-running processes
swiftlog run --project etl --group daily -- python process_data.py --date today
```

### CI/CD Integration

```bash
# GitHub Actions, GitLab CI, Jenkins, etc.
swiftlog run --project myapp --group ci-build -- make build
swiftlog run --project myapp --group ci-test -- make test
```

---

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

**Components:**
- **CLI Tool** - Wraps commands and streams logs via gRPC
- **Ingestor** - Receives logs and writes to Loki/PostgreSQL
- **REST API** - HTTP endpoints for querying and management
- **WebSocket** - Real-time log streaming to browsers
- **AI Worker** - Background AI analysis using OpenAI
- **Frontend** - Next.js 14 web interface

---

## 🎯 Key Features

### Smart Auto-Detection
- Project names from git repo, CI/CD vars, or directory
- Group names generated from commands
- Configurable defaults and project-level config files

### Flexible Configuration
- OpenAI-compatible API support (Azure, LocalAI, Ollama)
- Customizable AI models
- Environment-based configuration
- Docker Compose for easy deployment

### Security Built-In
- SHA-256 hashed API tokens
- JWT authentication
- Encrypted sensitive data
- Network isolation for infrastructure services
- CORS configuration

### Production Ready
- GitHub Actions CI/CD workflows
- Pre-built binaries for Linux, macOS, Windows
- Docker images on GitHub Container Registry
- Automated deployment scripts
- Health checks and monitoring

---

## 📚 Documentation

Comprehensive documentation is available in the `docs/` folder:

| Document | Description |
|----------|-------------|
| **[Getting Started](docs/getting-started.md)** | Installation and quick start guide |
| **[CLI Guide](docs/cli-guide.md)** | Complete CLI tool documentation |
| **[API Reference](docs/api-reference.md)** | REST and WebSocket API documentation |
| **[Architecture](docs/architecture.md)** | System design and components |
| **[Configuration](docs/configuration.md)** | Environment variables and settings |
| **[Deployment](docs/deployment.md)** | Production deployment guide |
| **[Development](docs/development.md)** | Contributing and local development |

**[📖 View All Documentation](docs/index.md)**

---

## 🎯 Access Points

| Service | URL | Description |
|---------|-----|-------------|
| **Frontend** | http://localhost:3000 | Web interface |
| **REST API** | http://localhost:8080 | HTTP API |
| **WebSocket** | ws://localhost:8081 | Real-time streaming |
| **gRPC Ingestor** | localhost:50051 | CLI connection |

---

## 🛠️ Technology Stack

| Layer | Technology |
|-------|-----------|
| **Backend** | Go 1.24+, gRPC, Gin, gorilla/websocket |
| **Frontend** | Next.js 14, TypeScript 5, Tailwind CSS |
| **Database** | PostgreSQL 16 |
| **Log Storage** | Grafana Loki 2.9 |
| **Pub/Sub** | Redis 7 |
| **AI** | OpenAI API |
| **Deployment** | Docker, Docker Compose |

---

## 📦 Installation Options

### Using Pre-Built Binaries

Download CLI binaries from [Releases](https://github.com/aliancn/swiftlog/releases):

```bash
# Linux
wget https://github.com/aliancn/swiftlog/releases/latest/download/swiftlog-linux-amd64
chmod +x swiftlog-linux-amd64
sudo mv swiftlog-linux-amd64 /usr/local/bin/swiftlog

# macOS (ARM)
wget https://github.com/aliancn/swiftlog/releases/latest/download/swiftlog-darwin-arm64
chmod +x swiftlog-darwin-arm64
sudo mv swiftlog-darwin-arm64 /usr/local/bin/swiftlog

# Windows
# Download swiftlog-windows-amd64.exe from releases
```

### Using Docker Images

Pre-built images available on GitHub Container Registry:

```bash
docker pull ghcr.io/aliancn/swiftlog/api:latest
docker pull ghcr.io/aliancn/swiftlog/ingestor:latest
docker pull ghcr.io/aliancn/swiftlog/websocket:latest
docker pull ghcr.io/aliancn/swiftlog/ai-worker:latest
docker pull ghcr.io/aliancn/swiftlog/frontend:latest
```

### Building from Source

```bash
git clone https://github.com/aliancn/swiftlog.git
cd swiftlog

# Build CLI
make cli

# Build all services
make build
```

---

## ⚙️ Configuration

### Minimal Configuration

```bash
# .env
POSTGRES_PASSWORD=your-password
JWT_SECRET=$(openssl rand -base64 32)
ENCRYPTION_KEY=$(openssl rand -base64 32)
```

### Production Configuration

```bash
# .env
POSTGRES_PASSWORD=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 32)
ENCRYPTION_KEY=$(openssl rand -base64 32)
PUBLIC_URL=https://logs.yourdomain.com
CORS_ORIGINS=https://logs.yourdomain.com
OPENAI_API_KEY=sk-your-key  # Optional for AI features
```

See [Configuration Guide](docs/configuration.md) for all options.

---

## 🧪 Testing

```bash
# Run integration tests
cd tests
./run_all_tests.sh

# Individual tests
./cli/swiftlog run --project test --group simple -- bash tests/01_simple_test.sh
```

---

## 🚢 Deployment

### Quick Deploy with GitHub Actions

1. Configure GitHub Secrets (see [Deployment Guide](docs/deployment.md))
2. Push a version tag:
```bash
git tag v1.0.0
git push origin v1.0.0
```

This automatically:
- Builds multi-platform CLI binaries
- Creates GitHub Release
- Builds and pushes Docker images
- Deploys to your server (optional)

### Manual Deploy

```bash
# On your server
git clone <repository-url>
cd swiftlog
cp .env.example .env
# Edit .env with production values
docker compose up -d
```

See [Deployment Guide](docs/deployment.md) for detailed instructions.

---

## 🤝 Contributing

We welcome contributions! Please see our [Development Guide](docs/development.md) for:

- Setting up local development environment
- Code style and standards
- Testing requirements
- Pull request process

---

## 📄 License

[MIT License](LICENSE)

---

## 🙋 Support

- **Documentation**: [docs/index.md](docs/index.md)
- **Issues**: [GitHub Issues](https://github.com/aliancn/swiftlog/issues)
- **Discussions**: [GitHub Discussions](https://github.com/aliancn/swiftlog/discussions)

---

## ⭐ Star History

If you find SwiftLog useful, please consider giving it a star! ⭐

---

**Built with ❤️ using Go, TypeScript, and Docker**
