package cmd

import (
	"fmt"
	"os"
	"syscall"

	"expenseVault/export"
	"expenseVault/services"

	"golang.org/x/term"

	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore transactions from a JSON file (with optional decryption)",
	RunE: func(cmd *cobra.Command, args []string) error {
		input, _ := cmd.Flags().GetString("file")
		decrypt, _ := cmd.Flags().GetBool("decrypt")

		if input == "" {
			return fmt.Errorf("file path is required")
		}

		var data []byte
		var err error

		if decrypt {
			// Read encrypted file.
			ciphertext, readErr := os.ReadFile(input)
			if readErr != nil {
				return readErr
			}

			// Prompt for decryption password.
			fmt.Print("Decryption password: ")
			passBytes, termErr := term.ReadPassword(int(syscall.Stdin))
			fmt.Println("")
			if termErr != nil {
				return termErr
			}

			// Decrypt with AES-256-GCM.
			data, err = services.Decrypt(ciphertext, string(passBytes))
			if err != nil {
				return fmt.Errorf("decryption failed: %w", err)
			}
			fmt.Println("🔓 Backup decrypted successfully")

			// Write decrypted data to temp file for import.
			tmpFile := input + ".tmp"
			if err := os.WriteFile(tmpFile, data, 0600); err != nil {
				return err
			}
			defer os.Remove(tmpFile)
			input = tmpFile
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
	restoreCmd.Flags().Bool("decrypt", false, "Decrypt the backup before restoring")
}
