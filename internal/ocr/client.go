package ocr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Client struct {
	endpoint   string
	httpClient *http.Client
}

type OCRResponse struct {
	Text  string `json:"text"`
	Error string `json:"error,omitempty"`
}

func NewClient(endpoint string) *Client {
	return &Client{
		endpoint:   endpoint,
		httpClient: &http.Client{},
	}
}

func (c *Client) ExtractText(imagePath string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Detect content type from file extension
	ext := strings.ToLower(filepath.Ext(imagePath))
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Create form file with proper Content-Type header
	h := make(map[string][]string)
	h["Content-Disposition"] = []string{fmt.Sprintf(`form-data; name="file"; filename="%s"`, filepath.Base(imagePath))}
	h["Content-Type"] = []string{contentType}

	part, err := writer.CreatePart(h)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest("POST", c.endpoint+"/ocr", &buf)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("OCR request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OCR service error (status %d): %s", resp.StatusCode, string(body))
	}

	var ocrResp OCRResponse
	if err := json.NewDecoder(resp.Body).Decode(&ocrResp); err != nil {
		return "", fmt.Errorf("failed to decode OCR response: %w", err)
	}

	if ocrResp.Error != "" {
		return "", fmt.Errorf("OCR error: %s", ocrResp.Error)
	}

	return ocrResp.Text, nil
}
