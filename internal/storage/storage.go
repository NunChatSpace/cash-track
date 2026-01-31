package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type LocalStorage struct {
	baseDir string
}

func NewLocalStorage(baseDir string) (*LocalStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}
	return &LocalStorage{baseDir: baseDir}, nil
}

func (s *LocalStorage) Save(filename string, reader io.Reader) (string, error) {
	ext := filepath.Ext(filename)
	newFilename := uuid.New().String() + ext
	fullPath := filepath.Join(s.baseDir, newFilename)

	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		os.Remove(fullPath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return newFilename, nil
}

func (s *LocalStorage) GetPath(filename string) string {
	return filepath.Join(s.baseDir, filename)
}

func (s *LocalStorage) Delete(filename string) error {
	return os.Remove(filepath.Join(s.baseDir, filename))
}
