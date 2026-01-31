package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"

	"cash-track/internal/database"
	"cash-track/internal/llm"
	"cash-track/internal/ocr"
	"cash-track/internal/storage"
)

type Handler struct {
	repo        *database.Repository
	storage     *storage.LocalStorage
	ocrClient   *ocr.Client
	llmClient   *llm.Client
	templateDir string
}

func New(repo *database.Repository, storage *storage.LocalStorage, ocrClient *ocr.Client, llmClient *llm.Client, templateDir string) (*Handler, error) {
	return &Handler{
		repo:        repo,
		storage:     storage,
		ocrClient:   ocrClient,
		llmClient:   llmClient,
		templateDir: templateDir,
	}, nil
}

func (h *Handler) renderTemplate(w http.ResponseWriter, page string, data interface{}) {
	tmpl, err := template.ParseFiles(
		filepath.Join(h.templateDir, "layout.html"),
		filepath.Join(h.templateDir, page),
	)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "layout", data)
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/chat", http.StatusFound)
}

func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	transactions, err := h.repo.ListTransactions(50, 0)
	if err != nil {
		http.Error(w, "Failed to load transactions", http.StatusInternalServerError)
		return
	}

	var views []interface{}
	for _, tx := range transactions {
		views = append(views, tx.ToView())
	}

	h.renderTemplate(w, "history.html", map[string]interface{}{
		"Transactions": views,
	})
}
