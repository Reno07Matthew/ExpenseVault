package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a transaction",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetInt("id")
		if id <= 0 {
			return fmt.Errorf("id is required")
		}
		userID, err := getCurrentUserID()
		if err != nil {
			return err
		}

		existing, err := store.GetTransaction(id)
		if err != nil {
			return err
		}

		if existing.UserID != userID {
			return fmt.Errorf("access denied")
		}

		if err := store.DeleteTransaction(id, userID); err != nil {
			return err
		}
		fmt.Println("Transaction deleted.")
		return nil
	},
}

func init() {
	deleteCmd.Flags().Int("id", 0, "Transaction ID")
}
