package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// ══════════════════════════════════════════════════════════
// UNIT 1 — Zero Values, Custom Types, Conversion
// ══════════════════════════════════════════════════════════

// TestShowZeroValues verifies Go's zero-value semantics.
// UNIT 1: Zero values — every type has a default zero value.
func TestShowZeroValues(t *testing.T) {
	result := ShowZeroValues()
	// UNIT 4: Table-driven test — checking multiple expected substrings.
	expected := []string{"int=0", "float=0.0", `string=""`, "bool=false", "rupees=0.00"}
	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("ShowZeroValues() = %q, missing %q", result, exp)
		}
	}
}

// TestRupeesConversion tests type conversion from Rupees to float64 and int.
// UNIT 1: Conversion (not casting) — explicit type conversion.
func TestRupeesConversion(t *testing.T) {
	r := Rupees(99.99)
	// UNIT 1: ToFloat64 — conversion from custom type to float64.
	if r.ToFloat64() != 99.99 {
		t.Errorf("ToFloat64(): got %v, want 99.99", r.ToFloat64())
	}
	// UNIT 1: ToInt — conversion truncates to int.
	if r.ToInt() != 99 {
		t.Errorf("ToInt(): got %d, want 99", r.ToInt())
	}
}

// TestRupeesAdd tests the Rupees.Add value receiver.
func TestRupeesAdd(t *testing.T) {
	a := Rupees(100.50)
	b := Rupees(200.25)
	sum := a.Add(b)
	if sum != 300.75 {
		t.Errorf("Add(): got %s, want 300.75", sum)
	}
}

// ══════════════════════════════════════════════════════════
// UNIT 2 — Array (fixed-size)
// ══════════════════════════════════════════════════════════

// TestValidCategoriesArray tests the fixed-size ValidCategories array.
// UNIT 2: Array — length is part of the type.
func TestValidCategoriesArray(t *testing.T) {
	// UNIT 2: Array — len() works on arrays.
	if len(ValidCategories) != 11 {
		t.Errorf("expected 11 valid categories, got %d", len(ValidCategories))
	}
}

// TestIsCategoryValid tests category validation using array iteration.
// UNIT 2: Array — for-range over fixed array.
func TestIsCategoryValid(t *testing.T) {
	tests := []struct {
		cat  Category
		want bool
	}{
		{CategoryFood, true},
		{CategorySalary, true},
		{Category("InvalidCat"), false},
		{Category(""), false},
	}
	for _, tc := range tests {
		t.Run(string(tc.cat), func(t *testing.T) {
			if got := IsCategoryValid(tc.cat); got != tc.want {
				t.Errorf("IsCategoryValid(%q) = %v, want %v", tc.cat, got, tc.want)
			}
		})
	}
}

// ══════════════════════════════════════════════════════════
// UNIT 2 — Slice: slicing, delete, make, multi-dimensional
// ══════════════════════════════════════════════════════════

// TestFilterTransactions tests callback-based slice filtering.
// UNIT 2: Slice — make, append. UNIT 3: Callback.
func TestFilterTransactions(t *testing.T) {
	txs := []Transaction{
		{Type: Expense, Amount: 50, Description: "A"},
		{Type: Income, Amount: 5000, Description: "B"},
		{Type: Expense, Amount: 200, Description: "C"},
	}

	// UNIT 3: Anonymous function passed as callback.
	expenses := FilterTransactions(txs, func(tx Transaction) bool {
		return tx.IsExpense()
	})
	if len(expenses) != 2 {
		t.Errorf("expected 2 expenses, got %d", len(expenses))
	}
}

// TestDeleteTransactionFromSlice tests slice element removal.
// UNIT 2: Delete from slice — using append + slicing.
func TestDeleteTransactionFromSlice(t *testing.T) {
	txs := []Transaction{
		{ID: 1, Description: "A"},
		{ID: 2, Description: "B"},
		{ID: 3, Description: "C"},
	}

	result := DeleteTransactionFromSlice(txs, 1) // remove "B"
	if len(result) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(result))
	}
	if result[0].Description != "A" || result[1].Description != "C" {
		t.Errorf("unexpected result: %v", result)
	}

	// UNIT 1: Control flow — boundary cases.
	result2 := DeleteTransactionFromSlice(txs, -1) // invalid index
	if len(result2) != len(txs) {
		t.Errorf("negative index should return original slice")
	}
}

