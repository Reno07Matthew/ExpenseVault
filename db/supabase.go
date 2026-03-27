package db

import (
	"database/sql"
	"fmt"
	"net/url"

	"expenseVault/models"

	_ "github.com/lib/pq"
)

// convertDSN converts a postgresql:// URI to lib/pq keyword/value format.
// This fixes Supabase pooler dotted usernames (e.g. postgres.projectid)
// which lib/pq misparses in URI format.
func convertDSN(dsn string) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", err
	}

	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "5432"
	}
	dbname := ""
	if len(u.Path) > 1 {
		dbname = u.Path[1:]
	}
	user := u.User.Username()
	password, _ := u.User.Password()

	connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=require",
		host, port, dbname, user, password)

	return connStr, nil
}

// NewSupabaseStore opens Supabase (PostgreSQL) database connection and initializes tables.
func NewSupabaseStore(dsn string) (*Store, error) {
	if dsn == "" {
		return nil, &models.DatabaseError{Operation: "open", Err: fmt.Errorf("Supabase DSN is empty")}
	}

	// Convert URI format to keyword/value format for lib/pq.
	// Supabase pooler uses dotted usernames (e.g. postgres.projectid)
	// which lib/pq misparses in URI format.
	connStr, err := convertDSN(dsn)
	if err != nil {
		return nil, &models.DatabaseError{Operation: "open", Err: fmt.Errorf("invalid DSN: %w", err)}
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, &models.DatabaseError{Operation: "open", Err: err}
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, &models.DatabaseError{Operation: "ping", Err: fmt.Errorf("Supabase connection failed: %w", err)}
	}

	store := &Store{db: db, isMySQL: false, isPostgres: true}

	if err := createSupabaseTables(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func createSupabaseTables(db *sql.DB) error {
	// PostgreSql uses SERIAL for auto-increment and different table syntax.
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			last_login TIMESTAMP WITH TIME ZONE NULL
		)`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
			type VARCHAR(50) NOT NULL,
			amount DECIMAL(15,2) NOT NULL,
			category VARCHAR(100) NOT NULL,
			description TEXT NOT NULL,
			date DATE NOT NULL,
			notes TEXT DEFAULT '',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS budgets (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			category VARCHAR(100) NOT NULL,
			amount DECIMAL(15,2) NOT NULL,
			month VARCHAR(7) NOT NULL,
			UNIQUE (user_id, category, month)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_tx_date ON transactions(date)`,
		`CREATE INDEX IF NOT EXISTS idx_tx_category ON transactions(category)`,
		`CREATE INDEX IF NOT EXISTS idx_tx_type ON transactions(type)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return &models.DatabaseError{Operation: "create_tables", Err: err}
		}
	}

	// Safe migration: Add user_id to transactions if missing
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'transactions' AND column_name = 'user_id'
		)`).Scan(&exists)
	
	if err == nil && !exists {
		_, _ = db.Exec("ALTER TABLE transactions ADD COLUMN user_id BIGINT DEFAULT 1 REFERENCES users(id) ON DELETE CASCADE")
	}

	return nil
}

