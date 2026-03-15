package cmd

import (
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync transactions with server",
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, err := getCurrentUserID()
		if err != nil {
			return err
		}

		serverURL, _ := cmd.Flags().GetString("server")
		if serverURL == "" {
			return fmt.Errorf("server URL is required")
		}

		fmt.Printf("Starting concurrent sync for user #%d to %s...\n", userID, serverURL)

		var wg sync.WaitGroup
		statusChan := make(chan string, 2)

		// Goroutine 1: Upload
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Simulate upload delay
			time.Sleep(1 * time.Second)
			statusChan <- "Upload complete"
		}()

		// Goroutine 2: Download
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Simulate download delay
			time.Sleep(1 * time.Second)
			statusChan <- "Download complete"
		}()

		// Wait and close channel in separate goroutine
		go func() {
			wg.Wait()
			close(statusChan)
		}()

		for status := range statusChan {
			fmt.Println("-", status)
		}

		fmt.Println("Sync successful.")
		return nil
	},
}

func init() {
	syncCmd.Flags().String("server", "", "Server URL")
	syncCmd.Flags().String("token", "", "JWT token")
}
