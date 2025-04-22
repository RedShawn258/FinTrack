package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"net/smtp"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
)

type Claims struct {
	UserID uint `json:"userId"`
	jwt.RegisteredClaims
}

type RegistrationRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required,min=6,max=100"`
}

type ResetPasswordRequest struct {
	Identifier      string `json:"identifier" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=6"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,min=6"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func RegisterHandler(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)

	var req RegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Invalid registration data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	username := strings.TrimSpace(req.Username)

	var existingUser models.User
	if err := db.DB.Where("username = ? OR email = ?", username, email).First(&existingUser).Error; err == nil {
		log.Warn("Attempt to register duplicate user", zap.String("username", username), zap.String("email", email))
		c.JSON(http.StatusConflict, gin.H{"error": "Username or email already in use"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Failed to hash password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

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

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		log.Warn("Login failed: invalid password", zap.String("identifier", identifier))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

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

func ResetPasswordHandler(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Invalid reset password request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		log.Warn("Password confirmation mismatch")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
		return
	}

	var user models.User
	if err := db.DB.Where("username = ? OR email = ?", req.Identifier, req.Identifier).First(&user).Error; err != nil {
		log.Warn("User not found for password reset", zap.String("identifier", req.Identifier))
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Failed to hash new password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	if err := db.DB.Model(&user).Update("password_hash", string(hashedPassword)).Error; err != nil {
		log.Error("Failed to update password in database", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update password"})
		return
	}

	log.Info("User password reset successfully", zap.Uint("userID", user.ID))
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
}

func ForgotPasswordHandler(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)

	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Invalid email format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	var user models.User
	if err := db.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// For security, don't reveal if email exists
		c.JSON(http.StatusOK, gin.H{"message": "If an account exists, you will receive a password reset email"})
		return
	}

	// Generate reset token
	resetToken := uuid.New().String()
	expiryTime := time.Now().Add(15 * time.Minute)

	// Save reset token to user
	if err := db.DB.Model(&user).Updates(map[string]interface{}{
		"reset_token":        resetToken,
		"reset_token_expiry": expiryTime,
	}).Error; err != nil {
		log.Error("Failed to save reset token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	// Create reset link
	resetLink := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", resetToken)

	// Send email
	emailFrom := os.Getenv("EMAIL_FROM")
	emailPassword := os.Getenv("EMAIL_PASSWORD")

	// Configure email settings
	smtpHost := "smtp.gmail.com"
	smtpPort := 587

	// Create email message with better formatting
	emailSubject := "FinTrack Password Reset"
	emailBody := fmt.Sprintf(`
		<html>
		<head>
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { padding: 20px; max-width: 600px; margin: 0 auto; }
				.button { 
					display: inline-block; 
					padding: 10px 20px; 
					background-color: #3b82f6; 
					color: white; 
					text-decoration: none; 
					border-radius: 5px; 
					margin: 20px 0;
				}
				.footer { color: #666; font-size: 0.9em; margin-top: 30px; }
			</style>
		</head>
		<body>
			<div class="container">
				<h2>Password Reset Request</h2>
				<p>Hello,</p>
				<p>You have requested to reset your password for your FinTrack account.</p>
				<p>Please click the button below to set a new password:</p>
				<p>
					<a href="%s" class="button">Reset Password</a>
				</p>
				<p>Or copy and paste this link in your browser:</p>
				<p>%s</p>
				<p>This link will expire in 15 minutes for security reasons.</p>
				<div class="footer">
					<p>If you didn't request this password reset, please ignore this email.</p>
					<p>For security, this request was received from your FinTrack account.</p>
				</div>
			</div>
		</body>
		</html>
	`, resetLink, resetLink)

	message := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", req.Email, emailSubject, emailBody))

	// Configure auth and send email
	auth := smtp.PlainAuth("", emailFrom, emailPassword, smtpHost)
	err := smtp.SendMail(fmt.Sprintf("%s:%d", smtpHost, smtpPort), auth, emailFrom, []string{req.Email}, message)
	if err != nil {
		log.Error("Failed to send reset email", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send reset email. Error: " + err.Error()})
		return
	}

	log.Info("Password reset email sent successfully", zap.String("email", req.Email))
	c.JSON(http.StatusOK, gin.H{"message": "Password reset email sent successfully"})
}

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

// GetProfileHandler retrieves the user profile
func GetProfileHandler(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	userID := c.MustGet("userID").(uint)

	// Use the global DB variable
	if db.DB == nil {
		logger.Error("Database connection not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		logger.Error("Failed to retrieve user", zap.Error(err), zap.Uint("userID", userID))
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	profile := models.ProfileResponse{
		Username:             user.Username,
		Email:                user.Email,
		FirstName:            user.FirstName,
		LastName:             user.LastName,
		ProfileImage:         user.ProfileImage,
		PhoneNumber:          user.PhoneNumber,
		Currency:             user.Currency,
		NotificationsEnabled: user.NotificationsEnabled,
		Theme:                user.Theme,
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateProfileRequest defines the request structure for updating a profile
type UpdateProfileRequest struct {
	FirstName            string `json:"firstName,omitempty"`
	LastName             string `json:"lastName,omitempty"`
	PhoneNumber          string `json:"phoneNumber,omitempty"`
	Currency             string `json:"currency,omitempty"`
	NotificationsEnabled *bool  `json:"notificationsEnabled,omitempty"`
	Theme                string `json:"theme,omitempty"`
}

// UpdateProfileHandler updates the user profile
func UpdateProfileHandler(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	userID := c.MustGet("userID").(uint)

	var updateReq UpdateProfileRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		logger.Error("Invalid update profile request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Use the global DB variable
	if db.DB == nil {
		logger.Error("Database connection not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		logger.Error("Failed to retrieve user", zap.Error(err), zap.Uint("userID", userID))
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update only non-empty fields
	if updateReq.FirstName != "" {
		user.FirstName = updateReq.FirstName
	}
	if updateReq.LastName != "" {
		user.LastName = updateReq.LastName
	}
	if updateReq.PhoneNumber != "" {
		user.PhoneNumber = updateReq.PhoneNumber
	}
	if updateReq.Currency != "" {
		user.Currency = updateReq.Currency
	}
	if updateReq.NotificationsEnabled != nil {
		user.NotificationsEnabled = *updateReq.NotificationsEnabled
	}
	if updateReq.Theme != "" {
		user.Theme = updateReq.Theme
	}

	if err := db.DB.Save(&user).Error; err != nil {
		logger.Error("Failed to update user profile", zap.Error(err), zap.Uint("userID", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	profile := models.ProfileResponse{
		Username:             user.Username,
		Email:                user.Email,
		FirstName:            user.FirstName,
		LastName:             user.LastName,
		ProfileImage:         user.ProfileImage,
		PhoneNumber:          user.PhoneNumber,
		Currency:             user.Currency,
		NotificationsEnabled: user.NotificationsEnabled,
		Theme:                user.Theme,
	}

	c.JSON(http.StatusOK, profile)
}

// ProfileImageUploadHandler handles profile image uploads
func ProfileImageUploadHandler(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	userID := c.MustGet("userID").(uint)

	// Get the file from form data
	file, err := c.FormFile("profileImage")
	if err != nil {
		logger.Error("Failed to get profile image", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}

	// Validate file type and size
	if file.Size > 5*1024*1024 { // 5MB limit
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 5MB limit"})
		return
	}

	// Check file extension
	filename := file.Filename
	ext := filepath.Ext(filename)
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	if !allowedExts[strings.ToLower(ext)] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only JPG, PNG and GIF images are allowed"})
		return
	}

	// Create unique filename with user ID
	newFilename := fmt.Sprintf("profile_%d_%s%s", userID, time.Now().Format("20060102150405"), ext)
	uploadPath := "./uploads/profiles/" // Adjust path as needed

	// Create directory if not exists
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		logger.Error("Failed to create upload directory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process image"})
		return
	}

	fullPath := filepath.Join(uploadPath, newFilename)

	// Save the file
	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		logger.Error("Failed to save uploaded file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	// Update the user's profile image field in the database
	if db.DB == nil {
		logger.Error("Database connection not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Get relative path for storage in DB
	relativeImagePath := fmt.Sprintf("/uploads/profiles/%s", newFilename)

	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Update("profile_image", relativeImagePath).Error; err != nil {
		logger.Error("Failed to update profile image in DB", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Profile image uploaded successfully",
		"imageUrl": relativeImagePath,
	})
}
