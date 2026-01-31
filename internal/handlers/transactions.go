package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"cash-track/internal/models"
)

func (h *Handler) UploadSlip(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("slip")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename, err := h.storage.Save(header.Filename, file)
	if err != nil {
		log.Printf("Failed to save file: %v", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	tx, err := h.repo.CreateTransaction(filename)
	if err != nil {
		log.Printf("Failed to create transaction: %v", err)
		http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
		return
	}

	go h.processOCR(tx.ID, filename)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         tx.ID,
		"image_path": filename,
		"redirect":   "/transactions/" + strconv.FormatInt(tx.ID, 10) + "/confirm",
	})
}

func (h *Handler) processOCR(txID int64, filename string) {
	imagePath := h.storage.GetPath(filename)

	// Step 1: Extract text with EasyOCR
	rawText, err := h.ocrClient.ExtractText(imagePath)
	if err != nil {
		log.Printf("OCR failed for transaction %d: %v", txID, err)
		return
	}

	log.Printf("OCR text for transaction %d: %s", txID, rawText)

	// Step 2: Parse with LLM (Ollama)
	parsed, err := h.llmClient.ParseSlipText(rawText)
	if err != nil {
		log.Printf("LLM parsing failed for transaction %d: %v", txID, err)
		// Still save the raw OCR text
		h.repo.UpdateOCRResult(txID, rawText, 0, "", "", "", "", 0)
		return
	}

	err = h.repo.UpdateOCRResult(
		txID,
		rawText,
		parsed.Amount,
		parsed.TxnDate,
		parsed.Channel,
		parsed.Category,
		parsed.Description,
		parsed.Confidence,
	)
	if err != nil {
		log.Printf("Failed to update OCR result for transaction %d: %v", txID, err)
	}
}

func (h *Handler) ConfirmPage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	tx, err := h.repo.GetTransaction(id)
	if err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	h.renderTemplate(w, "confirm.html", map[string]interface{}{
		"Transaction": tx.ToView(),
	})
}

func (h *Handler) ConfirmTransaction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	var req models.ConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.repo.ConfirmTransaction(id, req); err != nil {
		log.Printf("Failed to confirm transaction %d: %v", id, err)
		http.Error(w, "Failed to confirm transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"redirect": "/history",
	})
}

func (h *Handler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	tx, err := h.repo.GetTransaction(id)
	if err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tx.ToView())
}

func (h *Handler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	tx, err := h.repo.GetTransaction(id)
	if err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	if err := h.repo.DeleteTransaction(id); err != nil {
		log.Printf("Failed to delete transaction %d: %v", id, err)
		http.Error(w, "Failed to delete transaction", http.StatusInternalServerError)
		return
	}

	if tx.SlipImagePath.Valid && tx.SlipImagePath.String != "" {
		if err := h.storage.Delete(tx.SlipImagePath.String); err != nil {
			log.Printf("Failed to delete slip for transaction %d: %v", id, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"redirect": "/history",
	})
}

func (h *Handler) GetRecentTransactions(w http.ResponseWriter, r *http.Request) {
	transactions, err := h.repo.ListTransactions(20, 0)
	if err != nil {
		http.Error(w, "Failed to load transactions", http.StatusInternalServerError)
		return
	}

	var views []interface{}
	for _, tx := range transactions {
		views = append(views, tx.ToView())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(views)
}
