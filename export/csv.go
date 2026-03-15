package export

import (
	"encoding/csv"
	"os"

	"expenseVault/models"
)

// CSVExporter exports transactions to CSV.
// UNIT 3: Implements Exporter interface — polymorphism.
type CSVExporter struct{}

func (e *CSVExporter) Export(transactions []models.Transaction, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	// UNIT 3: Defer — ensures file is closed even if an error occurs below.
	defer file.Close()

	writer := csv.NewWriter(file)
	// UNIT 3: Defer — Flush is deferred so all buffered data is written.
	defer writer.Flush()

	// UNIT 2: Slice — composite literal ([]string{...}).
	_ = writer.Write([]string{"id", "type", "amount", "category", "description", "date", "notes"})
	for _, t := range transactions {
		// UNIT 1: Conversion — string(t.Type), t.Amount.String(), etc.
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
// UNIT 3: Implements Importer interface — polymorphism.
type CSVImporter struct{}

func (i *CSVImporter) Import(inputPath string) ([]models.Transaction, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	// UNIT 3: Defer — file.Close() runs when Import returns.
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// UNIT 1: var keyword — txs starts as nil slice (zero value for slice).
	var txs []models.Transaction
	for idx, row := range rows {
		if idx == 0 {
			continue // skip header
		}
		if len(row) < 7 {
			continue
		}
		amount, err := parseAmount(row[2])
		if err != nil {
			return nil, err
		}
		// UNIT 1: Type conversion — models.TransactionType(row[1]), models.Rupees(amount), etc.
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
