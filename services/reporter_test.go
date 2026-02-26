package services

import (
	"fmt"
	"strings"
	"testing"

	"expenseVault/models"
)

func sampleTransactions() []models.Transaction {
	return []models.Transaction{
		{ID: 1, Type: models.Expense, Amount: 200, Category: models.CategoryFood, Description: "Lunch", Date: "2026-01-15"},
		{ID: 2, Type: models.Expense, Amount: 150, Category: models.CategoryTravel, Description: "Bus", Date: "2026-01-20"},
		{ID: 3, Type: models.Income, Amount: 5000, Category: models.CategorySalary, Description: "Salary", Date: "2026-02-01"},
		{ID: 4, Type: models.Expense, Amount: 300, Category: models.CategoryFood, Description: "Dinner", Date: "2026-02-10"},
		{ID: 5, Type: models.Expense, Amount: 1000, Category: models.CategoryShopping, Description: "Clothes", Date: "2025-12-25"},
	}
}

// ══════════════════════════════════════════════════════════
// UNIT 3 — Interface / Polymorphism Tests
// ══════════════════════════════════════════════════════════

func TestMonthlyReporter(t *testing.T) {
	r := &MonthlyReporter{}
	txs := sampleTransactions()
	result := r.Generate(txs)

	tests := []struct {
		name     string
		contains string
	}{
		{"has 2026-01", "2026-01"},
		{"has 2026-02", "2026-02"},
		{"has 2025-12", "2025-12"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(result, tc.contains) {
				t.Errorf("expected report to contain %q, got:\n%s", tc.contains, result)
			}
		})
	}
}

func TestCategoryReporter(t *testing.T) {
	r := &CategoryReporter{}
	txs := sampleTransactions()
	result := r.Generate(txs)

	tests := []struct {
		name     string
		contains string
	}{
		{"has Food", "Food"},
		{"has Travel", "Travel"},
		{"has Salary", "Salary"},
		{"has Shopping", "Shopping"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(result, tc.contains) {
				t.Errorf("expected report to contain %q, got:\n%s", tc.contains, result)
			}
		})
	}
}

func TestYearlyReporter(t *testing.T) {
	r := &YearlyReporter{}
	txs := sampleTransactions()
	result := r.Generate(txs)

	tests := []struct {
		name     string
		contains string
	}{
		{"has 2026", "2026"},
		{"has 2025", "2025"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(result, tc.contains) {
				t.Errorf("expected report to contain %q, got:\n%s", tc.contains, result)
			}
		})
	}
}

func TestReporterEmptyInput(t *testing.T) {
	reporters := []struct {
		name string
		r    Reporter
	}{
		{"MonthlyReporter", &MonthlyReporter{}},
		{"CategoryReporter", &CategoryReporter{}},
		{"YearlyReporter", &YearlyReporter{}},
	}

	for _, tc := range reporters {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.r.Generate(nil)
			if result != "" {
				t.Errorf("expected empty string for nil input, got %q", result)
			}
		})
	}
}

// TestReporterInterface verifies all reporters satisfy the Reporter interface.
// UNIT 3: Interfaces & polymorphism — compile-time check.
func TestReporterInterface(t *testing.T) {
	var _ Reporter = &MonthlyReporter{}
	var _ Reporter = &CategoryReporter{}
	var _ Reporter = &YearlyReporter{}
}

// ══════════════════════════════════════════════════════════
// UNIT 3 — Variadic + Unfurling tests
// ══════════════════════════════════════════════════════════

// TestCombineReports tests variadic reporter parameter.
// UNIT 3: Variadic parameter — reporters ...Reporter.
func TestCombineReports(t *testing.T) {
	txs := sampleTransactions()

	// Call with multiple individual reporters (variadic).
	result := CombineReports(txs, &MonthlyReporter{}, &CategoryReporter{})
	if !strings.Contains(result, "2026-01") {
		t.Error("combined report should contain monthly data")
	}
	if !strings.Contains(result, "Food") {
		t.Error("combined report should contain category data")
	}
	if !strings.Contains(result, "---") {
		t.Error("combined report should contain separator")
	}
}

// TestRunAllReports tests unfurling a slice into variadic args.
// UNIT 3: Unfurling a slice — allReporters...
func TestRunAllReports(t *testing.T) {
	txs := sampleTransactions()
	result := RunAllReports(txs)

	// Should contain output from all three reporters.
	if !strings.Contains(result, "2026-01") {
		t.Error("RunAllReports should include monthly data")
	}
	if !strings.Contains(result, "Food") {
		t.Error("RunAllReports should include category data")
	}
	if !strings.Contains(result, "2025") || !strings.Contains(result, "2026") {
		t.Error("RunAllReports should include yearly data")
	}
}

// TestCombineReportsEmpty tests variadic with zero arguments.
func TestCombineReportsEmpty(t *testing.T) {
	result := CombineReports(sampleTransactions())
	if result != "" {
		t.Errorf("expected empty string with no reporters, got %q", result)
	}
}

// ══════════════════════════════════════════════════════════
// UNIT 3 — Function expression, Returning a function, Closure, Callback
// ══════════════════════════════════════════════════════════

// TestMakeAmountFilter tests closure-based filter generation.
// UNIT 3: Returning a function + Closure.
func TestMakeAmountFilter(t *testing.T) {
	filter := MakeAmountFilter(500)
	txs := sampleTransactions()

	filtered := ApplyFilter(txs, filter)
	for _, tx := range filtered {
		if tx.Amount < 500 {
			t.Errorf("filter should exclude amounts < 500, got %s", tx.Amount)
		}
	}
}

