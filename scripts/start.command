#!/usr/bin/env bash
set -euo pipefail

if ! command -v go >/dev/null 2>&1; then
  echo "Go is not installed. Please install Go: https://go.dev/dl/"
  exit 1
fi

if command -v ollama >/dev/null 2>&1; then
  echo "Ollama detected."
else
  echo "Ollama not found. Install Ollama: https://ollama.com/"
fi

if command -v docker >/dev/null 2>&1; then
  if docker compose version >/dev/null 2>&1; then
    docker compose up -d
  elif command -v docker-compose >/dev/null 2>&1; then
    docker-compose up -d
  else
    echo "Docker Compose not found. OCR will be unavailable."
  fi
else
  echo "Docker not found. OCR will be unavailable."
fi

echo "Starting server on http://localhost:8080"

go run cmd/server/main.go
