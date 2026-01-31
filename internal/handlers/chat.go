package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"cash-track/internal/llm"
)

// ChatRequest represents the incoming chat message
type ChatRequest struct {
	Message   string  `json:"message"`
	ImagePath *string `json:"image_path"`
}

// ChatResponse represents the chat response
type ChatResponse struct {
	ReplyText     string      `json:"reply_text"`
	TransactionID *int64      `json:"transaction_id,omitempty"`
	Debug         interface{} `json:"debug,omitempty"`
}

// Chat handles POST /api/chat
func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// If image path provided, run OCR first
	var ocrText *string
	if req.ImagePath != nil && *req.ImagePath != "" {
		text, err := h.ocrClient.ExtractText(*req.ImagePath)
		if err != nil {
			log.Printf("OCR failed: %v", err)
		} else {
			ocrText = &text
		}
	}

	// Parse with LLM
	llmResp, err := h.llmClient.ParseChatMessage(req.Message, ocrText)
	if err != nil {
		log.Printf("LLM parsing failed: %v", err)
		respondChat(w, "ขออภัย ไม่สามารถประมวลผลได้ กรุณาลองใหม่", nil, nil)
		return
	}

	// Handle based on intent
	switch llmResp.Intent {
	case "add_transaction", "bill_payment":
		h.handleAddTransaction(w, llmResp, req.ImagePath, ocrText)
	case "query_summary":
		h.handleQuerySummary(w, llmResp)
	default:
		respondChat(w, "ไม่เข้าใจคำสั่ง กรุณาลองพิมพ์ใหม่ เช่น 'วันนี้กินข้าวไป 50 บาท' หรือ 'เดือนนี้ใช้ไปเท่าไหร่'", nil, llmResp)
	}
}

func (h *Handler) handleAddTransaction(w http.ResponseWriter, resp *llm.ChatResponse, imagePath *string, ocrText *string) {
	if resp.Transaction == nil {
		respondChat(w, "ไม่พบข้อมูลรายการ กรุณาระบุจำนวนเงิน", nil, resp)
		return
	}

	tx := resp.Transaction

	// Set defaults
	if tx.Currency == "" {
		tx.Currency = "THB"
	}
	if tx.Direction == "" {
		tx.Direction = "expense"
	}

	// Determine status based on completeness - pending if missing important fields
	status := "confirmed"
	if tx.Amount == 0 || tx.Category == "" || tx.Channel == "" {
		status = "pending"
	}

	// Get image path and OCR text as strings
	var slipPath, rawOCR string
	if imagePath != nil {
		slipPath = *imagePath
	}
	if ocrText != nil {
		rawOCR = *ocrText
	}

	// Create transaction
	created, err := h.repo.CreateTransactionFromChat(
		tx.TxnDate,
		tx.Amount,
		tx.Currency,
		tx.Direction,
		tx.Channel,
		tx.AccountLabel,
		tx.Category,
		tx.Description,
		slipPath,
		rawOCR,
		resp.Confidence,
		status,
	)
	if err != nil {
		log.Printf("Failed to create transaction: %v", err)
		respondChat(w, "ไม่สามารถบันทึกรายการได้", nil, resp)
		return
	}

	// Build reply
	reply := buildTransactionReply(tx, status)
	respondChat(w, reply, &created.ID, resp)
}

func (h *Handler) handleQuerySummary(w http.ResponseWriter, resp *llm.ChatResponse) {
	if resp.Filters == nil {
		respondChat(w, "ไม่เข้าใจคำถาม กรุณาลองใหม่", nil, resp)
		return
	}

	filters := resp.Filters

	// Calculate date range based on period
	from, to := calculatePeriod(filters.Period)

	// Query database
	summary, err := h.repo.QuerySummary(filters.Direction, from, to, filters.Category, filters.Channel)
	if err != nil {
		log.Printf("Query failed: %v", err)
		respondChat(w, "ไม่สามารถดึงข้อมูลได้", nil, resp)
		return
	}

	// Build reply
	reply := buildSummaryReplyText(summary.TotalExpense, summary.TotalIncome, filters)
	respondChat(w, reply, nil, resp)
}

func respondChat(w http.ResponseWriter, text string, txID *int64, debug interface{}) {
	resp := ChatResponse{
		ReplyText:     text,
		TransactionID: txID,
		Debug:         debug,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func buildTransactionReply(tx *llm.ParsedTransaction, status string) string {
	var reply string

	if status == "confirmed" {
		reply = fmt.Sprintf("บันทึกแล้ว: %.2f บาท", tx.Amount)
		if tx.Category != "" {
			reply += fmt.Sprintf(" หมวด%s", categoryThai(tx.Category))
		}
		if tx.Channel != "" {
			reply += fmt.Sprintf(" (%s)", tx.Channel)
		}
		if tx.Description != "" {
			reply += fmt.Sprintf(" - %s", tx.Description)
		}
	} else {
		reply = fmt.Sprintf("บันทึก %.2f บาท - รอยืนยัน", tx.Amount)
		var missing []string
		if tx.Category == "" {
			missing = append(missing, "หมวดหมู่")
		}
		if tx.Channel == "" {
			missing = append(missing, "ช่องทาง")
		}
		if len(missing) > 0 {
			reply += fmt.Sprintf(" (กรุณาระบุ: %s)", strings.Join(missing, ", "))
		}
	}

	return reply
}

func buildSummaryReplyText(totalExpense, totalIncome float64, filters *llm.QueryFilters) string {
	var reply string

	if filters.Direction == "expense" || filters.Direction == "both" || filters.Direction == "" {
		reply = fmt.Sprintf("ใช้ไป %.2f บาท", totalExpense)
	}
	if filters.Direction == "income" {
		reply = fmt.Sprintf("รายรับ %.2f บาท", totalIncome)
	}

	if filters.Category != "" {
		reply += fmt.Sprintf(" (หมวด%s)", categoryThai(filters.Category))
	}
	if filters.Channel != "" {
		reply += fmt.Sprintf(" (%s)", filters.Channel)
	}

	return reply
}

func calculatePeriod(period llm.PeriodFilter) (string, string) {
	now := time.Now()

	switch period.Type {
	case "month":
		if period.From != "" && period.To != "" {
			return period.From, period.To
		}
		// Current month
		firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		lastDay := firstDay.AddDate(0, 1, -1)
		return firstDay.Format("2006-01-02"), lastDay.Format("2006-01-02")
	case "year":
		firstDay := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		lastDay := time.Date(now.Year(), 12, 31, 0, 0, 0, 0, now.Location())
		return firstDay.Format("2006-01-02"), lastDay.Format("2006-01-02")
	case "day":
		today := now.Format("2006-01-02")
		return today, today
	case "range":
		return period.From, period.To
	default:
		// Default to current month
		firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		lastDay := firstDay.AddDate(0, 1, -1)
		return firstDay.Format("2006-01-02"), lastDay.Format("2006-01-02")
	}
}

func categoryThai(category string) string {
	mapping := map[string]string{
		"food":      "อาหาร",
		"rent":      "ค่าเช่า",
		"shopping":  "ช้อปปิ้ง",
		"transport": "เดินทาง",
		"bill":      "ค่าบริการ",
		"debt":      "หนี้สิน",
		"other":     "อื่นๆ",
	}
	if thai, ok := mapping[category]; ok {
		return thai
	}
	return category
}

// ChatPage renders the chat UI
func (h *Handler) ChatPage(w http.ResponseWriter, r *http.Request) {
	h.renderTemplate(w, "chat.html", nil)
}
