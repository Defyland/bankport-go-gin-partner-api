package config

import (
	"strings"
	"testing"
)

func TestValidateRejectsProductionDefaults(t *testing.T) {
	t.Setenv("BANKPORT_ENV", "production")

	cfg := Load()

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected production config with sandbox defaults to be rejected")
	}
}

func TestValidateAcceptsProductionSecretsAndKeys(t *testing.T) {
	t.Setenv("BANKPORT_ENV", "production")
	t.Setenv("API_KEY_HASH_PEPPER", "0123456789abcdef0123456789abcdef")
	t.Setenv("WEBHOOK_SIGNING_KEY", "abcdef0123456789abcdef0123456789")
	t.Setenv("BANKPORT_FULL_ACCESS_API_KEY", "bp_live_full_access_key_for_validation")
	t.Setenv("BANKPORT_READONLY_API_KEY", "bp_live_readonly_key_for_validation")
	t.Setenv("BANKPORT_OTHER_PARTNER_API_KEY", "bp_live_other_partner_key_for_validation")

	cfg := Load()

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected production config with non-default secrets to pass: %v", err)
	}
}

func TestValidateRejectsInvalidOperationalLimits(t *testing.T) {
	cfg := Load()
	cfg.MaxRequestBodyBytes = 0

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid max request body size to be rejected")
	}
}

func TestValidateRejectsMalformedEnvironmentValues(t *testing.T) {
	t.Setenv("PORT", "eighty")
	t.Setenv("LOG_LEVEL", "verbose")
	t.Setenv("REQUEST_TIMEOUT", "fast")
	t.Setenv("READINESS_ENABLED", "sometimes")
	t.Setenv("RATE_LIMIT_PER_MINUTE", "many")
	t.Setenv("MAX_REQUEST_BODY_BYTES", "huge")

	cfg := Load()
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected malformed environment values to be rejected")
	}
	for _, expected := range []string{
		"PORT",
		"LOG_LEVEL",
		"REQUEST_TIMEOUT",
		"READINESS_ENABLED",
		"RATE_LIMIT_PER_MINUTE",
		"MAX_REQUEST_BODY_BYTES",
	} {
		if !strings.Contains(err.Error(), expected) {
			t.Fatalf("expected validation error to mention %s, got %v", expected, err)
		}
	}
}
