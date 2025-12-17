package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env               string
	DBType            string
	DBHost            string
	DBPort            string
	DBUser            string
	DBPass            string
	DBName            string
	JWTSecret         string
	RefreshTokenSecret string // Optional separate secret for refresh tokens
	ServerPort        string
	CORSOrigins       []string // Allowed CORS origins
	AccessTokenExpiry time.Duration
	RefreshTokenExpiry time.Duration
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	CacheEnabled      bool // Toggle to disable caching in dev
}

// LoadConfig loads environment variables into the Config struct.
// Note: godotenv.Load() should be called in main.go before calling this function
func LoadConfig() (*Config, error) {
	corsOriginsStr := getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:3001")
	corsOrigins := strings.Split(corsOriginsStr, ",")
	// Trim whitespace from each origin
	for i, origin := range corsOrigins {
		corsOrigins[i] = strings.TrimSpace(origin)
	}

	// Parse token expiration durations
	accessTokenExpiry, err := time.ParseDuration(getEnv("ACCESS_TOKEN_EXPIRY", "15m"))
	if err != nil {
		accessTokenExpiry = 15 * time.Minute // Default fallback
	}

	refreshTokenExpiry, err := time.ParseDuration(getEnv("REFRESH_TOKEN_EXPIRY", "168h"))
	if err != nil {
		refreshTokenExpiry = 7 * 24 * time.Hour // Default fallback (7 days)
	}

	// Use separate refresh token secret if provided, otherwise use JWT secret
	refreshTokenSecret := getEnv("REFRESH_TOKEN_SECRET", "")
	if refreshTokenSecret == "" {
		refreshTokenSecret = getEnv("JWT_SECRET", "your-secret-key")
	}

	// Parse Redis DB number
	redisDB := 0
	if redisDBStr := getEnv("REDIS_DB", "0"); redisDBStr != "" {
		if parsed, err := strconv.Atoi(redisDBStr); err == nil {
			redisDB = parsed
		}
	}

	// Cache enabled flag (default true, can disable in dev)
	cacheEnabled := getEnv("CACHE_ENABLED", "true") == "true"

	config := &Config{
		Env:                getEnv("ENV", "development"),
		DBType:             getEnv("DB_TYPE", "mysql"),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "3306"),
		DBUser:             getEnv("DB_USER", "root"),
		DBPass:             getEnv("DB_PASS", "password"),
		DBName:             getEnv("DB_NAME", "fintrack"),
		JWTSecret:          getEnv("JWT_SECRET", "your-secret-key"),
		RefreshTokenSecret: refreshTokenSecret,
		ServerPort:         getEnv("PORT", "8080"),
		CORSOrigins:        corsOrigins,
		AccessTokenExpiry:  accessTokenExpiry,
		RefreshTokenExpiry: refreshTokenExpiry,
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisDB:            redisDB,
		CacheEnabled:       cacheEnabled,
	}

	return config, nil
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.TrimSpace(value)
	}
	return defaultVal
}
