package cmd

import (
	"fmt"

	"expenseVault/export"

	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore transactions from a JSON file",
	RunE: func(cmd *cobra.Command, args []string) error {
		input, _ := cmd.Flags().GetString("file")
		if input == "" {
			return fmt.Errorf("file path is required")
		}

		importer := &export.JSONImporter{}
		txs, err := importer.Import(input)
		if err != nil {
			return err
		}

		count, err := store.BulkInsert(txs)
		if err != nil {
			return err
		}
		fmt.Printf("Restored %d transactions.\n", count)
		return nil
	},
}

func init() {
	restoreCmd.Flags().StringP("file", "f", "", "Backup file path")
}
