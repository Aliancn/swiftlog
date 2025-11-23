# GitHub Actions Setup Guide

This guide explains how to configure GitHub Actions for SwiftLog's CI/CD pipeline.

## Overview

SwiftLog uses GitHub Actions for:
1. **Building CLI binaries** for multiple platforms
2. **Creating GitHub Releases** with automatic artifact uploads
3. **Building and pushing Docker images** to GitHub Container Registry
4. **Automated deployment** to production/staging servers

## Required Secrets

Configure the following secrets in your GitHub repository settings (Settings → Secrets and variables → Actions):

### Deployment Secrets

| Secret Name | Description | Required | Example |
|------------|-------------|----------|---------|
| `DEPLOY_HOST` | Server hostname or IP | Yes (for deploy) | `swiftlog.example.com` or `192.168.1.100` |
| `DEPLOY_USER` | SSH username | Yes (for deploy) | `deploy` or `ubuntu` |
| `DEPLOY_SSH_KEY` | SSH private key | Yes (for deploy) | Full private key including headers |
| `DEPLOY_PATH` | Deployment directory | No | `/opt/swiftlog` (default) |

### Application Secrets

These secrets are automatically injected into the `.env` file during deployment:

| Secret Name | Description | Required | Example |
|------------|-------------|----------|---------|
| `POSTGRES_PASSWORD` | PostgreSQL database password | Yes | Use strong password (see generation below) |
| `JWT_SECRET` | JWT signing secret for authentication | Yes | Random 32-byte string (see generation below) |
| `ADMIN_PASSWORD` | Default admin user password | Yes | Use strong password |
| `ENVIRONMENT` | Environment name | No | `production` (default), `staging` |
| `LOG_LEVEL` | Application log level | No | `info` (default), `debug`, `warn`, `error` |

**Note:** OpenAI API key is NOT automatically injected. If you need AI analysis features, manually add `OPENAI_API_KEY` to the `.env` file on the server after deployment.

### Frontend Configuration (Optional)

| Secret Name | Description | Example |
|------------|-------------|---------|
| `NEXT_PUBLIC_API_URL` | Frontend API URL | `https://api.swiftlog.example.com/api/v1` |
| `NEXT_PUBLIC_WS_URL` | WebSocket URL | `wss://ws.swiftlog.example.com` |

If not set, defaults to `localhost` URLs.

### Generating Secure Values

```bash
# Generate a strong JWT secret (32 bytes, base64 encoded)
openssl rand -base64 32

# Generate a strong password (16 bytes, base64 encoded)
openssl rand -base64 16

# Or use a password generator
pwgen -s 32 1
```

## Setting Up Secrets

### 1. Generate SSH Key Pair

On your local machine:

```bash
# Generate a new SSH key for deployment
ssh-keygen -t ed25519 -C "github-actions-deploy" -f ~/.ssh/swiftlog_deploy

# This creates:
# - ~/.ssh/swiftlog_deploy (private key - for GitHub secret)
# - ~/.ssh/swiftlog_deploy.pub (public key - for server)
```

### 2. Add Public Key to Server

Copy the public key to your deployment server:

```bash
# Copy public key to server
ssh-copy-id -i ~/.ssh/swiftlog_deploy.pub deploy@your-server.com

# Or manually:
cat ~/.ssh/swiftlog_deploy.pub
# Copy the output and add to ~/.ssh/authorized_keys on the server
```

### 3. Add Private Key to GitHub

1. Go to your repository on GitHub
2. Click **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Name: `DEPLOY_SSH_KEY`
5. Value: Copy the **entire contents** of `~/.ssh/swiftlog_deploy` including the header and footer:

```
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtz
...
-----END OPENSSH PRIVATE KEY-----
```

### 4. Add Other Secrets

Repeat the process for other secrets:

- `DEPLOY_HOST`: Your server hostname or IP
- `DEPLOY_USER`: The SSH username (e.g., `deploy`)
- `DEPLOY_PATH`: (Optional) Custom deployment path

## Server Preparation

On your deployment server:

### 1. Create Deployment User

```bash
# Create deploy user
sudo useradd -m -s /bin/bash deploy

# Add to docker group
sudo usermod -aG docker deploy

# Allow sudo without password (optional, for docker commands)
echo "deploy ALL=(ALL) NOPASSWD: /usr/bin/docker, /usr/bin/docker-compose" | sudo tee /etc/sudoers.d/deploy
```

### 2. Install Dependencies

```bash
# Install Docker
curl -fsSL https://get.docker.com | sh
sudo systemctl enable docker
sudo systemctl start docker

# Install Docker Compose
sudo apt-get update
sudo apt-get install docker-compose-plugin
```

### 3. Create Deployment Directory

```bash
sudo mkdir -p /opt/swiftlog
sudo chown deploy:deploy /opt/swiftlog
```

### 4. Configure Firewall

```bash
# Allow SSH
sudo ufw allow 22/tcp

# Allow SwiftLog services
sudo ufw allow 3000/tcp  # Frontend
sudo ufw allow 8080/tcp  # API
sudo ufw allow 8081/tcp  # WebSocket
sudo ufw allow 50051/tcp # gRPC

# Enable firewall
sudo ufw enable
```

