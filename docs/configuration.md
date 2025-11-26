---
layout: default
title: Configuration
nav_order: 6
description: "Complete configuration reference for SwiftLog services and components"
---

# SwiftLog Configuration Guide

Complete reference for configuring SwiftLog services and components.

## Environment Variables

SwiftLog uses environment variables for configuration. All variables are defined in the `.env` file.

### Quick Setup

```bash
# Copy the example configuration
cp .env.example .env

# Edit the configuration
nano .env
```

---

## Required Variables

### Database Configuration

#### POSTGRES_PASSWORD

PostgreSQL database password.

**Required:** Yes (Production)
**Default:** None
**Example:** `$(openssl rand -base64 32)`

```bash
POSTGRES_PASSWORD=your-secure-password-here
```

**Security Note:** Use a strong password (32+ characters) in production. Generate with:
```bash
openssl rand -base64 32
```

---

### Security Configuration

#### JWT_SECRET

Secret key for JWT token signing.

**Required:** Yes (Production)
**Default:** None
**Example:** `$(openssl rand -base64 32)`

```bash
JWT_SECRET=your-jwt-secret-here
```

**Security Note:** Must be at least 32 characters. Generate with:
```bash
openssl rand -base64 32
```

#### ENCRYPTION_KEY

Key for encrypting sensitive data (API keys).

**Required:** Yes (Production)
**Default:** None
**Example:** `$(openssl rand -base64 32)`

```bash
ENCRYPTION_KEY=your-encryption-key-here
```

**Security Note:** Must be exactly 32 bytes. Generate with:
```bash
openssl rand -base64 32
```

---

## Optional Variables

### Application Configuration

#### ENVIRONMENT

Application environment mode.

**Required:** No
**Default:** `production`
**Options:** `production`, `staging`, `development`

```bash
ENVIRONMENT=production
```

#### LOG_LEVEL

Logging verbosity level.

**Required:** No
**Default:** `info`
**Options:** `debug`, `info`, `warn`, `error`

```bash
LOG_LEVEL=info
```

**Recommendations:**
- Development: `debug`
- Staging: `info`
- Production: `warn` or `info`

---

### CORS Configuration

#### CORS_ORIGINS

Comma-separated list of allowed origins for CORS.

**Required:** Strongly recommended for production
**Default:** `http://localhost,http://127.0.0.1`

```bash
# Single origin
CORS_ORIGINS=https://logs.yourdomain.com

# Multiple origins
CORS_ORIGINS=https://logs.yourdomain.com,https://app.yourdomain.com
```

#### PUBLIC_URL

Public URL of the application (used for CORS if CORS_ORIGINS not set).

**Required:** No
**Default:** None

```bash
PUBLIC_URL=https://logs.yourdomain.com
```

---

### Database Configuration

#### POSTGRES_USER

PostgreSQL username.

**Required:** No
**Default:** `swiftlog`

```bash
POSTGRES_USER=swiftlog
```

#### POSTGRES_DB

PostgreSQL database name.

**Required:** No
**Default:** `swiftlog`

```bash
POSTGRES_DB=swiftlog
```

#### POSTGRES_HOST

PostgreSQL host address.

**Required:** No
**Default:** `postgres` (Docker service name)

```bash
POSTGRES_HOST=postgres
```

#### POSTGRES_PORT

PostgreSQL port.

**Required:** No
**Default:** `5432`

```bash
POSTGRES_PORT=5432
```

---

### AI Configuration

#### OPENAI_API_KEY

OpenAI API key for AI analysis features.

**Required:** Only if using AI features
**Default:** None

```bash
OPENAI_API_KEY=sk-your-openai-key-here
```

**Note:** If not set, AI analysis features will be disabled.

#### OPENAI_BASE_URL

Custom OpenAI-compatible API endpoint.

**Required:** No
**Default:** `https://api.openai.com/v1`

```bash
# Azure OpenAI
OPENAI_BASE_URL=https://your-resource.openai.azure.com/openai/deployments/your-deployment

# LocalAI (self-hosted)
OPENAI_BASE_URL=http://localhost:8080/v1

# Other provider
OPENAI_BASE_URL=https://api.custom-provider.com/v1
```

**Note:** URL should end with `/v1` (without `/chat/completions`).

#### OPENAI_MODEL

