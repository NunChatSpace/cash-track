#!/usr/bin/env bash
set -euo pipefail

if ! command -v brew >/dev/null 2>&1; then
  echo "Homebrew not found. Install from https://brew.sh/"
  exit 1
fi

echo "Installing Go..."
brew install go

echo "Installing Ollama..."
brew install ollama

echo "Installing Docker Desktop (optional for OCR)..."
brew install --cask docker

echo "Setup complete. You may need to restart your terminal."
echo "Run scripts/start.command to launch."
