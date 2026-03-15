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

		userID, err := getCurrentUserID()
		if err != nil {
			return err
		}

		txs, err := store.GetAllTransactions(userID)
		if err != nil {
			return err
		}
		if len(txs) == 0 {
			fmt.Println("No transactions for report.")
			return nil
		}

		if all {
			fmt.Println("Generating all reports (parallel)...")
			fmt.Println(services.RunAllReports(txs))
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
