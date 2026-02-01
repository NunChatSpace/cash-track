package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"cash-track/internal/models"
)

// DashboardSummary handles GET /api/dashboard/summary
func (h *Handler) DashboardSummary(w http.ResponseWriter, r *http.Request) {
	from, to := getDateRange(r)
	userID, _ := h.currentUserID(w, r)

	summary, err := h.repo.GetDashboardSummary(userID, from, to)
	if err != nil {
		http.Error(w, "Failed to get summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// DashboardByCategory handles GET /api/dashboard/by-category
func (h *Handler) DashboardByCategory(w http.ResponseWriter, r *http.Request) {
	from, to := getDateRange(r)
	userID, _ := h.currentUserID(w, r)

	categories, err := h.repo.GetExpenseByCategory(userID, from, to)
	if err != nil {
		http.Error(w, "Failed to get category data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// DashboardByChannel handles GET /api/dashboard/by-channel
func (h *Handler) DashboardByChannel(w http.ResponseWriter, r *http.Request) {
	from, to := getDateRange(r)
	userID, _ := h.currentUserID(w, r)

	channels, err := h.repo.GetExpenseByChannel(userID, from, to)
	if err != nil {
		http.Error(w, "Failed to get channel data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(channels)
}

// DashboardTransactions handles GET /api/dashboard/transactions
func (h *Handler) DashboardTransactions(w http.ResponseWriter, r *http.Request) {
	from, to := getDateRange(r)
	userID, _ := h.currentUserID(w, r)

	limit := 200
	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			if parsed > 500 {
				parsed = 500
			}
			limit = parsed
		}
	}
	category := r.URL.Query().Get("category")
	channel := r.URL.Query().Get("channel")

	var (
		transactions []models.Transaction
		err          error
	)
	if category != "" || channel != "" {
		transactions, err = h.repo.ListTransactionsByRangeFiltered(userID, from, to, category, channel, limit)
	} else {
		transactions, err = h.repo.ListTransactionsByRange(userID, from, to, limit)
	}
	if err != nil {
		http.Error(w, "Failed to get transactions", http.StatusInternalServerError)
		return
	}

	var views []interface{}
	for _, tx := range transactions {
		views = append(views, tx.ToView())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(views)
}

// DashboardPage renders the dashboard UI
func (h *Handler) DashboardPage(w http.ResponseWriter, r *http.Request) {
	h.renderTemplate(w, "dashboard.html", h.withUserContext(w, r, nil))
}

// getDateRange extracts from/to dates from query params, defaults to current month
func getDateRange(r *http.Request) (string, string) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if from == "" || to == "" {
		// Default to current month
		now := time.Now()
		firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		lastDay := firstDay.AddDate(0, 1, -1)
		from = firstDay.Format("2006-01-02")
		to = lastDay.Format("2006-01-02")
	}

	return from, to
}