// TestSliceFirstN tests slicing a slice.
// UNIT 2: Slicing a slice.
func TestSliceFirstN(t *testing.T) {
	txs := []Transaction{
		{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5},
	}
	first3 := SliceFirstN(txs, 3)
	if len(first3) != 3 {
		t.Errorf("expected 3 transactions, got %d", len(first3))
	}
	// UNIT 2: Slicing — request more than length.
	all := SliceFirstN(txs, 100)
	if len(all) != 5 {
		t.Errorf("expected all 5, got %d", len(all))
	}
}

// TestGroupByCategory tests multi-dimensional slice (map of slices).
// UNIT 2: Map + slice of slices.
func TestGroupByCategory(t *testing.T) {
	txs := []Transaction{
		{Category: CategoryFood, Description: "A"},
		{Category: CategoryFood, Description: "B"},
		{Category: CategoryTravel, Description: "C"},
	}
	groups := GroupByCategory(txs)

	if len(groups[CategoryFood]) != 2 {
		t.Errorf("expected 2 food txs, got %d", len(groups[CategoryFood]))
	}
	if len(groups[CategoryTravel]) != 1 {
		t.Errorf("expected 1 travel tx, got %d", len(groups[CategoryTravel]))
	}
}

// TestPurgeCategoryFromMap tests map delete.
// UNIT 2: Map — delete built-in.
func TestPurgeCategoryFromMap(t *testing.T) {
	m := map[Category][]Transaction{
		CategoryFood:   {{ID: 1}},
		CategoryTravel: {{ID: 2}},
	}
	PurgeCategoryFromMap(m, CategoryFood)
	if _, exists := m[CategoryFood]; exists {
		t.Error("CategoryFood should have been deleted from map")
	}
	if len(m) != 1 {
		t.Errorf("expected 1 entry remaining, got %d", len(m))
	}
}

// TestBuildCategorySummary tests the 2D slice builder.
// UNIT 2: Multi-dimensional slice ([][]string).
func TestBuildCategorySummary(t *testing.T) {
	txs := []Transaction{
		{Category: CategoryFood, Amount: 100},
		{Category: CategoryFood, Amount: 200},
		{Category: CategoryTravel, Amount: 50},
	}
	rows := BuildCategorySummary(txs)
	if len(rows) != 2 {
		t.Errorf("expected 2 rows (2 categories), got %d", len(rows))
	}
	// UNIT 2: Multi-dimensional — each row is []string with 2 elements.
	for _, row := range rows {
		if len(row) != 2 {
			t.Errorf("each row should have 2 columns, got %d", len(row))
		}
	}
}

// ══════════════════════════════════════════════════════════
// UNIT 2 — Embedded struct + Anonymous struct
// ══════════════════════════════════════════════════════════

// TestEmbeddedStruct tests promoted fields and methods from Metadata.
// UNIT 2: Embedded struct — Metadata is embedded in TransactionWithMeta.
func TestEmbeddedStruct(t *testing.T) {
	tm := TransactionWithMeta{
		Transaction: Transaction{
			ID: 1, Type: Expense, Amount: 100, Description: "Lunch",
			Date: "2026-01-01", Category: CategoryFood,
		},
		Metadata: Metadata{
			CreatedBy: "alice",
			Tags:      []string{"work", "daily"},
		},
	}

	// UNIT 2: Promoted fields — direct access without nesting.
	if tm.Description != "Lunch" {
		t.Errorf("promoted field Description: got %s, want Lunch", tm.Description)
	}
	if tm.CreatedBy != "alice" {
		t.Errorf("promoted field CreatedBy: got %s, want alice", tm.CreatedBy)
	}

	// UNIT 2: Promoted method — HasTag from Metadata.
	if !tm.HasTag("work") {
		t.Error("HasTag(\"work\") should return true")
	}
	if tm.HasTag("personal") {
		t.Error("HasTag(\"personal\") should return false")
	}
}

