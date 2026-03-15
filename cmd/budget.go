package cmd

import (
	"fmt"
	"strconv"
	"time"

	"expenseVault/models"

	"github.com/spf13/cobra"
)

var budgetCmd = &cobra.Command{
	Use:   "budget [category] [amount]",
	Short: "Set a monthly budget target for a category",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, err := getCurrentUserID()
		if err != nil {
			return err
		}

		category := models.Category(args[0])
		if !models.IsCategoryValid(category) {
			return fmt.Errorf("invalid category: %s", args[0])
		}

		amount, err := strconv.ParseFloat(args[1], 64)
		if err != nil || amount < 0 {
			return fmt.Errorf("invalid amount: %s", args[1])
		}

		month, _ := cmd.Flags().GetString("month")
		if month == "" {
			month = time.Now().Format("2006-01")
		}

		err = store.SetBudget(userID, category, models.Rupees(amount*100), month)
		if err != nil {
			return err
		}

		fmt.Printf("✅ Budget for %s set to ₹%.2f for %s\n", category, amount, month)
		return nil
	},
}

func init() {
	budgetCmd.Flags().StringP("month", "m", "", "Month in YYYY-MM format (default: current month)")
	rootCmd.AddCommand(budgetCmd)
}
