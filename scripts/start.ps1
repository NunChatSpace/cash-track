$ErrorActionPreference = "Stop"

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
  Write-Host "Go is not installed. Please install Go: https://go.dev/dl/" -ForegroundColor Yellow
  exit 1
}

if (Get-Command ollama -ErrorAction SilentlyContinue) {
  Write-Host "Ollama detected."
} else {
  Write-Host "Ollama not found. Install Ollama: https://ollama.com/" -ForegroundColor Yellow
}

if (Get-Command docker -ErrorAction SilentlyContinue) {
  if (docker compose version 2>$null) {
    docker compose up -d
  } elseif (Get-Command docker-compose -ErrorAction SilentlyContinue) {
    docker-compose up -d
  } else {
    Write-Host "Docker Compose not found. OCR will be unavailable." -ForegroundColor Yellow
  }
} else {
  Write-Host "Docker not found. OCR will be unavailable." -ForegroundColor Yellow
}

Write-Host "Starting server on http://localhost:8080"

& go run cmd/server/main.go
