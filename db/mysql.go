package db

import (
	"database/sql"
	"fmt"

	"expenseVault/models"

	_ "github.com/go-sql-driver/mysql"
)

// NewMySQLStore opens MySQL database connection and initializes tables.
func NewMySQLStore(dsn string) (*Store, error) {
	dsn = ensureMySQLParams(dsn)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, &models.DatabaseError{Operation: "open", Err: err}
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, &models.DatabaseError{Operation: "ping", Err: fmt.Errorf("MySQL connection failed: %w", err)}
	}

	store := &Store{db: db, isMySQL: true}

	if err := createMySQLTables(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func ensureMySQLParams(dsn string) string {
	if dsn == "" {
		return dsn
	}
	if !hasParam(dsn) {
		return dsn + "?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"
	}
	return dsn + "&parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"
}

func hasParam(dsn string) bool {
	for i := len(dsn) - 1; i >= 0; i-- {
		switch dsn[i] {
		case '?':
			return true
		case '/':
			return false
		}
	}
	return false
}

func createMySQLTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_login TIMESTAMP NULL,
			INDEX idx_username (username)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT,
			type VARCHAR(50) NOT NULL,
			amount DECIMAL(15,2) NOT NULL,
			category VARCHAR(100) NOT NULL,
			description TEXT NOT NULL,
			date DATE NOT NULL,
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_tx_date (date),
			INDEX idx_tx_category (category),
			INDEX idx_tx_type (type),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS budgets (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			category VARCHAR(100) NOT NULL,
			amount DECIMAL(15,2) NOT NULL,
			month VARCHAR(7) NOT NULL,
			UNIQUE KEY uk_user_cat_month (user_id, category, month),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return &models.DatabaseError{Operation: "create_tables", Err: err}
		}
	}

	// Safe migration: Add user_id to transactions if missing
	var columnCount int
	err := db.QueryRow("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'transactions' AND column_name = 'user_id' AND table_schema = DATABASE()").Scan(&columnCount)
	if err == nil && columnCount == 0 {
		_, _ = db.Exec("ALTER TABLE transactions ADD COLUMN user_id BIGINT DEFAULT 1, ADD FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE")
	}

	return nil
}