// TestAnonymousStruct tests inline struct declarations.
// UNIT 2: Anonymous struct — declared without a named type.
func TestAnonymousStruct(t *testing.T) {
	result := ParsedDateRange("2026-01-01", "2026-12-31")
	if !result.Valid {
		t.Error("expected valid date range")
	}
	if result.Start != "2026-01-01" {
		t.Errorf("Start: got %s, want 2026-01-01", result.Start)
	}

	// Invalid range
	invalid := ParsedDateRange("2026-12-31", "2026-01-01")
	if invalid.Valid {
		t.Error("reversed dates should be invalid")
	}

	// Empty
	empty := ParsedDateRange("", "2026-01-01")
	if empty.Valid {
		t.Error("empty start should be invalid")
	}
}

// ══════════════════════════════════════════════════════════
// UNIT 3 — Error handling: errors.Is, errors.As, Unwrap, wrapping
// ══════════════════════════════════════════════════════════

// TestSentinelErrors tests errors.Is with sentinel errors.
// UNIT 3: Checking errors — errors.Is().
func TestSentinelErrors(t *testing.T) {
	if !errors.Is(ErrNotFound, ErrNotFound) {
		t.Error("ErrNotFound should match itself")
	}
	if errors.Is(ErrNotFound, ErrDuplicateUser) {
		t.Error("ErrNotFound should not match ErrDuplicateUser")
	}
}

// TestWrapDBError tests error wrapping with DatabaseError.
// UNIT 3: Error wrapping — Unwrap() enables errors.Is() chaining.
func TestWrapDBError(t *testing.T) {
	wrapped := WrapDBError("select", ErrNotFound)
	if wrapped == nil {
		t.Fatal("expected non-nil error")
	}

	// UNIT 3: errors.Is() walks the chain via Unwrap().
	if !errors.Is(wrapped, ErrNotFound) {
		t.Error("wrapped error should match ErrNotFound via errors.Is")
	}

	// UNIT 3: errors.As() extracts typed error.
	var dbErr *DatabaseError
	if !errors.As(wrapped, &dbErr) {
		t.Error("should extract *DatabaseError via errors.As")
	}
	if dbErr.Operation != "select" {
		t.Errorf("Operation: got %q, want %q", dbErr.Operation, "select")
	}
}

// TestWrapDBErrorNil tests that wrapping nil returns nil.
func TestWrapDBErrorNil(t *testing.T) {
	if WrapDBError("op", nil) != nil {
		t.Error("wrapping nil should return nil")
	}
}

// TestIsNotFound tests the convenience helper.
func TestIsNotFound(t *testing.T) {
	if !IsNotFound(ErrNotFound) {
		t.Error("IsNotFound should return true for ErrNotFound")
	}
	if IsNotFound(fmt.Errorf("other error")) {
		t.Error("IsNotFound should return false for other errors")
	}
}

// TestAsValidationError tests typed error extraction.
// UNIT 3: errors.As — extracts *ValidationError from chain.
func TestAsValidationError(t *testing.T) {
	err := &ValidationError{Field: "amount", Msg: "must be positive"}
	ve, ok := AsValidationError(err)
	if !ok {
		t.Fatal("expected AsValidationError to return true")
	}
	if ve.Field != "amount" {
		t.Errorf("Field: got %q, want %q", ve.Field, "amount")
	}
}

// TestValidateAll tests variadic batch validation.
// UNIT 3: Variadic parameter — txs ...*Transaction.
func TestValidateAll(t *testing.T) {
	good := NewTransaction(Expense, 100, CategoryFood, "Lunch", "2026-01-01")
	bad1 := &Transaction{Type: "invalid", Amount: 100, Description: "x", Date: "2026-01-01"}
	bad2 := &Transaction{Type: Expense, Amount: -1, Description: "x", Date: "2026-01-01"}

	// UNIT 3: Unfurling — passing multiple pointers to variadic function.
	errs := ValidateAll(good, bad1, bad2)
	if len(errs) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errs))
	}

	// All valid
	errs2 := ValidateAll(good)
	if len(errs2) != 0 {
		t.Errorf("expected 0 errors for valid tx, got %d", len(errs2))
	}
}

