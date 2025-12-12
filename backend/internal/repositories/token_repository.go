package repositories

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/RedShawn258/FinTrack/backend/internal/models"
)

type TokenRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewTokenRepository(db *gorm.DB, logger *zap.Logger) *TokenRepository {
	return &TokenRepository{
		db:     db,
		logger: logger,
	}
}

// Create stores a new refresh token in the database
func (r *TokenRepository) Create(token *models.RefreshToken) error {
	if err := r.db.Create(token).Error; err != nil {
		r.logger.Error("Failed to create refresh token", zap.Error(err))
		return err
	}
	return nil
}

// FindByTokenHash finds a refresh token by its hash
func (r *TokenRepository) FindByTokenHash(tokenHash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	if err := r.db.Where("token_hash = ? AND deleted_at IS NULL", tokenHash).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

// MarkAsReplaced marks a token as replaced by another token (for rotation)
func (r *TokenRepository) MarkAsReplaced(tokenID uint, replacedByID uint) error {
	if err := r.db.Model(&models.RefreshToken{}).
		Where("id = ?", tokenID).
		Update("replaced_by", replacedByID).Error; err != nil {
		r.logger.Error("Failed to mark token as replaced", zap.Error(err), zap.Uint("tokenID", tokenID))
		return err
	}
	return nil
}

// DeleteExpiredTokens removes expired tokens from the database
func (r *TokenRepository) DeleteExpiredTokens() error {
	if err := r.db.Where("expires_at < ?", time.Now()).Delete(&models.RefreshToken{}).Error; err != nil {
		r.logger.Error("Failed to delete expired tokens", zap.Error(err))
		return err
	}
	return nil
}

// DeleteByUserID deletes all refresh tokens for a user (e.g., on logout all devices)
func (r *TokenRepository) DeleteByUserID(userID uint) error {
	if err := r.db.Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error; err != nil {
		r.logger.Error("Failed to delete user tokens", zap.Error(err), zap.Uint("userID", userID))
		return err
	}
	return nil
}

// IsTokenReplaced checks if a token has been replaced (used for reuse detection)
func (r *TokenRepository) IsTokenReplaced(tokenID uint) (bool, error) {
	var token models.RefreshToken
	if err := r.db.Select("replaced_by").Where("id = ?", tokenID).First(&token).Error; err != nil {
		return false, err
	}
	return token.ReplacedBy != nil, nil
}