## Workflows

### Release Workflow

**File**: `.github/workflows/release.yml`

**Triggered by**: Pushing a version tag (e.g., `v0.1.0`)

**Actions**:
1. Builds CLI for Linux, macOS, Windows (amd64 and arm64)
2. Creates GitHub Release
3. Uploads CLI binaries with SHA256 checksums
4. Builds Docker images for all services
5. Pushes images to GitHub Container Registry

**Usage**:
```bash
git tag v0.1.0
git push origin v0.1.0
```

### Deployment Workflow

**File**: `.github/workflows/deploy.yml`

**Triggered by**:
- Pushing a version tag (automatic deployment)
- Manual workflow dispatch (Settings → Actions → Deploy to Server → Run workflow)

**Actions**:
1. Downloads Docker images
2. Creates deployment package
3. Copies files to server via SCP
4. Executes deployment via SSH
5. Runs health checks
6. Reports status

**Usage**:

**Automatic (on tag push):**
```bash
git tag v0.1.0
git push origin v0.1.0
```

**Manual:**
1. Go to Actions tab
2. Select "Deploy to Server"
3. Click "Run workflow"
4. Choose environment (production/staging)
5. Click "Run workflow"

## GitHub Container Registry

Docker images are automatically published to:
```
ghcr.io/YOUR_USERNAME/swiftlog/api:TAG
ghcr.io/YOUR_USERNAME/swiftlog/ingestor:TAG
ghcr.io/YOUR_USERNAME/swiftlog/websocket:TAG
ghcr.io/YOUR_USERNAME/swiftlog/ai-worker:TAG
ghcr.io/YOUR_USERNAME/swiftlog/frontend:TAG
```

### Making Images Public

By default, images are private. To make them public:

1. Go to your GitHub profile → Packages
2. Click on the package (e.g., `swiftlog/api`)
3. Click "Package settings"
4. Scroll down to "Danger Zone"
5. Click "Change visibility" → "Public"

## Environments (Optional)

For better control, you can set up GitHub environments:

1. Go to Settings → Environments
2. Create environments: `production`, `staging`
3. Configure environment-specific secrets
4. Add protection rules (e.g., require approval for production)

### Environment-Specific Secrets

In each environment, you can override secrets:
- `DEPLOY_HOST` - Different server for staging
- `DEPLOY_PATH` - Different path for staging
- `NEXT_PUBLIC_API_URL` - Different API URL

## Troubleshooting

### SSH Authentication Failed

**Problem**: `Permission denied (publickey)`

**Solution**:
1. Verify public key is in `~/.ssh/authorized_keys` on server
2. Check private key in GitHub secret is complete and unmodified
3. Test SSH connection manually:
   ```bash
   ssh -i ~/.ssh/swiftlog_deploy deploy@your-server
   ```

### Docker Permission Denied

**Problem**: `permission denied while trying to connect to the Docker daemon socket`

**Solution**:
```bash
# On the server
sudo usermod -aG docker deploy
# Log out and back in, or:
newgrp docker
```

### Workflow Not Triggering

**Problem**: Pushed tag but workflow doesn't run

**Solution**:
1. Check tag format matches `v*.*.*` (e.g., `v0.1.0`)
2. Verify workflow files are in `.github/workflows/` on main branch
3. Check Actions tab for errors
4. Ensure GitHub Actions are enabled (Settings → Actions → General)

### Build Fails

**Problem**: CLI build or Docker build fails

**Solution**:
1. Check Actions tab for detailed error logs
2. Verify `go.mod` dependencies are correct
3. Test build locally:
   ```bash
   cd cli && go build .
   docker build -f backend/Dockerfile.api backend/
   ```

### Deployment Fails

**Problem**: Deployment step fails

**Solution**:
1. SSH into server and check logs: `docker compose logs`
2. Verify `.env` file exists and is configured
3. Check disk space: `df -h`
4. Verify Docker is running: `docker ps`
5. Check server logs: `journalctl -u docker`

## Testing CI/CD Locally

### Test CLI Build

```bash
cd cli
GOOS=linux GOARCH=amd64 go build -o swiftlog-linux-amd64 .
./swiftlog-linux-amd64 --version
```

### Test Docker Build

```bash
cd backend
docker build -f Dockerfile.api -t swiftlog-api:test .
docker run --rm swiftlog-api:test --version
```

### Test Deployment Script

```bash
./deploy.sh --help
./deploy.sh -h localhost -u $USER -v latest
```

## Security Best Practices

1. **Never commit secrets** to the repository
2. **Use SSH keys** instead of passwords
3. **Rotate SSH keys** periodically
4. **Limit deploy user permissions** on server
5. **Use environment-specific secrets** for staging/production
6. **Enable branch protection** on main branch
7. **Require approval** for production deployments
8. **Monitor deployment logs** for suspicious activity

## Reference Links

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [GitHub Encrypted Secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
- [Deployment Environments](https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment)
