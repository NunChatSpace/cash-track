$ErrorActionPreference = "Stop"

if (-not (Get-Command winget -ErrorAction SilentlyContinue)) {
  Write-Host "winget not found. Install App Installer from Microsoft Store." -ForegroundColor Yellow
  exit 1
}

Write-Host "Installing Go..."
winget install --id GoLang.Go --silent --accept-source-agreements --accept-package-agreements

Write-Host "Installing Ollama..."
winget install --id Ollama.Ollama --silent --accept-source-agreements --accept-package-agreements

Write-Host "Installing Docker Desktop (optional for OCR)..."
winget install --id Docker.DockerDesktop --silent --accept-source-agreements --accept-package-agreements

Write-Host "Setup complete. You may need to restart your terminal."
Write-Host "Run scripts/start.bat or scripts/start.ps1 to launch."