AI model to use for analysis.

**Required:** No
**Default:** `gpt-4o-mini`

```bash
# OpenAI models
OPENAI_MODEL=gpt-4o-mini
OPENAI_MODEL=gpt-4o
OPENAI_MODEL=gpt-3.5-turbo

# Azure OpenAI
OPENAI_MODEL=gpt-35-turbo

# LocalAI/Ollama
OPENAI_MODEL=llama2
```

---

### Port Configuration

#### NGINX_PORT

Port for Nginx frontend proxy.

**Required:** No
**Default:** `80`

```bash
NGINX_PORT=80
# Or for HTTPS
NGINX_PORT=443
```

#### GRPC_PORT

Port for gRPC Ingestor service.

**Required:** No
**Default:** `50051`

```bash
GRPC_PORT=50051
```

#### API_PORT

Internal port for REST API service.

**Required:** No
**Default:** `8080`

```bash
API_PORT=8080
```

#### WS_PORT

Internal port for WebSocket service.

**Required:** No
**Default:** `8081`

```bash
WS_PORT=8081
```

---

### Authentication Configuration

#### ADMIN_USERNAME

Initial admin username.

**Required:** No
**Default:** `admin`

```bash
ADMIN_USERNAME=admin
```

#### ADMIN_PASSWORD

Initial admin password.

**Required:** Recommended
**Default:** None
**Example:** `$(openssl rand -base64 16)`

```bash
ADMIN_PASSWORD=your-admin-password
```

#### JWT_EXPIRATION

JWT token expiration duration.

**Required:** No
**Default:** `2h`

```bash
JWT_EXPIRATION=2h
# Other examples: 1h, 30m, 24h
```

**Recommendations:**
- Short-lived: `1h` or `2h` (more secure)
- Long-lived: `24h` (more convenient)

---

## Configuration Examples

### Development Environment

```bash
# .env for development
ENVIRONMENT=development
LOG_LEVEL=debug

POSTGRES_PASSWORD=devpassword
JWT_SECRET=dev-jwt-secret-change-in-prod
ENCRYPTION_KEY=dev-encryption-key-change-prod

# CORS for local frontend
CORS_ORIGINS=http://localhost:3000

# OpenAI (optional for testing)
OPENAI_API_KEY=sk-your-key-here
```

### Production Environment

```bash
# .env for production
ENVIRONMENT=production
LOG_LEVEL=info

# Strong passwords (generated)
POSTGRES_PASSWORD=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 32)
ENCRYPTION_KEY=$(openssl rand -base64 32)
ADMIN_PASSWORD=$(openssl rand -base64 16)

# Public access
PUBLIC_URL=https://logs.yourdomain.com
CORS_ORIGINS=https://logs.yourdomain.com

# Ports
NGINX_PORT=443  # HTTPS
GRPC_PORT=50051

# AI (if using)
OPENAI_API_KEY=sk-prod-key-here
OPENAI_MODEL=gpt-4o-mini

# JWT expiration (shorter for security)
JWT_EXPIRATION=1h
```

### Staging Environment

```bash
# .env for staging
ENVIRONMENT=staging
LOG_LEVEL=info

POSTGRES_PASSWORD=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 32)
ENCRYPTION_KEY=$(openssl rand -base64 32)

PUBLIC_URL=https://staging-logs.yourdomain.com
CORS_ORIGINS=https://staging-logs.yourdomain.com

# Test with cheaper model
OPENAI_MODEL=gpt-3.5-turbo
```

---

## CLI Configuration

The SwiftLog CLI tool has its own configuration file.

### Configuration File Location

- **Linux/macOS:** `~/.swiftlog/config.yaml`
- **Windows:** `%USERPROFILE%\.swiftlog\config.yaml`

### Configuration Format

```yaml
server: localhost:50051
token: your-api-token-here
default_project: myapp
default_group: main
```

### CLI Configuration Commands

```bash
# Set server and token
swiftlog config set --token YOUR_TOKEN --server localhost:50051

# Set default project and group
swiftlog config set --default-project myapp
swiftlog config set --default-group main

# View current configuration
swiftlog config get

# Show config file path
swiftlog config path
```

---

## Creating API Tokens

### For Testing (Database)

