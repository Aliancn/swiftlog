# Changelog

All notable changes to SwiftLog will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

#### DevOps & Deployment
- **Environment variable injection during deployment**
  - Automatic `.env` file management from GitHub Secrets
  - `scripts/update-env.sh` script for safe variable updates
  - Support for all configuration variables (database, security, AI, etc.)
  - Automatic backup before updates (`.env.backup.TIMESTAMP`)
  - Interactive mode for manual deployments
  - Selective override - only updates specified variables
  - Complete documentation in `ENV_OVERRIDE_GUIDE.md`

## [0.1.0] - 2025-11-22

### Added

#### Platform Core
- **Real-time log streaming platform** for script execution monitoring
- **Multi-tier architecture** with Project → Group → Run hierarchy
- **WebSocket support** for real-time log updates in the web interface
- **gRPC-based ingestion** for high-performance log streaming
- **PostgreSQL** for metadata storage and run management
- **Grafana Loki** integration for efficient log storage and querying
- **Redis** for caching and real-time communication

#### CLI Features
- **SwiftLog CLI** for wrapping and streaming script logs
- **Multi-platform support**: Linux, macOS (Intel/ARM), Windows
- **Smart project inference** from:
  - Git repository name
  - CI/CD environment variables (GitHub Actions, GitLab CI, Jenkins, CircleCI)
  - Project configuration files (`.swiftlog.json`, `.swiftlog.yaml`)
  - Directory name
  - User-configured defaults
- **Command-based group naming** - automatically generates group names from commands
- **Configuration management** with `swiftlog config` commands
- **Real-time output passthrough** - see logs in terminal while streaming
- **Exit code preservation** - maintains original command exit codes

#### Web Interface
- **Modern Next.js frontend** with React 18 and TypeScript
- **Dashboard** showing all projects and recent activity
- **Project details** with group listing and run history
- **Group details** with run listing and breadcrumb navigation
- **Run details** with:
  - Real-time log viewer with auto-scroll
  - STDOUT/STDERR differentiation with color coding
  - Log search and filtering
  - Full-screen mode support
- **AI-powered log analysis** with:
  - OpenAI integration for intelligent log analysis
  - Markdown rendering for AI reports
  - Language selection (Chinese/English)
  - Manual and automatic analysis modes
- **Resource management**:
  - CRUD operations for projects, groups, and runs
  - Bulk deletion support
  - Status monitoring
- **Responsive design** optimized for desktop and mobile

#### Backend Services
- **API service** - RESTful API for metadata and log queries
- **Ingestor service** - gRPC server for high-throughput log ingestion
- **WebSocket service** - Real-time log streaming to web clients
- **AI worker service** - Background job processing for log analysis
- **Health check endpoints** for all services
- **JWT authentication** support
- **Admin user management**

#### Developer Experience
- **Docker Compose** orchestration for easy local development
- **Makefile** with helpful commands for common operations
- **Development mode** supporting local service development
- **Comprehensive documentation**:
  - Main README with architecture overview
  - CLI README with detailed usage guide
  - API documentation
  - Contributing guidelines
- **Test suite** for CLI with multiple test scenarios

#### DevOps & CI/CD
- **GitHub Actions workflows** for:
  - Automated CLI builds for multiple platforms
  - Docker image builds and publishing to GHCR
  - Automated deployments to servers
  - Release management
- **Deployment script** for easy server deployment
- **Multi-architecture Docker images** (amd64, arm64)
- **Version management** and changelog automation

### Technical Details

#### Technologies Used
- **Frontend**: Next.js 16, React 18, TypeScript, TailwindCSS
- **Backend**: Go 1.21+, gRPC, Protocol Buffers
- **CLI**: Go 1.21+, Cobra, Viper
- **Database**: PostgreSQL 16
- **Log Storage**: Grafana Loki 2.9
- **Cache**: Redis 7
- **Container**: Docker, Docker Compose

#### Architecture Highlights
- **Microservices architecture** with service-specific containers
- **Event-driven design** using Redis pub/sub
- **Streaming-first** approach for real-time log delivery
- **Stateless services** for horizontal scalability
- **Health checks** for all services ensuring reliability

### Security
- **API token authentication** for CLI access
- **SHA-256 token hashing** in database
- **JWT-based** web authentication
- **Environment variable** configuration for secrets
- **No exposed credentials** in codebase

### Documentation
- Complete setup and installation guides
- CLI usage documentation with examples
- API endpoint documentation
- Architecture diagrams and explanations
- Troubleshooting guides
- Integration examples (cron, GitHub Actions, Docker)

---

## Release Notes

### v0.1.0 - Initial Release

This is the first public release of SwiftLog, a comprehensive platform for streaming, storing, and analyzing script execution logs in real-time.

**Highlights:**
- 📊 Beautiful web interface for viewing logs in real-time
- 🖥️ Cross-platform CLI for wrapping any command
- 🤖 AI-powered log analysis with OpenAI integration
- 🚀 Production-ready Docker deployment
- 📦 Easy installation with pre-built binaries

**Get Started:**
1. Download the CLI for your platform from the [Releases page](https://github.com/aliancn/swiftlog/releases)
2. Deploy the platform: `make start`
3. Configure the CLI: `./cli/swiftlog config set --token YOUR_TOKEN --server localhost:50051`
4. Start streaming logs: `./cli/swiftlog run -- your-command`

**Requirements:**
- Docker & Docker Compose (for platform deployment)
- Go 1.21+ (for building from source)
- Modern web browser (for web interface)

For detailed instructions, see the [README](README.md) and [CLI documentation](cli/README.md).

---

[0.1.0]: https://github.com/aliancn/swiftlog/releases/tag/v0.1.0
