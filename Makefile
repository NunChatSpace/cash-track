.PHONY: run build dev services-up services-down clean setup pull-model

# Run the Go server
run:
	go run cmd/server/main.go

# Build the Go binary
build:
	go build -o bin/cash-track cmd/server/main.go

# Run with live reload (requires air: go install github.com/air-verse/air@latest)
dev:
	air

# Start all services (OCR + Ollama)
services-up:
	docker-compose up -d

# Stop all services
services-down:
	docker-compose down

# Pull Ollama model (run after services-up)
pull-model:
	docker-compose exec ollama ollama pull llama3.2

# Run OCR service locally (requires Python venv)
ocr-local:
	cd ocr-service && python main.py

# Install Go dependencies
deps:
	go mod tidy

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f cash-track.db

# Full setup
setup: deps services-up pull-model
	@echo "Setup complete! Run 'make run' to start the server"
