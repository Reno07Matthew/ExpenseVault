package services

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"

	"expenseVault/models"
)

// ──────────────────────────────────────────────────────────
// UNIT 3 — Interfaces & Polymorphism
// ──────────────────────────────────────────────────────────

// Reporter generates a report summary.
// UNIT 3: Interface — defines behaviour; any type with Generate(...) satisfies it.
// UNIT 4: All concrete reporters use pointer receivers (method sets on *T).
type Reporter interface {
	Generate(transactions []models.Transaction) string
}

// ──────────────────────────────────────────────────────────
// UNIT 3 — Methods (pointer receivers) + UNIT 4 — Method sets
// ──────────────────────────────────────────────────────────

// MonthlyReporter groups by month (YYYY-MM).
type MonthlyReporter struct{}

func (r *MonthlyReporter) Generate(transactions []models.Transaction) string {
	monthly := map[string]models.Rupees{}
	for _, t := range transactions {
		key := ""
		if len(t.Date) >= 7 {
			key = t.Date[:7]
		}
		monthly[key] += t.Amount
	}
	keys := make([]string, 0, len(monthly))
	for k := range monthly {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("%s: %s\n", k, monthly[k]))
	}
	return sb.String()
}

// CategoryReporter groups by category.
type CategoryReporter struct{}

func (r *CategoryReporter) Generate(transactions []models.Transaction) string {
	byCat := map[models.Category]models.Rupees{}
	for _, t := range transactions {
		byCat[t.Category] += t.Amount
	}
	cats := make([]string, 0, len(byCat))
	for c := range byCat {
		cats = append(cats, string(c))
	}
	sort.Strings(cats)

	var sb strings.Builder
	for _, c := range cats {
		amt := byCat[models.Category(c)]
		sb.WriteString(fmt.Sprintf("%s: %s\n", c, amt))
	}
	return sb.String()
}

// YearlyReporter groups by year.
type YearlyReporter struct{}

func (r *YearlyReporter) Generate(transactions []models.Transaction) string {
	yearly := map[string]models.Rupees{}
	for _, t := range transactions {
		key := ""
		if len(t.Date) >= 4 {
			key = t.Date[:4]
		}
		yearly[key] += t.Amount
	}
	keys := make([]string, 0, len(yearly))
	for k := range yearly {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("%s: %s\n", k, yearly[k]))
	}
	return sb.String()
}

// ──────────────────────────────────────────────────────────
// UNIT 3 — Variadic parameter + Unfurling a slice
// ──────────────────────────────────────────────────────────

// CombineReports runs multiple reporters and joins their output.
// UNIT 3: Variadic parameter — reporters ...Reporter accepts zero or more Reporters.
func CombineReports(transactions []models.Transaction, reporters ...Reporter) string {
	var sb strings.Builder
	for i, r := range reporters {
		if i > 0 {
			sb.WriteString("---\n")
		}
		sb.WriteString(r.Generate(transactions))
	}
	return sb.String()
}

// RunAllReports is a convenience function demonstrating unfurling.
// UNIT 3: Unfurling a slice — allReporters... spreads the slice into variadic args.
// RunAllReports runs all reporters in parallel using goroutines.
// UNIT 5: Concurrency — WaitGroup & Channels.
func RunAllReports(transactions []models.Transaction) string {
	allReporters := []Reporter{
		&MonthlyReporter{},
		&CategoryReporter{},
		&YearlyReporter{},
	}

	type result struct {
		index  int
		output string
	}
	resChan := make(chan result, len(allReporters))
	var wg sync.WaitGroup

	for i, r := range allReporters {
		wg.Add(1)
		go func(idx int, rep Reporter) {
			defer wg.Done()
			resChan <- result{idx, rep.Generate(transactions)}
		}(i, r)
	}

	// Wait in a separate goroutine so we can close the channel.
	go func() {
		wg.Wait()
		close(resChan)
	}()

	// Collect results and sort by index.
	results := make([]string, len(allReporters))
	for r := range resChan {
		results[r.index] = r.output
	}

	return strings.Join(results, "---\n")
}

// ──────────────────────────────────────────────────────────
// UNIT 3 — Function expression, Returning a function, Closure
// ──────────────────────────────────────────────────────────

