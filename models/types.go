package models

import (
	"fmt"
	"strings"
)

// ──────────────────────────────────────────────────────────
// UNIT 1 — Custom Types and Conversion (not casting)
// ──────────────────────────────────────────────────────────

// Rupees is a custom type wrapping float64 for currency amounts.
// UNIT 1: Creating your own type.
type Rupees float64

// String uses fmt.Sprintf — UNIT 1: fmt package usage.
// UNIT 4: Value receiver (method set on T).
func (r Rupees) String() string {
	return fmt.Sprintf("%.2f", float64(r))
}

// ToFloat64 demonstrates type conversion (not casting).
// UNIT 1: Conversion — Go does not have casting, only explicit conversion.
func (r Rupees) ToFloat64() float64 {
	return float64(r) // conversion, NOT casting
}

// ToInt truncates to integer — another conversion example.
func (r Rupees) ToInt() int {
	return int(r) // conversion from float64-based type to int
}

// Add returns the sum as a new Rupees value (value receiver — immutable).
func (r Rupees) Add(other Rupees) Rupees {
	return r + other
}

// Category represents a transaction category.
// UNIT 1: Custom type wrapping string.
type Category string

// ──────────────────────────────────────────────────────────
// UNIT 2 — Array (fixed size) for valid categories
// ──────────────────────────────────────────────────────────

// UNIT 1: var keyword for package-level declaration.
// UNIT 2: Array — fixed-size collection. Unlike slices, the length is part of the type.
var ValidCategories = [11]Category{
	CategoryFood,
	CategoryTravel,
	CategoryShopping,
	CategoryBills,
	CategoryHealth,
	CategoryEducation,
	CategoryEntertainment,
	CategorySalary,
	CategoryFreelance,
	CategoryInvestment,
	CategoryOther,
}

const (
	CategoryFood          Category = "Food"
	CategoryTravel        Category = "Travel"
	CategoryShopping      Category = "Shopping"
	CategoryBills         Category = "Bills"
	CategoryHealth        Category = "Health"
	CategoryEducation     Category = "Education"
	CategoryEntertainment Category = "Entertainment"
	CategorySalary        Category = "Salary"
	CategoryFreelance     Category = "Freelance"
	CategoryInvestment    Category = "Investment"
	CategoryOther         Category = "Other"
)

// IsCategoryValid checks whether a category exists in the fixed ValidCategories array.
// UNIT 2: Array — iterating with for-range.
// UNIT 1: Control flow — loop + conditional.
func IsCategoryValid(c Category) bool {
	for _, valid := range ValidCategories {
		if valid == c {
			return true
		}
	}
	return false
}

// TransactionType indicates income or expense.
// UNIT 1: Custom type wrapping string.
type TransactionType string

const (
	Income  TransactionType = "income"
	Expense TransactionType = "expense"
)

// ──────────────────────────────────────────────────────────
// UNIT 1 — Zero Values demonstration
// ──────────────────────────────────────────────────────────

// ZeroValueDemo holds fields to show Go's zero values.
// UNIT 1: Every type has a zero value in Go.
type ZeroValueDemo struct {
	IntVal    int     // zero value: 0
	FloatVal  float64 // zero value: 0.0
	StringVal string  // zero value: ""
	BoolVal   bool    // zero value: false
	RupeesVal Rupees  // zero value: 0.0 (based on float64)
}

// ShowZeroValues returns a formatted string of all zero values.
// UNIT 1: Demonstrates zero values + fmt package.
func ShowZeroValues() string {
	// UNIT 1: var keyword — declares with zero values (no explicit initialization).
	var demo ZeroValueDemo
	return fmt.Sprintf("int=%d float=%.1f string=%q bool=%v rupees=%s",
		demo.IntVal, demo.FloatVal, demo.StringVal, demo.BoolVal, demo.RupeesVal)
}

// ──────────────────────────────────────────────────────────
// UNIT 2 — Embedded Struct
// ──────────────────────────────────────────────────────────

