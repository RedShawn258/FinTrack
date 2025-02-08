package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv" // Optional; if you want to load from a .env file
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
}

// LoadConfig loads environment variables into the Config struct.
// In real production, you'd want better error handling or fallback values.
func LoadConfig() (*Config, error) {
	// Optionally load .env file if present (godotenv)
	_ = godotenv.Load()

	config := &Config{
		Env:        getEnv("ENV", "development"),
		DBType:     getEnv("DB_TYPE", "mysql"), // or "postgres"
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"), // default MySQL port
		DBUser:     getEnv("DB_USER", "root"),
		DBPass:     getEnv("DB_PASS", ""),
		DBName:     getEnv("DB_NAME", "fintrack"),
		JWTSecret:  getEnv("JWT_SECRET", "super-secret-jwt-key"),
		ServerPort: getEnv("PORT", "8080"),
	}

	return config, nil
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.TrimSpace(value)
	}
	return defaultVal
}
