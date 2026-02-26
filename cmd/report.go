package cmd

import (
	"fmt"

	"expenseVault/services"

	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate reports",
	RunE: func(cmd *cobra.Command, args []string) error {
		reportType, _ := cmd.Flags().GetString("type")
		all, _ := cmd.Flags().GetBool("all")

		txs, err := store.GetAllTransactions()
		if err != nil {
			return err
		}
		if len(txs) == 0 {
			fmt.Println("No transactions for report.")
			return nil
		}

		if all {
			fmt.Println("Monthly Report:")
			fmt.Println((&services.MonthlyReporter{}).Generate(txs))
			fmt.Println("Category Report:")
			fmt.Println((&services.CategoryReporter{}).Generate(txs))
			fmt.Println("Yearly Report:")
			fmt.Println((&services.YearlyReporter{}).Generate(txs))
			return nil
		}

		var reporter services.Reporter
		switch reportType {
		case "monthly":
			reporter = &services.MonthlyReporter{}
		case "category":
			reporter = &services.CategoryReporter{}
		case "yearly":
			reporter = &services.YearlyReporter{}
		default:
			return fmt.Errorf("unknown report type: %s", reportType)
		}

		fmt.Println(reporter.Generate(txs))
		return nil
	},
}

func init() {
	reportCmd.Flags().String("type", "monthly", "Report type: monthly, category, yearly")
	reportCmd.Flags().Bool("all", false, "Generate all reports")
}