// ══════════════════════════════════════════════════════════
// LAB 4  — Pointer / Value Receiver & Pass-by-value Tests
// ══════════════════════════════════════════════════════════

// TestNewTransaction verifies the factory function returns a valid *Transaction.
func TestNewTransaction(t *testing.T) {
	tx := NewTransaction(Expense, 500, CategoryFood, "Lunch", "2026-02-26")

	if tx == nil {
		t.Fatal("NewTransaction returned nil")
	}
	if tx.Type != Expense {
		t.Errorf("expected type Expense, got %s", tx.Type)
	}
	if tx.Amount != 500 {
		t.Errorf("expected amount 500, got %s", tx.Amount)
	}
	if tx.Category != CategoryFood {
		t.Errorf("expected category Food, got %s", tx.Category)
	}
	if tx.Description != "Lunch" {
		t.Errorf("expected description Lunch, got %s", tx.Description)
	}
	if tx.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

// TestCloneTransaction verifies deep copy semantics.
func TestCloneTransaction(t *testing.T) {
	original := NewTransaction(Income, 10000, CategorySalary, "Salary", "2026-01-01")
	original.ID = 42

	cloned := CloneTransaction(original)

	// values should match
	if cloned.ID != original.ID || cloned.Amount != original.Amount {
		t.Error("cloned transaction should match original values")
	}

	// mutating clone must not affect original
	cloned.Amount = 99999
	if original.Amount == cloned.Amount {
		t.Error("modifying clone should NOT affect original (deep copy)")
	}
}

// TestValueReceiverString tests the value-receiver String() method.
func TestValueReceiverString(t *testing.T) {
	tx := Transaction{
		ID:          1,
		Type:        Expense,
		Amount:      250,
		Category:    CategoryFood,
		Description: "Lunch",
		Date:        "2026-02-26",
	}

	s := tx.String()
	if !strings.Contains(s, "Lunch") {
		t.Errorf("String() should contain description, got: %s", s)
	}
	if !strings.Contains(s, "250.00") {
		t.Errorf("String() should contain formatted amount, got: %s", s)
	}
}

// TestValueReceiverSummary tests Summary() value receiver.
func TestValueReceiverSummary(t *testing.T) {
	tx := Transaction{Description: "Coffee", Amount: 50, Category: CategoryFood}
	got := tx.Summary()
	if !strings.Contains(got, "Coffee") || !strings.Contains(got, "50.00") {
		t.Errorf("unexpected Summary: %s", got)
	}
}

// TestIsExpenseIsIncome tests boolean helpers.
func TestIsExpenseIsIncome(t *testing.T) {
	exp := Transaction{Type: Expense}
	inc := Transaction{Type: Income}

	if !exp.IsExpense() {
		t.Error("expected IsExpense() = true for expense")
	}
	if exp.IsIncome() {
		t.Error("expected IsIncome() = false for expense")
	}
	if !inc.IsIncome() {
		t.Error("expected IsIncome() = true for income")
	}
}

// TestPointerReceiverSetAmount tests pointer receiver mutation.
func TestPointerReceiverSetAmount(t *testing.T) {
	tx := NewTransaction(Expense, 100, CategoryFood, "Snack", "2026-02-26")
	tx.SetAmount(200)

	if tx.Amount != 200 {
		t.Errorf("expected 200 after SetAmount, got %s", tx.Amount)
	}
}

// TestPointerReceiverSetCategory tests pointer receiver mutation.
func TestPointerReceiverSetCategory(t *testing.T) {
	tx := NewTransaction(Expense, 100, CategoryOther, "Bus ticket", "2026-02-26")
	tx.SetCategory(CategoryTravel)

	if tx.Category != CategoryTravel {
		t.Errorf("expected Travel after SetCategory, got %s", tx.Category)
	}
}

// TestApplyDiscount tests discount calculation via pointer receiver.
func TestApplyDiscount(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		discount float64
		want     Rupees
	}{
		{"10% off 1000", 1000, 10, 900},
		{"50% off 500", 500, 50, 250},
		{"0% off 100", 100, 0, 100},     // 0 discount => no change
		{"110% off 100", 100, 110, 100}, // >100 => no change
		{"negative", 100, -5, 100},      // negative => no change
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tx := NewTransaction(Expense, tc.amount, CategoryShopping, "Item", "2026-01-01")
			tx.ApplyDiscount(tc.discount)
			if tx.Amount != tc.want {
				t.Errorf("ApplyDiscount(%v): got %s, want %s", tc.discount, tx.Amount, tc.want)
			}
		})
	}
}

