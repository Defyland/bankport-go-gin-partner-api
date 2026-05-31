package store

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/config"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/domain"
)

var (
	randomRead      = rand.Read
	fallbackCounter atomic.Uint64
)

type Repository struct {
	mu sync.RWMutex

	apiKeyHashPepper     string
	partnersByAPIKeyHash map[string]domain.Partner
	accounts             map[string]domain.Account
	statements           map[string][]domain.StatementEntry
	pixTransfers         map[string]domain.PixTransfer
	payouts              map[string]domain.Payout
	refunds              map[string]domain.Refund
	refundedByTransferID map[string]int64
	webhookEndpoints     map[string]domain.WebhookEndpoint
	events               []domain.Event
	webhookDeliveries    []domain.WebhookDelivery
	auditEntries         []domain.AuditEntry
	sandboxScenarios     []domain.SandboxScenario
}

func NewSeededRepository(cfg config.Config) *Repository {
	now := time.Now().UTC()
	repo := &Repository{
		apiKeyHashPepper:     cfg.APIKeyHashPepper,
		partnersByAPIKeyHash: make(map[string]domain.Partner),
		accounts:             make(map[string]domain.Account),
		statements:           make(map[string][]domain.StatementEntry),
		pixTransfers:         make(map[string]domain.PixTransfer),
		payouts:              make(map[string]domain.Payout),
		refunds:              make(map[string]domain.Refund),
		refundedByTransferID: make(map[string]int64),
		webhookEndpoints:     make(map[string]domain.WebhookEndpoint),
		sandboxScenarios: []domain.SandboxScenario{
			{ID: "scenario_pix_success", Name: "Pix transfer accepted", Description: "Financial write commands are accepted and queued for webhook delivery."},
			{ID: "scenario_low_balance", Name: "Insufficient balance", Description: "Requests above the available account balance fail with a deterministic error envelope."},
			{ID: "scenario_rate_limited", Name: "Rate limit exceeded", Description: "A partner crossing the per-minute budget receives HTTP 429 and retry metadata."},
		},
	}

	repo.addPartner(domain.Partner{
		ID:                 "partner_sandbox_bank",
		Name:               "Sandbox Bank",
		DeveloperAppID:     "app_sandbox_full",
		APIKeyHash:         HashAPIKey(cfg.FullAccessAPIKey, cfg.APIKeyHashPepper),
		RateLimitPerMinute: cfg.DefaultRateLimitRPM,
		Scopes: scopeSet(
			domain.ScopeAccountsRead,
			domain.ScopePixWrite,
			domain.ScopePayoutsWrite,
			domain.ScopeRefundsWrite,
			domain.ScopeWebhooksWrite,
			domain.ScopeAuditRead,
			domain.ScopeSandboxRead,
		),
	})
	repo.addPartner(domain.Partner{
		ID:                 "partner_sandbox_bank",
		Name:               "Sandbox Bank Read Only",
		DeveloperAppID:     "app_sandbox_readonly",
		APIKeyHash:         HashAPIKey(cfg.ReadOnlyAPIKey, cfg.APIKeyHashPepper),
		RateLimitPerMinute: cfg.DefaultRateLimitRPM,
		Scopes:             scopeSet(domain.ScopeAccountsRead, domain.ScopeSandboxRead),
	})
	repo.addPartner(domain.Partner{
		ID:                 "partner_other_bank",
		Name:               "Other Partner Bank",
		DeveloperAppID:     "app_other_partner",
		APIKeyHash:         HashAPIKey(cfg.OtherPartnerAPIKey, cfg.APIKeyHashPepper),
		RateLimitPerMinute: cfg.DefaultRateLimitRPM,
		Scopes:             scopeSet(domain.ScopeAccountsRead, domain.ScopePixWrite, domain.ScopeSandboxRead),
	})

	repo.accounts["acct_sandbox_001"] = domain.Account{
		ID:                  "acct_sandbox_001",
		PartnerID:           "partner_sandbox_bank",
		Currency:            "BRL",
		AvailableBalanceCts: 2_500_000,
		PendingBalanceCts:   150_000,
		UpdatedAt:           now,
	}
	repo.accounts["acct_sandbox_002"] = domain.Account{
		ID:                  "acct_sandbox_002",
		PartnerID:           "partner_sandbox_bank",
		Currency:            "BRL",
		AvailableBalanceCts: 750_000,
		PendingBalanceCts:   0,
		UpdatedAt:           now,
	}
	repo.accounts["acct_other_001"] = domain.Account{
		ID:                  "acct_other_001",
		PartnerID:           "partner_other_bank",
		Currency:            "BRL",
		AvailableBalanceCts: 1_000_000,
		PendingBalanceCts:   0,
		UpdatedAt:           now,
	}
	repo.statements["acct_sandbox_001"] = []domain.StatementEntry{
		{ID: "stmt_001", AccountID: "acct_sandbox_001", Type: "credit", Description: "Sandbox opening balance", AmountCents: 2_500_000, Currency: "BRL", OccurredAt: now.Add(-48 * time.Hour)},
		{ID: "stmt_002", AccountID: "acct_sandbox_001", Type: "debit", Description: "Card settlement fee", AmountCents: -3500, Currency: "BRL", OccurredAt: now.Add(-24 * time.Hour)},
	}
	return repo
}

