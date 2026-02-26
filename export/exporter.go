package export

import (
	"fmt"
	"strings"

	"expenseVault/models"
)

// ──────────────────────────────────────────────────────────
// UNIT 3 — Interfaces & Polymorphism
// ──────────────────────────────────────────────────────────

// Exporter exports transactions to a file.
// UNIT 3: Interface — polymorphism; CSVExporter and JSONExporter both implement this.
type Exporter interface {
	Export(transactions []models.Transaction, outputPath string) error
}

// Importer imports transactions from a file.
// UNIT 3: Interface — polymorphism; CSVImporter and JSONImporter both implement this.
type Importer interface {
	Import(inputPath string) ([]models.Transaction, error)
}

// GetExporter returns the appropriate exporter based on format string.
// UNIT 3: Polymorphism — returns different concrete types as the same interface.
func GetExporter(format string) (Exporter, error) {
	switch strings.ToLower(format) {
	case "csv":
		return &CSVExporter{}, nil
	case "json":
		return &JSONExporter{}, nil
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// GetImporter returns the appropriate importer based on format string.
// UNIT 3: Polymorphism — returns different concrete types as the same interface.
func GetImporter(format string) (Importer, error) {
	switch strings.ToLower(format) {
	case "csv":
		return &CSVImporter{}, nil
	case "json":
		return &JSONImporter{}, nil
	default:
		return nil, fmt.Errorf("unsupported import format: %s", format)
	}
}

// ──────────────────────────────────────────────────────────
// UNIT 3 — Function expression + Returning a function
// ──────────────────────────────────────────────────────────

// TransformFunc is a function type for transforming transactions before export.
// UNIT 3: Function expression — named function type.
type TransformFunc func(models.Transaction) models.Transaction

// MakeDescriptionPrefixer returns a TransformFunc that prefixes descriptions.
// UNIT 3: Returning a function + Closure — captures `prefix`.
func MakeDescriptionPrefixer(prefix string) TransformFunc {
	return func(tx models.Transaction) models.Transaction {
		tx.Description = prefix + tx.Description
		return tx
	}
}

// TransformAll applies a transform function to all transactions.
// UNIT 3: Callback — accepts a function as parameter.
// UNIT 2: Slice — make + append.
func TransformAll(txs []models.Transaction, fn TransformFunc) []models.Transaction {
	result := make([]models.Transaction, 0, len(txs))
	for _, tx := range txs {
		result = append(result, fn(tx))
	}
	return result
}
