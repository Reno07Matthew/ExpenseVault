package export

import "expenseVault/models"

// Exporter exports transactions to a file.
type Exporter interface {
	Export(transactions []models.Transaction, outputPath string) error
}

// Importer imports transactions from a file.
type Importer interface {
	Import(inputPath string) ([]models.Transaction, error)
}
