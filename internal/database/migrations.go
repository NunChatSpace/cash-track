package database

import "database/sql"

func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		txn_date TEXT,
		amount REAL,
		currency TEXT NOT NULL DEFAULT 'THB',
		direction TEXT NOT NULL DEFAULT 'expense',
		channel TEXT,
		account_label TEXT,
		category TEXT,
		description TEXT,
		slip_image_path TEXT,
		raw_ocr_text TEXT,
		llm_confidence REAL,
		status TEXT NOT NULL DEFAULT 'pending',
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	);

	CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
	CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at);
	CREATE INDEX IF NOT EXISTS idx_transactions_txn_date ON transactions(txn_date);
	CREATE INDEX IF NOT EXISTS idx_transactions_direction ON transactions(direction);
	CREATE INDEX IF NOT EXISTS idx_transactions_category ON transactions(category);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return err
	}

	// Migration: add new columns if they don't exist (for existing databases)
	migrations := []string{
		`ALTER TABLE transactions ADD COLUMN txn_date TEXT`,
		`ALTER TABLE transactions ADD COLUMN currency TEXT DEFAULT 'THB'`,
		`ALTER TABLE transactions ADD COLUMN direction TEXT DEFAULT 'expense'`,
		`ALTER TABLE transactions ADD COLUMN account_label TEXT`,
		`ALTER TABLE transactions ADD COLUMN category TEXT`,
		`ALTER TABLE transactions ADD COLUMN description TEXT`,
		`ALTER TABLE transactions ADD COLUMN slip_image_path TEXT`,
		`ALTER TABLE transactions ADD COLUMN llm_confidence REAL`,
	}

	for _, m := range migrations {
		// Ignore errors for columns that already exist
		db.Exec(m)
	}

	// Migrate old data: copy image_path to slip_image_path if exists
	db.Exec(`UPDATE transactions SET slip_image_path = image_path WHERE slip_image_path IS NULL AND image_path IS NOT NULL`)

	// Migrate old data: copy transaction_date to txn_date
	db.Exec(`UPDATE transactions SET txn_date = date(transaction_date) WHERE txn_date IS NULL AND transaction_date IS NOT NULL`)

	return nil
}
