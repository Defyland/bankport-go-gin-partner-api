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

	loadErrors []error
}

func Load() Config {
	loader := envLoader{}
	cfg := Config{
		Environment:         env("BANKPORT_ENV", "development"),
		Port:                loader.string("PORT", "8080"),
		ServiceName:         env("OTEL_SERVICE_NAME", "bankport-partner-api"),
		LogLevel:            loader.logLevel("LOG_LEVEL", slog.LevelInfo),
		RequestTimeout:      loader.duration("REQUEST_TIMEOUT", 3*time.Second),
		ShutdownTimeout:     loader.duration("SHUTDOWN_TIMEOUT", 8*time.Second),
		IdempotencyTTL:      loader.duration("IDEMPOTENCY_TTL", 24*time.Hour),
		ReadinessEnabled:    loader.bool("READINESS_ENABLED", true),
		DefaultRateLimitRPM: loader.int("RATE_LIMIT_PER_MINUTE", 120),
		MaxRequestBodyBytes: loader.int64("MAX_REQUEST_BODY_BYTES", 1<<20),
		APIKeyHashPepper:    env("API_KEY_HASH_PEPPER", defaultAPIKeyHashPepper),
		WebhookSigningKey:   env("WEBHOOK_SIGNING_KEY", defaultWebhookSigningKey),
		FullAccessAPIKey:    env("BANKPORT_FULL_ACCESS_API_KEY", defaultFullAccessAPIKey),
		ReadOnlyAPIKey:      env("BANKPORT_READONLY_API_KEY", defaultReadOnlyAPIKey),
		OtherPartnerAPIKey:  env("BANKPORT_OTHER_PARTNER_API_KEY", defaultOtherPartnerKey),
	}
	cfg.loadErrors = loader.errors
	return cfg
}

func (c Config) Validate() error {
	errs := append([]error(nil), c.loadErrors...)
	if !validPort(c.Port) {
		errs = append(errs, errors.New("PORT must be an integer between 1 and 65535"))
	}
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

type envLoader struct {
	errors []error
}

func (l *envLoader) string(key, fallback string) string {
	return env(key, fallback)
}

func (l *envLoader) int(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		l.errors = append(l.errors, invalidEnvError(key, "integer"))
		return fallback
	}
	return parsed
}

func (l *envLoader) int64(key string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		l.errors = append(l.errors, invalidEnvError(key, "integer"))
		return fallback
	}
	return parsed
}

func (l *envLoader) bool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		l.errors = append(l.errors, invalidEnvError(key, "boolean"))
		return fallback
	}
	return parsed
}

func (l *envLoader) duration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		l.errors = append(l.errors, invalidEnvError(key, "duration"))
		return fallback
	}
	return parsed
}

func (l *envLoader) logLevel(key string, fallback slog.Level) slog.Level {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	level, ok := parseLogLevel(value)
	if !ok {
		l.errors = append(l.errors, errors.New(key+" must be one of debug, info, warn, or error"))
		return fallback
	}
	return level
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func parseLogLevel(value string) (slog.Level, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug, true
	case "info":
		return slog.LevelInfo, true
	case "warn", "warning":
		return slog.LevelWarn, true
	case "error":
		return slog.LevelError, true
	default:
		return slog.LevelInfo, false
	}
}

func weakSecret(value, defaultValue string) bool {
	value = strings.TrimSpace(value)
	return value == "" || value == defaultValue || len(value) < 32
}

func isDefaultKey(value, defaultValue string) bool {
	return strings.TrimSpace(value) == defaultValue
}

func invalidEnvError(key, expected string) error {
	return errors.New(key + " must be a valid " + expected)
}

func validPort(value string) bool {
	port, err := strconv.Atoi(strings.TrimSpace(value))
	return err == nil && port >= 1 && port <= 65535
}
