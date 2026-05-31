package config

import (
	"errors"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAPIKeyHashPepper  = "dev-only-api-key-hash-pepper-change-me"
	defaultWebhookSigningKey = "dev-only-webhook-signing-key-change-me"
	defaultFullAccessAPIKey  = "bp_sandbox_full_access_key"
	defaultReadOnlyAPIKey    = "bp_sandbox_readonly_key"
	defaultOtherPartnerKey   = "bp_sandbox_other_partner_key"
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
	MaxRequestBodyBytes int64
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
		MaxRequestBodyBytes: envInt64("MAX_REQUEST_BODY_BYTES", 1<<20),
		APIKeyHashPepper:    env("API_KEY_HASH_PEPPER", defaultAPIKeyHashPepper),
		WebhookSigningKey:   env("WEBHOOK_SIGNING_KEY", defaultWebhookSigningKey),
		FullAccessAPIKey:    env("BANKPORT_FULL_ACCESS_API_KEY", defaultFullAccessAPIKey),
		ReadOnlyAPIKey:      env("BANKPORT_READONLY_API_KEY", defaultReadOnlyAPIKey),
		OtherPartnerAPIKey:  env("BANKPORT_OTHER_PARTNER_API_KEY", defaultOtherPartnerKey),
	}
}

func (c Config) Validate() error {
	var errs []error
	if c.RequestTimeout <= 0 {
		errs = append(errs, errors.New("REQUEST_TIMEOUT must be greater than zero"))
	}
	if c.ShutdownTimeout <= 0 {
		errs = append(errs, errors.New("SHUTDOWN_TIMEOUT must be greater than zero"))
	}
	if c.IdempotencyTTL <= 0 {
		errs = append(errs, errors.New("IDEMPOTENCY_TTL must be greater than zero"))
	}
	if c.DefaultRateLimitRPM < 0 {
		errs = append(errs, errors.New("RATE_LIMIT_PER_MINUTE must not be negative"))
	}
	if c.MaxRequestBodyBytes <= 0 {
		errs = append(errs, errors.New("MAX_REQUEST_BODY_BYTES must be greater than zero"))
	}

	if strings.EqualFold(c.Environment, "production") {
		if weakSecret(c.APIKeyHashPepper, defaultAPIKeyHashPepper) {
			errs = append(errs, errors.New("API_KEY_HASH_PEPPER must be set to a strong non-default value in production"))
		}
		if weakSecret(c.WebhookSigningKey, defaultWebhookSigningKey) {
			errs = append(errs, errors.New("WEBHOOK_SIGNING_KEY must be set to a strong non-default value in production"))
		}
		if isDefaultKey(c.FullAccessAPIKey, defaultFullAccessAPIKey) ||
			isDefaultKey(c.ReadOnlyAPIKey, defaultReadOnlyAPIKey) ||
			isDefaultKey(c.OtherPartnerAPIKey, defaultOtherPartnerKey) {
			errs = append(errs, errors.New("sandbox API keys must not be used in production"))
		}
	}
	return errors.Join(errs...)
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

func envInt64(key string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
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

func weakSecret(value, defaultValue string) bool {
	value = strings.TrimSpace(value)
	return value == "" || value == defaultValue || len(value) < 32
}

func isDefaultKey(value, defaultValue string) bool {
	return strings.TrimSpace(value) == defaultValue
}
