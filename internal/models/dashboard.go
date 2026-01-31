package models

// DashboardSummary represents the summary data for the dashboard
type DashboardSummary struct {
	Period       Period           `json:"period"`
	TotalExpense float64          `json:"total_expense"`
	TotalIncome  float64          `json:"total_income"`
	ByCategory   []CategoryAmount `json:"by_category"`
	ByChannel    []ChannelAmount  `json:"by_channel"`
}

// Period represents a date range
type Period struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// CategoryAmount represents spending by category
type CategoryAmount struct {
	Category         string  `json:"category"`
	Amount           float64 `json:"amount"`
	PercentOfExpense float64 `json:"percent_of_expense"`
}

// ChannelAmount represents spending by channel
type ChannelAmount struct {
	Channel string  `json:"channel"`
	Amount  float64 `json:"amount"`
}
