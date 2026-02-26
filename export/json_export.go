package export

import (
	"encoding/json"
	"os"

	"expenseVault/models"
)

// JSONExporter exports transactions to JSON.
type JSONExporter struct{}

func (e *JSONExporter) Export(transactions []models.Transaction, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(transactions)
}

// JSONImporter imports transactions from JSON.
type JSONImporter struct{}

func (i *JSONImporter) Import(inputPath string) ([]models.Transaction, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var txs []models.Transaction
	dec := json.NewDecoder(file)
	if err := dec.Decode(&txs); err != nil {
		return nil, err
	}
	return txs, nil
}
