package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"expenseVault/export"

	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup all transactions to a JSON file",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		if output == "" {
			home, _ := os.UserHomeDir()
			output = filepath.Join(home, "expensevault_backup.json")
		}

		txs, err := store.GetAllTransactions()
		if err != nil {
			return err
		}

		exporter := &export.JSONExporter{}
		if err := exporter.Export(txs, output); err != nil {
			return err
		}
		fmt.Printf("Backup saved to %s\n", output)
		return nil
	},
}

func init() {
	backupCmd.Flags().StringP("output", "o", "", "Output file path")
}
