package export

import (
	"os"

	"expenseVault/models"
)

// JSONExporter exports transactions to JSON.
// LAB 4.1: Uses models.MarshalTransactions for JSON serialization.
type JSONExporter struct{}

func (e *JSONExporter) Export(transactions []models.Transaction, outputPath string) error {
	// LAB 4.1: Marshal using centralized JSON helper.
	data, err := models.MarshalTransactions(transactions)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, data, 0644)
}

// JSONImporter imports transactions from JSON.
// LAB 4.1: Uses models.UnmarshalTransactions for JSON deserialization.
type JSONImporter struct{}

func (i *JSONImporter) Import(inputPath string) ([]models.Transaction, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, err
	}
	// LAB 4.1: Unmarshal using centralized JSON helper.
	return models.UnmarshalTransactions(data)
}
