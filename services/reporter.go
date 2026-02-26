package services

import (
	"fmt"
	"sort"
	"strings"

	"expenseVault/models"
)

// Reporter generates a report summary.
type Reporter interface {
	Generate(transactions []models.Transaction) string
}

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
