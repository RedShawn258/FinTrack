package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"github.com/RedShawn258/FinTrack/backend/internal/config"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
	"github.com/RedShawn258/FinTrack/backend/internal/repositories"
)

type AuthService struct {
	config           *config.Config
	tokenRepo        *repositories.TokenRepository
	logger           *zap.Logger
	accessTokenSecret  string
	refreshTokenSecret string
}

func NewAuthService(cfg *config.Config, tokenRepo *repositories.TokenRepository, logger *zap.Logger) *AuthService {
	return &AuthService{
		config:             cfg,
		tokenRepo:          tokenRepo,
		logger:             logger,
		accessTokenSecret:  cfg.JWTSecret,
		refreshTokenSecret: cfg.RefreshTokenSecret,
	}
}

// TokenPair represents both access and refresh tokens
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 // seconds until access token expires
}

// AccessTokenClaims represents the claims in an access token
type AccessTokenClaims struct {
	UserID uint `json:"userId"`
	jwt.RegisteredClaims
}

// RefreshTokenClaims represents the claims in a refresh token
type RefreshTokenClaims struct {
	UserID uint `json:"userId"`
	jwt.RegisteredClaims
}

// HashToken hashes a token using SHA-256
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// GenerateAccessToken creates a short-lived access token
func (s *AuthService) GenerateAccessToken(userID uint) (string, error) {
	expirationTime := time.Now().Add(s.config.AccessTokenExpiry)
	
	claims := &AccessTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.accessTokenSecret))
}

// GenerateRefreshToken creates a long-lived refresh token and stores it in DB
func (s *AuthService) GenerateRefreshToken(userID uint) (string, *models.RefreshToken, error) {
	// Generate a random token string
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		s.logger.Error("Failed to generate random token", zap.Error(err))
		return "", nil, errors.New("failed to generate refresh token")
	}
	
	// Create JWT with the random token as the ID
	tokenID := hex.EncodeToString(tokenBytes)
	expirationTime := time.Now().Add(s.config.RefreshTokenExpiry)
	
	claims := &RefreshTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.refreshTokenSecret))
	if err != nil {
		s.logger.Error("Failed to sign refresh token", zap.Error(err))
		return "", nil, err
	}

	// Hash and store the token
	tokenHash := HashToken(tokenString)
	refreshToken := &models.RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expirationTime,
	}

	if err := s.tokenRepo.Create(refreshToken); err != nil {
		s.logger.Error("Failed to store refresh token", zap.Error(err))
		return "", nil, errors.New("failed to store refresh token")
	}

	return tokenString, refreshToken, nil
}

// ValidateAccessToken validates and parses an access token
func (s *AuthService) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.accessTokenSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token and checks if it's been used/replaced
func (s *AuthService) ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, *models.RefreshToken, error) {
	// Parse the JWT
	claims := &RefreshTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.refreshTokenSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, nil, errors.New("invalid refresh token")
	}

	// Look up the token in the database
	tokenHash := HashToken(tokenString)
	dbToken, err := s.tokenRepo.FindByTokenHash(tokenHash)
	if err != nil {
		return nil, nil, errors.New("refresh token not found")
	}

	// Check if token has expired
	if time.Now().After(dbToken.ExpiresAt) {
		return nil, nil, errors.New("refresh token expired")
	}

	// Check if token has been replaced (reuse detection)
	if dbToken.ReplacedBy != nil {
		s.logger.Warn("Attempted reuse of replaced refresh token", 
			zap.Uint("tokenID", dbToken.ID),
			zap.Uint("userID", claims.UserID))
		return nil, nil, errors.New("refresh token has been revoked")
	}

	// Verify user ID matches
	if claims.UserID != dbToken.UserID {
		return nil, nil, errors.New("token user mismatch")
	}

	return claims, dbToken, nil
}

// ValidateAndRotateRefreshToken validates a refresh token and issues new tokens (rotation)
func (s *AuthService) ValidateAndRotateRefreshToken(refreshTokenString string) (*TokenPair, error) {
	// Validate the refresh token
	claims, dbToken, err := s.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	userID := claims.UserID

	// Generate new access token
	newAccessToken, err := s.GenerateAccessToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate new refresh token
	newRefreshTokenString, newDbToken, err := s.GenerateRefreshToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Mark old token as replaced
	if err := s.tokenRepo.MarkAsReplaced(dbToken.ID, newDbToken.ID); err != nil {
		s.logger.Error("Failed to mark old token as replaced", zap.Error(err))
		// Don't fail the request, but log the error
	}

	return &TokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshTokenString,
		ExpiresIn:    int64(s.config.AccessTokenExpiry.Seconds()),
	}, nil
}

// GenerateTokenPair generates both access and refresh tokens for a user
func (s *AuthService) GenerateTokenPair(userID uint) (*TokenPair, error) {
	accessToken, err := s.GenerateAccessToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, _, err := s.GenerateRefreshToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.config.AccessTokenExpiry.Seconds()),
	}, nil
}

