package database

import (
	"database/sql"

	"cash-track/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateTransaction creates a new transaction from a slip image
func (r *Repository) CreateTransaction(slipImagePath string) (*models.Transaction, error) {
	result, err := r.db.Exec(
		`INSERT INTO transactions (slip_image_path, direction, currency, status)
		 VALUES (?, 'expense', 'THB', 'pending')`,
		slipImagePath,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetTransaction(id)
}

// CreateTransactionFromChat creates a transaction from chat input (no image required)
func (r *Repository) CreateTransactionFromChat(
	txnDate string,
	amount float64,
	currency string,
	direction string,
	channel string,
	accountLabel string,
	category string,
	description string,
	slipImagePath string,
	rawOCRText string,
	llmConfidence float64,
	status string,
) (*models.Transaction, error) {
	result, err := r.db.Exec(`
		INSERT INTO transactions (
			txn_date, amount, currency, direction, channel, account_label,
			category, description, slip_image_path, raw_ocr_text, llm_confidence, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		nullString(txnDate), nullFloat(amount), currency, direction,
		nullString(channel), nullString(accountLabel), nullString(category),
		nullString(description), nullString(slipImagePath), nullString(rawOCRText),
		nullFloat(llmConfidence), status,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetTransaction(id)
}

func (r *Repository) GetTransaction(id int64) (*models.Transaction, error) {
	tx := &models.Transaction{}
	err := r.db.QueryRow(`
		SELECT id, txn_date, amount, currency, direction, channel, account_label,
		       category, description, slip_image_path, raw_ocr_text, llm_confidence,
		       status, created_at, updated_at
		FROM transactions WHERE id = ?
	`, id).Scan(
		&tx.ID, &tx.TxnDate, &tx.Amount, &tx.Currency, &tx.Direction,
		&tx.Channel, &tx.AccountLabel, &tx.Category, &tx.Description,
		&tx.SlipImagePath, &tx.RawOCRText, &tx.LLMConfidence,
		&tx.Status, &tx.CreatedAt, &tx.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// UpdateOCRResult updates a transaction with OCR/LLM parsed data
func (r *Repository) UpdateOCRResult(
	id int64,
	rawText string,
	amount float64,
	txnDate string,
	channel string,
	category string,
	description string,
	llmConfidence float64,
) error {
	_, err := r.db.Exec(`
		UPDATE transactions
		SET raw_ocr_text = ?, amount = ?, txn_date = ?, channel = ?,
		    category = ?, description = ?, llm_confidence = ?,
		    updated_at = datetime('now')
		WHERE id = ?
	`, rawText, nullFloat(amount), nullString(txnDate), nullString(channel),
		nullString(category), nullString(description), nullFloat(llmConfidence), id)
	return err
}

func (r *Repository) ConfirmTransaction(id int64, req models.ConfirmRequest) error {
	_, err := r.db.Exec(`
		UPDATE transactions
		SET amount = ?, txn_date = ?, direction = ?, channel = ?,
		    account_label = ?, category = ?, description = ?,
		    status = 'confirmed', updated_at = datetime('now')
		WHERE id = ?
	`, req.Amount, nullString(req.TxnDate), nullString(req.Direction),
		nullString(req.Channel), nullString(req.AccountLabel),
		nullString(req.Category), nullString(req.Description), id)
	return err
}

func (r *Repository) ListTransactions(limit, offset int) ([]models.Transaction, error) {
	rows, err := r.db.Query(`
		SELECT id, txn_date, amount, currency, direction, channel, account_label,
		       category, description, slip_image_path, raw_ocr_text, llm_confidence,
		       status, created_at, updated_at
		FROM transactions
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		err := rows.Scan(
			&tx.ID, &tx.TxnDate, &tx.Amount, &tx.Currency, &tx.Direction,
			&tx.Channel, &tx.AccountLabel, &tx.Category, &tx.Description,
			&tx.SlipImagePath, &tx.RawOCRText, &tx.LLMConfidence,
			&tx.Status, &tx.CreatedAt, &tx.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}
	return transactions, rows.Err()
}

// GetDashboardSummary returns aggregated data for the dashboard
func (r *Repository) GetDashboardSummary(from, to string) (*models.DashboardSummary, error) {
	summary := &models.DashboardSummary{
		Period: models.Period{From: from, To: to},
	}

	// Get totals - use COALESCE to fallback to created_at date if txn_date is empty
	err := r.db.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN direction = 'expense' THEN amount ELSE 0 END), 0) as total_expense,
			COALESCE(SUM(CASE WHEN direction = 'income' THEN amount ELSE 0 END), 0) as total_income
		FROM transactions
		WHERE status = 'confirmed'
		  AND COALESCE(NULLIF(txn_date, ''), date(created_at)) >= ?
		  AND COALESCE(NULLIF(txn_date, ''), date(created_at)) <= ?
	`, from, to).Scan(&summary.TotalExpense, &summary.TotalIncome)
	if err != nil {
		return nil, err
	}

	// Get by category
	summary.ByCategory, err = r.GetExpenseByCategory(from, to)
	if err != nil {
		return nil, err
	}

	// Calculate percentages
	for i := range summary.ByCategory {
		if summary.TotalExpense > 0 {
			summary.ByCategory[i].PercentOfExpense = (summary.ByCategory[i].Amount / summary.TotalExpense) * 100
		}
	}

	// Get by channel
	summary.ByChannel, err = r.GetExpenseByChannel(from, to)
	if err != nil {
		return nil, err
	}

	return summary, nil
}

// GetExpenseByCategory returns expense breakdown by category
func (r *Repository) GetExpenseByCategory(from, to string) ([]models.CategoryAmount, error) {
	rows, err := r.db.Query(`
		SELECT COALESCE(NULLIF(category, ''), 'uncategorized') as category, SUM(amount) as amount
		FROM transactions
		WHERE status = 'confirmed'
		  AND direction = 'expense'
		  AND COALESCE(NULLIF(txn_date, ''), date(created_at)) >= ?
		  AND COALESCE(NULLIF(txn_date, ''), date(created_at)) <= ?
		GROUP BY category
		ORDER BY amount DESC
	`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.CategoryAmount
	for rows.Next() {
		var ca models.CategoryAmount
		if err := rows.Scan(&ca.Category, &ca.Amount); err != nil {
			return nil, err
		}
		result = append(result, ca)
	}
	return result, rows.Err()
}

// GetExpenseByChannel returns expense breakdown by channel
func (r *Repository) GetExpenseByChannel(from, to string) ([]models.ChannelAmount, error) {
	rows, err := r.db.Query(`
		SELECT COALESCE(NULLIF(channel, ''), 'unknown') as channel, SUM(amount) as amount
		FROM transactions
		WHERE status = 'confirmed'
		  AND direction = 'expense'
		  AND COALESCE(NULLIF(txn_date, ''), date(created_at)) >= ?
		  AND COALESCE(NULLIF(txn_date, ''), date(created_at)) <= ?
		GROUP BY channel
		ORDER BY amount DESC
	`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.ChannelAmount
	for rows.Next() {
		var ca models.ChannelAmount
		if err := rows.Scan(&ca.Channel, &ca.Amount); err != nil {
			return nil, err
		}
		result = append(result, ca)
	}
	return result, rows.Err()
}

// QuerySummary returns summary data based on query filters (for chat)
func (r *Repository) QuerySummary(direction string, from, to string, category, channel string) (*models.DashboardSummary, error) {
	summary := &models.DashboardSummary{
		Period: models.Period{From: from, To: to},
	}

	// Build query based on filters - use created_at as fallback for txn_date
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE status = 'confirmed'
		  AND COALESCE(NULLIF(txn_date, ''), date(created_at)) >= ?
		  AND COALESCE(NULLIF(txn_date, ''), date(created_at)) <= ?`
	args := []interface{}{from, to}

	if direction == "expense" || direction == "income" {
		query += ` AND direction = ?`
		args = append(args, direction)
	}
	if category != "" {
		query += ` AND category = ?`
		args = append(args, category)
	}
	if channel != "" {
		query += ` AND channel = ?`
		args = append(args, channel)
	}

	var total float64
	err := r.db.QueryRow(query, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	if direction == "expense" {
		summary.TotalExpense = total
	} else if direction == "income" {
		summary.TotalIncome = total
	} else {
		// Get both
		summary.TotalExpense = total
		summary.TotalIncome = total
	}

	return summary, nil
}

// Helper functions for nullable fields
func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullFloat(f float64) interface{} {
	if f == 0 {
		return nil
	}
	return f
}
