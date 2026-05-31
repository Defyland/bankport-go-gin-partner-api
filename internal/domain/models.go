package domain

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	ScopeAccountsRead  = "accounts:read"
	ScopePixWrite      = "pix:write"
	ScopePayoutsWrite  = "payouts:write"
	ScopeRefundsWrite  = "refunds:write"
	ScopeWebhooksWrite = "webhooks:write"
	ScopeAuditRead     = "audit:read"
	ScopeSandboxRead   = "sandbox:read"
)

const (
	maxDescriptionLength = 280
	maxPixKeyLength      = 140
	maxReasonLength      = 280
)

var supportedWebhookEventTypes = map[string]bool{
	"pix.transfer.created.v1":       true,
	"payout.created.v1":             true,
	"refund.created.v1":             true,
	"webhook.delivery.requested.v1": true,
	"api.rate_limit_exceeded.v1":    true,
}

var (
	ErrAccountNotFound       = errors.New("account not found")
	ErrInsufficientFunds     = errors.New("insufficient available balance")
	ErrValidation            = errors.New("validation failed")
	ErrOriginalTxnNotFound   = errors.New("original transaction not found")
	ErrRefundExceedsOriginal = errors.New("refund exceeds original transaction amount")
)

type SignEventFunc func(event Event, endpoint WebhookEndpoint) string

func ErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrValidation):
		return "validation_failed"
	case errors.Is(err, ErrAccountNotFound):
		return "account_not_found"
	case errors.Is(err, ErrInsufficientFunds):
		return "insufficient_funds"
	case errors.Is(err, ErrOriginalTxnNotFound):
		return "original_transaction_not_found"
	case errors.Is(err, ErrRefundExceedsOriginal):
		return "refund_exceeds_original"
	default:
		return "internal_error"
	}
}

type Partner struct {
	ID                 string
	Name               string
	DeveloperAppID     string
	APIKeyHash         string
	Scopes             map[string]bool
	RateLimitPerMinute int
}

func (p Partner) HasScope(scope string) bool {
	return p.Scopes[scope]
}

type PartnerApp struct {
	PartnerID          string   `json:"partner_id"`
	PartnerName        string   `json:"partner_name"`
	DeveloperAppID     string   `json:"developer_app_id"`
	Scopes             []string `json:"scopes"`
	RateLimitPerMinute int      `json:"rate_limit_per_minute"`
}

type RateLimitPolicy struct {
	PartnerID          string `json:"partner_id"`
	DeveloperAppID     string `json:"developer_app_id"`
	LimitPerMinute     int    `json:"limit_per_minute"`
	PartitionStrategy  string `json:"partition_strategy"`
	DistributedBacking string `json:"distributed_backing"`
}

type UsageReport struct {
	GeneratedAt          time.Time `json:"generated_at"`
	PartnerCount         int       `json:"partner_count"`
	DeveloperAppCount    int       `json:"developer_app_count"`
	AccountCount         int       `json:"account_count"`
	PixTransferCount     int       `json:"pix_transfer_count"`
	PayoutCount          int       `json:"payout_count"`
	RefundCount          int       `json:"refund_count"`
	EventCount           int       `json:"event_count"`
	WebhookEndpointCount int       `json:"webhook_endpoint_count"`
	WebhookDeliveryCount int       `json:"webhook_delivery_count"`
	AuditEntryCount      int       `json:"audit_entry_count"`
}

type Account struct {
	ID                  string    `json:"account_id"`
	PartnerID           string    `json:"partner_id"`
	Currency            string    `json:"currency"`
	AvailableBalanceCts int64     `json:"available_balance_cents"`
	PendingBalanceCts   int64     `json:"pending_balance_cents"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type StatementEntry struct {
	ID          string    `json:"id"`
	AccountID   string    `json:"account_id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	AmountCents int64     `json:"amount_cents"`
	Currency    string    `json:"currency"`
	OccurredAt  time.Time `json:"occurred_at"`
}

type PixTransferRequest struct {
	SourceAccountID string `json:"source_account_id"`
	AmountCents     int64  `json:"amount_cents"`
	Currency        string `json:"currency"`
	PixKey          string `json:"pix_key"`
	Description     string `json:"description"`
}

