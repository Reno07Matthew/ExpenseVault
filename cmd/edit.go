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
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetInt("id")
		if id <= 0 {
			return fmt.Errorf("id is required")
		}

		existing, err := store.GetTransaction(id)
		if err != nil {
			return err
		}

		txType, _ := cmd.Flags().GetString("type")
		amount, _ := cmd.Flags().GetFloat64("amount")
		category, _ := cmd.Flags().GetString("category")
		desc, _ := cmd.Flags().GetString("description")
		date, _ := cmd.Flags().GetString("date")
		notes, _ := cmd.Flags().GetString("notes")

		if txType != "" {
			existing.Type = models.TransactionType(txType)
		}
		if amount > 0 {
			existing.Amount = models.Rupees(amount)
		}
		if category != "" {
			existing.Category = models.Category(category)
		}
		if desc != "" {
			existing.Description = desc
		}
		if date != "" {
			existing.Date = date
		} else if existing.Date == "" {
			existing.Date = time.Now().Format("2006-01-02")
		}
		if notes != "" {
			existing.Notes = notes
		}

		if err := store.UpdateTransaction(*existing); err != nil {
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
