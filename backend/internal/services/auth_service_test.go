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
		name           string
		setupToken     func() string
		expectedError  bool
		expectedUserID uint
	}{
		{
			name: "valid token",
			setupToken: func() string {
				token, _ := service.GenerateAccessToken(userID)
				return token
			},
			expectedError:  false,
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
				// Generate actual JWT token first (this will try to insert into DB)
				oldExpiry := service.config.RefreshTokenExpiry
				service.config.RefreshTokenExpiry = 7 * 24 * time.Hour

				// Mock the INSERT for GenerateRefreshToken
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `refresh_tokens`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				jwtToken, dbToken, err := service.GenerateRefreshToken(userID)
				service.config.RefreshTokenExpiry = oldExpiry

				if err != nil {
					// If generation fails, create a simple valid JWT manually for testing
					// This is a fallback - the real token should work
					return "", nil
				}

				// Get the hash of the generated token
				actualHash := HashToken(jwtToken)

				// Update dbToken with actual hash
				dbToken.TokenHash = actualHash

				// Mock the SELECT query that ValidateRefreshToken will make
				// GORM adds LIMIT parameter, so we need to match with sqlmock.AnyArg()
				rows := sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "created_at", "updated_at", "deleted_at", "replaced_by"}).
					AddRow(dbToken.ID, dbToken.UserID, dbToken.TokenHash, dbToken.ExpiresAt, time.Now(), time.Now(), nil, nil)
				mock.ExpectQuery("SELECT .* FROM `refresh_tokens`").
					WithArgs(actualHash, sqlmock.AnyArg()). // GORM adds LIMIT parameter
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
					WithArgs(tokenHash, sqlmock.AnyArg()). // GORM adds LIMIT parameter
					WillReturnRows(rows)

				return tokenString, dbToken
			},
			expectedError: true,
			description:   "should reject expired refresh token",
		},
		{
			name: "replaced refresh token",
			setupToken: func() (string, *models.RefreshToken) {
				// Generate a valid JWT first
				oldExpiry := service.config.RefreshTokenExpiry
				service.config.RefreshTokenExpiry = 7 * 24 * time.Hour

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `refresh_tokens`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				jwtToken, dbToken, err := service.GenerateRefreshToken(userID)
				service.config.RefreshTokenExpiry = oldExpiry

				if err != nil || dbToken == nil {
					return "", nil
				}

				// Mark as replaced
				replacedBy := uint(2)
				dbToken.ReplacedBy = &replacedBy
				tokenHash := HashToken(jwtToken)

				rows := sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "created_at", "updated_at", "deleted_at", "replaced_by"}).
					AddRow(dbToken.ID, dbToken.UserID, dbToken.TokenHash, dbToken.ExpiresAt, time.Now(), time.Now(), nil, replacedBy)
				mock.ExpectQuery("SELECT .* FROM `refresh_tokens`").
					WithArgs(tokenHash, sqlmock.AnyArg()).
					WillReturnRows(rows)

				return jwtToken, dbToken
			},
			expectedError: true,
			description:   "should reject replaced refresh token (reuse detection)",
		},
		{
			name: "refresh token not found in database",
			setupToken: func() (string, *models.RefreshToken) {
				// Generate a valid JWT token
				oldExpiry := service.config.RefreshTokenExpiry
				service.config.RefreshTokenExpiry = 7 * 24 * time.Hour

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `refresh_tokens`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				jwtToken, _, _ := service.GenerateRefreshToken(userID)
				service.config.RefreshTokenExpiry = oldExpiry

				// Mock DB to return no rows (token not found)
				tokenHash := HashToken(jwtToken)
				rows := sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "created_at", "updated_at", "deleted_at", "replaced_by"})
				mock.ExpectQuery("SELECT .* FROM `refresh_tokens`").
					WithArgs(tokenHash, sqlmock.AnyArg()).
					WillReturnRows(rows)

				return jwtToken, nil
			},
			expectedError: true,
			description:   "should reject refresh token not found in database",
		},
		{
			name: "invalid refresh token format",
			setupToken: func() (string, *models.RefreshToken) {
				// Return malformed JWT string
				return "not.a.valid.jwt.token", nil
			},
			expectedError: true,
			description:   "should reject malformed refresh token",
		},
		{
			name: "user ID mismatch",
			setupToken: func() (string, *models.RefreshToken) {
				// Generate token for one user
				oldExpiry := service.config.RefreshTokenExpiry
				service.config.RefreshTokenExpiry = 7 * 24 * time.Hour

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `refresh_tokens`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				jwtToken, dbToken, err := service.GenerateRefreshToken(userID)
				service.config.RefreshTokenExpiry = oldExpiry

				if err != nil || dbToken == nil {
					return "", nil
				}

				// But mock DB to return different user ID
				tokenHash := HashToken(jwtToken)
				differentUserID := uint(999)
				rows := sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "created_at", "updated_at", "deleted_at", "replaced_by"}).
					AddRow(dbToken.ID, differentUserID, dbToken.TokenHash, dbToken.ExpiresAt, time.Now(), time.Now(), nil, nil)
				mock.ExpectQuery("SELECT .* FROM `refresh_tokens`").
					WithArgs(tokenHash, sqlmock.AnyArg()).
					WillReturnRows(rows)

				return jwtToken, dbToken
			},
			expectedError: true,
			description:   "should reject token with mismatched user ID",
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

