package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
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
func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	config := &Config{
		Env:        getEnv("ENV", "development"),
		DBType:     getEnv("DB_TYPE", "mysql"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPass:     getEnv("DB_PASS", "password"),
		DBName:     getEnv("DB_NAME", "fintrack"),
		JWTSecret:  getEnv("JWT_SECRET", "your-secret-key"),
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
