package db

import (
	"database/sql"
	"fmt"

	"expenseVault/models"

	_ "modernc.org/sqlite"
)

// Store manages all database operations.
type Store struct {
	db *sql.DB
}

// NewStore opens/creates the SQLite database and initializes tables.
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, &models.DatabaseError{Operation: "open", Err: err}
	}

	store := &Store{db: db}
	if err := store.createTables(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL,
			amount REAL NOT NULL,
			category TEXT NOT NULL,
			description TEXT NOT NULL,
			date TEXT NOT NULL,
			notes TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_login DATETIME
		)`,
		`CREATE INDEX IF NOT EXISTS idx_tx_date ON transactions(date)`,
		`CREATE INDEX IF NOT EXISTS idx_tx_category ON transactions(category)`,
		`CREATE INDEX IF NOT EXISTS idx_tx_type ON transactions(type)`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return &models.DatabaseError{Operation: "create_tables", Err: err}
		}
	}
	return nil
}

// AddTransaction inserts a new transaction.
func (s *Store) AddTransaction(t models.Transaction) (int64, error) {
	if err := models.ValidateTransaction(t); err != nil {
		return 0, err
	}

	result, err := s.db.Exec(
		`INSERT INTO transactions (type, amount, category, description, date, notes)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		t.Type, t.Amount.ToFloat64(), t.Category, t.Description, t.Date, t.Notes,
	)
	if err != nil {
		return 0, &models.DatabaseError{Operation: "insert", Err: err}
	}
	return result.LastInsertId()
}

// GetTransaction retrieves a transaction by ID.
func (s *Store) GetTransaction(id int) (*models.Transaction, error) {
	row := s.db.QueryRow(
		`SELECT id, type, amount, category, description, date, notes, created_at, updated_at
		 FROM transactions WHERE id = ?`,
		id,
	)

	var t models.Transaction
	var amount float64
	if err := row.Scan(&t.ID, &t.Type, &amount, &t.Category, &t.Description, &t.Date, &t.Notes, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, &models.DatabaseError{Operation: "get", Err: err}
	}
	if t.Type != models.Income && t.Type != models.Expense {
		return nil, &models.ValidationError{Field: "type", Msg: "invalid type in database"}
	}
	if t.Category == "" {
		return nil, &models.ValidationError{Field: "category", Msg: "missing"}
	}
	t.Amount = models.Rupees(amount)
	return &t, nil
}

// ListTransactions lists filtered transactions.
func (s *Store) ListTransactions(txType, category, startDate, endDate string, limit int) ([]models.Transaction, error) {
	query := `SELECT id, type, amount, category, description, date, notes, created_at, updated_at FROM transactions WHERE 1=1`
	args := []any{}

	if txType != "" {
		query += " AND type = ?"
		args = append(args, txType)
	}
	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}
	if startDate != "" {
		query += " AND date >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND date <= ?"
		args = append(args, endDate)
	}
	query += " ORDER BY date DESC, id DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, &models.DatabaseError{Operation: "list", Err: err}
	}
	defer rows.Close()

	var txs []models.Transaction
	for rows.Next() {
		var t models.Transaction
		var amount float64
		if err := rows.Scan(&t.ID, &t.Type, &amount, &t.Category, &t.Description, &t.Date, &t.Notes, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, &models.DatabaseError{Operation: "scan", Err: err}
		}
		t.Amount = models.Rupees(amount)
		txs = append(txs, t)
	}
	return txs, nil
}

// UpdateTransaction updates an existing transaction.
func (s *Store) UpdateTransaction(t models.Transaction) error {
	if err := models.ValidateTransaction(t); err != nil {
		return err
	}

	_, err := s.db.Exec(
		`UPDATE transactions SET type = ?, amount = ?, category = ?, description = ?, date = ?, notes = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		t.Type, t.Amount.ToFloat64(), t.Category, t.Description, t.Date, t.Notes, t.ID,
	)
	if err != nil {
		return &models.DatabaseError{Operation: "update", Err: err}
	}
	return nil
}

// DeleteTransaction deletes a transaction by ID.
func (s *Store) DeleteTransaction(id int) error {
	_, err := s.db.Exec("DELETE FROM transactions WHERE id = ?", id)
	if err != nil {
		return &models.DatabaseError{Operation: "delete", Err: err}
	}
	return nil
}

// GetAllTransactions returns all transactions.
func (s *Store) GetAllTransactions() ([]models.Transaction, error) {
	return s.ListTransactions("", "", "", "", 0)
}

// BulkInsert inserts multiple transactions.
func (s *Store) BulkInsert(transactions []models.Transaction) (int, error) {
	count := 0
	for _, t := range transactions {
		if _, err := s.AddTransaction(t); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// CreateUser creates a new user.
func (s *Store) CreateUser(username, passwordHash string) (int64, error) {
	result, err := s.db.Exec(
		"INSERT INTO users (username, password_hash) VALUES (?, ?)",
		username, passwordHash,
	)
	if err != nil {
		return 0, &models.DatabaseError{Operation: "create_user", Err: err}
	}
	return result.LastInsertId()
}

// GetUserByUsername retrieves a user by username.
func (s *Store) GetUserByUsername(username string) (*models.User, error) {
	row := s.db.QueryRow(
		"SELECT id, username, password_hash, created_at, COALESCE(last_login, created_at) FROM users WHERE username = ?",
		username,
	)
	var u models.User
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt, &u.LastLogin); err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.AuthError{Reason: "user not found"}
		}
		return nil, &models.DatabaseError{Operation: "get_user", Err: err}
	}
	return &u, nil
}

// UpdateLastLogin updates the last_login timestamp for a user.
func (s *Store) UpdateLastLogin(userID int64) error {
	_, err := s.db.Exec(
		"UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = ?",
		userID,
	)
	if err != nil {
		return &models.DatabaseError{Operation: "update_last_login", Err: err}
	}
	return nil
}
