package config

import (
	"os"
	"strings"
)

type Config struct {
	Env        string
	DBType     string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPass     string
	DBName     string
	JWTSecret  string
	ServerPort string
	CORSOrigins []string // Allowed CORS origins
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

	config := &Config{
		Env:         getEnv("ENV", "development"),
		DBType:      getEnv("DB_TYPE", "mysql"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "3306"),
		DBUser:      getEnv("DB_USER", "root"),
		DBPass:      getEnv("DB_PASS", "password"),
		DBName:      getEnv("DB_NAME", "fintrack"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key"),
		ServerPort:  getEnv("PORT", "8080"),
		CORSOrigins: corsOrigins,
	}

	return config, nil
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.TrimSpace(value)
	}
	return defaultVal
}