// TestPassByValueVsPointer explicitly demonstrates the difference.
func TestPassByValueVsPointer(t *testing.T) {
	original := NewTransaction(Expense, 100, CategoryOther, "Test", "2026-01-01")

	// Pass by value — original should NOT change
	_ = ModifyByValue(*original, 9999)
	if original.Amount != 100 {
		t.Errorf("pass-by-value should NOT modify original, got amount=%s", original.Amount)
	}

	// Pass by pointer — original SHOULD change
	ModifyByPointer(original, 9999)
	if original.Amount != 9999 {
		t.Errorf("pass-by-pointer SHOULD modify original, got amount=%s", original.Amount)
	}
}

// TestModifyByValueReturnsModifiedCopy checks the returned copy has the new value.
func TestModifyByValueReturnsModifiedCopy(t *testing.T) {
	original := Transaction{Amount: 100, Type: Expense, Description: "x", Date: "2026-01-01"}
	modified := ModifyByValue(original, 200)

	if modified.Amount != 200 {
		t.Errorf("returned copy should have 200, got %s", modified.Amount)
	}
	if original.Amount != 100 {
		t.Errorf("original should remain 100, got %s", original.Amount)
	}
}

// TestEditTransactionFields tests the pointer-based bulk edit helper.
func TestEditTransactionFields(t *testing.T) {
	tx := NewTransaction(Expense, 100, CategoryFood, "Lunch", "2026-01-01")

	EditTransactionFields(tx, Income, 5000, CategorySalary, "Monthly pay", "2026-02-01", "first salary")

	if tx.Type != Income {
		t.Errorf("type: got %s, want income", tx.Type)
	}
	if tx.Amount != 5000 {
		t.Errorf("amount: got %s, want 5000", tx.Amount)
	}
	if tx.Category != CategorySalary {
		t.Errorf("category: got %s, want Salary", tx.Category)
	}
	if tx.Description != "Monthly pay" {
		t.Errorf("description: got %s, want Monthly pay", tx.Description)
	}
	if tx.Date != "2026-02-01" {
		t.Errorf("date: got %s, want 2026-02-01", tx.Date)
	}
	if tx.Notes != "first salary" {
		t.Errorf("notes: got %s, want first salary", tx.Notes)
	}
}

// TestEditTransactionFieldsPartial only updates non-empty fields.
func TestEditTransactionFieldsPartial(t *testing.T) {
	tx := NewTransaction(Expense, 100, CategoryFood, "Lunch", "2026-01-01")
	tx.Notes = "original note"

	// only change amount, leave everything else
	EditTransactionFields(tx, "", 999, "", "", "", "")

	if tx.Amount != 999 {
		t.Errorf("amount should be 999, got %s", tx.Amount)
	}
	if tx.Type != Expense {
		t.Errorf("type should remain expense, got %s", tx.Type)
	}
	if tx.Category != CategoryFood {
		t.Errorf("category should remain Food, got %s", tx.Category)
	}
	if tx.Description != "Lunch" {
		t.Errorf("description should remain Lunch, got %s", tx.Description)
	}
	if tx.Notes != "original note" {
		t.Errorf("notes should remain original note, got %s", tx.Notes)
	}
}

// ══════════════════════════════════════════════════════════
// LAB 4  — Validation Tests
// ══════════════════════════════════════════════════════════

