package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
)

// JWT claims structure
type Claims struct {
	UserID uint `json:"userId"`
	jwt.RegisteredClaims
}

// Payloads for registration/login requests
type RegistrationRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"` // can be username OR email
	Password   string `json:"password" binding:"required,min=6,max=100"`
}

// ResetPasswordRequest struct for binding JSON request
type ResetPasswordRequest struct {
	Identifier      string `json:"identifier" binding:"required"` // Can be username or email
	NewPassword     string `json:"newPassword" binding:"required,min=6"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,min=6"`
}

// RegisterHandler handles new user registration.
func RegisterHandler(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)

	var req RegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Invalid registration data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Trim and lowercase the email for consistency
	email := strings.ToLower(strings.TrimSpace(req.Email))
	username := strings.TrimSpace(req.Username)

	// Check if username or email already exists
	var existingUser models.User
	if err := db.DB.Where("username = ? OR email = ?", username, email).First(&existingUser).Error; err == nil {
		// Found a user with the same username or email
		log.Warn("Attempt to register duplicate user", zap.String("username", username), zap.String("email", email))
		c.JSON(http.StatusConflict, gin.H{"error": "Username or email already in use"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Failed to hash password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Create new user
	newUser := models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	if err := db.DB.Create(&newUser).Error; err != nil {
		log.Error("Failed to create user in DB", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}

	log.Info("User registered successfully", zap.Uint("userID", newUser.ID))
	c.JSON(http.StatusCreated, gin.H{
		"message": "User registration successful",
		"user": gin.H{
			"id":       newUser.ID,
			"username": newUser.Username,
			"email":    newUser.Email,
		},
	})
}

// LoginHandler allows login with either username OR email + password.
func LoginHandler(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Invalid login data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	identifier := strings.TrimSpace(req.Identifier)
	password := req.Password

	// Attempt to find user by username OR email
	var user models.User
	if err := db.DB.Where("username = ? OR email = ?", identifier, identifier).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn("Login failed: user not found", zap.String("identifier", identifier))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		log.Error("Database error during login", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		log.Warn("Login failed: invalid password", zap.String("identifier", identifier))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT
	tokenString, err := generateJWT(user.ID, c)
	if err != nil {
		log.Error("Failed to generate JWT", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	log.Info("User logged in successfully", zap.Uint("userID", user.ID))
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   tokenString,
	})
}

// ResetPasswordHandler allows users to reset their password
func ResetPasswordHandler(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Invalid reset password request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Validate password match
	if req.NewPassword != req.ConfirmPassword {
		log.Warn("Password confirmation mismatch")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
		return
	}

	// Search for user by username or email
	var user models.User
	if err := db.DB.Where("username = ? OR email = ?", req.Identifier, req.Identifier).First(&user).Error; err != nil {
		log.Warn("User not found for password reset", zap.String("identifier", req.Identifier))
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Failed to hash new password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	// Update user's password
	if err := db.DB.Model(&user).Update("password_hash", string(hashedPassword)).Error; err != nil {
		log.Error("Failed to update password in database", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update password"})
		return
	}

	log.Info("User password reset successfully", zap.Uint("userID", user.ID))
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
}

// generateJWT creates a JWT token for the given user ID.
func generateJWT(userID uint, c *gin.Context) (string, error) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)

	secret, exists := c.Get("jwtSecret")
	if !exists {
		log.Error("JWT secret not found in context")
		return "", nil
	}
	jwtSecret, ok := secret.(string)
	if !ok {
		log.Error("JWT secret type assertion failed")
		return "", nil
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
