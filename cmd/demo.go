package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Demo placeholder",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Demo command is available for coursework concepts.")
	},
}
