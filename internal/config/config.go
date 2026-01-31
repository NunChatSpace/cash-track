package config

import (
	"os"
)

type Config struct {
	ServerPort    string
	DatabaseURL   string
	UploadDir     string
	OCREndpoint   string
	OllamaURL     string
	OllamaModel   string
}

func Load() *Config {
	return &Config{
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "./cash-track.db"),
		UploadDir:     getEnv("UPLOAD_DIR", "./uploads"),
		OCREndpoint:   getEnv("OCR_ENDPOINT", "http://localhost:8001"),
		OllamaURL:     getEnv("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel:   getEnv("OLLAMA_MODEL", "llama3.2"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
