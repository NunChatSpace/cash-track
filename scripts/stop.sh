#!/usr/bin/env bash
set -euo pipefail

if command -v docker >/dev/null 2>&1; then
  if docker compose version >/dev/null 2>&1; then
    docker compose down
  elif command -v docker-compose >/dev/null 2>&1; then
    docker-compose down
  fi
fi

echo "Services stopped."