func HashAPIKey(apiKey, pepper string) string {
	mac := hmac.New(sha256.New, []byte(pepper))
	_, _ = mac.Write([]byte(apiKey))
	return hex.EncodeToString(mac.Sum(nil))
}

func (r *Repository) AuthenticateAPIKey(apiKey string) (domain.Partner, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	partner, ok := r.partnersByAPIKeyHash[HashAPIKey(strings.TrimSpace(apiKey), r.apiKeyHashPepper)]
	return partner, ok
}

func (r *Repository) PartnerApps(ctx context.Context) ([]domain.PartnerApp, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	apps := make([]domain.PartnerApp, 0, len(r.partnersByAPIKeyHash))
	for _, partner := range r.partnersByAPIKeyHash {
		apps = append(apps, domain.PartnerApp{
			PartnerID:          partner.ID,
			PartnerName:        partner.Name,
			DeveloperAppID:     partner.DeveloperAppID,
			Scopes:             sortedScopes(partner.Scopes),
			RateLimitPerMinute: partner.RateLimitPerMinute,
		})
	}
	sort.Slice(apps, func(i, j int) bool {
		return apps[i].DeveloperAppID < apps[j].DeveloperAppID
	})
	return apps, nil
}

func (r *Repository) RateLimitPolicies(ctx context.Context) ([]domain.RateLimitPolicy, error) {
	apps, err := r.PartnerApps(ctx)
	if err != nil {
		return nil, err
	}

	policies := make([]domain.RateLimitPolicy, 0, len(apps))
	for _, app := range apps {
		policies = append(policies, domain.RateLimitPolicy{
			PartnerID:          app.PartnerID,
			DeveloperAppID:     app.DeveloperAppID,
			LimitPerMinute:     app.RateLimitPerMinute,
			PartitionStrategy:  "partner_id + route_pattern + fixed_1m_window",
			DistributedBacking: "in_memory_sandbox",
		})
	}
	return policies, nil
}

