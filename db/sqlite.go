package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"expenseVault/models"

	_ "modernc.org/sqlite"
)

// ──────────────────────────────────────────────────────────
// UNIT 4 — Pointer-based struct & Factory
// ──────────────────────────────────────────────────────────

// Store manages all database operations.
// UNIT 2: Struct — groups related fields.
type Store struct {
	db         *sql.DB
	isMySQL    bool
	isPostgres bool
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
	store := &Store{db: db, isMySQL: false}
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
	// Enable foreign key support
	if _, err := s.db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return &models.DatabaseError{Operation: "enable_fks", Err: err}
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_login DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER REFERENCES users(id),
			type TEXT NOT NULL,
			amount REAL NOT NULL,
			category TEXT NOT NULL,
			description TEXT NOT NULL,
			date TEXT NOT NULL,
			notes TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS budgets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id),
			category TEXT NOT NULL,
			amount REAL NOT NULL,
			month TEXT NOT NULL,
			UNIQUE(user_id, category, month)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_tx_date ON transactions(date)`,
		`CREATE INDEX IF NOT EXISTS idx_tx_category ON transactions(category)`,
		`CREATE INDEX IF NOT EXISTS idx_tx_type ON transactions(type)`,
		`CREATE INDEX IF NOT EXISTS idx_budget_lookup ON budgets(user_id, month)`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return &models.DatabaseError{Operation: "create_tables", Err: err}
		}
	}

	// Safe migration: Add user_id to transactions if missing
	hasUserID := false
	rows, err := s.db.Query("PRAGMA table_info(transactions)")
	if err == nil {
		for rows.Next() {
			var cid int
			var name, dtype string
			var notnull, pk int
			var dflt any
			if err := rows.Scan(&cid, &name, &dtype, &notnull, &dflt, &pk); err == nil {
				if name == "user_id" {
					hasUserID = true
					break
				}
			}
		}
		rows.Close()
	}

	if !hasUserID {
		// Use DEFAULT 0 so it doesn't fail FK until a real user is needed
		_, _ = s.db.Exec("ALTER TABLE transactions ADD COLUMN user_id INTEGER DEFAULT 0")
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

	query := `INSERT INTO transactions (user_id, type, amount, category, description, date, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`
	
	if s.isPostgres {
		var lastInsertID int64
		err := s.db.QueryRow(s.rebind(query+" RETURNING id"),
			t.UserID, t.Type, t.Amount.ToFloat64(), t.Category, t.Description, t.Date, t.Notes,
		).Scan(&lastInsertID)
		if err != nil {
			return 0, &models.DatabaseError{Operation: "insert", Err: err}
		}
		return lastInsertID, nil
	}

	result, err := s.db.Exec(
		s.rebind(query),
		t.UserID, t.Type, t.Amount.ToFloat64(), t.Category, t.Description, t.Date, t.Notes,
	)
	if err != nil {
		return 0, &models.DatabaseError{Operation: "insert", Err: err}
	}
	return result.LastInsertId()
}

// GetTransaction retrieves a transaction by ID.
// UNIT 4: Returns *Transaction (pointer).
func (s *Store) GetTransaction(id int) (*models.Transaction, error) {
	query := `SELECT id, user_id, type, amount, category, description, date, notes, created_at, updated_at
		 FROM transactions WHERE id = ?`
	row := s.db.QueryRow(s.rebind(query), id)

	// UNIT 1: var keyword — t starts with zero values.
	var t models.Transaction
	var amount float64
	if err := row.Scan(&t.ID, &t.UserID, &t.Type, &amount, &t.Category, &t.Description, &t.Date, &t.Notes, &t.CreatedAt, &t.UpdatedAt); err != nil {
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

// ListTransactions lists filtered transactions for a specific user.
// UNIT 2: Slice — dynamic queries with append.
func (s *Store) ListTransactions(userID int64, txType, category, startDate, endDate string, limit int) ([]models.Transaction, error) {
	query := `SELECT id, user_id, type, amount, category, description, date, notes, created_at, updated_at FROM transactions WHERE user_id = ?`
	// UNIT 2: Slice — using []any{} (empty composite literal).
	args := []any{userID}

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
	rows, err := s.db.Query(s.rebind(query), args...)
	if err != nil {
		return nil, &models.DatabaseError{Operation: "list", Err: err}
	}
	// UNIT 3: Defer — ensures rows.Close() runs when function returns.
	defer rows.Close()

	var txs []models.Transaction
	for rows.Next() {
		var t models.Transaction
		var amount float64
		if err := rows.Scan(&t.ID, &t.UserID, &t.Type, &amount, &t.Category, &t.Description, &t.Date, &t.Notes, &t.CreatedAt, &t.UpdatedAt); err != nil {
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

	query := `UPDATE transactions SET type = ?, amount = ?, category = ?, description = ?, date = ?, notes = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?`
	_, err := s.db.Exec(
		s.rebind(query),
		t.Type, t.Amount.ToFloat64(), t.Category, t.Description, t.Date, t.Notes, t.ID, t.UserID,
	)
	if err != nil {
		return &models.DatabaseError{Operation: "update", Err: err}
	}
	return nil
}

// DeleteTransaction deletes a transaction by ID and user ID.
func (s *Store) DeleteTransaction(id int, userID int64) error {
	query := "DELETE FROM transactions WHERE id = ? AND user_id = ?"
	_, err := s.db.Exec(s.rebind(query), id, userID)
	if err != nil {
		return &models.DatabaseError{Operation: "delete", Err: err}
	}
	return nil
}

// GetAllTransactions returns all transactions for a specific user.
func (s *Store) GetAllTransactions(userID int64) ([]models.Transaction, error) {
	return s.ListTransactions(userID, "", "", "", "", 0)
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
	query := "INSERT INTO users (username, password_hash) VALUES (?, ?)"
	
	if s.isPostgres {
		var lastInsertID int64
		err := s.db.QueryRow(s.rebind(query+" RETURNING id"), username, passwordHash).Scan(&lastInsertID)
		if err != nil {
			return 0, &models.DatabaseError{Operation: "create_user", Err: err}
		}
		return lastInsertID, nil
	}

	result, err := s.db.Exec(s.rebind(query), username, passwordHash)
	if err != nil {
		return 0, &models.DatabaseError{Operation: "create_user", Err: err}
	}
	return result.LastInsertId()
}

// GetUserByUsername retrieves a user by username.
// UNIT 4: Returns *models.User (pointer).
func (s *Store) GetUserByUsername(username string) (*models.User, error) {
	query := "SELECT id, username, password_hash, created_at, COALESCE(last_login, created_at) FROM users WHERE username = ?"
	row := s.db.QueryRow(s.rebind(query), username)
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
	query := "UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = ?"
	_, err := s.db.Exec(s.rebind(query), userID)
	if err != nil {
		return &models.DatabaseError{Operation: "update_last_login", Err: err}
	}
	return nil
}
// SetBudget creates or updates a budget for a user.
func (s *Store) SetBudget(userID int64, category models.Category, amount models.Rupees, month string) error {
	var query string
	if s.isMySQL {
		query = `INSERT INTO budgets (user_id, category, amount, month)
				 VALUES (?, ?, ?, ?)
				 ON DUPLICATE KEY UPDATE amount = VALUES(amount)`
	} else {
		query = `INSERT INTO budgets (user_id, category, amount, month)
				 VALUES (?, ?, ?, ?)
				 ON CONFLICT(user_id, category, month) DO UPDATE SET amount = excluded.amount`
	}

	_, err := s.db.Exec(s.rebind(query), userID, category, amount.ToFloat64(), month)
	if err != nil {
		return &models.DatabaseError{Operation: "set_budget", Err: err}
	}
	return nil
}

// GetBudgets retrieves all budgets for a specific user and month.
func (s *Store) GetBudgets(userID int64, month string) (map[models.Category]models.Rupees, error) {
	query := "SELECT category, amount FROM budgets WHERE user_id = ? AND month = ?"
	rows, err := s.db.Query(s.rebind(query), userID, month)
	if err != nil {
		return nil, &models.DatabaseError{Operation: "get_budgets", Err: err}
	}
	defer rows.Close()

	budgets := make(map[models.Category]models.Rupees)
	for rows.Next() {
		var cat string
		var amt float64
		if err := rows.Scan(&cat, &amt); err != nil {
			return nil, err
		}
		budgets[models.Category(cat)] = models.Rupees(amt * 100)
	}
	return budgets, nil
}

// ExecuteReadQuery safely executes a read-only SQL query and returns dynamic rows.
func (s *Store) ExecuteReadQuery(query string, params ...interface{}) ([]map[string]interface{}, error) {
	// Basic safety check: ensure the query starts with SELECT
	qUpper := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(qUpper, "SELECT") {
		return nil, fmt.Errorf("only SELECT queries are allowed for safety")
	}

	rows, err := s.db.Query(s.rebind(query), params...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}

	for rows.Next() {
		columnsData := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columnsData {
			columnPointers[i] = &columnsData[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		rowData := make(map[string]interface{})
		for i, colName := range columns {
			val := columnPointers[i].(*interface{})
			if b, ok := (*val).([]byte); ok {
				rowData[colName] = string(b)
			} else {
				rowData[colName] = *val
			}
		}
		result = append(result, rowData)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Store) rebind(query string) string {
	if !s.isPostgres {
		return query
	}

	var sb strings.Builder
	paramIdx := 1
	for {
		idx := strings.Index(query, "?")
		if idx == -1 {
			sb.WriteString(query)
			break
		}
		sb.WriteString(query[:idx])
		fmt.Fprintf(&sb, "$%d", paramIdx)
		paramIdx++
		query = query[idx+1:]
	}
	return sb.String()
}
