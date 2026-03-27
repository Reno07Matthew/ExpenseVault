package cmd

import (
	"fmt"
	"os"

	"expenseVault/db"
	"expenseVault/utils"

	"github.com/spf13/cobra"
)

var store *db.Store
var appConfig *utils.Config

// rootCmd is the base command for ExpenseVault CLI.
var rootCmd = &cobra.Command{
	Use:   "expensevault",
	Short: "ExpenseVault CLI",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "server" {
			return nil
		}

		cfg, err := utils.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		appConfig = cfg

		if cfg.DBType == "mysql" {
			store, err = db.NewMySQLStore(cfg.MySQLDSN)
			if err != nil {
				return fmt.Errorf("failed to connect to MySQL: %w", err)
			}
		} else if cfg.DBType == "supabase" || cfg.DBType == "postgres" {
			store, err = db.NewSupabaseStore(cfg.SupabaseDSN)
			if err != nil {
				return fmt.Errorf("failed to connect to Supabase/PostgreSQL: %w", err)
			}
		} else {
			store, err = db.NewStore(cfg.SQLitePath)
			if err != nil {
				return fmt.Errorf("failed to open SQLite database: %w", err)
			}
		}

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if store != nil {
			_ = store.Close()
		}
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(restoreCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(demoCmd)
	rootCmd.AddCommand(signupCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(askCmd)
}