// TransactionFilter is a function type used as callback.
// UNIT 3: Function expression — defining a named function type.
type TransactionFilter func(models.Transaction) bool

// MakeAmountFilter returns a function that filters by minimum amount.
// UNIT 3: Returning a function — the returned closure captures `min`.
// UNIT 3: Closure — the inner function closes over `min`.
func MakeAmountFilter(min models.Rupees) TransactionFilter {
	// UNIT 3: Closure — min is captured from the enclosing scope.
	return func(tx models.Transaction) bool {
		return tx.Amount >= min
	}
}

// MakeCategoryFilter returns a filter that matches a given category.
// UNIT 3: Returning a function + Closure.
func MakeCategoryFilter(cat models.Category) TransactionFilter {
	return func(tx models.Transaction) bool {
		return tx.Category == cat
	}
}

// MakeTypeFilter returns a filter by transaction type.
// UNIT 3: Returning a function + Closure.
func MakeTypeFilter(tt models.TransactionType) TransactionFilter {
	return func(tx models.Transaction) bool {
		return tx.Type == tt
	}
}

// ChainFilters combines multiple filters with AND logic.
// UNIT 3: Variadic parameter + Returning a function + Closure.
func ChainFilters(filters ...TransactionFilter) TransactionFilter {
	// UNIT 3: Closure — captures the `filters` slice.
	return func(tx models.Transaction) bool {
		for _, f := range filters {
			if !f(tx) {
				return false
			}
		}
		return true
	}
}

// ApplyFilter applies a callback filter to transactions in parallel.
// UNIT 5: Concurrency — Chunked processing with goroutines.
func ApplyFilter(txs []models.Transaction, filter TransactionFilter) []models.Transaction {
	if len(txs) < 100 { // Only parallelize if significant enough
		result := make([]models.Transaction, 0, len(txs))
		for _, tx := range txs {
			if filter(tx) {
				result = append(result, tx)
			}
		}
		return result
	}

	numCPU := runtime.NumCPU()
	if numCPU <= 0 {
		numCPU = 1
	}
	
	chunkSize := (len(txs) + numCPU - 1) / numCPU
	resChan := make(chan []models.Transaction, numCPU)
	var wg sync.WaitGroup

	for i := 0; i < len(txs); i += chunkSize {
		end := i + chunkSize
		if end > len(txs) {
			end = len(txs)
		}

		wg.Add(1)
		go func(chunk []models.Transaction) {
			defer wg.Done()
			filtered := make([]models.Transaction, 0, len(chunk))
			for _, tx := range chunk {
				if filter(tx) {
					filtered = append(filtered, tx)
				}
			}
			resChan <- filtered
		}(txs[i:end])
	}

	go func() {
		wg.Wait()
		close(resChan)
	}()

	var finalResult []models.Transaction
	for filtered := range resChan {
		finalResult = append(finalResult, filtered...)
	}
	return finalResult
}

// ──────────────────────────────────────────────────────────
// UNIT 3 — Recursion
// ──────────────────────────────────────────────────────────

// SumTransactionsRecursive recursively sums transaction amounts.
// UNIT 3: Recursion — function calls itself with a smaller slice.
func SumTransactionsRecursive(txs []models.Transaction) models.Rupees {
	// Base case — empty slice.
	if len(txs) == 0 {
		return 0
	}
	// Recursive case — first element + sum of rest.
	// UNIT 2: Slicing a slice — txs[1:] to get remainder.
	return txs[0].Amount + SumTransactionsRecursive(txs[1:])
}

// ──────────────────────────────────────────────────────────
// UNIT 3 — Anonymous function
// ──────────────────────────────────────────────────────────

// SortByAmountDesc sorts transactions in-place by amount descending.
// UNIT 3: Anonymous function — passed directly to sort.Slice.
func SortByAmountDesc(txs []models.Transaction) {
	// UNIT 3: Anonymous function — the func literal is defined inline.
	sort.Slice(txs, func(i, j int) bool {
		return txs[i].Amount > txs[j].Amount
	})
}

// SortByDate sorts transactions in-place by date ascending.
func SortByDate(txs []models.Transaction) {
	sort.Slice(txs, func(i, j int) bool {
		return txs[i].Date < txs[j].Date
	})
}