func (r PixTransferRequest) Validate() error {
	if strings.TrimSpace(r.SourceAccountID) == "" {
		return fieldError("source_account_id", "is required")
	}
	if r.AmountCents <= 0 {
		return fieldError("amount_cents", "must be greater than zero")
	}
	if normalizeCurrency(r.Currency) != "BRL" {
		return fieldError("currency", "must be BRL")
	}
	if strings.TrimSpace(r.PixKey) == "" {
		return fieldError("pix_key", "is required")
	}
	if len(strings.TrimSpace(r.PixKey)) > maxPixKeyLength {
		return fieldError("pix_key", "must be at most 140 characters")
	}
	if len(strings.TrimSpace(r.Description)) > maxDescriptionLength {
		return fieldError("description", "must be at most 280 characters")
	}
	return nil
}

type PixTransfer struct {
	ID              string    `json:"id"`
	PartnerID       string    `json:"partner_id"`
	SourceAccountID string    `json:"source_account_id"`
	AmountCents     int64     `json:"amount_cents"`
	Currency        string    `json:"currency"`
	PixKey          string    `json:"pix_key"`
	Status          string    `json:"status"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
}

type PayoutRequest struct {
	AccountID     string `json:"account_id"`
	AmountCents   int64  `json:"amount_cents"`
	Currency      string `json:"currency"`
	BankCode      string `json:"bank_code"`
	Branch        string `json:"branch"`
	AccountNumber string `json:"account_number"`
	Document      string `json:"document"`
	Description   string `json:"description"`
}

func (r PayoutRequest) Validate() error {
	if strings.TrimSpace(r.AccountID) == "" {
		return fieldError("account_id", "is required")
	}
	if r.AmountCents <= 0 {
		return fieldError("amount_cents", "must be greater than zero")
	}
	if normalizeCurrency(r.Currency) != "BRL" {
		return fieldError("currency", "must be BRL")
	}
	if !fixedDigits(r.BankCode, 3) {
		return fieldError("bank_code", "must contain exactly 3 digits")
	}
	if !digitsBetween(r.Branch, 1, 6) {
		return fieldError("branch", "must contain 1 to 6 digits")
	}
	if !accountNumber(r.AccountNumber) {
		return fieldError("account_number", "must contain 1 to 20 digits, hyphens, or uppercase X")
	}
	if !documentNumber(r.Document) {
		return fieldError("document", "must contain 11 or 14 digits")
	}
	if len(strings.TrimSpace(r.Description)) > maxDescriptionLength {
		return fieldError("description", "must be at most 280 characters")
	}
	return nil
}

type Payout struct {
	ID            string    `json:"id"`
	PartnerID     string    `json:"partner_id"`
	AccountID     string    `json:"account_id"`
	AmountCents   int64     `json:"amount_cents"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	BankCode      string    `json:"bank_code"`
	Branch        string    `json:"branch"`
	AccountNumber string    `json:"account_number"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

type RefundRequest struct {
	OriginalTransactionID string `json:"original_transaction_id"`
	AccountID             string `json:"account_id"`
	AmountCents           int64  `json:"amount_cents"`
	Currency              string `json:"currency"`
	Reason                string `json:"reason"`
}

func (r RefundRequest) Validate() error {
	if strings.TrimSpace(r.OriginalTransactionID) == "" {
		return fieldError("original_transaction_id", "is required")
	}
	if strings.TrimSpace(r.AccountID) == "" {
		return fieldError("account_id", "is required")
	}
	if r.AmountCents <= 0 {
		return fieldError("amount_cents", "must be greater than zero")
	}
	if normalizeCurrency(r.Currency) != "BRL" {
		return fieldError("currency", "must be BRL")
	}
	if strings.TrimSpace(r.Reason) == "" {
		return fieldError("reason", "is required")
	}
	if len(strings.TrimSpace(r.Reason)) > maxReasonLength {
		return fieldError("reason", "must be at most 280 characters")
	}
	return nil
}

type Refund struct {
	ID                    string    `json:"id"`
	PartnerID             string    `json:"partner_id"`
	AccountID             string    `json:"account_id"`
	OriginalTransactionID string    `json:"original_transaction_id"`
	AmountCents           int64     `json:"amount_cents"`
	Currency              string    `json:"currency"`
	Status                string    `json:"status"`
	Reason                string    `json:"reason"`
	CreatedAt             time.Time `json:"created_at"`
}

type WebhookEndpointRequest struct {
	URL         string   `json:"url"`
	EventTypes  []string `json:"event_types"`
	Description string   `json:"description"`
}

func (r WebhookEndpointRequest) Validate() error {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(r.URL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fieldError("url", "must be an absolute URL")
	}
	if parsed.User != nil {
		return fieldError("url", "must not include user info")
	}
	if parsed.Scheme != "https" && !isLocalWebhookHost(parsed.Hostname()) {
		return fieldError("url", "must use https outside localhost")
	}
	if len(r.EventTypes) == 0 {
		return fieldError("event_types", "must contain at least one event type")
	}
	if len(strings.TrimSpace(r.Description)) > maxDescriptionLength {
		return fieldError("description", "must be at most 280 characters")
	}
	for _, eventType := range r.EventTypes {
		normalized := strings.TrimSpace(eventType)
		if normalized == "" {
			return fieldError("event_types", "must not contain blank values")
		}
		if normalized != "*" && !supportedWebhookEventTypes[normalized] {
			return fieldError("event_types", "contains unsupported event type "+normalized)
		}
	}
	return nil
}

type WebhookEndpoint struct {
	ID          string    `json:"id"`
	PartnerID   string    `json:"partner_id"`
	URL         string    `json:"url"`
	EventTypes  []string  `json:"event_types"`
	Description string    `json:"description"`
	SecretID    string    `json:"secret_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type Event struct {
	ID             string         `json:"event_id"`
	Type           string         `json:"event_type"`
	SchemaVersion  string         `json:"schema_version"`
	OccurredAt     time.Time      `json:"occurred_at"`
	Producer       string         `json:"producer"`
	PartnerID      string         `json:"partner_id"`
	DeveloperAppID string         `json:"developer_app_id"`
	CorrelationID  string         `json:"correlation_id"`
	Payload        map[string]any `json:"payload"`
}

