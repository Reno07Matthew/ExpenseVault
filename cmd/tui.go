package cmd

import (
	"fmt"

	"expenseVault/tui"
	"expenseVault/utils"

	"github.com/spf13/cobra"
)

// tuiCmd launches the interactive TUI dashboard.
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive terminal dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := utils.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		return tui.RunTUI(store, cfg)
	},
}
