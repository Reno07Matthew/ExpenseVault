package cmd

import (
	"fmt"
	"time"

	"expenseVault/models"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit an existing transaction",
	// LAB 4: Uses pointer-based EditTransactionFields to mutate
	// the transaction through a pointer instead of manual field copies.
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetInt("id")
		if id <= 0 {
			return fmt.Errorf("id is required")
		}

		userID, err := getCurrentUserID()
		if err != nil {
			return err
		}

		// GetTransaction returns *models.Transaction (pointer from DB query)
		existing, err := store.GetTransaction(id)
		if err != nil {
			return err
		}

		if existing.UserID != userID {
			return fmt.Errorf("access denied")
		}

		txType, _ := cmd.Flags().GetString("type")
		amount, _ := cmd.Flags().GetFloat64("amount")
		category, _ := cmd.Flags().GetString("category")
		desc, _ := cmd.Flags().GetString("description")
		date, _ := cmd.Flags().GetString("date")
		notes, _ := cmd.Flags().GetString("notes")

		if date == "" && existing.Date == "" {
			date = time.Now().Format("2006-01-02")
		}

		// LAB 4: Pass pointer to mutate transaction in-place
		models.EditTransactionFields(
			existing, // *Transaction — pointer, mutations affect caller
			models.TransactionType(txType),
			amount,
			models.Category(category),
			desc,
			date,
			notes,
		)

		if err := store.UpdateTransaction(existing); err != nil {
			return err
		}
		fmt.Println("Transaction updated.")
		return nil
	},
}

func init() {
	editCmd.Flags().Int("id", 0, "Transaction ID")
	editCmd.Flags().String("type", "", "Transaction type: income or expense")
	editCmd.Flags().Float64("amount", 0, "Amount")
	editCmd.Flags().String("category", "", "Category")
	editCmd.Flags().String("description", "", "Description")
	editCmd.Flags().String("date", "", "Date (YYYY-MM-DD)")
	editCmd.Flags().String("notes", "", "Notes")
}
