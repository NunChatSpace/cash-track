@echo off
setlocal

where go >nul 2>nul
if %errorlevel% neq 0 (
  echo Go is not installed. Please install Go: https://go.dev/dl/
  exit /b 1
)

where ollama >nul 2>nul
if %errorlevel% neq 0 (
  echo Ollama not found. Install Ollama: https://ollama.com/
) else (
  echo Ollama detected.
)

where docker >nul 2>nul
if %errorlevel% neq 0 (
  echo Docker not found. OCR will be unavailable.
) else (
  docker compose version >nul 2>nul
  if %errorlevel% neq 0 (
    where docker-compose >nul 2>nul
    if %errorlevel% neq 0 (
      echo Docker Compose not found. OCR will be unavailable.
    ) else (
      docker-compose up -d
    )
  ) else (
    docker compose up -d
  )
)

echo Starting server on http://localhost:8080

go run cmd/server/main.go
endlocal
