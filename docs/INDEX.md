# SwiftLog Documentation Index

Welcome to the SwiftLog documentation! This index helps you find the right documentation for your needs.

## 🚀 Getting Started

New to SwiftLog? Start here:

1. **[README.md](../README.md)** - Project overview and features
2. **[QUICKSTART.md](../QUICKSTART.md)** - Get up and running in 5 minutes
3. **[DEPLOYMENT_SUMMARY.md](../DEPLOYMENT_SUMMARY.md)** - Deployment guide (中文)

## 📖 User Guides

### Command-Line Interface (CLI)

**[cli/README.md](../cli/README.md)** - Complete CLI documentation

Learn how to:
- Install and configure the CLI
- Wrap commands for log collection
- Use different project and group hierarchies
- Troubleshoot connection issues
- Integrate with CI/CD pipelines

**Key Sections:**
- Installation and Setup
- Configuration Management
- Command Examples
- Advanced Usage Patterns
- Troubleshooting Guide

### Web Interface

**[frontend/README.md](../frontend/README.md)** - Frontend development guide

Learn about:
- Project structure
- Component architecture
- API client usage
- Development workflow
- Building and deployment

**Key Sections:**
- Tech Stack Overview
- Key Components (AIReport, LogViewer)
- API Client Documentation
- Development Guidelines
- Troubleshooting

## 🔧 Developer Guides

### API Reference

**[docs/API.md](./API.md)** - Complete REST and WebSocket API documentation

Comprehensive reference including:
- Authentication methods
- All REST endpoints with examples
- WebSocket protocol
- Error handling
- Rate limiting
- Client library examples (Python, Node.js)

**Key Sections:**
- REST Endpoints (Projects, Groups, Runs, Logs)
- WebSocket Streaming
- Authentication & Authorization
- Complete Workflow Examples
- Client SDK Examples

### Architecture

**[docs/ARCHITECTURE.md](./ARCHITECTURE.md)** - System design and architecture

Deep dive into:
- High-level system design
- Component responsibilities
- Data flow diagrams
- Technology stack
- Database schema
- Security model
- Scalability considerations

**Key Sections:**
- System Design Overview
- Component Architecture
- Data Flow Patterns
- Database Schema
- Security & Authentication
- Performance & Scalability

### Contributing

**[CONTRIBUTING.md](../CONTRIBUTING.md)** - How to contribute to SwiftLog

Everything you need to know about:
- Code of conduct
- Development environment setup
- Coding standards (Go, TypeScript, SQL)
- Testing requirements
- Pull request process
- Documentation guidelines

**Key Sections:**
- Getting Started
- Development Workflow
- Coding Standards
- Testing Guidelines
- Submitting Changes

### Testing

**[tests/README.md](../tests/README.md)** - Integration test suite documentation

Learn about:
- Available test scripts
- Running the test suite
- Test scenarios covered
- Creating new tests
- Interpreting results

**Key Sections:**
- Test Scripts Overview
- Running Tests
- Expected Behavior
- Adding New Tests

## 📦 Component Documentation

### Backend Services

Each backend service has inline documentation:

- **Ingestor** (`backend/cmd/ingestor/`) - gRPC log ingestion service
- **API** (`backend/cmd/api/`) - REST API server
- **WebSocket** (`backend/cmd/websocket/`) - Real-time streaming server
- **AI Worker** (`backend/cmd/ai-worker/`) - Background AI analysis

See [Architecture Documentation](./ARCHITECTURE.md#components) for detailed component descriptions.

### Internal Packages

- **`backend/internal/auth/`** - Authentication logic
- **`backend/internal/database/`** - Database connections
- **`backend/internal/models/`** - Data models
- **`backend/internal/repository/`** - Data access layer
- **`backend/internal/loki/`** - Loki client
- **`backend/internal/ai/`** - AI analysis

## 📋 Quick Reference

### Common Tasks

| Task | Documentation |
|------|---------------|
| Install and run SwiftLog | [QUICKSTART.md](../QUICKSTART.md) |
| Use the CLI | [cli/README.md](../cli/README.md) |
| Call the API | [docs/API.md](./API.md) |
| Understand architecture | [docs/ARCHITECTURE.md](./ARCHITECTURE.md) |
| Run tests | [tests/README.md](../tests/README.md) |
| Contribute code | [CONTRIBUTING.md](../CONTRIBUTING.md) |
| Deploy to production | [DEPLOYMENT_SUMMARY.md](../DEPLOYMENT_SUMMARY.md) |

### Quick Links

- **Configuration**: [README.md - Environment Variables](../README.md#environment-variables)
- **Troubleshooting**: [README.md - Troubleshooting](../README.md#troubleshooting)
- **CLI Commands**: [cli/README.md - Commands](../cli/README.md#commands)
- **API Endpoints**: [docs/API.md - Endpoints](./API.md#endpoints)
- **Database Schema**: [docs/ARCHITECTURE.md - Database Schema](./ARCHITECTURE.md#database-schema)
- **WebSocket Protocol**: [docs/API.md - WebSocket API](./API.md#websocket-api)

## 🎯 Documentation by Role

### For End Users

1. [README.md](../README.md) - What is SwiftLog?
2. [QUICKSTART.md](../QUICKSTART.md) - How do I use it?
3. [cli/README.md](../cli/README.md) - CLI usage guide

### For Frontend Developers

1. [frontend/README.md](../frontend/README.md) - Frontend development
2. [docs/API.md](./API.md) - API reference
3. [CONTRIBUTING.md](../CONTRIBUTING.md) - Contribution guidelines

### For Backend Developers

1. [docs/ARCHITECTURE.md](./ARCHITECTURE.md) - System architecture
2. [docs/API.md](./API.md) - API design
3. [CONTRIBUTING.md](../CONTRIBUTING.md) - Coding standards
4. [tests/README.md](../tests/README.md) - Testing guide

### For DevOps/SRE

1. [DEPLOYMENT_SUMMARY.md](../DEPLOYMENT_SUMMARY.md) - Deployment guide
2. [README.md - Docker Compose](../README.md#docker-compose-commands) - Container management
3. [docs/ARCHITECTURE.md - Scalability](./ARCHITECTURE.md#scalability) - Scaling strategies

### For API Consumers

1. [docs/API.md](./API.md) - Complete API reference
2. [docs/API.md - Examples](./API.md#examples) - Code examples
3. [docs/API.md - Authentication](./API.md#authentication) - Auth guide

## 🔍 Search Tips

Can't find what you're looking for? Try:

1. **README.md** - Start with the main README for high-level overview
2. **Search in docs/** - All detailed documentation is in the `docs/` folder
3. **Component README** - Each major component has its own README
4. **GitHub Issues** - Check if your question has been answered
5. **GitHub Discussions** - Ask the community

## 📝 Documentation Standards

All SwiftLog documentation follows these standards:

- **Clear structure** with table of contents
- **Code examples** for all features
- **Troubleshooting sections** where applicable
- **Cross-references** to related documentation
- **Up-to-date** with the latest release

## 🆘 Getting Help

If you can't find what you need in the documentation:

1. **Check the docs again** - Use Ctrl+F to search
2. **GitHub Issues** - Search existing issues
3. **GitHub Discussions** - Ask a question
4. **Contributing** - Help improve the docs!

## 📅 Last Updated

This index was last updated: **2025-11-19**

---

**Start here:** [README.md](../README.md) | [Quick Start](../QUICKSTART.md)
