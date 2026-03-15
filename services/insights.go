package services

import (
	"fmt"
	"math/rand"
	"time"

	"expenseVault/models"
)

// CalculateDashboardData processes transactions and budgets to generate metrics.
func CalculateDashboardData(txs []models.Transaction, budgets map[models.Category]models.Rupees) models.DashboardData {
	data := models.DashboardData{}
	
	categoryTotals := make(map[models.Category]models.Rupees)
	
	for _, tx := range txs {
		if tx.IsIncome() {
			data.MonthlyIncome += tx.Amount
		} else if tx.IsExpense() {
			data.TotalExpenses += tx.Amount
			categoryTotals[tx.Category] += tx.Amount
		}
	}

	data.Savings = data.MonthlyIncome - data.TotalExpenses
	
	// Determine normalization factor (Income or Total Budgeted)
	var base float64 = float64(data.MonthlyIncome)
	if base == 0 {
		var totalBudget models.Rupees
		for _, b := range budgets {
			totalBudget += b
		}
		if totalBudget > 0 {
			base = float64(totalBudget)
			data.UsingBudget = true
		}
	}

	if base > 0 {
		data.ExpenseRatio = float64(data.TotalExpenses) / base
		data.SavingsRatio = float64(data.Savings) / base
	}

	// Prepare breakdown
	for _, cat := range models.ValidCategories {
		amt := categoryTotals[cat]
		target := budgets[cat]
		
		if amt > 0 || target > 0 {
			percent := 0.0
			div := base
			if target > 0 {
				div = float64(target)
			}
			if div > 0 {
				percent = float64(amt) / div
			}
			data.Breakdown = append(data.Breakdown, models.CategoryBreakdown{
				Category: cat,
				Amount:   amt,
				Target:   target,
				Percent:  percent,
			})
		}
	}

	generateInsights(&data, categoryTotals, budgets)
	
	return data
}

func generateInsights(data *models.DashboardData, categoryTotals map[models.Category]models.Rupees, budgets map[models.Category]models.Rupees) {
	if data.MonthlyIncome == 0 && !data.UsingBudget {
		data.SmartInsight = "💡 Set an income or category budgets to see how you're tracking!"
		data.HasMajorIssues = true
		return
	}

	incomeF := float64(data.MonthlyIncome)
	if data.UsingBudget {
		var totalBudget models.Rupees
		for _, b := range budgets {
			totalBudget += b
		}
		incomeF = float64(totalBudget)
	}
	
	// Rule: Over individual category budget
	for cat, target := range budgets {
		actual := categoryTotals[cat]
		if actual > target && target > 0 {
			data.SmartInsight = fmt.Sprintf("⚠️ Budget exceeded for %s! You've spent %.0f%% of your target.", cat, (float64(actual)/float64(target))*100)
			data.HasMajorIssues = true
			break
		}
	}

	if data.SmartInsight == "" {
		// Legacy rules (weighted by income or total budget)
		if float64(categoryTotals[models.CategoryShopping]) > incomeF*0.30 {
			data.SmartInsight = "😅 Shopping spree detected! Maybe chill on Amazon?"
			data.HasMajorIssues = true
		} else if float64(categoryTotals[models.CategoryFood]) > incomeF*0.25 {
			data.SmartInsight = "🍔 That's a lot of food! Cooking at home could save money."
			data.HasMajorIssues = true
		} else if data.SavingsRatio > 0.30 {
			data.SmartInsight = "🔥 You're a saving machine!"
		}
	}

	if !data.HasMajorIssues && data.SmartInsight == "" {
		tips := []string{
			"Your future self thanks you for saving today.",
			"Small expenses leak big budgets.",
			"Coffee outside every day = vacation money gone.",
			"Track every penny, the pounds will take care of themselves.",
			"A budget is telling your money where to go instead of wondering where it went.",
		}
		rand.Seed(time.Now().UnixNano())
		data.DailyTip = tips[rand.Intn(len(tips))]
	}
}