// Metadata holds common audit fields.
// UNIT 2: This struct is embedded into TransactionWithMeta.
type Metadata struct {
	CreatedBy string
	Tags      []string // UNIT 2: Slice inside struct
}

// HasTag checks whether a tag exists.
// UNIT 2: for-range over slice field.
func (m Metadata) HasTag(tag string) bool {
	for _, t := range m.Tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}

// TransactionWithMeta demonstrates embedded struct.
// UNIT 2: Embedded struct — Metadata fields/methods are promoted.
type TransactionWithMeta struct {
	Transaction // UNIT 2: Embedded struct (promoted fields)
	Metadata    // UNIT 2: Another embedded struct
}

// ──────────────────────────────────────────────────────────
// UNIT 2 — Slice operations: slicing, append, delete, make, multi-dimensional
// ──────────────────────────────────────────────────────────

// FilterTransactions returns a filtered subset of transactions.
// UNIT 2: Slice — make, append, for-range.
// UNIT 3: Callback — accepts a function parameter for filtering.
func FilterTransactions(txs []Transaction, predicate func(Transaction) bool) []Transaction {
	// UNIT 2: make — pre-allocate a slice with length 0 and estimated capacity.
	result := make([]Transaction, 0, len(txs))
	for _, tx := range txs {
		// UNIT 3: Callback — predicate is called for each transaction.
		if predicate(tx) {
			// UNIT 2: append — grow the slice.
			result = append(result, tx)
		}
	}
	return result
}

// DeleteTransactionFromSlice removes a transaction at index i from a slice.
// UNIT 2: Delete from slice using append + slicing.
func DeleteTransactionFromSlice(txs []Transaction, i int) []Transaction {
	if i < 0 || i >= len(txs) {
		return txs
	}
	// UNIT 2: Slicing a slice — txs[:i] and txs[i+1:]
	// UNIT 3: Unfurling a slice with ... (variadic spread)
	return append(txs[:i], txs[i+1:]...)
}

// SliceFirstN returns the first n elements of a slice.
// UNIT 2: Slicing a slice.
func SliceFirstN(txs []Transaction, n int) []Transaction {
	if n > len(txs) {
		n = len(txs)
	}
	// UNIT 2: Slicing — creates a sub-slice without copying.
	return txs[:n]
}

// GroupByCategory groups transactions into a multi-dimensional structure.
// UNIT 2: Multi-dimensional slice (slice of slices) and Map with delete.
func GroupByCategory(txs []Transaction) map[Category][]Transaction {
	// UNIT 2: Map — create, add elements.
	groups := make(map[Category][]Transaction)
	for _, tx := range txs {
		// UNIT 2: Map — add element; append to slice value.
		groups[tx.Category] = append(groups[tx.Category], tx)
	}
	return groups
}

// PurgeCategoryFromMap removes a category key from a map.
// UNIT 2: Map — delete.
func PurgeCategoryFromMap(m map[Category][]Transaction, cat Category) {
	delete(m, cat) // UNIT 2: built-in delete for maps
}

// BuildCategorySummary builds a 2D slice: each row is [category, totalAmount].
// UNIT 2: Multi-dimensional slice ([][]string).
func BuildCategorySummary(txs []Transaction) [][]string {
	totals := make(map[Category]Rupees)
	for _, tx := range txs {
		totals[tx.Category] += tx.Amount
	}
	// UNIT 2: Multi-dimensional slice — composite literal.
	var rows [][]string
	for cat, amt := range totals {
		rows = append(rows, []string{string(cat), amt.String()})
	}
	return rows
}

// ──────────────────────────────────────────────────────────
// UNIT 2 — Anonymous Struct
// ──────────────────────────────────────────────────────────

// ParsedDateRange returns start/end dates as an anonymous struct.
// UNIT 2: Anonymous struct — declared and used inline without a named type.
func ParsedDateRange(start, end string) struct {
	Start string
	End   string
	Valid bool
} {
	valid := start != "" && end != "" && start <= end
	// UNIT 2: Anonymous struct — composite literal.
	return struct {
		Start string
		End   string
		Valid bool
	}{
		Start: start,
		End:   end,
		Valid: valid,
	}
}
