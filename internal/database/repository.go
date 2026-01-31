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

func (r *Repository) EnsureDefaultUser() (int64, error) {
	_, err := r.db.Exec(`INSERT OR IGNORE INTO users (name) VALUES ('default')`)
	if err != nil {
		return 0, err
	}

	var id int64
	if err := r.db.QueryRow(`SELECT id FROM users WHERE name = 'default'`).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) ListUsers() ([]models.User, error) {
	rows, err := r.db.Query(`SELECT id, name, cutoff_day FROM users ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.CutoffDay); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *Repository) CreateUser(name string) (*models.User, error) {
	result, err := r.db.Exec(`INSERT INTO users (name, cutoff_day) VALUES (?, 1)`, name)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.User{ID: id, Name: name, CutoffDay: 1}, nil
}

func (r *Repository) GetUser(id int64) (*models.User, error) {
	var user models.User
	if err := r.db.QueryRow(`SELECT id, name, cutoff_day FROM users WHERE id = ?`, id).Scan(&user.ID, &user.Name, &user.CutoffDay); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) UpdateUserCutoff(id int64, cutoffDay int) (*models.User, error) {
	_, err := r.db.Exec(`UPDATE users SET cutoff_day = ? WHERE id = ?`, cutoffDay, id)
	if err != nil {
		return nil, err
	}
	return r.GetUser(id)
}

// CreateTransaction creates a new transaction from a slip image
func (r *Repository) CreateTransaction(userID int64, slipImagePath string) (*models.Transaction, error) {
	result, err := r.db.Exec(
		`INSERT INTO transactions (user_id, slip_image_path, direction, currency, status)
		 VALUES (?, ?, 'expense', 'THB', 'pending')`,
		userID, slipImagePath,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetTransaction(userID, id)
}

// CreateTransactionFromChat creates a transaction from chat input (no image required)
func (r *Repository) CreateTransactionFromChat(
	userID int64,
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
			user_id, txn_date, amount, currency, direction, channel, account_label,
			category, description, slip_image_path, raw_ocr_text, llm_confidence, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, nullString(txnDate), nullFloat(amount), currency, direction,
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

	return r.GetTransaction(userID, id)
}

func (r *Repository) GetTransaction(userID, id int64) (*models.Transaction, error) {
	tx := &models.Transaction{}
	err := r.db.QueryRow(`
		SELECT id, user_id, txn_date, amount, currency, direction, channel, account_label,
		       category, description, slip_image_path, raw_ocr_text, llm_confidence,
		       status, created_at, updated_at
		FROM transactions WHERE id = ? AND user_id = ?
	`, id, userID).Scan(
		&tx.ID, &tx.UserID, &tx.TxnDate, &tx.Amount, &tx.Currency, &tx.Direction,
		&tx.Channel, &tx.AccountLabel, &tx.Category, &tx.Description,
		&tx.SlipImagePath, &tx.RawOCRText, &tx.LLMConfidence,
		&tx.Status, &tx.CreatedAt, &tx.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (r *Repository) DeleteTransaction(userID, id int64) error {
	result, err := r.db.Exec(`DELETE FROM transactions WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
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

func (r *Repository) ConfirmTransaction(userID, id int64, req models.ConfirmRequest) error {
	_, err := r.db.Exec(`
		UPDATE transactions
		SET amount = ?, txn_date = ?, direction = ?, channel = ?,
		    account_label = ?, category = ?, description = ?,
		    status = 'confirmed', updated_at = datetime('now')
		WHERE id = ? AND user_id = ?
	`, req.Amount, nullString(req.TxnDate), nullString(req.Direction),
		nullString(req.Channel), nullString(req.AccountLabel),
		nullString(req.Category), nullString(req.Description), id, userID)
	return err
}

func (r *Repository) ListTransactions(userID int64, limit, offset int) ([]models.Transaction, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, txn_date, amount, currency, direction, channel, account_label,
		       category, description, slip_image_path, raw_ocr_text, llm_confidence,
		       status, created_at, updated_at
		FROM transactions
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		err := rows.Scan(
			&tx.ID, &tx.UserID, &tx.TxnDate, &tx.Amount, &tx.Currency, &tx.Direction,
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
func (r *Repository) GetDashboardSummary(userID int64, from, to string) (*models.DashboardSummary, error) {
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
		  AND user_id = ?
		  AND COALESCE(NULLIF(txn_date, ''), date(created_at)) >= ?
		  AND COALESCE(NULLIF(txn_date, ''), date(created_at)) <= ?
	`, userID, from, to).Scan(&summary.TotalExpense, &summary.TotalIncome)
	if err != nil {
		return nil, err
	}

	// Get by category
	summary.ByCategory, err = r.GetExpenseByCategory(userID, from, to)
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
	summary.ByChannel, err = r.GetExpenseByChannel(userID, from, to)
	if err != nil {
		return nil, err
	}

	return summary, nil
}

// GetExpenseByCategory returns expense breakdown by category
func (r *Repository) GetExpenseByCategory(userID int64, from, to string) ([]models.CategoryAmount, error) {
	rows, err := r.db.Query(`
		SELECT COALESCE(NULLIF(category, ''), 'uncategorized') as category, SUM(amount) as amount
		FROM transactions
		WHERE status = 'confirmed'
		  AND user_id = ?
		  AND direction = 'expense'
		  AND COALESCE(NULLIF(txn_date, ''), date(created_at)) >= ?
		  AND COALESCE(NULLIF(txn_date, ''), date(created_at)) <= ?
		GROUP BY category
		ORDER BY amount DESC
	`, userID, from, to)
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
func (r *Repository) GetExpenseByChannel(userID int64, from, to string) ([]models.ChannelAmount, error) {
	rows, err := r.db.Query(`
		SELECT COALESCE(NULLIF(channel, ''), 'unknown') as channel, SUM(amount) as amount
		FROM transactions
		WHERE status = 'confirmed'
		  AND user_id = ?
		  AND direction = 'expense'
		  AND COALESCE(NULLIF(txn_date, ''), date(created_at)) >= ?
		  AND COALESCE(NULLIF(txn_date, ''), date(created_at)) <= ?
		GROUP BY channel
		ORDER BY amount DESC
	`, userID, from, to)
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
func (r *Repository) QuerySummary(userID int64, direction string, from, to string, category, channel string) (*models.DashboardSummary, error) {
	summary := &models.DashboardSummary{
		Period: models.Period{From: from, To: to},
	}

	// Build query based on filters - use created_at as fallback for txn_date
	baseQuery := `
		FROM transactions
		WHERE status = 'confirmed'
		  AND user_id = ?
		  AND COALESCE(NULLIF(txn_date, ''), date(created_at)) >= ?
		  AND COALESCE(NULLIF(txn_date, ''), date(created_at)) <= ?`
	args := []interface{}{userID, from, to}

	if category != "" {
		baseQuery += ` AND category = ?`
		args = append(args, category)
	}
	if channel != "" {
		baseQuery += ` AND channel = ?`
		args = append(args, channel)
	}

	var err error
	switch direction {
	case "expense":
		query := `SELECT COALESCE(SUM(amount), 0) ` + baseQuery + ` AND direction = 'expense'`
		err = r.db.QueryRow(query, args...).Scan(&summary.TotalExpense)
	case "income":
		query := `SELECT COALESCE(SUM(amount), 0) ` + baseQuery + ` AND direction = 'income'`
		err = r.db.QueryRow(query, args...).Scan(&summary.TotalIncome)
	default:
		query := `
			SELECT
				COALESCE(SUM(CASE WHEN direction = 'expense' THEN amount ELSE 0 END), 0),
				COALESCE(SUM(CASE WHEN direction = 'income' THEN amount ELSE 0 END), 0)
			` + baseQuery
		err = r.db.QueryRow(query, args...).Scan(&summary.TotalExpense, &summary.TotalIncome)
	}
	if err != nil {
		return nil, err
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
