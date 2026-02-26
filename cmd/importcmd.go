package cmd

import (
	"fmt"
	"strings"

	"expenseVault/export"

	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import transactions from a file",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("file")
		format, _ := cmd.Flags().GetString("format")
		if path == "" {
			return fmt.Errorf("file path is required")
		}

		var importer export.Importer
		switch strings.ToLower(format) {
		case "csv":
			importer = &export.CSVImporter{}
		case "json":
			importer = &export.JSONImporter{}
		default:
			return fmt.Errorf("unknown format: %s", format)
		}

		txs, err := importer.Import(path)
		if err != nil {
			return err
		}

		count, err := store.BulkInsert(txs)
		if err != nil {
			return err
		}

		fmt.Printf("Imported %d transactions.\n", count)
		return nil
	},
}

func init() {
	importCmd.Flags().StringP("file", "f", "", "Input file path")
	importCmd.Flags().StringP("format", "t", "csv", "Format: csv or json")
}
