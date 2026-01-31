package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type createUserRequest struct {
	Name string `json:"name"`
}

type selectUserRequest struct {
	UserID int64 `json:"user_id"`
}

type updateCutoffRequest struct {
	CutoffDay int `json:"cutoff_day"`
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.repo.ListUsers()
	if err != nil {
		http.Error(w, "Failed to load users", http.StatusInternalServerError)
		return
	}

	currentUserID, _ := h.currentUserID(w, r)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"current_user_id": currentUserID,
		"users":           users,
	})
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	user, err := h.repo.CreateUser(name)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusBadRequest)
		return
	}

	h.setCurrentUserCookie(w, user.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) SelectUser(w http.ResponseWriter, r *http.Request) {
	var req selectUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == 0 {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	if _, err := h.repo.GetUser(req.UserID); err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	h.setCurrentUserCookie(w, req.UserID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func (h *Handler) UpdateCutoff(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req updateCutoffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CutoffDay < 1 || req.CutoffDay > 30 {
		http.Error(w, "Cutoff day must be between 1 and 30", http.StatusBadRequest)
		return
	}

	user, err := h.repo.UpdateUserCutoff(id, req.CutoffDay)
	if err != nil {
		http.Error(w, "Failed to update cutoff", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