func (r *Repository) UsageReport(ctx context.Context) (domain.UsageReport, error) {
	if err := ctx.Err(); err != nil {
		return domain.UsageReport{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := ctx.Err(); err != nil {
		return domain.UsageReport{}, err
	}
	partnerIDs := make(map[string]bool)
	for _, partner := range r.partnersByAPIKeyHash {
		partnerIDs[partner.ID] = true
	}
	return domain.UsageReport{
		GeneratedAt:          time.Now().UTC(),
		PartnerCount:         len(partnerIDs),
		DeveloperAppCount:    len(r.partnersByAPIKeyHash),
		AccountCount:         len(r.accounts),
		PixTransferCount:     len(r.pixTransfers),
		PayoutCount:          len(r.payouts),
		RefundCount:          len(r.refunds),
		EventCount:           len(r.events),
		WebhookEndpointCount: len(r.webhookEndpoints),
		WebhookDeliveryCount: len(r.webhookDeliveries),
		AuditEntryCount:      len(r.auditEntries),
	}, nil
}

func (r *Repository) GetAccount(ctx context.Context, partnerID, accountID string) (domain.Account, error) {
	if err := ctx.Err(); err != nil {
		return domain.Account{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := ctx.Err(); err != nil {
		return domain.Account{}, err
	}
	account, ok := r.accounts[accountID]
	if !ok || account.PartnerID != partnerID {
		return domain.Account{}, domain.ErrAccountNotFound
	}
	return account, nil
}

func (r *Repository) ListStatements(ctx context.Context, partnerID, accountID string) ([]domain.StatementEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	account, ok := r.accounts[accountID]
	if !ok || account.PartnerID != partnerID {
		return nil, domain.ErrAccountNotFound
	}
	entries := append([]domain.StatementEntry(nil), r.statements[accountID]...)
	return entries, nil
}

func (r *Repository) CreatePixTransfer(ctx context.Context, partner domain.Partner, request domain.PixTransferRequest, correlationID string, sign SignEventFunc) (domain.PixTransfer, int, error) {
	if err := ctx.Err(); err != nil {
		return domain.PixTransfer{}, 0, err
	}
	if err := request.Validate(); err != nil {
		return domain.PixTransfer{}, 0, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return domain.PixTransfer{}, 0, err
	}
	account, ok := r.accounts[request.SourceAccountID]
	if !ok || account.PartnerID != partner.ID {
		return domain.PixTransfer{}, 0, domain.ErrAccountNotFound
	}
	if account.AvailableBalanceCts < request.AmountCents {
		return domain.PixTransfer{}, 0, domain.ErrInsufficientFunds
	}

	now := time.Now().UTC()
	account.AvailableBalanceCts -= request.AmountCents
	account.UpdatedAt = now
	r.accounts[account.ID] = account

	transfer := domain.PixTransfer{
		ID:              newID("pix"),
		PartnerID:       partner.ID,
		SourceAccountID: account.ID,
		AmountCents:     request.AmountCents,
		Currency:        "BRL",
		PixKey:          strings.TrimSpace(request.PixKey),
		Status:          "accepted",
		Description:     strings.TrimSpace(request.Description),
		CreatedAt:       now,
	}
	r.pixTransfers[transfer.ID] = transfer
	r.statements[account.ID] = append(r.statements[account.ID], domain.StatementEntry{
		ID:          newID("stmt"),
		AccountID:   account.ID,
		Type:        "debit",
		Description: "Pix transfer " + transfer.ID,
		AmountCents: -request.AmountCents,
		Currency:    "BRL",
		OccurredAt:  now,
	})

	event := r.appendEventLocked(partner, correlationID, "pix.transfer.created.v1", map[string]any{
		"transfer_id":       transfer.ID,
		"source_account_id": transfer.SourceAccountID,
		"amount_cents":      transfer.AmountCents,
		"currency":          transfer.Currency,
		"status":            transfer.Status,
	})
	deliveries := r.queueDeliveriesLocked(ctx, partner.ID, event, sign)
	return transfer, deliveries, nil
}

func (r *Repository) CreatePayout(ctx context.Context, partner domain.Partner, request domain.PayoutRequest, correlationID string, sign SignEventFunc) (domain.Payout, int, error) {
	if err := ctx.Err(); err != nil {
		return domain.Payout{}, 0, err
	}
	if err := request.Validate(); err != nil {
		return domain.Payout{}, 0, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return domain.Payout{}, 0, err
	}
	account, ok := r.accounts[request.AccountID]
	if !ok || account.PartnerID != partner.ID {
		return domain.Payout{}, 0, domain.ErrAccountNotFound
	}
	if account.AvailableBalanceCts < request.AmountCents {
		return domain.Payout{}, 0, domain.ErrInsufficientFunds
	}

	now := time.Now().UTC()
	account.AvailableBalanceCts -= request.AmountCents
	account.UpdatedAt = now
	r.accounts[account.ID] = account

	payout := domain.Payout{
		ID:            newID("payout"),
		PartnerID:     partner.ID,
		AccountID:     account.ID,
		AmountCents:   request.AmountCents,
		Currency:      "BRL",
		Status:        "queued",
		BankCode:      strings.TrimSpace(request.BankCode),
		Branch:        strings.TrimSpace(request.Branch),
		AccountNumber: strings.TrimSpace(request.AccountNumber),
		Description:   strings.TrimSpace(request.Description),
		CreatedAt:     now,
	}
	r.payouts[payout.ID] = payout
	r.statements[account.ID] = append(r.statements[account.ID], domain.StatementEntry{
		ID:          newID("stmt"),
		AccountID:   account.ID,
		Type:        "debit",
		Description: "Payout " + payout.ID,
		AmountCents: -request.AmountCents,
		Currency:    "BRL",
		OccurredAt:  now,
	})

	event := r.appendEventLocked(partner, correlationID, "payout.created.v1", map[string]any{
		"payout_id":    payout.ID,
		"account_id":   payout.AccountID,
		"amount_cents": payout.AmountCents,
		"currency":     payout.Currency,
		"status":       payout.Status,
	})
	deliveries := r.queueDeliveriesLocked(ctx, partner.ID, event, sign)
	return payout, deliveries, nil
}

func (r *Repository) CreateRefund(ctx context.Context, partner domain.Partner, request domain.RefundRequest, correlationID string, sign SignEventFunc) (domain.Refund, int, error) {
	if err := ctx.Err(); err != nil {
		return domain.Refund{}, 0, err
	}
	if err := request.Validate(); err != nil {
		return domain.Refund{}, 0, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return domain.Refund{}, 0, err
	}
	account, ok := r.accounts[request.AccountID]
	if !ok || account.PartnerID != partner.ID {
		return domain.Refund{}, 0, domain.ErrAccountNotFound
	}
	original, ok := r.pixTransfers[request.OriginalTransactionID]
	if !ok || original.PartnerID != partner.ID || original.SourceAccountID != request.AccountID {
		return domain.Refund{}, 0, domain.ErrOriginalTxnNotFound
	}
	refundedAmount := r.refundedByTransferID[original.ID]
	if refundedAmount+request.AmountCents > original.AmountCents {
		return domain.Refund{}, 0, domain.ErrRefundExceedsOriginal
	}
	remainingAfterRefund := original.AmountCents - refundedAmount - request.AmountCents

	now := time.Now().UTC()
	account.AvailableBalanceCts += request.AmountCents
	account.UpdatedAt = now
	r.accounts[account.ID] = account
	r.refundedByTransferID[original.ID] = refundedAmount + request.AmountCents

	refund := domain.Refund{
		ID:                    newID("refund"),
		PartnerID:             partner.ID,
		AccountID:             account.ID,
		OriginalTransactionID: original.ID,
		AmountCents:           request.AmountCents,
		Currency:              "BRL",
		Status:                "accepted",
		Reason:                strings.TrimSpace(request.Reason),
		CreatedAt:             now,
	}
	r.refunds[refund.ID] = refund
	r.statements[account.ID] = append(r.statements[account.ID], domain.StatementEntry{
		ID:          newID("stmt"),
		AccountID:   account.ID,
		Type:        "credit",
		Description: "Refund " + refund.ID,
		AmountCents: request.AmountCents,
		Currency:    "BRL",
		OccurredAt:  now,
	})

	event := r.appendEventLocked(partner, correlationID, "refund.created.v1", map[string]any{
		"refund_id":                         refund.ID,
		"original_transaction_id":           refund.OriginalTransactionID,
		"account_id":                        refund.AccountID,
		"amount_cents":                      refund.AmountCents,
		"currency":                          refund.Currency,
		"status":                            refund.Status,
		"remaining_refundable_amount_cents": remainingAfterRefund,
	})
	deliveries := r.queueDeliveriesLocked(ctx, partner.ID, event, sign)
	return refund, deliveries, nil
}

func (r *Repository) RegisterWebhookEndpoint(ctx context.Context, partner domain.Partner, request domain.WebhookEndpointRequest) (domain.WebhookEndpoint, error) {
	if err := ctx.Err(); err != nil {
		return domain.WebhookEndpoint{}, err
	}
	if err := request.Validate(); err != nil {
		return domain.WebhookEndpoint{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return domain.WebhookEndpoint{}, err
	}
	endpoint := domain.WebhookEndpoint{
		ID:          newID("wh"),
		PartnerID:   partner.ID,
		URL:         strings.TrimSpace(request.URL),
		EventTypes:  append([]string(nil), request.EventTypes...),
		Description: strings.TrimSpace(request.Description),
		SecretID:    "whsec_" + newToken(10),
		CreatedAt:   time.Now().UTC(),
	}
	r.webhookEndpoints[endpoint.ID] = endpoint
	return endpoint, nil
}

func (r *Repository) ListAuditEntries(_ context.Context, partnerID string) []domain.AuditEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := make([]domain.AuditEntry, 0, len(r.auditEntries))
	for _, entry := range r.auditEntries {
		if entry.PartnerID == partnerID {
			entries = append(entries, entry)
		}
	}
	return entries
}

func (r *Repository) AddAuditEntry(entry domain.AuditEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry.ID = newID("audit")
	entry.OccurredAt = time.Now().UTC()
	r.auditEntries = append(r.auditEntries, entry)
}

func (r *Repository) SandboxScenarios() []domain.SandboxScenario {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return append([]domain.SandboxScenario(nil), r.sandboxScenarios...)
}

func (r *Repository) Events() []domain.Event {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return append([]domain.Event(nil), r.events...)
}

func (r *Repository) WebhookDeliveries() []domain.WebhookDelivery {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return append([]domain.WebhookDelivery(nil), r.webhookDeliveries...)
}

func (r *Repository) addPartner(partner domain.Partner) {
	r.partnersByAPIKeyHash[partner.APIKeyHash] = partner
}

func (r *Repository) appendEventLocked(partner domain.Partner, correlationID, eventType string, payload map[string]any) domain.Event {
	event := domain.Event{
		ID:             newID("evt"),
		Type:           eventType,
		SchemaVersion:  "1.0.0",
		OccurredAt:     time.Now().UTC(),
		Producer:       "bankport-partner-api",
		PartnerID:      partner.ID,
		DeveloperAppID: partner.DeveloperAppID,
		CorrelationID:  correlationID,
		Payload:        payload,
	}
	r.events = append(r.events, event)
	return event
}

func (r *Repository) queueDeliveriesLocked(ctx context.Context, partnerID string, event domain.Event, sign SignEventFunc) int {
	queued := 0
	for _, endpoint := range r.webhookEndpoints {
		if ctx.Err() != nil {
			return queued
		}
		if endpoint.PartnerID != partnerID || !endpointAccepts(endpoint, event.Type) {
			continue
		}
		delivery := domain.WebhookDelivery{
			ID:           newID("whd"),
			EndpointID:   endpoint.ID,
			EventID:      event.ID,
			PartnerID:    partnerID,
			Status:       "queued",
			Signature:    sign(event, endpoint),
			NextAttempt:  time.Now().UTC(),
			AttemptCount: 0,
			CreatedAt:    time.Now().UTC(),
		}
		r.webhookDeliveries = append(r.webhookDeliveries, delivery)
		r.events = append(r.events, domain.Event{
			ID:             newID("evt"),
			Type:           "webhook.delivery.requested.v1",
			SchemaVersion:  "1.0.0",
			OccurredAt:     time.Now().UTC(),
			Producer:       "bankport-partner-api",
			PartnerID:      partnerID,
			DeveloperAppID: event.DeveloperAppID,
			CorrelationID:  event.CorrelationID,
			Payload: map[string]any{
				"delivery_id": delivery.ID,
				"endpoint_id": endpoint.ID,
				"event_id":    event.ID,
				"status":      delivery.Status,
			},
		})
		queued++
	}
	return queued
}

func endpointAccepts(endpoint domain.WebhookEndpoint, eventType string) bool {
	for _, accepted := range endpoint.EventTypes {
		if accepted == eventType || accepted == "*" {
			return true
		}
	}
	return false
}

func scopeSet(scopes ...string) map[string]bool {
	set := make(map[string]bool, len(scopes))
	for _, scope := range scopes {
		set[scope] = true
	}
	return set
}

func sortedScopes(scopes map[string]bool) []string {
	result := make([]string, 0, len(scopes))
	for scope, enabled := range scopes {
		if enabled {
			result = append(result, scope)
		}
	}
	sort.Strings(result)
	return result
}

func newID(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, newToken(12))
}

func newToken(size int) string {
	bytes := make([]byte, size)
	if _, err := randomRead(bytes); err != nil {
		seed := fmt.Sprintf("%d:%d", time.Now().UnixNano(), fallbackCounter.Add(1))
		sum := sha256.Sum256([]byte(seed))
		return hex.EncodeToString(sum[:])[:size*2]
	}
	return hex.EncodeToString(bytes)
}

type SignEventFunc func(event domain.Event, endpoint domain.WebhookEndpoint) string

func ErrorCode(err error) string {
	switch {
	case errors.Is(err, domain.ErrValidation):
		return "validation_failed"
	case errors.Is(err, domain.ErrAccountNotFound):
		return "account_not_found"
	case errors.Is(err, domain.ErrInsufficientFunds):
		return "insufficient_funds"
	case errors.Is(err, domain.ErrOriginalTxnNotFound):
		return "original_transaction_not_found"
	case errors.Is(err, domain.ErrRefundExceedsOriginal):
		return "refund_exceeds_original"
	default:
		return "internal_error"
	}
}
