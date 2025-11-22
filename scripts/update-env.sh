#!/bin/bash
# SwiftLog Environment Variables Update Script
# This script updates .env file with values from environment variables or command line

set -e

ENV_FILE="${1:-.env}"
BACKUP_SUFFIX=".backup.$(date +%Y%m%d_%H%M%S)"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "🔧 Updating environment variables in: $ENV_FILE"

# Backup existing .env if it exists
if [ -f "$ENV_FILE" ]; then
  cp "$ENV_FILE" "${ENV_FILE}${BACKUP_SUFFIX}"
  echo -e "${GREEN}✓${NC} Backed up existing file to: ${ENV_FILE}${BACKUP_SUFFIX}"
fi

# Create .env from .env.example if it doesn't exist
if [ ! -f "$ENV_FILE" ]; then
  if [ -f ".env.example" ]; then
    cp .env.example "$ENV_FILE"
    echo -e "${GREEN}✓${NC} Created $ENV_FILE from .env.example"
  else
    echo -e "${YELLOW}⚠${NC}  No .env.example found, creating empty $ENV_FILE"
    touch "$ENV_FILE"
  fi
fi

# Function to update or add an environment variable
update_env() {
  local key="$1"
  local value="$2"

  if [ -z "$value" ]; then
    echo -e "${YELLOW}⊘${NC} Skipping $key (empty value)"
    return
  fi

  # Escape special characters in value for sed
  local escaped_value=$(echo "$value" | sed 's/[&/\]/\\&/g')

  # Check if key exists
  if grep -q "^${key}=" "$ENV_FILE"; then
    # Update existing key
    sed -i.tmp "s|^${key}=.*|${key}=${escaped_value}|" "$ENV_FILE"
    echo -e "${GREEN}✓${NC} Updated: $key"
  else
    # Add new key
    echo "${key}=${value}" >> "$ENV_FILE"
    echo -e "${GREEN}+${NC} Added: $key"
  fi

  # Remove sed backup file
  rm -f "${ENV_FILE}.tmp"
}

# Update environment variables from environment or arguments
# Format: VAR_NAME or VAR_NAME=value

# Database
if [ -n "$POSTGRES_PASSWORD" ]; then
  update_env "POSTGRES_PASSWORD" "$POSTGRES_PASSWORD"
fi

# Security
if [ -n "$JWT_SECRET" ]; then
  update_env "JWT_SECRET" "$JWT_SECRET"
fi

# Admin
if [ -n "$ADMIN_USERNAME" ]; then
  update_env "ADMIN_USERNAME" "$ADMIN_USERNAME"
fi

if [ -n "$ADMIN_PASSWORD" ]; then
  update_env "ADMIN_PASSWORD" "$ADMIN_PASSWORD"
fi

# Application
if [ -n "$ENVIRONMENT" ]; then
  update_env "ENVIRONMENT" "$ENVIRONMENT"
fi

if [ -n "$LOG_LEVEL" ]; then
  update_env "LOG_LEVEL" "$LOG_LEVEL"
fi

# Frontend URLs (for build args)
if [ -n "$NEXT_PUBLIC_API_URL" ]; then
  update_env "NEXT_PUBLIC_API_URL" "$NEXT_PUBLIC_API_URL"
fi

if [ -n "$NEXT_PUBLIC_WS_URL" ]; then
  update_env "NEXT_PUBLIC_WS_URL" "$NEXT_PUBLIC_WS_URL"
fi

echo ""
echo -e "${GREEN}✅ Environment file updated successfully!${NC}"
echo ""
echo "Current configuration:"
echo "---"
# Display non-sensitive parts
grep -E "^(ENVIRONMENT|LOG_LEVEL|ADMIN_USERNAME)=" "$ENV_FILE" 2>/dev/null || true
echo "---"
echo ""
echo "⚠️  Remember to review $ENV_FILE and set any missing values!"
