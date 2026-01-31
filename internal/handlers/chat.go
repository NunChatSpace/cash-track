package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"cash-track/internal/llm"
)

// ChatRequest represents the incoming chat message
type ChatRequest struct {
	Message   string  `json:"message"`
	ImagePath *string `json:"image_path"`
	Lang      string  `json:"lang"`
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

	userID, _ := h.currentUserID(w, r)
	log.Printf("Chat request user_id=%d message=%q image=%v", userID, req.Message, req.ImagePath != nil && *req.ImagePath != "")

	// If image path provided, run OCR first
	var ocrText *string
	if req.ImagePath != nil && *req.ImagePath != "" {
		imagePath := *req.ImagePath
		if !filepath.IsAbs(imagePath) && !strings.Contains(imagePath, ":\\") {
			imagePath = h.storage.GetPath(imagePath)
		}
		text, err := h.ocrClient.ExtractText(imagePath)
		if err != nil {
			log.Printf("OCR failed: %v", err)
		} else {
			ocrText = &text
		}
	}

	lang := normalizeLang(req.Lang)

	// Parse with LLM
	llmResp, err := h.llmClient.ParseChatMessage(req.Message, ocrText, lang)
	if err != nil {
		log.Printf("LLM parsing failed: %v", err)
		respondChat(w, chatText(lang, "error_processing"), nil, nil)
		return
	}
	log.Printf("Chat parsed intent=%s confidence=%.2f", llmResp.Intent, llmResp.Confidence)

	// Handle based on intent
	switch llmResp.Intent {
	case "add_transaction", "bill_payment":
		h.handleAddTransaction(w, r, req.Message, llmResp, req.ImagePath, ocrText, lang)
	case "query_summary":
		h.handleQuerySummary(w, r, llmResp, lang)
	default:
		respondChat(w, chatText(lang, "error_unknown"), nil, llmResp)
	}
}

func (h *Handler) handleAddTransaction(w http.ResponseWriter, r *http.Request, message string, resp *llm.ChatResponse, imagePath *string, ocrText *string, lang string) {
	if resp.Transaction == nil {
		respondChat(w, chatText(lang, "missing_amount"), nil, resp)
		return
	}

	tx := resp.Transaction

	if tx.Amount == 0 {
		if message != "" {
			tx.Amount = llm.ExtractAmountFromText(message)
		} else if ocrText != nil {
			tx.Amount = llm.ExtractAmountFromText(*ocrText)
		}
	}

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
	userID, _ := h.currentUserID(w, r)
	created, err := h.repo.CreateTransactionFromChat(
		userID,
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
		respondChat(w, chatText(lang, "save_failed"), nil, resp)
		return
	}
	log.Printf("Transaction created id=%d status=%s amount=%.2f", created.ID, status, tx.Amount)

	// Build reply
	reply := buildTransactionReply(tx, status, lang)
	respondChat(w, reply, &created.ID, resp)
}

