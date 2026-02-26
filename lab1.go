//go:build ignore

package main

import (
	"fmt"
	"strings"
	"time"
)

// ────────────────────────────────────────────────
// Custom Types
// ────────────────────────────────────────────────
type Rupees float64

func (r Rupees) String() string {
	return fmt.Sprintf("₹%.2f", float64(r))
}

type Category string

const (
	Food     Category = "Food"
	Travel   Category = "Travel"
	Shopping Category = "Shopping"
	Other    Category = "Other"
)

// Fixed category array
var categoryList = [4]Category{Food, Travel, Shopping, Other}

// ────────────────────────────────────────────────
// Structs
// ────────────────────────────────────────────────
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

type Expense struct {
	Date        Date
	Category    Category
	Description string
	Amount      Rupees
}

type Summary struct {
	Income     Rupees
	TotalSpent Rupees
	Remaining  Rupees
	ByCategory map[Category]Rupees
}

// ────────────────────────────────────────────────
// Global slice
// ────────────────────────────────────────────────
var expenses []Expense

// ────────────────────────────────────────────────
// Auto category detection
// ────────────────────────────────────────────────
func autoCategory(desc string) Category {
	desc = strings.ToLower(desc)

	if strings.Contains(desc, "food") || strings.Contains(desc, "lunch") {
		return Food
	} else if strings.Contains(desc, "bus") || strings.Contains(desc, "uber") {
		return Travel
	} else if strings.Contains(desc, "amazon") || strings.Contains(desc, "shop") {
		return Shopping
	}
	return Other
}

// ────────────────────────────────────────────────
// Category selection
// ────────────────────────────────────────────────
func chooseCategory(desc string) Category {
	fmt.Println("\nChoose Category:")
	for i, c := range categoryList {
		fmt.Printf("%d. %s\n", i+1, c)
	}
	fmt.Println("0. Auto detect")

	var choice int
	fmt.Print("Enter choice: ")
	fmt.Scanln(&choice)

	if choice == 0 {
		return autoCategory(desc)
	}

	if choice >= 1 && choice <= len(categoryList) {
		return categoryList[choice-1]
	}

	fmt.Println("Invalid choice. Defaulting to Other.")
	return Other
}

// ────────────────────────────────────────────────
// MAIN
// ────────────────────────────────────────────────
func main() {
	fmt.Println("💰 Income & Expense Tracker")
	fmt.Println("----------------------------")

	// Income input
	var incomeType string
	var incomeAmount float64

	fmt.Print("Enter income type (monthly/weekly): ")
	fmt.Scanln(&incomeType)

	fmt.Print("Enter income amount: ")
	fmt.Scanln(&incomeAmount)

	summary := Summary{
		Income:     Rupees(incomeAmount),
		ByCategory: make(map[Category]Rupees),
	}

	// Expense input loop
	for {
		var choice string
		fmt.Print("\nDo you want to add an expense? (yes/no): ")
		fmt.Scanln(&choice)

		if strings.ToLower(choice) != "yes" {
			break
		}

		var y, m, d int
		var desc string
		var amt float64

		fmt.Print("Enter date (YYYY MM DD): ")
		fmt.Scanln(&y, &m, &d)

		fmt.Print("Enter description: ")
		fmt.Scanln(&desc)

		cat := chooseCategory(desc)

		fmt.Print("Enter amount spent: ")
		fmt.Scanln(&amt)

		exp := Expense{
			Date:        Date{y, time.Month(m), d},
			Category:    cat,
			Description: desc,
			Amount:      Rupees(amt),
		}

		expenses = append(expenses, exp)
	}

	// ────────────────────────────────────────────────
	// Summary calculation
	// ────────────────────────────────────────────────
	for _, e := range expenses {
		summary.TotalSpent += e.Amount
		summary.ByCategory[e.Category] += e.Amount
	}

	summary.Remaining = summary.Income - summary.TotalSpent

	// ────────────────────────────────────────────────
	// Output
	// ────────────────────────────────────────────────
	fmt.Println("\n📊 Overall Summary")
	fmt.Println("------------------")
	fmt.Printf("Income (%s): %s\n", incomeType, summary.Income)
	fmt.Printf("Total Spent: %s\n", summary.TotalSpent)
	fmt.Printf("Remaining Balance: %s\n", summary.Remaining)

	fmt.Println("\nSpending by Category:")
	for cat, amt := range summary.ByCategory {
		fmt.Printf("  %-10s → %s\n", cat, amt)
	}

	fmt.Println("\nAll Expenses:")
	for i, e := range expenses {
		fmt.Printf(
			"%2d | %04d-%02d-%02d | %-8s | %-8s | %s\n",
			i+1,
			e.Date.Year,
			e.Date.Month,
			e.Date.Day,
			e.Category,
			e.Amount,
			e.Description,
		)
	}

	fmt.Println("\n--- End ---")
}
