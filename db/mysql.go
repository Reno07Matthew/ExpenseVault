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

	store := &Store{db: db}

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
		`CREATE TABLE IF NOT EXISTS transactions (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
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
			INDEX idx_tx_type (type)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_login TIMESTAMP NULL,
			INDEX idx_username (username)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return &models.DatabaseError{Operation: "create_tables", Err: err}
		}
	}
	return nil
}
