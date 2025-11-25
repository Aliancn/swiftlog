#!/bin/bash
set -e

command -v gh &> /dev/null || { echo "Install gh CLI first"; exit 1; }
gh auth status &> /dev/null || { echo "Run: gh auth login"; exit 1; }

echo "=== SwiftLog GitHub Secrets Setup ==="
echo ""
echo "This script will configure all required GitHub secrets for deployment."
echo ""

# Generate random secure keys
echo "Generating secure random keys..."
POSTGRES_PASSWORD=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 32)
ENCRYPTION_KEY=$(openssl rand -base64 32)
ADMIN_PASSWORD=$(openssl rand -base64 16)

# Deployment configuration
echo ""
echo "--- Deployment Configuration ---"
read -p "Server IP/domain: " DEPLOY_HOST
read -p "SSH user [ubuntu]: " DEPLOY_USER
DEPLOY_USER=${DEPLOY_USER:-ubuntu}
read -p "Deploy path [/opt/swiftlog]: " DEPLOY_PATH
DEPLOY_PATH=${DEPLOY_PATH:-/opt/swiftlog}
read -p "SSH key path [~/.ssh/id_ed25519]: " SSH_KEY_PATH
SSH_KEY_PATH="${SSH_KEY_PATH/#\~/$HOME}"
SSH_KEY_PATH=${SSH_KEY_PATH:-$HOME/.ssh/id_ed25519}
[ ! -f "$SSH_KEY_PATH" ] && { echo "SSH key not found: $SSH_KEY_PATH"; exit 1; }

# Database configuration
echo ""
echo "--- Database Configuration ---"
read -p "PostgreSQL user [swiftlog]: " POSTGRES_USER
POSTGRES_USER=${POSTGRES_USER:-swiftlog}
read -p "PostgreSQL database name [swiftlog]: " POSTGRES_DB
POSTGRES_DB=${POSTGRES_DB:-swiftlog}
echo "PostgreSQL password: (auto-generated)"

# Port configuration
echo ""
echo "--- Port Configuration ---"
read -p "NGINX port [80]: " NGINX_PORT
NGINX_PORT=${NGINX_PORT:-80}
read -p "gRPC port [50051]: " GRPC_PORT
GRPC_PORT=${GRPC_PORT:-50051}
read -p "API port [8080]: " API_PORT
API_PORT=${API_PORT:-8080}
read -p "WebSocket port [8081]: " WS_PORT
WS_PORT=${WS_PORT:-8081}

# Application configuration
echo ""
echo "--- Application Configuration ---"
read -p "Public URL (optional, e.g., https://logs.example.com): " PUBLIC_URL
read -p "CORS origins (optional, defaults to PUBLIC_URL): " CORS_ORIGINS
read -p "Environment [production]: " ENVIRONMENT
ENVIRONMENT=${ENVIRONMENT:-production}
read -p "Log level [info]: " LOG_LEVEL
LOG_LEVEL=${LOG_LEVEL:-info}

# Security configuration
echo ""
echo "--- Security Configuration ---"
read -p "Admin username [admin]: " ADMIN_USERNAME
ADMIN_USERNAME=${ADMIN_USERNAME:-admin}
echo "Admin password: (auto-generated)"
read -p "JWT expiration [2h]: " JWT_EXPIRATION
JWT_EXPIRATION=${JWT_EXPIRATION:-2h}
echo "JWT secret: (auto-generated)"
echo "Encryption key: (auto-generated)"

echo ""
echo "Clearing existing secrets..."
gh secret list | awk '{print $1}' | xargs -I {} gh secret delete {} 2>/dev/null || true

echo ""
echo "Setting GitHub secrets..."

# Deployment secrets
gh secret set DEPLOY_HOST --body "$DEPLOY_HOST"
gh secret set DEPLOY_USER --body "$DEPLOY_USER"
gh secret set DEPLOY_PATH --body "$DEPLOY_PATH"
gh secret set DEPLOY_SSH_KEY < "$SSH_KEY_PATH"

# Database secrets
gh secret set POSTGRES_USER --body "$POSTGRES_USER"
gh secret set POSTGRES_PASSWORD --body "$POSTGRES_PASSWORD"
gh secret set POSTGRES_DB --body "$POSTGRES_DB"

# Port configuration
gh secret set NGINX_PORT --body "$NGINX_PORT"
gh secret set GRPC_PORT --body "$GRPC_PORT"
gh secret set API_PORT --body "$API_PORT"
gh secret set WS_PORT --body "$WS_PORT"

# Security secrets
gh secret set JWT_SECRET --body "$JWT_SECRET"
gh secret set JWT_EXPIRATION --body "$JWT_EXPIRATION"
gh secret set ENCRYPTION_KEY --body "$ENCRYPTION_KEY"
gh secret set ADMIN_USERNAME --body "$ADMIN_USERNAME"
gh secret set ADMIN_PASSWORD --body "$ADMIN_PASSWORD"

# Application configuration
gh secret set ENVIRONMENT --body "$ENVIRONMENT"
gh secret set LOG_LEVEL --body "$LOG_LEVEL"
[ -n "$PUBLIC_URL" ] && gh secret set PUBLIC_URL --body "$PUBLIC_URL"
[ -n "$CORS_ORIGINS" ] && gh secret set CORS_ORIGINS --body "$CORS_ORIGINS"

echo ""
echo "✓ Secrets configured successfully!"
echo ""
echo "Configured secrets:"
gh secret list

echo ""
echo "=========================================="
echo "IMPORTANT - Save these credentials:"
echo "=========================================="
echo "Server: $DEPLOY_HOST"
echo "Admin Username: $ADMIN_USERNAME"
echo "Admin Password: $ADMIN_PASSWORD"
echo "PostgreSQL User: $POSTGRES_USER"
echo "PostgreSQL Password: $POSTGRES_PASSWORD"
echo "PostgreSQL Database: $POSTGRES_DB"
echo "JWT Secret: $JWT_SECRET"
echo "Encryption Key: $ENCRYPTION_KEY"
echo ""
echo "Store these credentials securely!"
echo "=========================================="
