package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List transactions",
	RunE: func(cmd *cobra.Command, args []string) error {
		txType, _ := cmd.Flags().GetString("type")
		category, _ := cmd.Flags().GetString("category")
		startDate, _ := cmd.Flags().GetString("start")
		endDate, _ := cmd.Flags().GetString("end")
		limit, _ := cmd.Flags().GetInt("limit")

		userID, err := getCurrentUserID()
		if err != nil {
			return err
		}

		txs, err := store.ListTransactions(userID, txType, category, startDate, endDate, limit)
		if err != nil {
			return err
		}
		if len(txs) == 0 {
			fmt.Println("No transactions found.")
			return nil
		}

		for _, t := range txs {
			// LAB 4: Uses value-receiver String() method on Transaction.
			fmt.Println(t.String())
		}
		return nil
	},
}

func init() {
	listCmd.Flags().String("type", "", "Filter by type")
	listCmd.Flags().String("category", "", "Filter by category")
	listCmd.Flags().String("start", "", "Start date (YYYY-MM-DD)")
	listCmd.Flags().String("end", "", "End date (YYYY-MM-DD)")
	listCmd.Flags().Int("limit", 0, "Limit results")
}
