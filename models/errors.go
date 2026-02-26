package models

import "fmt"

// DatabaseError wraps DB errors with context.
type DatabaseError struct {
	Operation string
	Err       error
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("database error during %s: %v", e.Operation, e.Err)
}

// AuthError is returned when authentication fails.
type AuthError struct {
	Reason string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("auth error: %s", e.Reason)
}

// ValidationError represents validation failures.
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Msg)
}

// ValidateTransaction checks basic transaction fields.
func ValidateTransaction(t Transaction) error {
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

// QuickSummary calculates a quick overview.
func QuickSummary(txs []Transaction) MonthlySummary {
	s := MonthlySummary{ByCategory: make(map[Category]Rupees)}
	for _, tx := range txs {
		if tx.Type == Income {
			s.Income += tx.Amount
		} else {
			s.Expenses += tx.Amount
		}
		s.ByCategory[tx.Category] += tx.Amount
	}
	s.Balance = s.Income - s.Expenses
	return s
}
