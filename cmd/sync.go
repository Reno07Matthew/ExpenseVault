package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync transactions with server",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, _ := cmd.Flags().GetString("server")
		if serverURL == "" {
			return fmt.Errorf("server URL is required")
		}

		fmt.Printf("Sync not implemented. Server: %s\n", serverURL)
		return nil
	},
}

func init() {
	syncCmd.Flags().String("server", "", "Server URL")
	syncCmd.Flags().String("token", "", "JWT token")
}
