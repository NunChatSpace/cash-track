package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	endpoint   string
	model      string
	httpClient *http.Client
}

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type GenerateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// ChatResponse represents the parsed LLM response for chat messages
type ChatResponse struct {
	Intent      string             `json:"intent"` // add_transaction | bill_payment | query_summary | unknown
	Transaction *ParsedTransaction `json:"transaction,omitempty"`
	Filters     *QueryFilters      `json:"filters,omitempty"`
	Confidence  float64            `json:"confidence,omitempty"`
}

// ParsedTransaction represents a transaction extracted by LLM
type ParsedTransaction struct {
	TxnDate      string  `json:"txn_date"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	Direction    string  `json:"direction"`
	Channel      string  `json:"channel"`
	AccountLabel string  `json:"account_label"`
	Category     string  `json:"category"`
	Description  string  `json:"description"`
	Confidence   float64 `json:"confidence"`
}

// QueryFilters represents filters for summary queries
type QueryFilters struct {
	Direction string       `json:"direction"` // income | expense | both
	Period    PeriodFilter `json:"period"`
	Category  string       `json:"category"`
	Channel   string       `json:"channel"`
}

// PeriodFilter represents a time period for queries
type PeriodFilter struct {
	Type string `json:"type"` // month | day | range | year | all
	From string `json:"from"`
	To   string `json:"to"`
}

func NewClient(endpoint, model string) *Client {
	return &Client{
		endpoint: endpoint,
		model:    model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ParseChatMessage parses a chat message (with optional OCR text) and returns structured data
func (c *Client) ParseChatMessage(message string, ocrText *string) (*ChatResponse, error) {
	var prompt string
	today := time.Now().Format("2006-01-02")

	if ocrText != nil && *ocrText != "" {
		// Use OCR prompt for slip parsing
		prompt = fmt.Sprintf(OCRPromptTemplate, *ocrText)
	} else {
		// Use text prompt for regular messages
		prompt = fmt.Sprintf(TextPromptTemplate, today, message)
	}

	response, err := c.generate(prompt)
	if err != nil {
		return nil, err
	}

	// Extract JSON from response
	jsonStr := extractJSON(response)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in LLM response")
	}

	var chatResp ChatResponse
	if err := json.Unmarshal([]byte(jsonStr), &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse LLM JSON: %w", err)
	}

	return &chatResp, nil
}

// ParseSlipText parses OCR text from a slip and returns transaction data
// This is a convenience method that wraps ParseChatMessage for backward compatibility
func (c *Client) ParseSlipText(ocrText string) (*ParsedTransaction, error) {
	resp, err := c.ParseChatMessage("", &ocrText)
	if err != nil {
		return nil, err
	}

	if resp.Transaction == nil {
		return nil, fmt.Errorf("no transaction data in LLM response")
	}

	// Copy confidence from response to transaction
	resp.Transaction.Confidence = resp.Confidence

	return resp.Transaction, nil
}

func (c *Client) generate(prompt string) (string, error) {
	reqBody := GenerateRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.endpoint+"/api/generate",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama error (status %d): %s", resp.StatusCode, string(body))
	}

	var genResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return genResp.Response, nil
}

func extractJSON(text string) string {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start == -1 || end == -1 || end <= start {
		return ""
	}
	return text[start : end+1]
}
