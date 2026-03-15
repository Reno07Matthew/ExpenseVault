package cmd

import (
	"fmt"
	"strings"

	"expenseVault/export"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export transactions to a file",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("output")
		format, _ := cmd.Flags().GetString("format")
		if path == "" {
			return fmt.Errorf("output path is required")
		}

		userID, err := getCurrentUserID()
		if err != nil {
			return err
		}

		txs, err := store.GetAllTransactions(userID)
		if err != nil {
			return err
		}

		var exporter export.Exporter
		switch strings.ToLower(format) {
		case "csv":
			exporter = &export.CSVExporter{}
		case "json":
			exporter = &export.JSONExporter{}
		default:
			return fmt.Errorf("unknown format: %s", format)
		}

		if err := exporter.Export(txs, path); err != nil {
			return err
		}

		fmt.Printf("Exported %d transactions.\n", len(txs))
		return nil
	},
}

func init() {
	exportCmd.Flags().StringP("output", "o", "", "Output file path")
	exportCmd.Flags().StringP("format", "f", "csv", "Format: csv or json")
}