// TestMakeCategoryFilter tests category-based closure filter.
func TestMakeCategoryFilter(t *testing.T) {
	filter := MakeCategoryFilter(models.CategoryFood)
	txs := sampleTransactions()

	filtered := ApplyFilter(txs, filter)
	if len(filtered) != 2 {
		t.Errorf("expected 2 food transactions, got %d", len(filtered))
	}
}

// TestMakeTypeFilter tests type-based closure filter.
func TestMakeTypeFilter(t *testing.T) {
	filter := MakeTypeFilter(models.Income)
	txs := sampleTransactions()

	filtered := ApplyFilter(txs, filter)
	if len(filtered) != 1 {
		t.Errorf("expected 1 income transaction, got %d", len(filtered))
	}
}

// TestChainFilters tests composing multiple filters.
// UNIT 3: Variadic + Returning a function + Closure.
func TestChainFilters(t *testing.T) {
	txs := sampleTransactions()

	// Chain: expense AND amount >= 200
	combined := ChainFilters(
		MakeTypeFilter(models.Expense),
		MakeAmountFilter(200),
	)
	filtered := ApplyFilter(txs, combined)
	for _, tx := range filtered {
		if tx.Type != models.Expense {
			t.Error("chained filter should only return expenses")
		}
		if tx.Amount < 200 {
			t.Errorf("chained filter should have amount >= 200, got %s", tx.Amount)
		}
	}
}

// TestApplyFilterWithAnonymousFunc tests inline anonymous function as callback.
// UNIT 3: Anonymous function + Callback.
func TestApplyFilterWithAnonymousFunc(t *testing.T) {
	txs := sampleTransactions()

	// UNIT 3: Anonymous function — defined inline, no named type.
	result := ApplyFilter(txs, func(tx models.Transaction) bool {
		return strings.Contains(tx.Description, "Lu")
	})
	if len(result) != 1 || result[0].Description != "Lunch" {
		t.Errorf("expected 1 result (Lunch), got %d", len(result))
	}
}

// ══════════════════════════════════════════════════════════
// UNIT 3 — Recursion
// ══════════════════════════════════════════════════════════

// TestSumTransactionsRecursive tests recursive summing.
// UNIT 3: Recursion — function calls itself.
func TestSumTransactionsRecursive(t *testing.T) {
	txs := sampleTransactions()
	// Expected: 200+150+5000+300+1000 = 6650
	got := SumTransactionsRecursive(txs)
	if got != 6650 {
		t.Errorf("SumRecursive: got %s, want 6650.00", got)
	}
}

// TestSumTransactionsRecursiveEmpty tests base case.
func TestSumTransactionsRecursiveEmpty(t *testing.T) {
	got := SumTransactionsRecursive(nil)
	if got != 0 {
		t.Errorf("SumRecursive(nil): got %s, want 0.00", got)
	}
}

// ══════════════════════════════════════════════════════════
// UNIT 3 — Anonymous function (sort)
// ══════════════════════════════════════════════════════════

// TestSortByAmountDesc tests in-place sort with anonymous function.
// UNIT 3: Anonymous function passed to sort.Slice.
func TestSortByAmountDesc(t *testing.T) {
	txs := sampleTransactions()
	SortByAmountDesc(txs)

	for i := 1; i < len(txs); i++ {
		if txs[i].Amount > txs[i-1].Amount {
			t.Errorf("not sorted descending at index %d: %s > %s", i, txs[i].Amount, txs[i-1].Amount)
		}
	}
}

// TestSortByDate tests date-based sort.
func TestSortByDate(t *testing.T) {
	txs := sampleTransactions()
	SortByDate(txs)

	for i := 1; i < len(txs); i++ {
		if txs[i].Date < txs[i-1].Date {
			t.Errorf("not sorted ascending at index %d: %s < %s", i, txs[i].Date, txs[i-1].Date)
		}
	}
}

// ══════════════════════════════════════════════════════════
// Benchmarks
// ══════════════════════════════════════════════════════════

func BenchmarkMonthlyReporter(b *testing.B) {
	txs := generateBenchTxs(1000)
	r := &MonthlyReporter{}
	for i := 0; i < b.N; i++ {
		r.Generate(txs)
	}
}

func BenchmarkCategoryReporter(b *testing.B) {
	txs := generateBenchTxs(1000)
	r := &CategoryReporter{}
	for i := 0; i < b.N; i++ {
		r.Generate(txs)
	}
}

func BenchmarkYearlyReporter(b *testing.B) {
	txs := generateBenchTxs(1000)
	r := &YearlyReporter{}
	for i := 0; i < b.N; i++ {
		r.Generate(txs)
	}
}

// BenchmarkSumRecursive benchmarks recursive sum vs iterative approach.
func BenchmarkSumRecursive(b *testing.B) {
	sizes := []int{10, 100, 500}
	for _, n := range sizes {
		txs := generateBenchTxs(n)
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				SumTransactionsRecursive(txs)
			}
		})
	}
}

// BenchmarkChainFilters benchmarks chained filter application.
func BenchmarkChainFilters(b *testing.B) {
	txs := generateBenchTxs(1000)
	filter := ChainFilters(
		MakeTypeFilter(models.Expense),
		MakeAmountFilter(100),
	)
	for i := 0; i < b.N; i++ {
		ApplyFilter(txs, filter)
	}
}

func generateBenchTxs(n int) []models.Transaction {
	cats := []models.Category{models.CategoryFood, models.CategoryTravel, models.CategoryShopping}
	txs := make([]models.Transaction, n)
	for i := 0; i < n; i++ {
		txs[i] = models.Transaction{
			ID:          int64(i + 1),
			Type:        models.Expense,
			Amount:      models.Rupees(float64(i)*5.5 + 1),
			Category:    cats[i%len(cats)],
			Description: "Bench",
			Date:        "2026-01-15",
		}
	}
	return txs
}
