package cmd

import (
	"fmt"
	"time"

	"expenseVault/models"
	"expenseVault/services"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new transaction",
	RunE: func(cmd *cobra.Command, args []string) error {
		txType, _ := cmd.Flags().GetString("type")
		amount, _ := cmd.Flags().GetFloat64("amount")
		category, _ := cmd.Flags().GetString("category")
		desc, _ := cmd.Flags().GetString("description")
		date, _ := cmd.Flags().GetString("date")
		notes, _ := cmd.Flags().GetString("notes")
		autoCat, _ := cmd.Flags().GetBool("auto-cat")

		if date == "" {
			date = time.Now().Format("2006-01-02")
		}

		cat := models.Category(category)
		if autoCat || category == "" {
			cat = services.NewCategorizer().AutoCategorize(desc)
		}

		tx := models.Transaction{
			Type:        models.TransactionType(txType),
			Amount:      models.Rupees(amount),
			Category:    cat,
			Description: desc,
			Date:        date,
			Notes:       notes,
		}

		id, err := store.AddTransaction(tx)
		if err != nil {
			return err
		}
		fmt.Printf("Added transaction #%d\n", id)
		return nil
	},
}

func init() {
	addCmd.Flags().StringP("type", "t", "", "Transaction type: income or expense")
	addCmd.Flags().Float64P("amount", "a", 0, "Amount")
	addCmd.Flags().StringP("category", "c", "", "Category")
	addCmd.Flags().StringP("description", "d", "", "Description")
	addCmd.Flags().String("date", "", "Date (YYYY-MM-DD)")
	addCmd.Flags().String("notes", "", "Notes")
	addCmd.Flags().Bool("auto-cat", false, "Auto-categorize based on description")
	_ = addCmd.MarkFlagRequired("type")
	_ = addCmd.MarkFlagRequired("amount")
	_ = addCmd.MarkFlagRequired("description")
}
