package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// Transaction represents a financial transaction.
type Transaction struct {
	ID          int64           `json:"id"`
	UserID      int64           `json:"user_id"`
	Type        TransactionType `json:"type"`
	Amount      Rupees          `json:"amount"`
	Category    Category        `json:"category"`
	Description string          `json:"description"`
	Date        string          `json:"date"`
	Notes       string          `json:"notes"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// ──────────────────────────────────────────────────────────
// LAB 4 — Value Receiver: String() (read-only, works on copy)
// ──────────────────────────────────────────────────────────

// String returns a human-readable summary of the transaction.
// Value receiver — does NOT modify the original; operates on a copy.
func (t Transaction) String() string {
	return fmt.Sprintf("[%d] %s | %s | %s | %s | %s",
		t.ID, t.Type, t.Amount, t.Category, t.Description, t.Date)
}

// Summary returns a short one-line summary (value receiver).
func (t Transaction) Summary() string {
	return fmt.Sprintf("%s: %s (%s)", t.Description, t.Amount, t.Category)
}

// IsExpense checks if the transaction is an expense (value receiver).
func (t Transaction) IsExpense() bool {
	return t.Type == Expense
}

// IsIncome checks if the transaction is income (value receiver).
func (t Transaction) IsIncome() bool {
	return t.Type == Income
}

// ──────────────────────────────────────────────────────────
// LAB 4 — Pointer Receivers: mutation methods (modify in-place)
// ──────────────────────────────────────────────────────────

// SetAmount updates the transaction amount in-place.
// Pointer receiver — mutates the original struct.
func (t *Transaction) SetAmount(a Rupees) {
	t.Amount = a
	t.UpdatedAt = time.Now()
}

// SetCategory updates the category in-place.
// Pointer receiver — mutates the original struct.
func (t *Transaction) SetCategory(c Category) {
	t.Category = c
	t.UpdatedAt = time.Now()
}

// SetDescription updates the description in-place.
// Pointer receiver — mutates the original struct.
func (t *Transaction) SetDescription(desc string) {
	t.Description = desc
	t.UpdatedAt = time.Now()
}

// SetDate updates the date in-place.
// Pointer receiver — mutates the original struct.
func (t *Transaction) SetDate(date string) {
	t.Date = date
	t.UpdatedAt = time.Now()
}

// SetNotes updates the notes in-place.
// Pointer receiver — mutates the original struct.
func (t *Transaction) SetNotes(notes string) {
	t.Notes = notes
	t.UpdatedAt = time.Now()
}

// ApplyDiscount reduces amount by a percentage (0–100). Pointer receiver.
func (t *Transaction) ApplyDiscount(pct float64) {
	if pct > 0 && pct <= 100 {
		t.Amount = Rupees(float64(t.Amount) * (1 - pct/100))
		t.UpdatedAt = time.Now()
	}
}

// ──────────────────────────────────────────────────────────
// LAB 4 — Factory function returning *Transaction (pointer)
// ──────────────────────────────────────────────────────────

// NewTransaction creates and returns a pointer to a new Transaction.
// Returning a pointer avoids copying and lets callers mutate directly.
func NewTransaction(userID int64, txType TransactionType, amount float64, category Category, desc, date string) *Transaction {
	now := time.Now()
	return &Transaction{
		UserID:      userID,
		Type:        txType,
		Amount:      Rupees(amount),
		Category:    category,
		Description: desc,
		Date:        date,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// CloneTransaction returns a deep copy (new pointer) of the transaction.
// Demonstrates returning a pointer from a function.
func CloneTransaction(src *Transaction) *Transaction {
	copy := *src // value copy
	return &copy // return pointer to the new copy
}

// ──────────────────────────────────────────────────────────
// LAB 4 — Pass-by-pointer helpers for edit operations
// ──────────────────────────────────────────────────────────

// EditTransactionFields modifies a transaction through a pointer.
// Only non-zero/non-empty values are applied.
func EditTransactionFields(t *Transaction, txType TransactionType, amount float64, category Category, desc, date, notes string) {
	if txType != "" {
		t.Type = txType
	}
	if amount > 0 {
		t.Amount = Rupees(amount)
	}
	if category != "" {
		t.Category = category
	}
	if desc != "" {
		t.Description = desc
	}
	if date != "" {
		t.Date = date
	}
	if notes != "" {
		t.Notes = notes
	}
	t.UpdatedAt = time.Now()
}

// ModifyByValue takes a Transaction by value — changes do NOT affect caller.
// This exists to demonstrate that pass-by-value creates a copy.
func ModifyByValue(t Transaction, newAmount Rupees) Transaction {
	t.Amount = newAmount // modifies the local copy only
	return t
}

// ModifyByPointer takes a *Transaction — changes DO affect caller.
// This exists to demonstrate that pass-by-pointer mutates the original.
func ModifyByPointer(t *Transaction, newAmount Rupees) {
	t.Amount = newAmount // modifies the original
}

// ──────────────────────────────────────────────────────────
// LAB 4.1 — JSON Marshal / Unmarshal helpers
// ──────────────────────────────────────────────────────────

// MarshalTransaction serializes a transaction to JSON bytes.
func MarshalTransaction(t *Transaction) ([]byte, error) {
	return json.Marshal(t)
}

// UnmarshalTransaction deserializes JSON bytes into a *Transaction.
func UnmarshalTransaction(data []byte) (*Transaction, error) {
	var t Transaction
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// MarshalTransactions serializes a slice of transactions to JSON.
func MarshalTransactions(txs []Transaction) ([]byte, error) {
	return json.MarshalIndent(txs, "", "  ")
}

// UnmarshalTransactions deserializes JSON bytes into a slice of transactions.
func UnmarshalTransactions(data []byte) ([]Transaction, error) {
	var txs []Transaction
	if err := json.Unmarshal(data, &txs); err != nil {
		return nil, err
	}
	return txs, nil
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

// Budget represents a monthly spending target for a category.
type Budget struct {
	ID       int64    `json:"id"`
	UserID   int64    `json:"user_id"`
	Category Category `json:"category"`
	Amount   Rupees   `json:"amount"`
	Month    string   `json:"month"` // YYYY-MM
}

// CategoryBreakdown holds spending info for a single category.
type CategoryBreakdown struct {
	Category Category
	Amount   Rupees
	Target   Rupees // Budgeted amount
	Percent  float64
}

// DashboardData holds all processed metrics for the dashboard view.
type DashboardData struct {
	MonthlyIncome   Rupees
	TotalExpenses   Rupees
	Savings         Rupees
	ExpenseRatio    float64
	SavingsRatio    float64
	Breakdown       []CategoryBreakdown
	SmartInsight    string
	DailyTip        string
	HasMajorIssues  bool
	UsingBudget     bool // True if calculations are based on budget rather than income
}
