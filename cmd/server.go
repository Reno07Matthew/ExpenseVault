package cmd

import (
	"fmt"

	"expenseVault/api"
	"expenseVault/utils"

	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start sync server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := utils.LoadConfig()
		if err != nil {
			return err
		}
		port, _ := cmd.Flags().GetInt("port")
		if port == 0 {
			port = cfg.ServerPort
		}
		addr := fmt.Sprintf(":%d", port)
		fmt.Printf("Starting server on %s\n", addr)
		return api.StartServer(addr)
	},
}

func init() {
	serverCmd.Flags().Int("port", 0, "Server port")
}