func (h *Handler) handleQuerySummary(w http.ResponseWriter, r *http.Request, resp *llm.ChatResponse, lang string) {
	if resp.Filters == nil {
		respondChat(w, chatText(lang, "error_unknown"), nil, resp)
		return
	}

	filters := resp.Filters

	// Calculate date range based on period
	userID, _ := h.currentUserID(w, r)
	cutoff := 1
	if user, err := h.repo.GetUser(userID); err == nil && user.CutoffDay >= 1 {
		cutoff = user.CutoffDay
	}
	from, to := calculatePeriod(filters.Period, cutoff)

	// Query database
	summary, err := h.repo.QuerySummary(userID, filters.Direction, from, to, filters.Category, filters.Channel)
	if err != nil {
		log.Printf("Query failed: %v", err)
		respondChat(w, chatText(lang, "fetch_failed"), nil, resp)
		return
	}

	// Build reply
	reply := buildSummaryReplyText(summary.TotalExpense, summary.TotalIncome, filters, from, to, lang)
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

func buildTransactionReply(tx *llm.ParsedTransaction, status string, lang string) string {
	var reply string

	if status == "confirmed" {
		if lang == "en" {
			reply = fmt.Sprintf("Saved: %.2f THB", tx.Amount)
		} else {
			reply = fmt.Sprintf("บันทึกแล้ว: %.2f บาท", tx.Amount)
		}
		if tx.Category != "" {
			if lang == "en" {
				reply += fmt.Sprintf(" (%s)", categoryLabel(tx.Category, lang))
			} else {
				reply += fmt.Sprintf(" หมวด%s", categoryLabel(tx.Category, lang))
			}
		}
		if tx.Channel != "" {
			reply += fmt.Sprintf(" (%s)", tx.Channel)
		}
		if tx.Description != "" {
			reply += fmt.Sprintf(" - %s", tx.Description)
		}
	} else {
		if lang == "en" {
			reply = fmt.Sprintf("Saved %.2f THB - pending", tx.Amount)
		} else {
			reply = fmt.Sprintf("บันทึก %.2f บาท - รอยืนยัน", tx.Amount)
		}
		var missing []string
		if tx.Category == "" {
			if lang == "en" {
				missing = append(missing, "category")
			} else {
				missing = append(missing, "หมวดหมู่")
			}
		}
		if tx.Channel == "" {
			if lang == "en" {
				missing = append(missing, "channel")
			} else {
				missing = append(missing, "ช่องทาง")
			}
		}
		if len(missing) > 0 {
			if lang == "en" {
				reply += fmt.Sprintf(" (please provide: %s)", strings.Join(missing, ", "))
			} else {
				reply += fmt.Sprintf(" (กรุณาระบุ: %s)", strings.Join(missing, ", "))
			}
		}
	}

	return reply
}

func buildSummaryReplyText(totalExpense, totalIncome float64, filters *llm.QueryFilters, from, to string, lang string) string {
	var reply string

	if filters.Direction == "expense" || filters.Direction == "" {
		if lang == "en" {
			reply = fmt.Sprintf("Spent %.2f THB", totalExpense)
		} else {
			reply = fmt.Sprintf("ใช้ไป %.2f บาท", totalExpense)
		}
	} else if filters.Direction == "income" {
		if lang == "en" {
			reply = fmt.Sprintf("Income %.2f THB", totalIncome)
		} else {
			reply = fmt.Sprintf("รายรับ %.2f บาท", totalIncome)
		}
	} else {
		if lang == "en" {
			reply = fmt.Sprintf("Expense %.2f THB, Income %.2f THB", totalExpense, totalIncome)
		} else {
			reply = fmt.Sprintf("รายจ่าย %.2f บาท, รายรับ %.2f บาท", totalExpense, totalIncome)
		}
	}

	if filters.Category != "" {
		if lang == "en" {
			reply += fmt.Sprintf(" (%s)", categoryLabel(filters.Category, lang))
		} else {
			reply += fmt.Sprintf(" (หมวด%s)", categoryLabel(filters.Category, lang))
		}
	}
	if filters.Channel != "" {
		reply += fmt.Sprintf(" (%s)", filters.Channel)
	}
	if from != "" && to != "" {
		if lang == "en" {
			reply += fmt.Sprintf(" from %s to %s", from, to)
		} else {
			reply += fmt.Sprintf(" ช่วง %s ถึง %s", from, to)
		}
	}

	return reply
}

func calculatePeriod(period llm.PeriodFilter, cutoffDay int) (string, string) {
	now := time.Now()

	switch period.Type {
	case "month":
		if period.From != "" && period.To != "" {
			return period.From, period.To
		}
		return cutoffRange(now, cutoffDay, 0)
	case "year":
		firstDay := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		lastDay := time.Date(now.Year(), 12, 31, 0, 0, 0, 0, now.Location())
		return firstDay.Format("2006-01-02"), lastDay.Format("2006-01-02")
	case "day":
		today := now.Format("2006-01-02")
		return today, today
	case "range":
		return period.From, period.To
	case "all":
		return "1970-01-01", now.Format("2006-01-02")
	default:
		return cutoffRange(now, cutoffDay, 0)
	}
}

func cutoffRange(now time.Time, cutoffDay int, offsetMonths int) (string, string) {
	if cutoffDay < 1 || cutoffDay > 30 {
		cutoffDay = 1
	}
	year, month := now.Year(), now.Month()
	if offsetMonths != 0 {
		date := time.Date(year, month, 1, 0, 0, 0, 0, now.Location()).AddDate(0, offsetMonths, 0)
		year, month = date.Year(), date.Month()
	}

	var from, to time.Time
	if now.Day() >= cutoffDay {
		from = time.Date(year, month, clampDay(year, month, cutoffDay), 0, 0, 0, 0, now.Location())
		to = time.Date(year, month+1, clampDay(year, month+1, cutoffDay-1), 0, 0, 0, 0, now.Location())
	} else {
		from = time.Date(year, month-1, clampDay(year, month-1, cutoffDay), 0, 0, 0, 0, now.Location())
		to = time.Date(year, month, clampDay(year, month, cutoffDay-1), 0, 0, 0, 0, now.Location())
	}

	return from.Format("2006-01-02"), to.Format("2006-01-02")
}

func clampDay(year int, month time.Month, day int) int {
	if day < 1 {
		return 1
	}
	last := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
	if day > last {
		return last
	}
	return day
}

func categoryLabel(category string, lang string) string {
	if lang == "en" {
		mapping := map[string]string{
			"food":      "food",
			"rent":      "rent",
			"shopping":  "shopping",
			"transport": "transport",
			"bill":      "bills",
			"debt":      "debt",
			"other":     "other",
		}
		if en, ok := mapping[category]; ok {
			return en
		}
		return category
	}
	mapping := map[string]string{
		"food":      "อาหาร",
		"rent":      "ค่าเช่า",
		"shopping":  "ช้อปปิ้ง",
		"transport": "เดินทาง",
		"bill":      "ค่าบริการ",
		"debt":      "หนี้สิน",
		"other":     "อื่นๆ",
	}
	if th, ok := mapping[category]; ok {
		return th
	}
	return category
}

func normalizeLang(lang string) string {
	if strings.ToLower(strings.TrimSpace(lang)) == "en" {
		return "en"
	}
	return "th"
}

func chatText(lang string, key string) string {
	if lang == "en" {
		switch key {
		case "error_processing":
			return "Sorry, I couldn't process that. Please try again."
		case "error_unknown":
			return "I didn't understand. Try something like 'lunch 50' or 'how much did I spend this month?'"
		case "missing_amount":
			return "Missing amount. Please specify the amount."
		case "save_failed":
			return "Unable to save the transaction."
		case "fetch_failed":
			return "Unable to fetch data."
		}
	}

	switch key {
	case "error_processing":
		return "ขออภัย ไม่สามารถประมวลผลได้ กรุณาลองใหม่"
	case "error_unknown":
		return "ไม่เข้าใจคำสั่ง กรุณาลองพิมพ์ใหม่ เช่น 'วันนี้กินข้าวไป 50 บาท' หรือ 'เดือนนี้ใช้ไปเท่าไหร่'"
	case "missing_amount":
		return "ไม่พบข้อมูลรายการ กรุณาระบุจำนวนเงิน"
	case "save_failed":
		return "ไม่สามารถบันทึกรายการได้"
	case "fetch_failed":
		return "ไม่สามารถดึงข้อมูลได้"
	}
	return ""
}

// ChatPage renders the chat UI
func (h *Handler) ChatPage(w http.ResponseWriter, r *http.Request) {
	h.renderTemplate(w, "chat.html", h.withUserContext(w, r, nil))
}
