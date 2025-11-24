#!/bin/bash
set -e

ENV_FILE="${1:-.env}"

if [ -f "$ENV_FILE" ]; then
  cp "$ENV_FILE" "${ENV_FILE}.backup.$(date +%Y%m%d_%H%M%S)"
fi

if [ ! -f "$ENV_FILE" ] && [ -f ".env.example" ]; then
  cp .env.example "$ENV_FILE"
fi

update_env() {
  local key="$1"
  local value="$2"
  [ -z "$value" ] && return

  local escaped_value=$(echo "$value" | sed 's/[&/\]/\\&/g')

  if grep -q "^${key}=" "$ENV_FILE"; then
    sed -i.tmp "s|^${key}=.*|${key}=${escaped_value}|" "$ENV_FILE"
  else
    echo "${key}=${value}" >> "$ENV_FILE"
  fi
  rm -f "${ENV_FILE}.tmp"
}

[ -n "$POSTGRES_PASSWORD" ] && update_env "POSTGRES_PASSWORD" "$POSTGRES_PASSWORD"
[ -n "$JWT_SECRET" ] && update_env "JWT_SECRET" "$JWT_SECRET"
[ -n "$ADMIN_USERNAME" ] && update_env "ADMIN_USERNAME" "$ADMIN_USERNAME"
[ -n "$ADMIN_PASSWORD" ] && update_env "ADMIN_PASSWORD" "$ADMIN_PASSWORD"
[ -n "$ENVIRONMENT" ] && update_env "ENVIRONMENT" "$ENVIRONMENT"
[ -n "$LOG_LEVEL" ] && update_env "LOG_LEVEL" "$LOG_LEVEL"
[ -n "$NEXT_PUBLIC_API_URL" ] && update_env "NEXT_PUBLIC_API_URL" "$NEXT_PUBLIC_API_URL"
[ -n "$NEXT_PUBLIC_WS_URL" ] && update_env "NEXT_PUBLIC_WS_URL" "$NEXT_PUBLIC_WS_URL"