func TestValidateTransaction(t *testing.T) {
	tests := []struct {
		name    string
		tx      Transaction
		wantErr bool
		field   string
	}{
		{
			name:    "valid expense",
			tx:      Transaction{Type: Expense, Amount: 100, Description: "Lunch", Date: "2026-01-01"},
			wantErr: false,
		},
		{
			name:    "valid income",
			tx:      Transaction{Type: Income, Amount: 5000, Description: "Salary", Date: "2026-01-01"},
			wantErr: false,
		},
		{
			name:    "invalid type",
			tx:      Transaction{Type: "transfer", Amount: 100, Description: "x", Date: "2026-01-01"},
			wantErr: true,
			field:   "type",
		},
		{
			name:    "zero amount",
			tx:      Transaction{Type: Expense, Amount: 0, Description: "x", Date: "2026-01-01"},
			wantErr: true,
			field:   "amount",
		},
		{
			name:    "negative amount",
			tx:      Transaction{Type: Expense, Amount: -50, Description: "x", Date: "2026-01-01"},
			wantErr: true,
			field:   "amount",
		},
		{
			name:    "empty description",
			tx:      Transaction{Type: Expense, Amount: 100, Description: "", Date: "2026-01-01"},
			wantErr: true,
			field:   "description",
		},
		{
			name:    "empty date",
			tx:      Transaction{Type: Expense, Amount: 100, Description: "x", Date: ""},
			wantErr: true,
			field:   "date",
		},
		{
			name:    "empty type",
			tx:      Transaction{Type: "", Amount: 100, Description: "x", Date: "2026-01-01"},
			wantErr: true,
			field:   "type",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateTransaction(&tc.tx)
			if tc.wantErr && err == nil {
				t.Errorf("expected error for field %q, got nil", tc.field)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if tc.wantErr && err != nil {
				ve, ok := err.(*ValidationError)
				if ok && ve.Field != tc.field {
					t.Errorf("expected field %q, got %q", tc.field, ve.Field)
				}
			}
		})
	}
}

// ══════════════════════════════════════════════════════════
// LAB 4.1 — JSON Marshal / Unmarshal Table-Driven Tests
// ══════════════════════════════════════════════════════════

func TestMarshalUnmarshalTransaction(t *testing.T) {
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		tx   Transaction
	}{
		{
			name: "basic expense",
			tx: Transaction{
				ID: 1, Type: Expense, Amount: 250.50, Category: CategoryFood,
				Description: "Lunch", Date: "2026-02-26", Notes: "",
				CreatedAt: now, UpdatedAt: now,
			},
		},
		{
			name: "income with notes",
			tx: Transaction{
				ID: 2, Type: Income, Amount: 50000, Category: CategorySalary,
				Description: "Monthly Salary", Date: "2026-01-01", Notes: "January pay",
				CreatedAt: now, UpdatedAt: now,
			},
		},
		{
			name: "zero amount edge case",
			tx: Transaction{
				ID: 3, Type: Expense, Amount: 0, Category: CategoryOther,
				Description: "Free sample", Date: "2026-03-01",
				CreatedAt: now, UpdatedAt: now,
			},
		},
		{
			name: "large amount",
			tx: Transaction{
				ID: 4, Type: Income, Amount: 9999999.99, Category: CategoryInvestment,
				Description: "Stock dividend", Date: "2026-06-15",
				CreatedAt: now, UpdatedAt: now,
			},
		},
		{
			name: "special characters in description",
			tx: Transaction{
				ID: 5, Type: Expense, Amount: 100, Category: CategoryEntertainment,
				Description: `Movie "Inception" — 3D & IMAX`, Date: "2026-04-10",
				Notes:     "with friends <3 & family",
				CreatedAt: now, UpdatedAt: now,
			},
		},
		{
			name: "unicode description",
			tx: Transaction{
				ID: 6, Type: Expense, Amount: 500, Category: CategoryFood,
				Description: "খাবার ₹500", Date: "2026-05-05",
				CreatedAt: now, UpdatedAt: now,
			},
		},
		{
			name: "empty optional fields",
			tx: Transaction{
				ID: 7, Type: Expense, Amount: 10, Category: "",
				Description: "Misc", Date: "2026-01-01",
			},
		},
		{
			name: "fractional rupees",
			tx: Transaction{
				ID: 8, Type: Expense, Amount: 99.99, Category: CategoryShopping,
				Description: "Online purchase", Date: "2026-07-07",
				CreatedAt: now, UpdatedAt: now,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := MarshalTransaction(&tc.tx)
			if err != nil {
				t.Fatalf("MarshalTransaction() error: %v", err)
			}

			got, err := UnmarshalTransaction(data)
			if err != nil {
				t.Fatalf("UnmarshalTransaction() error: %v", err)
			}

			// Compare key fields (time precision may vary with JSON)
			if got.ID != tc.tx.ID {
				t.Errorf("ID: got %d, want %d", got.ID, tc.tx.ID)
			}
			if got.Type != tc.tx.Type {
				t.Errorf("Type: got %s, want %s", got.Type, tc.tx.Type)
			}
			if got.Amount != tc.tx.Amount {
				t.Errorf("Amount: got %v, want %v", got.Amount, tc.tx.Amount)
			}
			if got.Category != tc.tx.Category {
				t.Errorf("Category: got %s, want %s", got.Category, tc.tx.Category)
			}
			if got.Description != tc.tx.Description {
				t.Errorf("Description: got %q, want %q", got.Description, tc.tx.Description)
			}
			if got.Date != tc.tx.Date {
				t.Errorf("Date: got %s, want %s", got.Date, tc.tx.Date)
			}
			if got.Notes != tc.tx.Notes {
				t.Errorf("Notes: got %q, want %q", got.Notes, tc.tx.Notes)
			}
		})
	}
}

