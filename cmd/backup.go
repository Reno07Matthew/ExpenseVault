package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"expenseVault/export"
	"expenseVault/services"

	"golang.org/x/term"

	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup all transactions to a JSON file (with optional encryption)",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		encrypt, _ := cmd.Flags().GetBool("encrypt")

		if output == "" {
			home, _ := os.UserHomeDir()
			if encrypt {
				output = filepath.Join(home, "expensevault_backup.enc")
			} else {
				output = filepath.Join(home, "expensevault_backup.json")
			}
		}

		userID, err := getCurrentUserID()
		if err != nil {
			return err
		}

		txs, err := store.GetAllTransactions(userID)
		if err != nil {
			return err
		}

		exporter := &export.JSONExporter{}
		if err := exporter.Export(txs, output+".tmp"); err != nil {
			return err
		}

		if encrypt {
			// Read the temporary JSON file.
			plaintext, err := os.ReadFile(output + ".tmp")
			if err != nil {
				return err
			}
			_ = os.Remove(output + ".tmp")

			// Prompt for encryption password.
			fmt.Print("Encryption password: ")
			passBytes, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Println("")
			if err != nil {
				return err
			}

			// Encrypt with AES-256-GCM.
			ciphertext, err := services.Encrypt(plaintext, string(passBytes))
			if err != nil {
				return fmt.Errorf("encryption failed: %w", err)
			}

			if err := os.WriteFile(output, ciphertext, 0600); err != nil {
				return err
			}
			fmt.Printf("🔒 Encrypted backup saved to %s (%d transactions)\n", output, len(txs))
		} else {
			// Rename temp file to final output.
			_ = os.Rename(output+".tmp", output)
			fmt.Printf("Backup saved to %s (%d transactions)\n", output, len(txs))
		}

		return nil
	},
}

func init() {
	backupCmd.Flags().StringP("output", "o", "", "Output file path")
	backupCmd.Flags().Bool("encrypt", false, "Encrypt the backup with AES-256-GCM")
}
