# SwiftLog Deployment Guide

This guide covers deploying SwiftLog to production environments.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Deployment Methods](#deployment-methods)
- [Quick Start](#quick-start)
- [Using GitHub Actions](#using-github-actions)
- [Manual Deployment](#manual-deployment)
- [Configuration](#configuration)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)
- [Upgrading](#upgrading)

## Prerequisites

### Server Requirements

- **OS**: Linux (Ubuntu 20.04+ or similar)
- **RAM**: Minimum 2GB, recommended 4GB+
- **Disk**: Minimum 10GB free space
- **CPU**: 2+ cores recommended
- **Network**: Stable internet connection, ports 3000, 8080, 8081, 50051 accessible

### Software Requirements

- Docker 20.10+
- Docker Compose 2.0+
- SSH access to the server

### Optional

- Domain name with DNS configured
- SSL/TLS certificate (for production HTTPS)
- Reverse proxy (Nginx, Caddy, or Traefik)

## Deployment Methods

SwiftLog supports multiple deployment methods:

1. **GitHub Actions (Recommended)** - Automated CI/CD pipeline
2. **Manual Deployment Script** - Interactive deployment script
3. **Docker Compose** - Direct deployment on server

## Quick Start

### Option 1: GitHub Actions (Automated)

1. **Configure GitHub Secrets**

   Go to your repository Settings → Secrets and add:

   ```
   DEPLOY_HOST         - Your server's IP or hostname
   DEPLOY_USER         - SSH user (e.g., deploy)
   DEPLOY_SSH_KEY      - SSH private key for authentication
   DEPLOY_PATH         - Deployment path (optional, default: /opt/swiftlog)
   ```

   Optional secrets for frontend configuration:
   ```
   NEXT_PUBLIC_API_URL - API URL (e.g., https://api.example.com/api/v1)
   NEXT_PUBLIC_WS_URL  - WebSocket URL (e.g., wss://ws.example.com)
   ```

2. **Create a Release**

   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

   This will automatically:
   - Build CLI binaries for all platforms
   - Build and push Docker images
   - Create a GitHub Release
   - Deploy to your server (if secrets are configured)

### Option 2: Manual Deployment Script

1. **Run the deployment script**

   ```bash
   chmod +x deploy.sh
   ./deploy.sh -h your-server.com -u deploy -v v0.1.0
   ```

   With SSH key:
   ```bash
   ./deploy.sh -h your-server.com -u deploy -k ~/.ssh/id_rsa -v v0.1.0
   ```

2. **Configure environment variables on server**

   SSH into your server and edit the `.env` file:
   ```bash
   ssh deploy@your-server.com
   cd /opt/swiftlog
   nano .env
   ```

   Set these critical values:
   ```env
   POSTGRES_PASSWORD=<strong-password>
   JWT_SECRET=<random-secret-key>
   ADMIN_PASSWORD=<admin-password>
   ```

3. **Restart services**

   ```bash
   cd /opt/swiftlog
   ./update.sh
   ```

### Option 3: Direct Docker Compose

1. **Copy files to server**

   ```bash
   scp -r docker-compose.yaml loki-config.yaml .env.example deploy@your-server.com:/opt/swiftlog/
   ```

2. **Configure environment**

   ```bash
   ssh deploy@your-server.com
   cd /opt/swiftlog
   cp .env.example .env
   nano .env  # Edit configuration
   ```

3. **Start services**

   ```bash
   docker compose up -d
   ```

## Using GitHub Actions

### Release Workflow

The release workflow (`.github/workflows/release.yml`) is triggered when you push a tag:

```bash
# Create and push a tag
git tag v0.1.0
git push origin v0.1.0
```

This workflow will:
1. Build CLI binaries for Linux, macOS, Windows (amd64 and arm64)
2. Create a GitHub Release with all binaries
3. Build and push Docker images to GitHub Container Registry
4. Generate checksums for all artifacts

### Deployment Workflow

The deployment workflow (`.github/workflows/deploy.yml`) can be triggered:

**On Tag Push (Automatic):**
```bash
git tag v0.1.0
git push origin v0.1.0
```

**Manual Trigger:**
1. Go to Actions → Deploy to Server
2. Click "Run workflow"
3. Select environment (production/staging)

### Configuring GitHub Secrets

Required secrets:
```
DEPLOY_HOST         - Server hostname or IP
DEPLOY_USER         - SSH username
DEPLOY_SSH_KEY      - Private SSH key (entire key including headers)
```

Optional secrets:
```
DEPLOY_PATH              - Deployment directory (default: /opt/swiftlog)
NEXT_PUBLIC_API_URL      - Frontend API URL
NEXT_PUBLIC_WS_URL       - Frontend WebSocket URL
```

Example SSH key format:
```
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAABlwAAAAdzc2gtcn
...
-----END OPENSSH PRIVATE KEY-----
```

## Manual Deployment

### Step-by-Step Guide

1. **Prepare the server**

   ```bash
   # Install Docker
   curl -fsSL https://get.docker.com | sh
   sudo usermod -aG docker $USER

   # Install Docker Compose
   sudo apt-get update
   sudo apt-get install docker-compose-plugin

   # Create deployment directory
   sudo mkdir -p /opt/swiftlog
   sudo chown $USER:$USER /opt/swiftlog
   ```

2. **Clone or copy files**

   **Option A: Clone repository**
   ```bash
   cd /opt/swiftlog
   git clone https://github.com/aliancn/swiftlog.git .
   git checkout v0.1.0
   ```

   **Option B: Copy files**
   ```bash
   scp docker-compose.yaml deploy@server:/opt/swiftlog/
   scp loki-config.yaml deploy@server:/opt/swiftlog/
   scp .env.example deploy@server:/opt/swiftlog/.env
   ```

3. **Configure environment**

   ```bash
   cd /opt/swiftlog
   nano .env
   ```

   Minimal configuration:
   ```env
   # Database
   POSTGRES_PASSWORD=changeme_to_strong_password

   # Security
   JWT_SECRET=generate_random_secret_here

   # Admin
   ADMIN_USERNAME=admin
   ADMIN_PASSWORD=changeme_to_strong_password

   # Environment
   ENVIRONMENT=production
   LOG_LEVEL=info
   ```

4. **Start services**

   ```bash
   docker compose up -d
   ```

5. **Verify deployment**

   ```bash
   # Check service status
   docker compose ps

   # Check logs
   docker compose logs -f

   # Test API
   curl http://localhost:8080/health
   ```

## Configuration

### Environment Variables

Full list of configuration options:

```env
# Database Configuration
POSTGRES_PASSWORD=changeme          # PostgreSQL password (REQUIRED)

# Security
JWT_SECRET=dev-secret-key          # JWT signing secret (REQUIRED for production)

# Admin User
ADMIN_USERNAME=admin               # Default admin username
ADMIN_PASSWORD=admin123           # Default admin password (CHANGE THIS)

# Application Settings
ENVIRONMENT=production            # Environment: production, staging, development
LOG_LEVEL=info                   # Log level: debug, info, warn, error

# Frontend URLs (for Docker build)
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_WS_URL=ws://localhost:8081
```

### Reverse Proxy Configuration

**Nginx Example:**

```nginx
# Frontend
server {
    listen 80;
    server_name swiftlog.example.com;

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}

# API
server {
    listen 80;
    server_name api.swiftlog.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

# WebSocket
server {
    listen 80;
    server_name ws.swiftlog.example.com;

    location / {
        proxy_pass http://localhost:8081;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Host $host;
    }
}
```

**Caddy Example:**

```caddy
swiftlog.example.com {
    reverse_proxy localhost:3000
}

api.swiftlog.example.com {
    reverse_proxy localhost:8080
}

ws.swiftlog.example.com {
    reverse_proxy localhost:8081
}
```

## Monitoring

### Health Checks

All services expose health check endpoints:

```bash
# API health
curl http://localhost:8080/health

# Check container health
docker compose ps
```

### Viewing Logs

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f api
docker compose logs -f ingestor
docker compose logs -f frontend
```

### Resource Usage

```bash
# Container stats
docker stats

# Disk usage
docker system df
```

## Troubleshooting

### Services Won't Start

**Check logs:**
```bash
docker compose logs
```

**Common issues:**
- Port conflicts: Check if ports 3000, 8080, 8081, 50051 are available
- Database not ready: Wait for PostgreSQL to initialize (first run takes longer)
- Missing .env: Copy from .env.example and configure

### Cannot Connect to API

**Check service is running:**
```bash
docker compose ps api
curl http://localhost:8080/health
```

**Check firewall:**
```bash
sudo ufw status
sudo ufw allow 8080/tcp
```

### Logs Not Appearing

**Check Loki:**
```bash
docker compose logs loki
curl http://localhost:3100/ready
```

**Check Ingestor:**
```bash
docker compose logs ingestor
```

### High Memory Usage

**Check resource limits:**
```bash
docker stats
```

**Adjust in docker-compose.yaml:**
```yaml
services:
  api:
    deploy:
      resources:
        limits:
          memory: 512M
        reservations:
          memory: 256M
```

## Upgrading

### Using GitHub Actions

1. Create and push a new tag:
   ```bash
   git tag v0.2.0
   git push origin v0.2.0
   ```

2. GitHub Actions will automatically deploy

### Manual Update

1. **Pull latest images:**
   ```bash
   cd /opt/swiftlog
   docker compose pull
   ```

2. **Backup database (recommended):**
   ```bash
   docker compose exec postgres pg_dump -U swiftlog swiftlog > backup.sql
   ```

3. **Restart services:**
   ```bash
   docker compose down
   docker compose up -d
   ```

4. **Verify:**
   ```bash
   docker compose ps
   docker compose logs
   ```

### Rollback

If something goes wrong:

```bash
# Use specific version
docker compose down
# Edit docker-compose.yaml to use previous version tag
docker compose up -d
```

## Security Best Practices

1. **Change default passwords** in `.env`
2. **Use strong JWT secret** (generate with `openssl rand -base64 32`)
3. **Enable HTTPS** with reverse proxy
4. **Restrict ports** with firewall (only expose 80/443 externally)
5. **Regular backups** of PostgreSQL database
6. **Keep Docker images updated**
7. **Use secrets management** for sensitive data
8. **Enable log rotation** to prevent disk filling

## Backup and Restore

### Backup

```bash
# Database
docker compose exec postgres pg_dump -U swiftlog swiftlog > swiftlog-backup-$(date +%Y%m%d).sql

# Configuration
cp .env .env.backup
```

### Restore

```bash
# Database
cat backup.sql | docker compose exec -T postgres psql -U swiftlog swiftlog
```

## Support

- **Documentation**: [README.md](../README.md)
- **CLI Guide**: [cli/README.md](../cli/README.md)
- **Issues**: [GitHub Issues](https://github.com/aliancn/swiftlog/issues)
- **Discussions**: [GitHub Discussions](https://github.com/aliancn/swiftlog/discussions)