// TestUnmarshalInvalidJSON tests error handling for bad JSON input.
func TestUnmarshalInvalidJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"invalid json", "{not json}"},
		{"wrong type (array)", "[]"},
		{"incomplete json", `{"id": 1, "type": `},
		{"null", "null"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := UnmarshalTransaction([]byte(tc.input))
			// null is valid JSON and produces a zero-value struct
			if tc.name == "null" {
				if err != nil {
					t.Errorf("null should unmarshal without error, got %v", err)
				}
				return
			}
			if err == nil && tc.name != "null" {
				t.Errorf("expected error for input %q, got nil (result=%+v)", tc.input, result)
			}
		})
	}
}

// TestMarshalUnmarshalTransactions tests bulk marshal/unmarshal.
func TestMarshalUnmarshalTransactions(t *testing.T) {
	txs := []Transaction{
		{ID: 1, Type: Expense, Amount: 100, Category: CategoryFood, Description: "A", Date: "2026-01-01"},
		{ID: 2, Type: Income, Amount: 5000, Category: CategorySalary, Description: "B", Date: "2026-01-02"},
		{ID: 3, Type: Expense, Amount: 200, Category: CategoryTravel, Description: "C", Date: "2026-01-03"},
	}

	data, err := MarshalTransactions(txs)
	if err != nil {
		t.Fatalf("MarshalTransactions error: %v", err)
	}

	got, err := UnmarshalTransactions(data)
	if err != nil {
		t.Fatalf("UnmarshalTransactions error: %v", err)
	}

	if len(got) != len(txs) {
		t.Fatalf("expected %d transactions, got %d", len(txs), len(got))
	}

	for i := range txs {
		if got[i].ID != txs[i].ID || got[i].Amount != txs[i].Amount {
			t.Errorf("transaction %d mismatch: got %+v, want %+v", i, got[i], txs[i])
		}
	}
}

// TestMarshalUnmarshalEmptySlice tests empty list edge case.
func TestMarshalUnmarshalEmptySlice(t *testing.T) {
	data, err := MarshalTransactions([]Transaction{})
	if err != nil {
		t.Fatalf("MarshalTransactions error: %v", err)
	}

	got, err := UnmarshalTransactions(data)
	if err != nil {
		t.Fatalf("UnmarshalTransactions error: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(got))
	}
}

// TestJSONFieldNames verifies JSON keys match expected names.
func TestJSONFieldNames(t *testing.T) {
	tx := Transaction{
		ID: 1, Type: Expense, Amount: 100, Category: CategoryFood,
		Description: "Test", Date: "2026-01-01", Notes: "n",
	}

	data, _ := json.Marshal(tx)
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}

	expectedKeys := []string{"id", "type", "amount", "category", "description", "date", "notes", "created_at", "updated_at"}
	for _, key := range expectedKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("expected JSON key %q not found", key)
		}
	}
}

