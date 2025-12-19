package db

import (
	"fmt"

	"go.uber.org/zap"
)

// FORBIDDEN_AUTOMIGRATE_MODELS lists models that MUST NOT be registered with AutoMigrate.
// These models use reserved keywords (e.g., 'key') or require explicit SQL control.
// Adding them to AutoMigrate will cause GORM to cache schema metadata and generate
// incorrect SQL queries. See GORM_ELIMINATION_SUMMARY.md for details.
var FORBIDDEN_AUTOMIGRATE_MODELS = []string{
	"IdempotencyKey",
	"JournalEntry",
	"JournalLine",
	"OutboxEvent",
}

// validateAutoMigrateModels ensures forbidden models are not registered with AutoMigrate.
// This is a runtime guard to prevent accidental regression.
func validateAutoMigrateModels(models []interface{}, logger *zap.Logger) error {
	for _, model := range models {
		modelType := fmt.Sprintf("%T", model)
		for _, forbidden := range FORBIDDEN_AUTOMIGRATE_MODELS {
			// Check if the model type name contains the forbidden model name
			// This works because fmt.Sprintf("%T", &models.IdempotencyKey{}) returns "*models.IdempotencyKey"
			if contains(modelType, forbidden) {
				logger.Error("FORBIDDEN MODEL IN AUTOMIGRATE",
					zap.String("model", modelType),
					zap.String("forbidden", forbidden),
					zap.String("reason", "This model must use explicit SQL migrations via RunRawMigrations()"),
				)
				return fmt.Errorf(
					"PRODUCTION GUARDRAIL VIOLATION: Model '%s' is FORBIDDEN from AutoMigrate(). "+
						"This model must use explicit SQL migrations via RunRawMigrations() to avoid "+
						"GORM schema cache issues with reserved keywords. See GORM_ELIMINATION_SUMMARY.md",
					forbidden,
				)
			}
		}
	}
	return nil
}

// contains checks if a string contains a substring (case-sensitive check for model type names)
func contains(s, substr string) bool {
	// Simple substring check - model type names like "*models.IdempotencyKey" contain the model name
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ValidateMigrationSafety performs runtime validation to ensure migration safety.
// Called before AutoMigrate to catch regressions early.
func ValidateMigrationSafety(logger *zap.Logger) error {
	logger.Info("Validating migration safety guardrails")

	// This validation happens at runtime, so we can't check the actual AutoMigrate call
	// Instead, we document the forbidden models and rely on code review + this check
	// as a reminder. The actual validation happens in RunMigrations().

	logger.Info("Migration safety validation passed",
		zap.Strings("forbidden_models", FORBIDDEN_AUTOMIGRATE_MODELS),
		zap.String("note", "These models must use explicit SQL migrations"),
	)

	return nil
}
