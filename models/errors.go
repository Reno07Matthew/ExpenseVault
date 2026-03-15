package models

import (
	"errors"
	"fmt"
)

// ──────────────────────────────────────────────────────────
// UNIT 3 — Error handling: errors with info, checking errors
// ──────────────────────────────────────────────────────────

// Sentinel errors — UNIT 3: Checking errors with errors.Is().
var (
	ErrNotFound      = errors.New("record not found")
	ErrDuplicateUser = errors.New("username already exists")
	ErrUnauthorized  = errors.New("unauthorized")
)

// DatabaseError wraps DB errors with context.
// UNIT 3: Errors with info — custom error type carrying contextual fields.
type DatabaseError struct {
	Operation string
	Err       error
}

// Error satisfies the error interface.
// UNIT 3: Methods + Interfaces — implements the built-in error interface.
func (e *DatabaseError) Error() string {
	return fmt.Sprintf("database error during %s: %v", e.Operation, e.Err)
}

// Unwrap enables errors.Is / errors.As to see the wrapped error.
// UNIT 3: Error handling — error wrapping / unwrapping.
func (e *DatabaseError) Unwrap() error {
	return e.Err
}

// AuthError is returned when authentication fails.
// UNIT 3: Errors with info — captures the reason for auth failure.
type AuthError struct {
	Reason string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("auth error: %s", e.Reason)
}

// ValidationError represents validation failures.
// UNIT 3: Errors with info — carries Field and Msg for detailed feedback.
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Msg)
}

// WrapDBError wraps a raw error with database operation context.
// UNIT 3: Error handling — wrapping errors with additional info.
func WrapDBError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return &DatabaseError{Operation: operation, Err: err}
}

// IsNotFound checks if an error chain contains ErrNotFound.
// UNIT 3: Checking errors — errors.Is() walks the error chain.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// AsValidationError extracts a *ValidationError from the error chain.
// UNIT 3: Checking errors — errors.As() for typed error inspection.
func AsValidationError(err error) (*ValidationError, bool) {
	var ve *ValidationError
	ok := errors.As(err, &ve)
	return ve, ok
}

// ──────────────────────────────────────────────────────────
// UNIT 3 — Variadic parameter for batch validation
// ──────────────────────────────────────────────────────────

// ValidateTransaction checks basic transaction fields.
// UNIT 4: Accepts *Transaction (pointer) — avoids copying the struct.
func ValidateTransaction(t *Transaction) error {
	if t.Type != Income && t.Type != Expense {
		return &ValidationError{Field: "type", Msg: "must be income or expense"}
	}
	if t.Amount <= 0 {
		return &ValidationError{Field: "amount", Msg: "must be positive"}
	}
	if t.Description == "" {
		return &ValidationError{Field: "description", Msg: "required"}
	}
	if t.Date == "" {
		return &ValidationError{Field: "date", Msg: "required"}
	}
	return nil
}

// ValidateAll validates multiple transactions and returns all errors.
// UNIT 3: Variadic parameter — txs ...*Transaction.
func ValidateAll(txs ...*Transaction) []error {
	// UNIT 1: var keyword — errs starts as nil slice (zero value).
	var errs []error
	for _, t := range txs {
		if err := ValidateTransaction(t); err != nil {
			// UNIT 3: Error handling — wrapping with fmt.Errorf %w.
			errs = append(errs, fmt.Errorf("tx %q: %w", t.Description, err))
		}
	}
	return errs
}

// ──────────────────────────────────────────────────────────
// QuickSummary — existing function with concept labels
// ──────────────────────────────────────────────────────────

// QuickSummary calculates a quick overview.
// UNIT 4: Uses value-receiver IsIncome() and IsExpense() methods.
// UNIT 2: Map — make, add elements, for-range.
func QuickSummary(txs []Transaction) MonthlySummary {
	// UNIT 2: Map — make creates a map.
	s := MonthlySummary{ByCategory: make(map[Category]Rupees)}
	for _, tx := range txs {
		// UNIT 4: Value receiver methods — read-only, operate on copy.
		if tx.IsIncome() {
			s.Income += tx.Amount
		} else if tx.IsExpense() {
			s.Expenses += tx.Amount
		}
		// UNIT 2: Map — add element (key may or may not exist).
		s.ByCategory[tx.Category] += tx.Amount
	}
	s.Balance = s.Income - s.Expenses
	return s
}
