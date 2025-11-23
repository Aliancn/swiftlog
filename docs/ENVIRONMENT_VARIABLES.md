# Environment Variables Configuration Guide

This guide explains how to manage environment variables in SwiftLog, especially during CI/CD deployments.

## Table of Contents

- [Available Environment Variables](#available-environment-variables)
- [Configuration Methods](#configuration-methods)
- [CI/CD Integration](#cicd-integration)
- [Local Development](#local-development)
- [Security Best Practices](#security-best-practices)

## Available Environment Variables

### Database

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `POSTGRES_PASSWORD` | Yes | `changeme` | PostgreSQL password |

### Security

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `JWT_SECRET` | Yes (production) | `dev-secret-key-change-in-production` | JWT signing secret |
| `ADMIN_USERNAME` | No | `admin` | Default admin username |
| `ADMIN_PASSWORD` | Yes (production) | `admin123` | Default admin password |

### Application

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ENVIRONMENT` | No | `production` | Environment: `production`, `staging`, `development` |
| `LOG_LEVEL` | No | `info` | Log level: `debug`, `info`, `warn`, `error` |

### AI Integration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OPENAI_API_KEY` | No* | - | OpenAI API key for AI analysis features (manual configuration only) |
| `OPENAI_BASE_URL` | No | `https://api.openai.com/v1` | OpenAI API endpoint (for Azure OpenAI, etc.) |
| `OPENAI_MODEL` | No | `gpt-4o-mini` | OpenAI model to use |

*Required if you want to use AI analysis features. **Note:** This must be configured manually on the server - it is NOT automatically injected during deployment.

### Frontend (Build-time)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `NEXT_PUBLIC_API_URL` | No | `http://localhost:8080/api/v1` | API URL for frontend |
| `NEXT_PUBLIC_WS_URL` | No | `ws://localhost:8081` | WebSocket URL for frontend |

## Configuration Methods

### Method 1: GitHub Actions (Recommended for Production)

Configure environment variables through GitHub Secrets for automated deployments.

#### Setup Steps

1. **Add Secrets to GitHub**

   Go to: `Settings → Secrets and variables → Actions → New repository secret`

   Add these secrets:
   ```
   POSTGRES_PASSWORD      - Strong database password
   JWT_SECRET            - Random secret key (generate with: openssl rand -base64 32)
   ADMIN_PASSWORD        - Strong admin password
   ENVIRONMENT           - Environment name (optional, default: production)
   LOG_LEVEL             - Log level (optional, default: info)
   ```

   **Note:** `OPENAI_API_KEY` is NOT configured via GitHub Secrets. If needed, add it manually to the `.env` file on the server after deployment.

2. **Deploy with Automatic Variable Injection**

   ```bash
   # Create and push a tag
   git tag v0.1.0
   git push origin v0.1.0
   ```

   GitHub Actions will automatically:
   - Create `.env` file from `.env.example`
   - Inject values from GitHub Secrets
   - Deploy to your server

#### How It Works

The deployment workflow (`.github/workflows/deploy.yml`) automatically:

1. Copies `scripts/update-env.sh` to the deployment package
2. Transfers environment variables from GitHub Secrets via SSH
3. Runs `update-env.sh` on the server to update `.env` file
4. Preserves existing values not specified in secrets

### Method 2: Manual Deployment Script

Use the interactive deployment script for manual deployments.

#### Interactive Mode

```bash
./deploy.sh -h your-server.com -u deploy -v v0.1.0
```

The script will:
1. Prompt you to update environment variables
2. Ask for values interactively
3. Apply them to the `.env` file on the server

#### Pre-set Environment Variables

```bash
# Export variables before running the script
export POSTGRES_PASSWORD="my-secure-password"
export JWT_SECRET="my-secret-key"
export ADMIN_PASSWORD="my-admin-password"
export OPENAI_API_KEY="sk-..."

# Deploy
./deploy.sh -h your-server.com -u deploy -v v0.1.0
```

### Method 3: Direct Server Configuration

#### Using the update-env.sh Script

SSH to your server and use the update script:

```bash
# SSH to server
ssh deploy@your-server.com

# Navigate to deployment directory
cd /opt/swiftlog

# Set environment variables
export POSTGRES_PASSWORD="new-password"
export JWT_SECRET="new-secret"
export ADMIN_PASSWORD="new-password"

# Update .env file
./update-env.sh .env

# Restart services
docker compose restart
```

#### Manual Editing

```bash
# SSH to server
ssh deploy@your-server.com

# Edit .env file
nano /opt/swiftlog/.env

# Restart services
cd /opt/swiftlog
docker compose restart
```

### Method 4: Environment-Specific Files

For managing multiple environments (staging, production):

```bash
# Create environment-specific files
.env.production
.env.staging

# Use specific file
cp .env.production .env
docker compose up -d
```

## CI/CD Integration

### GitHub Actions Workflow

The deployment workflow automatically handles environment variables:

```yaml
# .github/workflows/deploy.yml
env:
  POSTGRES_PASSWORD: ${{ secrets.POSTGRES_PASSWORD }}
  JWT_SECRET: ${{ secrets.JWT_SECRET }}
  ADMIN_PASSWORD: ${{ secrets.ADMIN_PASSWORD }}
  OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
  ENVIRONMENT: ${{ secrets.ENVIRONMENT || 'production' }}
  LOG_LEVEL: ${{ secrets.LOG_LEVEL || 'info' }}
```

### Variable Precedence

When deploying, variables are merged in this order (later overrides earlier):

1. `.env.example` (base template)
2. Existing `.env` (preserved values)
3. GitHub Secrets / Environment variables (override specific keys)

### Backup and Restore

Deployments automatically:
- Backup existing `.env` before updates
- Restore on failure
- Keep backup with timestamp: `.env.backup.YYYYMMDD_HHMMSS`

## Local Development

### Quick Start

```bash
# Copy example
cp .env.example .env

# Edit values
nano .env

# Start services
docker compose up -d
```

### Using update-env.sh

```bash
# Set variables
export POSTGRES_PASSWORD="dev-password"
export JWT_SECRET="dev-secret"

# Update .env
./scripts/update-env.sh .env

# Start services
docker compose up -d
```

### Development Overrides

For local development, you can create `.env.local`:

```bash
# .env.local (not committed to git)
POSTGRES_PASSWORD=dev-password
JWT_SECRET=dev-secret
LOG_LEVEL=debug
```

## Security Best Practices

### 1. Strong Passwords

```bash
# Generate strong passwords
openssl rand -base64 32    # For JWT_SECRET
openssl rand -base64 16    # For passwords
```

### 2. Never Commit Secrets

Ensure `.env` is in `.gitignore`:

```gitignore
.env
.env.local
.env.*.local
.env.production
.env.staging
```

### 3. Rotate Secrets Regularly

```bash
# Update JWT_SECRET
export JWT_SECRET="$(openssl rand -base64 32)"
./scripts/update-env.sh .env

# Restart services
docker compose restart
```

### 4. Use GitHub Secrets

- Never put secrets in workflow files
- Use GitHub Secrets for all sensitive data
- Enable secret scanning in repository settings

### 5. Limit Access

- Use environment-specific secrets for staging/production
- Configure GitHub environments with approval requirements
- Use separate SSH keys for different environments

## Troubleshooting

### Variables Not Updating

**Problem**: Changed a secret but deployment still uses old value

**Solution**:
```bash
# SSH to server
ssh deploy@your-server.com

# Check current values (non-sensitive only)
cd /opt/swiftlog
grep ENVIRONMENT .env

# Force update
export POSTGRES_PASSWORD="new-value"
./update-env.sh .env

# Restart
docker compose restart
```

### Script Permissions

**Problem**: `update-env.sh: permission denied`

**Solution**:
```bash
chmod +x scripts/update-env.sh
# or on server:
chmod +x /opt/swiftlog/update-env.sh
```

### Escaped Characters

**Problem**: Special characters in values not working

**Solution**:
The `update-env.sh` script automatically escapes special characters. If you're manually editing `.env`, quote values:

```bash
# Good
JWT_SECRET="secret/with/slashes"
POSTGRES_PASSWORD="pass&word#123"

# Bad (may break)
JWT_SECRET=secret/with/slashes
```

## Examples

### Example 1: First Deployment

```bash
# On GitHub
1. Add secrets: POSTGRES_PASSWORD, JWT_SECRET, ADMIN_PASSWORD
2. Push tag: git tag v0.1.0 && git push origin v0.1.0
3. Wait for deployment to complete
```

### Example 2: Update Single Variable

```bash
# Update OPENAI_API_KEY (manual only)
ssh deploy@server
cd /opt/swiftlog
nano .env  # Add or update: OPENAI_API_KEY=sk-new-key
docker compose restart ai-worker

# Or using update-env.sh:
ssh deploy@server
cd /opt/swiftlog
export OPENAI_API_KEY="sk-new-key"
./update-env.sh .env
docker compose restart ai-worker
```

### Example 3: Multiple Environments

```bash
# Create environment-specific secrets in GitHub
production:
  - POSTGRES_PASSWORD: prod-password
  - JWT_SECRET: prod-secret

staging:
  - POSTGRES_PASSWORD: staging-password
  - JWT_SECRET: staging-secret

# Deploy to specific environment
# GitHub Actions → Deploy → Run workflow → Select environment
```

## Reference

### Complete .env Template

See [`.env.example`](../.env.example) for the complete template with all available variables and their descriptions.

### Related Documentation

- [Deployment Guide](DEPLOYMENT.md)
- [GitHub Actions Setup](.github/SETUP.md)
- [Security Best Practices](../CONTRIBUTING.md#security)

## Support

- **Environment Issues**: Check logs with `docker compose logs`
- **Variable Problems**: Verify with `docker compose config`
- **GitHub Actions**: Check workflow logs in Actions tab
