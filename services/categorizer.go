package services

import (
	"strings"

	"expenseVault/models"
)

// Categorizer auto-categorizes transactions.
type Categorizer struct{}

// NewCategorizer is a factory function that returns a *Categorizer (pointer).
// LAB 4: Factory function returning pointer.
func NewCategorizer() *Categorizer {
	return &Categorizer{}
}

// AutoCategorize uses a pointer receiver to match the method set convention.
// LAB 4: Pointer receiver on Categorizer.
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
