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

		txs, err := store.ListTransactions(txType, category, startDate, endDate, limit)
		if err != nil {
			return err
		}
		if len(txs) == 0 {
			fmt.Println("No transactions found.")
			return nil
		}

		for _, t := range txs {
			fmt.Printf("%d | %s | %s | %s | %s | %s\n", t.ID, t.Type, t.Amount, t.Category, t.Description, t.Date)
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
