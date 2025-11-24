#!/bin/bash
set -e

command -v gh &> /dev/null || { echo "Install gh CLI first"; exit 1; }
gh auth status &> /dev/null || { echo "Run: gh auth login"; exit 1; }

echo "Clearing existing secrets..."
gh secret list | awk '{print $1}' | xargs -I {} gh secret delete {} 2>/dev/null || true

echo "Generating random keys..."
POSTGRES_PASSWORD=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 32)
ENCRYPTION_KEY=$(openssl rand -base64 32)
ADMIN_PASSWORD=$(openssl rand -base64 16)

read -p "Server IP/domain: " DEPLOY_HOST
read -p "SSH user [ubuntu]: " DEPLOY_USER
DEPLOY_USER=${DEPLOY_USER:-ubuntu}
read -p "Deploy path [/opt/swiftlog]: " DEPLOY_PATH
DEPLOY_PATH=${DEPLOY_PATH:-/opt/swiftlog}
read -p "SSH key path [~/.ssh/id_ed25519]: " SSH_KEY_PATH
SSH_KEY_PATH="${SSH_KEY_PATH/#\~/$HOME}"
SSH_KEY_PATH=${SSH_KEY_PATH:-$HOME/.ssh/id_ed25519}
[ ! -f "$SSH_KEY_PATH" ] && { echo "SSH key not found: $SSH_KEY_PATH"; exit 1; }

read -p "Public URL (optional): " PUBLIC_URL
read -p "NGINX port [80]: " NGINX_PORT
NGINX_PORT=${NGINX_PORT:-80}
read -p "gRPC port [50051]: " GRPC_PORT
GRPC_PORT=${GRPC_PORT:-50051}
read -p "CORS origins (optional): " CORS_ORIGINS
read -p "Admin username [admin]: " ADMIN_USERNAME
ADMIN_USERNAME=${ADMIN_USERNAME:-admin}
read -p "Log level [info]: " LOG_LEVEL
LOG_LEVEL=${LOG_LEVEL:-info}

echo ""
echo "Setting GitHub secrets..."
gh secret set DEPLOY_HOST --body "$DEPLOY_HOST"
gh secret set DEPLOY_USER --body "$DEPLOY_USER"
gh secret set DEPLOY_PATH --body "$DEPLOY_PATH"
gh secret set DEPLOY_SSH_KEY < "$SSH_KEY_PATH"
gh secret set POSTGRES_PASSWORD --body "$POSTGRES_PASSWORD"
gh secret set JWT_SECRET --body "$JWT_SECRET"
gh secret set ENCRYPTION_KEY --body "$ENCRYPTION_KEY"
gh secret set ADMIN_USERNAME --body "$ADMIN_USERNAME"
gh secret set ADMIN_PASSWORD --body "$ADMIN_PASSWORD"
gh secret set LOG_LEVEL --body "$LOG_LEVEL"
[ -n "$PUBLIC_URL" ] && gh secret set PUBLIC_URL --body "$PUBLIC_URL"
[ -n "$NGINX_PORT" ] && gh secret set NGINX_PORT --body "$NGINX_PORT"
[ -n "$GRPC_PORT" ] && gh secret set GRPC_PORT --body "$GRPC_PORT"
[ -n "$CORS_ORIGINS" ] && gh secret set CORS_ORIGINS --body "$CORS_ORIGINS"

echo ""
echo "Secrets configured:"
gh secret list

echo ""
echo "IMPORTANT - Save these credentials:"
echo "Server: $DEPLOY_HOST"
echo "Admin: $ADMIN_USERNAME"
echo "Password: $ADMIN_PASSWORD"
echo "DB Password: $POSTGRES_PASSWORD"
echo "JWT Secret: $JWT_SECRET"
echo "Encryption Key: $ENCRYPTION_KEY"
