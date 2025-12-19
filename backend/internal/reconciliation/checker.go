package reconciliation

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
)

// Checker performs periodic reconciliation of account balances
type Checker struct {
	logger   *zap.Logger
	interval time.Duration
	metrics  *Metrics
	stopChan chan struct{}
}

// ReconciliationResult represents the result of checking a single account
type ReconciliationResult struct {
	AccountID           uint
	MaterializedBalance float64
	ComputedBalance     float64
	Difference          float64
	IsConsistent        bool
}

// NewChecker creates a new reconciliation checker
func NewChecker(logger *zap.Logger, interval time.Duration) *Checker {
	return &Checker{
		logger:   logger,
		interval: interval,
		metrics:  NewMetrics(),
		stopChan: make(chan struct{}),
	}
}

// Start begins periodic reconciliation checks
func (c *Checker) Start(ctx context.Context) {
	c.logger.Info("Starting reconciliation checker",
		zap.Duration("interval", c.interval),
	)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// Initial check
	c.RunCheck(ctx)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Reconciliation checker stopping")
			return
		case <-c.stopChan:
			c.logger.Info("Reconciliation checker stopped")
			return
		case <-ticker.C:
			c.RunCheck(ctx)
		}
	}
}

// Stop gracefully stops the checker
func (c *Checker) Stop() {
	close(c.stopChan)
}

// RunCheck performs a single reconciliation check across all accounts
func (c *Checker) RunCheck(ctx context.Context) {
	startTime := time.Now()
	c.logger.Info("Starting reconciliation check")

	var accounts []models.Account
	if err := db.DB.Find(&accounts).Error; err != nil {
		c.logger.Error("Failed to fetch accounts for reconciliation", zap.Error(err))
		return
	}

	var failures int
	var totalChecked int

	for _, account := range accounts {
		result := c.CheckAccount(account.ID)
		totalChecked++

		if !result.IsConsistent {
			failures++
			c.logger.Warn("Account balance inconsistency detected",
				zap.Uint("accountId", result.AccountID),
				zap.Float64("materializedBalance", result.MaterializedBalance),
				zap.Float64("computedBalance", result.ComputedBalance),
				zap.Float64("difference", result.Difference),
			)
		}
	}

	duration := time.Since(startTime)
	c.metrics.ReconciliationFailures.Add(float64(failures))
	c.metrics.ReconciliationTotal.Add(float64(totalChecked))
	c.metrics.ReconciliationDuration.Observe(duration.Seconds())

	if failures > 0 {
		c.logger.Error("Reconciliation check completed with failures",
			zap.Int("failures", failures),
			zap.Int("totalChecked", totalChecked),
			zap.Duration("duration", duration),
		)
	} else {
		c.logger.Info("Reconciliation check completed successfully",
			zap.Int("accountsChecked", totalChecked),
			zap.Duration("duration", duration),
		)
	}
}

// CheckAccount verifies that an account's materialized balance matches computed balance
func (c *Checker) CheckAccount(accountID uint) ReconciliationResult {
	var account models.Account
	if err := db.DB.First(&account, accountID).Error; err != nil {
		c.logger.Error("Failed to fetch account for reconciliation",
			zap.Uint("accountId", accountID),
			zap.Error(err),
		)
		return ReconciliationResult{
			AccountID:    accountID,
			IsConsistent: false,
		}
	}

	// Compute balance from journal lines
	var computedBalance float64
	var totalCredits, totalDebits float64

	// Sum all credits (money in)
	db.DB.Model(&models.JournalLine{}).
		Where("account_id = ? AND direction = ?", accountID, models.DirectionCredit).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalCredits)

	// Sum all debits (money out)
	db.DB.Model(&models.JournalLine{}).
		Where("account_id = ? AND direction = ?", accountID, models.DirectionDebit).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalDebits)

	// Balance = credits - debits (money in - money out)
	computedBalance = totalCredits - totalDebits

	// Compare with materialized balance
	materializedBalance := account.Balance
	difference := materializedBalance - computedBalance

	// Allow small floating-point differences (0.01)
	isConsistent := abs(difference) < 0.01

	return ReconciliationResult{
		AccountID:           accountID,
		MaterializedBalance: materializedBalance,
		ComputedBalance:     computedBalance,
		Difference:          difference,
		IsConsistent:        isConsistent,
	}
}

// abs returns absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// CheckAllAccounts performs reconciliation check and returns results
func (c *Checker) CheckAllAccounts() ([]ReconciliationResult, error) {
	var accounts []models.Account
	if err := db.DB.Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %w", err)
	}

	results := make([]ReconciliationResult, 0, len(accounts))
	for _, account := range accounts {
		result := c.CheckAccount(account.ID)
		results = append(results, result)
	}

	return results, nil
}
