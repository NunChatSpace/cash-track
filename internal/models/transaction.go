package models

import (
	"database/sql"
)

type Transaction struct {
	ID            int64           `json:"id"`
	TxnDate       sql.NullString  `json:"txn_date"`
	Amount        sql.NullFloat64 `json:"amount"`
	Currency      string          `json:"currency"`
	Direction     string          `json:"direction"`
	Channel       sql.NullString  `json:"channel"`
	AccountLabel  sql.NullString  `json:"account_label"`
	Category      sql.NullString  `json:"category"`
	Description   sql.NullString  `json:"description"`
	SlipImagePath sql.NullString  `json:"slip_image_path"`
	RawOCRText    sql.NullString  `json:"raw_ocr_text"`
	LLMConfidence sql.NullFloat64 `json:"llm_confidence"`
	Status        string          `json:"status"`
	CreatedAt     string          `json:"created_at"`
	UpdatedAt     string          `json:"updated_at"`
}

type TransactionView struct {
	ID            int64   `json:"id"`
	TxnDate       string  `json:"txn_date"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Direction     string  `json:"direction"`
	Channel       string  `json:"channel"`
	AccountLabel  string  `json:"account_label"`
	Category      string  `json:"category"`
	Description   string  `json:"description"`
	SlipImagePath string  `json:"slip_image_path"`
	RawOCRText    string  `json:"raw_ocr_text"`
	LLMConfidence float64 `json:"llm_confidence"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
	// Legacy fields for template compatibility
	ImagePath       string `json:"image_path"`
	TransactionDate string `json:"transaction_date"`
}

func (t *Transaction) ToView() TransactionView {
	view := TransactionView{
		ID:        t.ID,
		Currency:  t.Currency,
		Direction: t.Direction,
		Status:    t.Status,
		CreatedAt: t.CreatedAt,
	}

	if t.TxnDate.Valid {
		view.TxnDate = t.TxnDate.String
		view.TransactionDate = t.TxnDate.String // Legacy
	}
	if t.Amount.Valid {
		view.Amount = t.Amount.Float64
	}
	if t.Channel.Valid {
		view.Channel = t.Channel.String
	}
	if t.AccountLabel.Valid {
		view.AccountLabel = t.AccountLabel.String
	}
	if t.Category.Valid {
		view.Category = t.Category.String
	}
	if t.Description.Valid {
		view.Description = t.Description.String
	}
	if t.SlipImagePath.Valid {
		view.SlipImagePath = t.SlipImagePath.String
		view.ImagePath = t.SlipImagePath.String // Legacy
	}
	if t.RawOCRText.Valid {
		view.RawOCRText = t.RawOCRText.String
	}
	if t.LLMConfidence.Valid {
		view.LLMConfidence = t.LLMConfidence.Float64
	}

	return view
}

type ConfirmRequest struct {
	Amount       float64 `json:"amount"`
	TxnDate      string  `json:"txn_date"`
	Direction    string  `json:"direction"`
	Channel      string  `json:"channel"`
	AccountLabel string  `json:"account_label"`
	Category     string  `json:"category"`
	Description  string  `json:"description"`
}
