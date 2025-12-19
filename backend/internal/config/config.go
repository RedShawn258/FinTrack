package config

import (
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env                string
	DBType             string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPass             string
	DBName             string
	JWTSecret          string
	RefreshTokenSecret string // Optional separate secret for refresh tokens
	ServerPort         string
	CORSOrigins        []string // Allowed CORS origins
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	RedisAddr          string
	RedisPassword      string
	RedisDB            int
	CacheEnabled       bool // Toggle to disable caching in dev
	// Outbox/Event configuration
	OutboxEnabled      bool          // Enable outbox pattern
	OutboxPollInterval time.Duration // How often to poll for pending events
	OutboxMaxRetries   int           // Maximum retry attempts before marking as FAILED
	EventPublisherType string        // "pubsub" or "webhook"
	PubSubTopic        string        // Google Pub/Sub topic name (if EventPublisherType = "pubsub")
	WebhookURL         string        // HTTP webhook URL (if EventPublisherType = "webhook")
	// Reconciliation configuration
	ReconciliationEnabled  bool          // Enable reconciliation checker
	ReconciliationInterval time.Duration // How often to run reconciliation checks
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

	// Parse database configuration
	// Support DATABASE_URL format: mysql://user:password@host:port/dbname
	// Falls back to individual DB_* env vars if DATABASE_URL is not set
	var dbHost, dbPort, dbUser, dbPass, dbName string
	if databaseURL := getEnv("DATABASE_URL", ""); databaseURL != "" {
		parsedURL, err := url.Parse(databaseURL)
		if err == nil && parsedURL.Scheme == "mysql" {
			dbHost = parsedURL.Hostname()
			if parsedURL.Port() != "" {
				dbPort = parsedURL.Port()
			} else {
				dbPort = "3306"
			}
			if parsedURL.User != nil {
				dbUser = parsedURL.User.Username()
				if password, ok := parsedURL.User.Password(); ok {
					dbPass = password
				}
			}
			dbName = strings.TrimPrefix(parsedURL.Path, "/")
		}
	}
	// Fallback to individual env vars if DATABASE_URL was not set or parsing failed
	if dbHost == "" {
		dbHost = getEnv("DB_HOST", "localhost")
		dbPort = getEnv("DB_PORT", "3306")
		dbUser = getEnv("DB_USER", "root")
		dbPass = getEnv("DB_PASS", "password")
		dbName = getEnv("DB_NAME", "fintrack")
	}

	// Parse Redis configuration
	// Support REDIS_URL format: redis://:password@host:port/dbnumber
	// Falls back to individual REDIS_* env vars if REDIS_URL is not set
	var redisAddr, redisPassword string
	var redisDB int
	if redisURL := getEnv("REDIS_URL", ""); redisURL != "" {
		parsedURL, err := url.Parse(redisURL)
		if err == nil && parsedURL.Scheme == "redis" {
			redisAddr = parsedURL.Host
			if parsedURL.User != nil {
				if password, ok := parsedURL.User.Password(); ok {
					redisPassword = password
				}
			}
			if parsedURL.Path != "" {
				if dbStr := strings.TrimPrefix(parsedURL.Path, "/"); dbStr != "" {
					if parsed, err := strconv.Atoi(dbStr); err == nil {
						redisDB = parsed
					}
				}
			}
		}
	}
	// Fallback to individual env vars if REDIS_URL was not set or parsing failed
	if redisAddr == "" {
		redisAddr = getEnv("REDIS_ADDR", "localhost:6379")
		redisPassword = getEnv("REDIS_PASSWORD", "")
		redisDBStr := getEnv("REDIS_DB", "0")
		if parsed, err := strconv.Atoi(redisDBStr); err == nil {
			redisDB = parsed
		}
	}

	// Cache enabled flag (default true, can disable in dev)
	cacheEnabled := getEnv("CACHE_ENABLED", "true") == "true"

	// Outbox/Event configuration
	outboxEnabled := getEnv("OUTBOX_ENABLED", "true") == "true"
	outboxPollIntervalStr := getEnv("OUTBOX_POLL_INTERVAL", "5s")
	outboxPollInterval, err := time.ParseDuration(outboxPollIntervalStr)
	if err != nil {
		outboxPollInterval = 5 * time.Second // Default fallback
	}
	outboxMaxRetriesStr := getEnv("OUTBOX_MAX_RETRIES", "5")
	outboxMaxRetries, err := strconv.Atoi(outboxMaxRetriesStr)
	if err != nil {
		outboxMaxRetries = 5 // Default fallback
	}
	eventPublisherType := getEnv("EVENT_PUBLISHER_TYPE", "webhook") // Default to webhook for local dev
	pubSubTopic := getEnv("PUBSUB_TOPIC", "")
	webhookURL := getEnv("WEBHOOK_URL", "http://localhost:8081/events")

	// Reconciliation configuration
	reconciliationEnabled := getEnv("RECONCILIATION_ENABLED", "true") == "true"
	reconciliationIntervalStr := getEnv("RECONCILIATION_INTERVAL", "5m")
	reconciliationInterval, err := time.ParseDuration(reconciliationIntervalStr)
	if err != nil {
		reconciliationInterval = 5 * time.Minute // Default fallback
	}

	config := &Config{
		Env:                    getEnv("ENV", "development"),
		DBType:                 getEnv("DB_TYPE", "mysql"),
		DBHost:                 dbHost,
		DBPort:                 dbPort,
		DBUser:                 dbUser,
		DBPass:                 dbPass,
		DBName:                 dbName,
		JWTSecret:              getEnv("JWT_SECRET", "your-secret-key"),
		RefreshTokenSecret:     refreshTokenSecret,
		ServerPort:             getEnv("PORT", "8080"),
		CORSOrigins:            corsOrigins,
		AccessTokenExpiry:      accessTokenExpiry,
		RefreshTokenExpiry:     refreshTokenExpiry,
		RedisAddr:              redisAddr,
		RedisPassword:          redisPassword,
		RedisDB:                redisDB,
		CacheEnabled:           cacheEnabled,
		OutboxEnabled:          outboxEnabled,
		OutboxPollInterval:     outboxPollInterval,
		OutboxMaxRetries:       outboxMaxRetries,
		EventPublisherType:     eventPublisherType,
		PubSubTopic:            pubSubTopic,
		WebhookURL:             webhookURL,
		ReconciliationEnabled:  reconciliationEnabled,
		ReconciliationInterval: reconciliationInterval,
	}

	return config, nil
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.TrimSpace(value)
	}
	return defaultVal
}
