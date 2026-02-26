package models

import "time"

// Transaction represents a financial transaction.
type Transaction struct {
	ID          int64           `json:"id"`
	Type        TransactionType `json:"type"`
	Amount      Rupees          `json:"amount"`
	Category    Category        `json:"category"`
	Description string          `json:"description"`
	Date        string          `json:"date"`
	Notes       string          `json:"notes"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// User represents an authenticated user.
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	LastLogin    time.Time `json:"last_login"`
}

// ReportEntry represents a single line in a report.
type ReportEntry struct {
	Label  string
	Amount Rupees
	Count  int
}

// MonthlySummary holds data for a monthly report.
type MonthlySummary struct {
	Month        string
	Income       Rupees
	Expenses     Rupees
	Balance      Rupees
	ByCategory   map[Category]Rupees
	Transactions []Transaction
}