// TestPasswordHashOmittedFromJSON verifies User.PasswordHash has json:"-".
func TestPasswordHashOmittedFromJSON(t *testing.T) {
	u := User{ID: 1, Username: "alice", PasswordHash: "secret123"}
	data, _ := json.Marshal(u)

	if strings.Contains(string(data), "secret123") {
		t.Error("PasswordHash should be omitted from JSON (json:\"-\" tag)")
	}
	if strings.Contains(string(data), "password_hash") {
		t.Error("password_hash key should not appear in JSON")
	}
}

// TestRupeesString tests the Rupees String() value receiver.
func TestRupeesString(t *testing.T) {
	tests := []struct {
		input Rupees
		want  string
	}{
		{100, "100.00"},
		{99.99, "99.99"},
		{0, "0.00"},
		{1000000, "1000000.00"},
		{0.1, "0.10"},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%v", tc.input), func(t *testing.T) {
			got := tc.input.String()
			if got != tc.want {
				t.Errorf("Rupees(%v).String() = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// ══════════════════════════════════════════════════════════
// LAB 4.1 — Benchmarks for large transaction sets
// ══════════════════════════════════════════════════════════

func generateTransactions(n int) []Transaction {
	txs := make([]Transaction, n)
	cats := []Category{CategoryFood, CategoryTravel, CategoryShopping, CategoryBills, CategorySalary}
	for i := 0; i < n; i++ {
		txs[i] = Transaction{
			ID:          int64(i + 1),
			Type:        Expense,
			Amount:      Rupees(float64(i)*10.5 + 1),
			Category:    cats[i%len(cats)],
			Description: fmt.Sprintf("Transaction-%d", i+1),
			Date:        fmt.Sprintf("2026-%02d-%02d", (i%12)+1, (i%28)+1),
			Notes:       "benchmark",
		}
	}
	return txs
}

// BenchmarkMarshalTransactions benchmarks JSON marshalling of large sets.
func BenchmarkMarshalTransactions(b *testing.B) {
	sizes := []int{10, 100, 1000, 5000}
	for _, size := range sizes {
		txs := generateTransactions(size)
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := MarshalTransactions(txs)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkUnmarshalTransactions benchmarks JSON unmarshalling of large sets.
func BenchmarkUnmarshalTransactions(b *testing.B) {
	sizes := []int{10, 100, 1000, 5000}
	for _, size := range sizes {
		txs := generateTransactions(size)
		data, _ := MarshalTransactions(txs)
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := UnmarshalTransactions(data)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkMarshalSingleTransaction benchmarks serialization of one transaction.
func BenchmarkMarshalSingleTransaction(b *testing.B) {
	tx := NewTransaction(Expense, 500, CategoryFood, "Benchmark item", "2026-01-01")
	for i := 0; i < b.N; i++ {
		_, _ = MarshalTransaction(tx)
	}
}

// BenchmarkUnmarshalSingleTransaction benchmarks deserialization of one transaction.
func BenchmarkUnmarshalSingleTransaction(b *testing.B) {
	tx := NewTransaction(Expense, 500, CategoryFood, "Benchmark item", "2026-01-01")
	data, _ := MarshalTransaction(tx)
	for i := 0; i < b.N; i++ {
		_, _ = UnmarshalTransaction(data)
	}
}

// BenchmarkNewTransaction benchmarks the factory function.
func BenchmarkNewTransaction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewTransaction(Expense, 100, CategoryFood, "Item", "2026-01-01")
	}
}

// BenchmarkCloneTransaction benchmarks the clone function.
func BenchmarkCloneTransaction(b *testing.B) {
	tx := NewTransaction(Expense, 100, CategoryFood, "Item", "2026-01-01")
	for i := 0; i < b.N; i++ {
		_ = CloneTransaction(tx)
	}
}
