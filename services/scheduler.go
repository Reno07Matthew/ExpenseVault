package services

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"expenseVault/models"
)

// ──────────────────────────────────────────────────────────
// ADVANCED FEATURE: Recurring Transaction Scheduler
// Demonstrates: Background goroutines, Ticker, time handling
// ──────────────────────────────────────────────────────────

var schedLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

// Frequency defines how often a recurring transaction repeats.
type Frequency string

const (
	FrequencyDaily   Frequency = "daily"
	FrequencyWeekly  Frequency = "weekly"
	FrequencyMonthly Frequency = "monthly"
)

// RecurringRule defines a rule for automatic transaction creation.
type RecurringRule struct {
	ID          int64
	UserID      int64
	Type        models.TransactionType
	Amount      models.Rupees
	Category    models.Category
	Description string
	Frequency   Frequency
	NextDue     time.Time
	Active      bool
}

// TransactionAdder is an interface for adding transactions (decouples from db.Store).
type TransactionAdder interface {
	AddTransaction(t *models.Transaction) (int64, error)
}

// Scheduler checks for due recurring transactions and creates them.
type Scheduler struct {
	rules    []RecurringRule
	store    TransactionAdder
	stopChan chan struct{}
}

// NewScheduler creates a new recurring transaction scheduler.
func NewScheduler(store TransactionAdder) *Scheduler {
	return &Scheduler{
		store:    store,
		stopChan: make(chan struct{}),
	}
}

// AddRule adds a recurring transaction rule.
func (s *Scheduler) AddRule(rule RecurringRule) {
	s.rules = append(s.rules, rule)
	schedLogger.Info("Recurring rule added",
		slog.String("description", rule.Description),
		slog.String("frequency", string(rule.Frequency)),
		slog.Float64("amount", rule.Amount.ToFloat64()),
	)
}

// ProcessDueTransactions checks all rules and creates transactions that are due.
// Returns the number of transactions created.
func (s *Scheduler) ProcessDueTransactions() int {
	now := time.Now()
	created := 0

	for i := range s.rules {
		rule := &s.rules[i]
		if !rule.Active {
			continue
		}

		// Create all transactions that are past due.
		for rule.NextDue.Before(now) || rule.NextDue.Equal(now) {
			tx := models.NewTransaction(
				rule.UserID,
				rule.Type,
				rule.Amount.ToFloat64(),
				rule.Category,
				rule.Description,
				rule.NextDue.Format("2006-01-02"),
			)
			tx.SetNotes(fmt.Sprintf("Auto-created by recurring rule (%s)", rule.Frequency))

			id, err := s.store.AddTransaction(tx)
			if err != nil {
				schedLogger.Error("Failed to create recurring transaction",
					slog.String("error", err.Error()),
					slog.String("description", rule.Description),
				)
				break
			}

			schedLogger.Info("Recurring transaction created",
				slog.Int64("tx_id", id),
				slog.String("description", rule.Description),
				slog.String("date", rule.NextDue.Format("2006-01-02")),
			)

			created++

			// Advance to the next due date.
			switch rule.Frequency {
			case FrequencyDaily:
				rule.NextDue = rule.NextDue.AddDate(0, 0, 1)
			case FrequencyWeekly:
				rule.NextDue = rule.NextDue.AddDate(0, 0, 7)
			case FrequencyMonthly:
				rule.NextDue = rule.NextDue.AddDate(0, 1, 0)
			}
		}
	}

	return created
}

// StartBackground runs the scheduler as a background goroutine.
// It checks for due transactions every hour.
func (s *Scheduler) StartBackground() {
	go func() {
		// Process on startup.
		created := s.ProcessDueTransactions()
		if created > 0 {
			schedLogger.Info("Startup: created recurring transactions",
				slog.Int("count", created),
			)
		}

		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				created := s.ProcessDueTransactions()
				if created > 0 {
					schedLogger.Info("Periodic: created recurring transactions",
						slog.Int("count", created),
					)
				}
			case <-s.stopChan:
				schedLogger.Info("Scheduler stopped")
				return
			}
		}
	}()

	schedLogger.Info("Recurring transaction scheduler started")
}

// Stop gracefully stops the background scheduler.
func (s *Scheduler) Stop() {
	close(s.stopChan)
}