```bash
docker compose exec postgres psql -U swiftlog -d swiftlog -c \
  "INSERT INTO api_tokens (user_id, token_hash, name)
   SELECT id, encode(sha256('test-token'::bytea), 'hex'), 'Test Token'
   FROM users LIMIT 1
   RETURNING id;"
```

Then use `test-token` as your API token.

### For Production

Use the web interface or API endpoint to create tokens securely (future feature).

---

## Docker Compose Configuration

### Service Configuration

Services are configured in `docker-compose.yaml`. Most settings are pulled from `.env`.

### Overriding Defaults

You can override any service configuration:

```yaml
# docker-compose.override.yaml
services:
  api:
    environment:
      - LOG_LEVEL=debug
    ports:
      - "8080:8080"  # Expose for debugging
```

### Resource Limits

Add resource limits for production:

```yaml
services:
  postgres:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G
```

---

## Security Best Practices

### 1. Strong Secrets

Generate all secrets with proper entropy:

```bash
# Generate strong secrets (32 bytes)
openssl rand -base64 32

# Or using /dev/urandom
head -c 32 /dev/urandom | base64
```

### 2. Rotate Secrets

Regularly rotate sensitive keys:
- JWT_SECRET: Every 3-6 months
- ENCRYPTION_KEY: Every 6-12 months (requires data migration)
- ADMIN_PASSWORD: Every 3 months

### 3. Restrict CORS

In production, always set specific origins:

```bash
# Good
CORS_ORIGINS=https://logs.yourdomain.com

# Bad (too permissive)
CORS_ORIGINS=*
```

### 4. Secure Passwords

- Use password managers for storing credentials
- Never commit `.env` files to version control
- Use different passwords for each environment

### 5. Network Security

- Use TLS/HTTPS in production
- Restrict database access to internal network
- Use firewall rules to limit access

---

## Environment-Specific Configuration

### GitHub Actions Deployment

Configure via GitHub Secrets. See [Deployment Guide](deployment) for details.

**Required Secrets:**
- `POSTGRES_PASSWORD`
- `JWT_SECRET`
- `ENCRYPTION_KEY`
- `ADMIN_PASSWORD`
- `DEPLOY_HOST`
- `DEPLOY_USER`
- `DEPLOY_SSH_KEY`

**Recommended Secrets:**
- `PUBLIC_URL`
- `CORS_ORIGINS`
- `OPENAI_API_KEY`

### Docker Swarm / Kubernetes

For orchestrated deployments, use secrets management:

**Docker Swarm:**
```bash
echo "my-secret" | docker secret create postgres_password -
```

**Kubernetes:**
```bash
kubectl create secret generic swiftlog-secrets \
  --from-literal=postgres-password='...' \
  --from-literal=jwt-secret='...'
```

---

## Troubleshooting

### Configuration Not Loading

1. **Check file exists:**
   ```bash
   ls -la .env
   ```

2. **Verify format:**
   - No spaces around `=`
   - No quotes needed (Docker Compose handles them)
   ```bash
   # Good
   JWT_SECRET=abc123

   # Bad
   JWT_SECRET = "abc123"
   ```

3. **Restart services:**
   ```bash
   docker compose down
   docker compose up -d
   ```

### CORS Errors

If you see CORS errors in the browser:

1. **Check CORS_ORIGINS:**
   ```bash
   docker compose exec api env | grep CORS
   ```

2. **Update and restart:**
   ```bash
   # Edit .env
   CORS_ORIGINS=https://your-domain.com

   # Restart
   docker compose restart api websocket
   ```

### Database Connection Errors

1. **Verify credentials:**
   ```bash
   docker compose exec postgres psql -U swiftlog -d swiftlog
   ```

2. **Check environment:**
   ```bash
   docker compose exec api env | grep POSTGRES
   ```

---

## Configuration Validation

### Check All Services

```bash
# View environment for all services
docker compose config

# Check specific service
docker compose exec api env
```

### Validate API Configuration

```bash
# Test API health
curl http://localhost:8080/health

# Check CORS headers
curl -H "Origin: https://your-domain.com" \
     -H "Access-Control-Request-Method: GET" \
     -X OPTIONS \
     http://localhost:8080/health
```

---

## Related Documentation

- [Getting Started](getting-started)
- [Deployment Guide](deployment)
- [Architecture](architecture)
- [API Reference](api-reference)