type WebhookDelivery struct {
	ID           string    `json:"id"`
	EndpointID   string    `json:"endpoint_id"`
	EventID      string    `json:"event_id"`
	PartnerID    string    `json:"partner_id"`
	Status       string    `json:"status"`
	Signature    string    `json:"signature"`
	NextAttempt  time.Time `json:"next_attempt_at"`
	AttemptCount int       `json:"attempt_count"`
	CreatedAt    time.Time `json:"created_at"`
}

type AuditEntry struct {
	ID            string    `json:"id"`
	PartnerID     string    `json:"partner_id"`
	RequestID     string    `json:"request_id"`
	CorrelationID string    `json:"correlation_id"`
	Action        string    `json:"action"`
	ResourceID    string    `json:"resource_id"`
	Status        string    `json:"status"`
	Reason        string    `json:"reason,omitempty"`
	OccurredAt    time.Time `json:"occurred_at"`
}

type SandboxScenario struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func normalizeCurrency(currency string) string {
	return strings.ToUpper(strings.TrimSpace(currency))
}

func isLocalWebhookHost(host string) bool {
	switch strings.ToLower(strings.TrimSpace(host)) {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}

func fixedDigits(value string, length int) bool {
	value = strings.TrimSpace(value)
	return len(value) == length && onlyDigits(value)
}

func digitsBetween(value string, minLength, maxLength int) bool {
	value = strings.TrimSpace(value)
	return len(value) >= minLength && len(value) <= maxLength && onlyDigits(value)
}

func documentNumber(value string) bool {
	value = strings.TrimSpace(value)
	return (len(value) == 11 || len(value) == 14) && onlyDigits(value)
}

func onlyDigits(value string) bool {
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return value != ""
}

func accountNumber(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) < 1 || len(value) > 20 {
		return false
	}
	for _, char := range value {
		switch {
		case char >= '0' && char <= '9':
			continue
		case char == '-':
			continue
		case char == 'X':
			continue
		default:
			return false
		}
	}
	return true
}

func fieldError(field, message string) error {
	return fmt.Errorf("%w: %s %s", ErrValidation, field, message)
}
