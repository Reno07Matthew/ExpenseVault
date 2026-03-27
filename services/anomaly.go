package services

import (
	"fmt"
	"math"
	"sort"
	"time"

	"expenseVault/models"
)

// ──────────────────────────────────────────────────────────
// ADVANCED FEATURE: Anomaly Detection
// Demonstrates: Statistical analysis, evaluation, innovation
// ──────────────────────────────────────────────────────────

// AnomalyAlert represents a flagged unusual transaction.
type AnomalyAlert struct {
	Transaction models.Transaction
	ZScore      float64 // How many standard deviations from the mean
	Message     string
}

// DetectAnomalies analyzes transactions and flags outliers.
// Uses standard deviation to identify transactions that are
// significantly higher than the category average (> 2σ).
func DetectAnomalies(txs []models.Transaction) []AnomalyAlert {
	// Group expenses by category.
	categoryTxs := make(map[models.Category][]models.Transaction)
	for _, tx := range txs {
		if tx.IsExpense() {
			categoryTxs[tx.Category] = append(categoryTxs[tx.Category], tx)
		}
	}

	var alerts []AnomalyAlert

	for cat, catTxs := range categoryTxs {
		if len(catTxs) < 3 {
			continue // Need at least 3 data points for meaningful stats
		}

		// Calculate mean.
		var sum float64
		for _, tx := range catTxs {
			sum += tx.Amount.ToFloat64()
		}
		mean := sum / float64(len(catTxs))

		// Calculate standard deviation.
		var varianceSum float64
		for _, tx := range catTxs {
			diff := tx.Amount.ToFloat64() - mean
			varianceSum += diff * diff
		}
		stdDev := math.Sqrt(varianceSum / float64(len(catTxs)))

		if stdDev == 0 {
			continue // All amounts are identical, no anomalies
		}

		// Flag transactions > 2 standard deviations above mean.
		threshold := 2.0
		for _, tx := range catTxs {
			zScore := (tx.Amount.ToFloat64() - mean) / stdDev
			if zScore > threshold {
				alerts = append(alerts, AnomalyAlert{
					Transaction: tx,
					ZScore:      math.Round(zScore*100) / 100,
					Message: fmt.Sprintf(
						"⚠️ Unusual %s expense: ₹%.0f (%.1fx above average ₹%.0f)",
						cat, tx.Amount.ToFloat64(), zScore, mean,
					),
				})
			}
		}
	}

	// Sort by severity (highest z-score first).
	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].ZScore > alerts[j].ZScore
	})

	return alerts
}

// GetSpendingTrend analyzes whether spending is increasing or decreasing.
// Returns a trend message comparing recent vs older spending.
func GetSpendingTrend(txs []models.Transaction) string {
	expenses := make([]models.Transaction, 0)
	for _, tx := range txs {
		if tx.IsExpense() {
			expenses = append(expenses, tx)
		}
	}

	if len(expenses) < 6 {
		return "📊 Not enough data for trend analysis yet."
	}

	// Sort by date.
	sort.Slice(expenses, func(i, j int) bool {
		return expenses[i].Date < expenses[j].Date
	})

	// Split into two halves.
	mid := len(expenses) / 2
	olderHalf := expenses[:mid]
	recentHalf := expenses[mid:]

	var olderSum, recentSum float64
	for _, tx := range olderHalf {
		olderSum += tx.Amount.ToFloat64()
	}
	for _, tx := range recentHalf {
		recentSum += tx.Amount.ToFloat64()
	}

	olderAvg := olderSum / float64(len(olderHalf))
	recentAvg := recentSum / float64(len(recentHalf))

	if olderAvg == 0 {
		return "📊 Trend data unavailable."
	}

	changePercent := ((recentAvg - olderAvg) / olderAvg) * 100

	if changePercent > 15 {
		return fmt.Sprintf("📈 Spending is UP %.0f%% — watch your expenses!", changePercent)
	} else if changePercent < -15 {
		return fmt.Sprintf("📉 Spending is DOWN %.0f%% — great discipline!", math.Abs(changePercent))
	}
	return "📊 Spending is stable. Keep it up!"
}

// ──────────────────────────────────────────────────────────
// Predictive Budgeting — End-of-Month Balance Forecast
// ──────────────────────────────────────────────────────────

// PredictionData holds the predictive budget analysis results.
type PredictionData struct {
	DailyBurnRate     float64
	DaysElapsed       int
	DaysRemaining     int
	ProjectedExpenses float64
	ProjectedSavings  float64
	Confidence        string // "High", "Medium", "Low"
}

// PredictEndOfMonth forecasts the end-of-month balance.
// Uses linear projection based on current spending velocity.
func PredictEndOfMonth(txs []models.Transaction, monthlyIncome float64) PredictionData {
	now := time.Now()
	daysInMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.Local).Day()
	daysElapsed := now.Day()
	daysRemaining := daysInMonth - daysElapsed

	// Sum expenses in the current month.
	currentMonth := now.Format("2006-01")
	var monthExpenses float64
	var txCount int
	for _, tx := range txs {
		if tx.IsExpense() && len(tx.Date) >= 7 && tx.Date[:7] == currentMonth {
			monthExpenses += tx.Amount.ToFloat64()
			txCount++
		}
	}

	pred := PredictionData{
		DaysElapsed:   daysElapsed,
		DaysRemaining: daysRemaining,
	}

	if daysElapsed == 0 {
		pred.Confidence = "Low"
		pred.ProjectedSavings = monthlyIncome
		return pred
	}

	// Daily burn rate = total spent / days elapsed.
	pred.DailyBurnRate = monthExpenses / float64(daysElapsed)

	// Project total month expenses.
	pred.ProjectedExpenses = pred.DailyBurnRate * float64(daysInMonth)
	pred.ProjectedSavings = monthlyIncome - pred.ProjectedExpenses

	// Confidence based on data points.
	switch {
	case txCount >= 15 && daysElapsed >= 15:
		pred.Confidence = "High"
	case txCount >= 7 && daysElapsed >= 7:
		pred.Confidence = "Medium"
	default:
		pred.Confidence = "Low"
	}

	return pred
}

// FormatPrediction returns a formatted string for the dashboard.
func FormatPrediction(pred PredictionData) string {
	if pred.Confidence == "Low" {
		return "🔮 Not enough data to predict — keep logging!"
	}

	icon := "🟢"
	if pred.ProjectedSavings < 0 {
		icon = "🔴"
	} else if pred.ProjectedSavings < pred.ProjectedExpenses*0.1 {
		icon = "🟡"
	}

	return fmt.Sprintf(
		"%s Predicted EOM Savings: ₹%.0f | Burn Rate: ₹%.0f/day | %d days left [%s confidence]",
		icon, pred.ProjectedSavings, pred.DailyBurnRate, pred.DaysRemaining, pred.Confidence,
	)
}
