package services

import (
	"fmt"
	"strings"
	"time"

	"expenseVault/models"
)

// ──────────────────────────────────────────────────────────
// ADVANCED FEATURE: Custom Query Language (CQL)
// Demonstrates: Lexer/parser, string processing, evaluation
// ──────────────────────────────────────────────────────────

// QueryFilter represents a parsed query filter.
type QueryFilter struct {
	Category    string
	AmountOp    string  // ">", "<", "=", ""
	AmountVal   float64
	DateRange   string  // "last-week", "last-month", or "YYYY-MM-DD"
	FuzzyText   string  // free text for fuzzy matching
}

// ParseQuery parses a CQL query string into a QueryFilter.
// Supported syntax: cat:food amt:>500 date:last-week free text
func ParseQuery(query string) QueryFilter {
	var filter QueryFilter
	tokens := strings.Fields(query)

	var freeText []string

	for _, token := range tokens {
		lower := strings.ToLower(token)

		switch {
		case strings.HasPrefix(lower, "cat:"):
			filter.Category = strings.TrimPrefix(lower, "cat:")

		case strings.HasPrefix(lower, "amt:"):
			amtStr := strings.TrimPrefix(lower, "amt:")
			if strings.HasPrefix(amtStr, ">") {
				filter.AmountOp = ">"
				fmt.Sscanf(amtStr[1:], "%f", &filter.AmountVal)
			} else if strings.HasPrefix(amtStr, "<") {
				filter.AmountOp = "<"
				fmt.Sscanf(amtStr[1:], "%f", &filter.AmountVal)
			} else {
				filter.AmountOp = "="
				fmt.Sscanf(amtStr, "%f", &filter.AmountVal)
			}

		case strings.HasPrefix(lower, "date:"):
			filter.DateRange = strings.TrimPrefix(lower, "date:")

		default:
			freeText = append(freeText, token)
		}
	}

	filter.FuzzyText = strings.Join(freeText, " ")
	return filter
}

// ApplyQueryFilter filters transactions based on a QueryFilter.
func ApplyQueryFilter(txs []models.Transaction, filter QueryFilter) []models.Transaction {
	var results []models.Transaction

	for _, tx := range txs {
		if !matchesFilter(tx, filter) {
			continue
		}
		results = append(results, tx)
	}

	return results
}

func matchesFilter(tx models.Transaction, f QueryFilter) bool {
	// Category filter.
	if f.Category != "" {
		if !strings.EqualFold(string(tx.Category), f.Category) {
			return false
		}
	}

	// Amount filter.
	if f.AmountOp != "" {
		amt := tx.Amount.ToFloat64()
		switch f.AmountOp {
		case ">":
			if amt <= f.AmountVal {
				return false
			}
		case "<":
			if amt >= f.AmountVal {
				return false
			}
		case "=":
			if amt != f.AmountVal {
				return false
			}
		}
	}

	// Date filter.
	if f.DateRange != "" {
		if !matchesDate(tx, f.DateRange) {
			return false
		}
	}

	// Fuzzy text filter (case-insensitive substring match).
	if f.FuzzyText != "" {
		searchLower := strings.ToLower(f.FuzzyText)
		descLower := strings.ToLower(tx.Description)
		catLower := strings.ToLower(string(tx.Category))
		notesLower := strings.ToLower(tx.Notes)

		if !strings.Contains(descLower, searchLower) &&
			!strings.Contains(catLower, searchLower) &&
			!strings.Contains(notesLower, searchLower) {
			return false
		}
	}

	return true
}

func matchesDate(tx models.Transaction, dateRange string) bool {
	now := time.Now()

	switch dateRange {
	case "today":
		return tx.Date == now.Format("2006-01-02")
	case "last-week":
		weekAgo := now.AddDate(0, 0, -7).Format("2006-01-02")
		return tx.Date >= weekAgo
	case "last-month":
		monthAgo := now.AddDate(0, -1, 0).Format("2006-01-02")
		return tx.Date >= monthAgo
	case "this-month":
		return len(tx.Date) >= 7 && tx.Date[:7] == now.Format("2006-01")
	default:
		// Exact date match.
		return tx.Date == dateRange
	}
}
