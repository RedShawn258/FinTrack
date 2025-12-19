package utils

import (
	"fmt"

	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
)

// VerifyJournalEntryBalance verifies that a journal entry satisfies the double-entry invariant:
// SUM(debits) == SUM(credits)
// Returns error if invariant is violated
func VerifyJournalEntryBalance(entryID uint) error {
	var totalDebits, totalCredits float64

	// Calculate total debits
	if err := db.DB.Model(&models.JournalLine{}).
		Where("entry_id = ? AND direction = ?", entryID, models.DirectionDebit).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalDebits).Error; err != nil {
		return fmt.Errorf("failed to calculate debits: %w", err)
	}

	// Calculate total credits
	if err := db.DB.Model(&models.JournalLine{}).
		Where("entry_id = ? AND direction = ?", entryID, models.DirectionCredit).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalCredits).Error; err != nil {
		return fmt.Errorf("failed to calculate credits: %w", err)
	}

	// Verify invariant
	if totalDebits != totalCredits {
		return fmt.Errorf("double-entry invariant violated: entry %d has debits=%.2f, credits=%.2f (difference=%.2f)",
			entryID, totalDebits, totalCredits, totalDebits-totalCredits)
	}

	return nil
}

// ReconcileAccountBalance recalculates account balance from journal lines
// Useful for verification and reconciliation jobs
func ReconcileAccountBalance(accountID uint) (float64, error) {
	var balance float64

	// Balance = SUM(credits) - SUM(debits)
	// Credits increase balance, debits decrease balance
	if err := db.DB.Model(&models.JournalLine{}).
		Where("account_id = ?", accountID).
		Select("COALESCE(SUM(CASE WHEN direction = 'credit' THEN amount ELSE -amount END), 0)").
		Scan(&balance).Error; err != nil {
		return 0, fmt.Errorf("failed to reconcile balance: %w", err)
	}

	return balance, nil
}