func TestGenerateRefreshToken(t *testing.T) {
	service, _, mock := setupTestService(t)
	userID := uint(1)

	tests := []struct {
		name          string
		setupMock     func()
		expectedError bool
		description   string
	}{
		{
			name: "successful refresh token generation",
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `refresh_tokens`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedError: false,
			description:   "should generate valid refresh token and store in DB",
		},
		{
			name: "database insert failure",
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `refresh_tokens`").
					WillReturnError(gorm.ErrInvalidDB)
				mock.ExpectRollback()
			},
			expectedError: true,
			description:   "should return error when DB insert fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			token, dbToken, err := service.GenerateRefreshToken(userID)

			if tt.expectedError {
				assert.Error(t, err, tt.description)
				assert.Empty(t, token)
				assert.Nil(t, dbToken)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotEmpty(t, token)
				assert.NotNil(t, dbToken)
				assert.Equal(t, userID, dbToken.UserID)
				assert.NotEmpty(t, dbToken.TokenHash)
			}
		})
	}
}

func TestValidateAccessToken_EdgeCases(t *testing.T) {
	service, _, _ := setupTestService(t)
	userID := uint(1)

	tests := []struct {
		name          string
		setupToken    func() string
		expectedError bool
		description   string
	}{
		{
			name: "token with wrong signing method",
			setupToken: func() string {
				// Create a token signed with wrong secret to simulate wrong method
				// In practice, this would be caught by the signing method check
				return "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOjF9.invalid"
			},
			expectedError: true,
			description:   "should reject token with unexpected signing method",
		},
		{
			name: "empty token string",
			setupToken: func() string {
				return ""
			},
			expectedError: true,
			description:   "should reject empty token string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupToken()
			claims, err := service.ValidateAccessToken(token)

			if tt.expectedError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, claims)
				assert.Equal(t, userID, claims.UserID)
			}
		})
	}
}

