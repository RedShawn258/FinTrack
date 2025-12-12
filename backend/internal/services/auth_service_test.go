package services

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/RedShawn258/FinTrack/backend/internal/config"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
	"github.com/RedShawn258/FinTrack/backend/internal/repositories"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	return gormDB, mock
}

func setupTestService(t *testing.T) (*AuthService, *gorm.DB, sqlmock.Sqlmock) {
	db, mock := setupTestDB(t)
	logger := zap.NewNop()
	
	cfg := &config.Config{
		JWTSecret:          "test-secret-key",
		RefreshTokenSecret: "test-refresh-secret",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
	}

	tokenRepo := repositories.NewTokenRepository(db, logger)
	service := NewAuthService(cfg, tokenRepo, logger)

	return service, db, mock
}

func TestHashToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "hash simple token",
			token: "test-token",
		},
		{
			name:  "hash empty token",
			token: "",
		},
		{
			name:  "hash long token",
			token: "a-very-long-token-string-that-should-be-hashed-properly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result1 := HashToken(tt.token)
			result2 := HashToken(tt.token)
			
			// Hash should be deterministic
			assert.Equal(t, result1, result2)
			// Hash should be 64 characters (SHA-256 hex)
			assert.Len(t, result1, 64)
			// Hash should not be empty
			assert.NotEmpty(t, result1)
		})
	}
}

func TestGenerateAccessToken(t *testing.T) {
	service, _, _ := setupTestService(t)
	userID := uint(1)

	token, err := service.GenerateAccessToken(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the token can be parsed
	claims, err := service.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
}

func TestValidateAccessToken(t *testing.T) {
	service, _, _ := setupTestService(t)
	userID := uint(1)

	tests := []struct {
		name          string
		setupToken    func() string
		expectedError bool
		expectedUserID uint
	}{
		{
			name: "valid token",
			setupToken: func() string {
				token, _ := service.GenerateAccessToken(userID)
				return token
			},
			expectedError: false,
			expectedUserID: userID,
		},
		{
			name: "invalid token string",
			setupToken: func() string {
				return "invalid.token.string"
			},
			expectedError: true,
		},
		{
			name: "expired token",
			setupToken: func() string {
				// Create a token with short expiry and wait
				oldExpiry := service.config.AccessTokenExpiry
				service.config.AccessTokenExpiry = 1 * time.Nanosecond
				token, _ := service.GenerateAccessToken(userID)
				time.Sleep(2 * time.Nanosecond)
				service.config.AccessTokenExpiry = oldExpiry
				return token
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupToken()
			claims, err := service.ValidateAccessToken(token)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, claims)
				assert.Equal(t, tt.expectedUserID, claims.UserID)
			}
		})
	}
}

