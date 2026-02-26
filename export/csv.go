package export

import (
	"encoding/csv"
	"os"

	"expenseVault/models"
)

// CSVExporter exports transactions to CSV.
type CSVExporter struct{}

func (e *CSVExporter) Export(transactions []models.Transaction, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	_ = writer.Write([]string{"id", "type", "amount", "category", "description", "date", "notes"})
	for _, t := range transactions {
		_ = writer.Write([]string{
			int64ToString(t.ID),
			string(t.Type),
			t.Amount.String(),
			string(t.Category),
			t.Description,
			t.Date,
			t.Notes,
		})
	}
	return writer.Error()
}

// CSVImporter imports transactions from CSV.
type CSVImporter struct{}

func (i *CSVImporter) Import(inputPath string) ([]models.Transaction, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var txs []models.Transaction
	for idx, row := range rows {
		if idx == 0 {
			continue
		}
		if len(row) < 7 {
			continue
		}
		amount, err := parseAmount(row[2])
		if err != nil {
			return nil, err
		}
		txs = append(txs, models.Transaction{
			Type:        models.TransactionType(row[1]),
			Amount:      models.Rupees(amount),
			Category:    models.Category(row[3]),
			Description: row[4],
			Date:        row[5],
			Notes:       row[6],
		})
	}
	return txs, nil
}
