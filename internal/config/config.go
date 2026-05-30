package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment         string
	Port                string
	ServiceName         string
	LogLevel            slog.Level
	RequestTimeout      time.Duration
	ShutdownTimeout     time.Duration
	IdempotencyTTL      time.Duration
	ReadinessEnabled    bool
	DefaultRateLimitRPM int
	APIKeyHashPepper    string
	WebhookSigningKey   string
	FullAccessAPIKey    string
	ReadOnlyAPIKey      string
	OtherPartnerAPIKey  string
}

func Load() Config {
	return Config{
		Environment:         env("BANKPORT_ENV", "development"),
		Port:                env("PORT", "8080"),
		ServiceName:         env("OTEL_SERVICE_NAME", "bankport-partner-api"),
		LogLevel:            parseLogLevel(env("LOG_LEVEL", "info")),
		RequestTimeout:      envDuration("REQUEST_TIMEOUT", 3*time.Second),
		ShutdownTimeout:     envDuration("SHUTDOWN_TIMEOUT", 8*time.Second),
		IdempotencyTTL:      envDuration("IDEMPOTENCY_TTL", 24*time.Hour),
		ReadinessEnabled:    envBool("READINESS_ENABLED", true),
		DefaultRateLimitRPM: envInt("RATE_LIMIT_PER_MINUTE", 120),
		APIKeyHashPepper:    env("API_KEY_HASH_PEPPER", "dev-only-api-key-hash-pepper-change-me"),
		WebhookSigningKey:   env("WEBHOOK_SIGNING_KEY", "dev-only-webhook-signing-key-change-me"),
		FullAccessAPIKey:    env("BANKPORT_FULL_ACCESS_API_KEY", "bp_sandbox_full_access_key"),
		ReadOnlyAPIKey:      env("BANKPORT_READONLY_API_KEY", "bp_sandbox_readonly_key"),
		OtherPartnerAPIKey:  env("BANKPORT_OTHER_PARTNER_API_KEY", "bp_sandbox_other_partner_key"),
	}
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func parseLogLevel(value string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