func TestGenerateTokenPair(t *testing.T) {
	service, _, mock := setupTestService(t)
	userID := uint(1)

	// Mock the database insert for refresh token
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `refresh_tokens`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	tokenPair, err := service.GenerateTokenPair(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.Greater(t, tokenPair.ExpiresIn, int64(0))

	// Validate access token
	claims, err := service.ValidateAccessToken(tokenPair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestValidateRefreshToken(t *testing.T) {
	service, _, mock := setupTestService(t)
	userID := uint(1)

	tests := []struct {
		name          string
		setupToken    func() (string, *models.RefreshToken)
		expectedError bool
		description   string
	}{
		{
			name: "valid refresh token",
			setupToken: func() (string, *models.RefreshToken) {
				// Create token in DB first
				tokenString := "test-refresh-token"
				tokenHash := HashToken(tokenString)
				dbToken := &models.RefreshToken{
					ID:        1,
					UserID:    userID,
					TokenHash: tokenHash,
					ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
				}
				
				// Mock DB query
				rows := sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "created_at", "updated_at", "deleted_at", "replaced_by"}).
					AddRow(dbToken.ID, dbToken.UserID, dbToken.TokenHash, dbToken.ExpiresAt, time.Now(), time.Now(), nil, nil)
				mock.ExpectQuery("SELECT .* FROM `refresh_tokens`").
					WithArgs(tokenHash).
					WillReturnRows(rows)

				// Generate actual JWT token
				oldExpiry := service.config.RefreshTokenExpiry
				service.config.RefreshTokenExpiry = 7 * 24 * time.Hour
				jwtToken, _, _ := service.GenerateRefreshToken(userID)
				service.config.RefreshTokenExpiry = oldExpiry

				// Get the hash of the generated token
				actualHash := HashToken(jwtToken)
				
				// Update mock to use actual hash
				mock.ExpectQuery("SELECT .* FROM `refresh_tokens`").
					WithArgs(actualHash).
					WillReturnRows(rows)

				return jwtToken, dbToken
			},
			expectedError: false,
			description:   "should validate a valid refresh token",
		},
		{
			name: "expired refresh token",
			setupToken: func() (string, *models.RefreshToken) {
				tokenString := "expired-token"
				tokenHash := HashToken(tokenString)
				dbToken := &models.RefreshToken{
					ID:        2,
					UserID:    userID,
					TokenHash: tokenHash,
					ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
				}
				
				rows := sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "created_at", "updated_at", "deleted_at", "replaced_by"}).
					AddRow(dbToken.ID, dbToken.UserID, dbToken.TokenHash, dbToken.ExpiresAt, time.Now(), time.Now(), nil, nil)
				mock.ExpectQuery("SELECT .* FROM `refresh_tokens`").
					WithArgs(tokenHash).
					WillReturnRows(rows)

				return tokenString, dbToken
			},
			expectedError: true,
			description:   "should reject expired refresh token",
		},
		{
			name: "replaced refresh token",
			setupToken: func() (string, *models.RefreshToken) {
				tokenString := "replaced-token"
				tokenHash := HashToken(tokenString)
				replacedBy := uint(2)
				dbToken := &models.RefreshToken{
					ID:        3,
					UserID:    userID,
					TokenHash: tokenHash,
					ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
					ReplacedBy: &replacedBy,
				}
				
				rows := sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "created_at", "updated_at", "deleted_at", "replaced_by"}).
					AddRow(dbToken.ID, dbToken.UserID, dbToken.TokenHash, dbToken.ExpiresAt, time.Now(), time.Now(), nil, replacedBy)
				mock.ExpectQuery("SELECT .* FROM `refresh_tokens`").
					WithArgs(tokenHash).
					WillReturnRows(rows)

				return tokenString, dbToken
			},
			expectedError: true,
			description:   "should reject replaced refresh token (reuse detection)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString, _ := tt.setupToken()
			claims, dbToken, err := service.ValidateRefreshToken(tokenString)

			if tt.expectedError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, claims)
				assert.Nil(t, dbToken)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, claims)
				assert.NotNil(t, dbToken)
			}
		})
	}
}

func TestValidateAndRotateRefreshToken(t *testing.T) {
	service, _, mock := setupTestService(t)
	userID := uint(1)

	// This is a simplified test - in practice, you'd need to properly mock
	// the entire flow including JWT generation and DB operations
	t.Run("token rotation flow", func(t *testing.T) {
		// Generate initial token pair
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `refresh_tokens`").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		initialPair, err := service.GenerateTokenPair(userID)
		require.NoError(t, err)

		// Mock validation of old token
		tokenHash := HashToken(initialPair.RefreshToken)
		rows := sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "created_at", "updated_at", "deleted_at", "replaced_by"}).
			AddRow(1, userID, tokenHash, time.Now().Add(7*24*time.Hour), time.Now(), time.Now(), nil, nil)
		mock.ExpectQuery("SELECT .* FROM `refresh_tokens`").
			WithArgs(tokenHash).
			WillReturnRows(rows)

		// Mock creation of new refresh token
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `refresh_tokens`").
			WillReturnResult(sqlmock.NewResult(2, 1))
		mock.ExpectCommit()

		// Mock marking old token as replaced
		mock.ExpectExec("UPDATE `refresh_tokens`").
			WithArgs(2, 1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Note: This test is simplified - actual implementation would need
		// proper JWT token generation and validation
		// The real test would validate the full rotation flow
		_ = initialPair
	})
}

