# Cash Track

Personal finance tracker with chat + slip OCR. Run locally on your PC.

## Quick Start (low-tech)

1) Install:
- Go (https://go.dev/dl/) ✅ required
- Ollama (https://ollama.com/) ✅ required for LLM
- Docker Desktop (https://www.docker.com/products/docker-desktop/) ✅ optional (required only for OCR + containerized LLM)

2) Start everything:

```bash
./scripts/start.sh
```

### One-click launchers

- Windows: `scripts/start.bat`
- Windows (PowerShell): `scripts/start.ps1`
- macOS: `scripts/start.command`
- Linux: `scripts/start.sh`

### One-click installers (install required packages)

- Windows (PowerShell): `scripts/setup-windows.ps1` (uses winget)
- macOS: `scripts/setup-macos.sh` (uses Homebrew)
- Linux (Debian/Ubuntu): `scripts/setup-linux.sh` (uses apt)

## Release versioning

Use the release script to bump version and tag:

```bash
make release_major
make release_minor
make release_patch

```
then 
```bash
git push origin release/vX.X.X
git push origin release
```

3) Open in browser:

```
http://localhost:8080
```

## Example Usage

### Chat
- `กินข้าว 50 บาท เงินสด`
- `เมื่อวาน Shopee 320 บาท`
- `เดือนนี้ใช้ไปเท่าไหร่`

### Slip upload (OCR)
- Upload a receipt image in chat → the system extracts amount/date/channel

### Screenshots
Add images here:
- `docs/screenshots/chat.png`
- `docs/screenshots/dashboard.png`
- `docs/screenshots/history.png`

## Stop services

```bash
./scripts/stop.sh
```

## Notes

- Docker is **optional**. Without Docker, OCR will be unavailable.
- First run may take time to download the Ollama model.
- If Ollama is not running, the app falls back to regex parsing.
- OCR requires the Docker OCR service to be running.
- To access from another device on the same network, use your PC's LAN IP and port 8080.

## LLM setup (Ollama)

If you don't have Ollama yet:

1) Install Ollama: https://ollama.com/
2) Pull the model:

```bash
ollama pull llama3.2
```

If you use Docker:

```bash
make services-up
make pull-model
```

## Makefile shortcuts

```bash
make setup   # install deps + start services + pull model
make run     # start server
```
