package db

import (
	"database/sql"
	"fmt"
	"log"

	"expenseVault/models"

	_ "modernc.org/sqlite"
)

// ──────────────────────────────────────────────────────────
// UNIT 4 — Pointer-based struct & Factory
// ──────────────────────────────────────────────────────────

// Store manages all database operations.
// UNIT 2: Struct — groups related fields.
type Store struct {
	db *sql.DB // UNIT 4: Pointer field — *sql.DB
}

// NewStore opens/creates the SQLite database and initializes tables.
// UNIT 4: Factory function returning *Store (pointer).
// UNIT 3: Error handling — wrapping with models.DatabaseError.
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, &models.DatabaseError{Operation: "open", Err: err}
	}

	// UNIT 1: Short declaration operator — store := &Store{...}
	store := &Store{db: db}
	if err := store.createTables(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

// Close closes the database connection.
// UNIT 4: Pointer receiver — method on *Store.
func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) createTables() error {
	// UNIT 2: Slice — composite literal of SQL strings.
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

	// UNIT 2: for-range over slice.
	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			// UNIT 3: Errors with info — wrapping with context.
			return &models.DatabaseError{Operation: "create_tables", Err: err}
		}
	}
	return nil
}

// AddTransaction inserts a new transaction.
// UNIT 4: Accepts *Transaction (pointer) to avoid struct copy.
func (s *Store) AddTransaction(t *models.Transaction) (int64, error) {
	// UNIT 3: Error handling — checking errors.
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
// UNIT 4: Returns *Transaction (pointer).
func (s *Store) GetTransaction(id int) (*models.Transaction, error) {
	row := s.db.QueryRow(
		`SELECT id, type, amount, category, description, date, notes, created_at, updated_at
		 FROM transactions WHERE id = ?`,
		id,
	)

	// UNIT 1: var keyword — t starts with zero values.
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
	// UNIT 1: Conversion — models.Rupees(amount) converts float64 to custom type.
	t.Amount = models.Rupees(amount)
	return &t, nil
}

// ListTransactions lists filtered transactions.
// UNIT 2: Slice — dynamic queries with append.
func (s *Store) ListTransactions(txType, category, startDate, endDate string, limit int) ([]models.Transaction, error) {
	query := `SELECT id, type, amount, category, description, date, notes, created_at, updated_at FROM transactions WHERE 1=1`
	// UNIT 2: Slice — using []any{} (empty composite literal).
	args := []any{}

	// UNIT 1: Control flow — conditionals building dynamic query.
	if txType != "" {
		query += " AND type = ?"
		// UNIT 2: append — grows the slice.
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

	// UNIT 3: Unfurling a slice — args... spreads the slice into variadic Query().
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, &models.DatabaseError{Operation: "list", Err: err}
	}
	// UNIT 3: Defer — ensures rows.Close() runs when function returns.
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
// UNIT 4: Accepts *Transaction (pointer) to avoid struct copy.
func (s *Store) UpdateTransaction(t *models.Transaction) error {
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

// ──────────────────────────────────────────────────────────
// UNIT 3 — Panic / Recover in BulkInsert
// ──────────────────────────────────────────────────────────

// BulkInsert inserts multiple transactions.
// UNIT 3: Defer + Recover — recovers from panics during bulk insert.
// UNIT 4: Passes *Transaction (pointer) to AddTransaction.
func (s *Store) BulkInsert(transactions []models.Transaction) (count int, err error) {
	// UNIT 3: Defer + Recover — safety net for unexpected panics.
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PANIC RECOVERED] BulkInsert: %v", r)
			err = fmt.Errorf("bulk insert panicked after %d records: %v", count, r)
		}
	}()

	for i := range transactions {
		// UNIT 4: Pass by pointer — &transactions[i] avoids copy.
		if _, insertErr := s.AddTransaction(&transactions[i]); insertErr != nil {
			return count, insertErr
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
// UNIT 4: Returns *models.User (pointer).
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
