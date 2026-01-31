package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// DashboardSummary handles GET /api/dashboard/summary
func (h *Handler) DashboardSummary(w http.ResponseWriter, r *http.Request) {
	from, to := getDateRange(r)

	summary, err := h.repo.GetDashboardSummary(from, to)
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

	categories, err := h.repo.GetExpenseByCategory(from, to)
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

	channels, err := h.repo.GetExpenseByChannel(from, to)
	if err != nil {
		http.Error(w, "Failed to get channel data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(channels)
}

// DashboardPage renders the dashboard UI
func (h *Handler) DashboardPage(w http.ResponseWriter, r *http.Request) {
	h.renderTemplate(w, "dashboard.html", nil)
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