func TestValidateAndRotateRefreshToken(t *testing.T) {
	userID := uint(1)

	tests := []struct {
		name          string
		setupMock     func(*testing.T, *AuthService, sqlmock.Sqlmock) string
		expectedError bool
		description   string
	}{
		{
			name: "successful token rotation",
			setupMock: func(t *testing.T, service *AuthService, mock sqlmock.Sqlmock) string {
				// Generate initial refresh token
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `refresh_tokens`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				refreshToken, _, err := service.GenerateRefreshToken(userID)
				require.NoError(t, err)

				// Mock validation query
				tokenHash := HashToken(refreshToken)
				rows := sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "created_at", "updated_at", "deleted_at", "replaced_by"}).
					AddRow(1, userID, tokenHash, time.Now().Add(7*24*time.Hour), time.Now(), time.Now(), nil, nil)
				mock.ExpectQuery("SELECT .* FROM `refresh_tokens`").
					WithArgs(tokenHash, sqlmock.AnyArg()).
					WillReturnRows(rows)

				// Mock new refresh token creation
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `refresh_tokens`").
					WillReturnResult(sqlmock.NewResult(2, 1))
				mock.ExpectCommit()

				// Mock marking old token as replaced
				mock.ExpectExec("UPDATE `refresh_tokens`").
					WithArgs(sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(0, 1))

				return refreshToken
			},
			expectedError: false,
			description:   "should successfully rotate tokens and mark old token as replaced",
		},
		{
			name: "invalid refresh token",
			setupMock: func(t *testing.T, service *AuthService, mock sqlmock.Sqlmock) string {
				return "invalid.token.string"
			},
			expectedError: true,
			description:   "should return error for invalid refresh token",
		},
		{
			name: "expired refresh token",
			setupMock: func(t *testing.T, service *AuthService, mock sqlmock.Sqlmock) string {
				// Generate a valid JWT token first
				oldExpiry := service.config.RefreshTokenExpiry
				service.config.RefreshTokenExpiry = 7 * 24 * time.Hour

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `refresh_tokens`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				refreshToken, _, err := service.GenerateRefreshToken(userID)
				service.config.RefreshTokenExpiry = oldExpiry
				require.NoError(t, err)

				// Mock validation query with expired token (expires_at in the past)
				tokenHash := HashToken(refreshToken)
				rows := sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "created_at", "updated_at", "deleted_at", "replaced_by"}).
					AddRow(1, userID, tokenHash, time.Now().Add(-1*time.Hour), time.Now(), time.Now(), nil, nil)
				mock.ExpectQuery("SELECT .* FROM `refresh_tokens`").
					WithArgs(tokenHash, sqlmock.AnyArg()).
					WillReturnRows(rows)

				return refreshToken
			},
			expectedError: true,
			description:   "should return error for expired refresh token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, _, mock := setupTestService(t)
			refreshToken := tt.setupMock(t, service, mock)
			tokenPair, err := service.ValidateAndRotateRefreshToken(refreshToken)

			if tt.expectedError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, tokenPair)
			} else {
				assert.NoError(t, err, tt.description)
				require.NotNil(t, tokenPair)
				assert.NotEmpty(t, tokenPair.AccessToken)
				assert.NotEmpty(t, tokenPair.RefreshToken)
				assert.Greater(t, tokenPair.ExpiresIn, int64(0))

				// Validate new access token
				claims, err := service.ValidateAccessToken(tokenPair.AccessToken)
				assert.NoError(t, err)
				assert.Equal(t, userID, claims.UserID)
			}
		})
	}
}

func TestGenerateTokenPair_ErrorCases(t *testing.T) {
	service, _, mock := setupTestService(t)
	userID := uint(1)

	t.Run("refresh token generation failure", func(t *testing.T) {
		// Mock DB failure for refresh token
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `refresh_tokens`").
			WillReturnError(gorm.ErrInvalidDB)
		mock.ExpectRollback()

		tokenPair, err := service.GenerateTokenPair(userID)
		assert.Error(t, err)
		assert.Nil(t, tokenPair)
	})
}

func TestValidateAndRotateRefreshToken_ErrorCases(t *testing.T) {
	userID := uint(1)

	// Note: Access token generation failure is hard to test without mocking crypto/rand
	// The GenerateAccessToken function has very few error paths (only JWT signing errors)
	// which are difficult to simulate. The current test coverage for ValidateAndRotateRefreshToken
	// already covers the main error paths (invalid token, expired token, refresh token generation failure)

	t.Run("refresh token generation failure during rotation", func(t *testing.T) {
		service, _, mock := setupTestService(t)

		// Generate initial refresh token
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `refresh_tokens`").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		refreshToken, _, err := service.GenerateRefreshToken(userID)
		require.NoError(t, err)

		// Mock validation query
		tokenHash := HashToken(refreshToken)
		rows := sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "created_at", "updated_at", "deleted_at", "replaced_by"}).
			AddRow(1, userID, tokenHash, time.Now().Add(7*24*time.Hour), time.Now(), time.Now(), nil, nil)
		mock.ExpectQuery("SELECT .* FROM `refresh_tokens`").
			WithArgs(tokenHash, sqlmock.AnyArg()).
			WillReturnRows(rows)

		// Mock DB failure for new refresh token creation
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `refresh_tokens`").
			WillReturnError(gorm.ErrInvalidDB)
		mock.ExpectRollback()

		tokenPair, err := service.ValidateAndRotateRefreshToken(refreshToken)
		assert.Error(t, err)
		assert.Nil(t, tokenPair)
		assert.Contains(t, err.Error(), "failed to generate refresh token")
	})
}
