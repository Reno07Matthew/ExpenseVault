package services

import (
	"strings"

	"expenseVault/models"
)

// Categorizer auto-categorizes transactions.
type Categorizer struct{}

func NewCategorizer() *Categorizer {
	return &Categorizer{}
}

func (c *Categorizer) AutoCategorize(desc string) models.Category {
	d := strings.ToLower(desc)

	switch {
	case strings.Contains(d, "food") || strings.Contains(d, "lunch"):
		return models.CategoryFood
	case strings.Contains(d, "uber") || strings.Contains(d, "taxi") || strings.Contains(d, "bus"):
		return models.CategoryTravel
	case strings.Contains(d, "bill") || strings.Contains(d, "electric") || strings.Contains(d, "water"):
		return models.CategoryBills
	case strings.Contains(d, "netflix") || strings.Contains(d, "movie"):
		return models.CategoryEntertainment
	case strings.Contains(d, "salary") || strings.Contains(d, "pay"):
		return models.CategorySalary
	default:
		return models.CategoryOther
	}
}
