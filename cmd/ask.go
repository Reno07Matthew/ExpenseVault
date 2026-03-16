package cmd

import (
	"fmt"
	"strings"

	"expenseVault/services"

	"github.com/spf13/cobra"
)

// askCmd allows users to query their database using natural language.
var askCmd = &cobra.Command{
	Use:   "ask <query>",
	Short: "Ask a question about your expenses using natural language",
	Long: `Ask translates your natural language query into a secure SQL read-only statement, 
queries your data, and returns a conversational summary.

Example:
  expensevault ask "How much did I spend on food this month?"
  expensevault ask "What were my top 3 categories last year?"`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Authenticate user
		userID, err := getCurrentUserID()
		if err != nil {
			fmt.Println("Error: You must be logged in to use the ask feature.")
			fmt.Println("Please run 'expensevault login' first.")
			return
		}

		query := strings.Join(args, " ")
		fmt.Printf("Analyzing your query: %q...\n", query)

		// 2. Translate NL -> SQL via LLM
		sqlQuery, err := services.GenerateSQL(query, userID)
		if err != nil {
			fmt.Printf("Error generating SQL from LLM: %v\n", err)
			return
		}

		fmt.Printf("[DEBUG] Generated SQL: %s\n", sqlQuery)

		// 3. Run the SQL securely
		results, err := store.ExecuteReadQuery(sqlQuery)
		if err != nil {
			fmt.Printf("Error executing query against database: %v\n", err)
			return
		}

		// 4. Summarize results back into Natural Language
		summary, err := services.SummarizeData(query, results)
		if err != nil {
			fmt.Printf("Error summarizing results: %v\n", err)
			return
		}

		// 5. Output the conversational response
		fmt.Println("\n--- Answer ---")
		fmt.Println(summary)
	},
}
