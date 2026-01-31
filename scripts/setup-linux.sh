#!/usr/bin/env bash
set -euo pipefail

if ! command -v apt >/dev/null 2>&1; then
  echo "apt not found. This installer supports Debian/Ubuntu only."
  exit 1
fi

echo "Updating package list..."
sudo apt update

echo "Installing Go..."
sudo apt install -y golang-go

echo "Installing Docker (optional for OCR)..."
sudo apt install -y docker.io docker-compose-plugin

if ! command -v ollama >/dev/null 2>&1; then
  echo "Installing Ollama..."
  curl -fsSL https://ollama.com/install.sh | sh
fi

echo "Setup complete. You may need to log out/in for Docker permissions."
echo "Run scripts/start.sh to launch."
