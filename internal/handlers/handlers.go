package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

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
	defaultUser int64
}

func New(repo *database.Repository, storage *storage.LocalStorage, ocrClient *ocr.Client, llmClient *llm.Client, templateDir string) (*Handler, error) {
	defaultUser, err := repo.EnsureDefaultUser()
	if err != nil {
		return nil, err
	}
	return &Handler{
		repo:        repo,
		storage:     storage,
		ocrClient:   ocrClient,
		llmClient:   llmClient,
		templateDir: templateDir,
		defaultUser: defaultUser,
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

func (h *Handler) withUserContext(w http.ResponseWriter, r *http.Request, data interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	if dataMap, ok := data.(map[string]interface{}); ok {
		for k, v := range dataMap {
			result[k] = v
		}
	}

	currentUserID, _ := h.currentUserID(w, r)
	users, _ := h.repo.ListUsers()
	currentUser, _ := h.repo.GetUser(currentUserID)
	result["CurrentUserID"] = currentUserID
	result["CurrentUser"] = currentUser
	result["Users"] = users

	return result
}

func (h *Handler) currentUserID(w http.ResponseWriter, r *http.Request) (int64, error) {
	if cookie, err := r.Cookie("ct_user_id"); err == nil {
		if id, err := strconv.ParseInt(cookie.Value, 10, 64); err == nil && id > 0 {
			if _, err := h.repo.GetUser(id); err == nil {
				return id, nil
			}
		}
	}

	h.setCurrentUserCookie(w, h.defaultUser)
	return h.defaultUser, nil
}

func (h *Handler) setCurrentUserCookie(w http.ResponseWriter, userID int64) {
	http.SetCookie(w, &http.Cookie{
		Name:     "ct_user_id",
		Value:    strconv.FormatInt(userID, 10),
		Path:     "/",
		HttpOnly: false,
	})
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/chat", http.StatusFound)
}

func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.currentUserID(w, r)
	transactions, err := h.repo.ListTransactions(userID, 50, 0)
	if err != nil {
		http.Error(w, "Failed to load transactions", http.StatusInternalServerError)
		return
	}

	var views []interface{}
	for _, tx := range transactions {
		views = append(views, tx.ToView())
	}

	h.renderTemplate(w, "history.html", h.withUserContext(w, r, map[string]interface{}{
		"Transactions": views,
	}))
}

func (h *Handler) UsersPage(w http.ResponseWriter, r *http.Request) {
	h.renderTemplate(w, "users.html", h.withUserContext(w, r, nil))
}
