package usecase

import (
	"context"
	"time"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/domain"
)

type AccountReader interface {
	GetAccount(ctx context.Context, partnerID, accountID string) (domain.Account, error)
	ListStatements(ctx context.Context, partnerID, accountID string) ([]domain.StatementEntry, error)
}

type FinancialCommandStore interface {
	CreatePixTransfer(ctx context.Context, partner domain.Partner, request domain.PixTransferRequest, correlationID string, sign domain.SignEventFunc) (domain.PixTransfer, int, error)
	CreatePayout(ctx context.Context, partner domain.Partner, request domain.PayoutRequest, correlationID string, sign domain.SignEventFunc) (domain.Payout, int, error)
	CreateRefund(ctx context.Context, partner domain.Partner, request domain.RefundRequest, correlationID string, sign domain.SignEventFunc) (domain.Refund, int, error)
	AddAuditEntry(entry domain.AuditEntry)
}

type WebhookStore interface {
	RegisterWebhookEndpoint(ctx context.Context, partner domain.Partner, request domain.WebhookEndpointRequest) (domain.WebhookEndpoint, error)
	AddAuditEntry(entry domain.AuditEntry)
}

type PlatformReader interface {
	ListAuditEntries(ctx context.Context, partnerID string) []domain.AuditEntry
	SandboxScenarios() []domain.SandboxScenario
}

type MetricsRecorder interface {
	RecordFinancialCommand(command, outcome string)
	RecordWebhookDeliveries(eventType, status string, count int)
}

type Service struct {
	accounts  AccountReader
	financial FinancialCommandStore
	webhooks  WebhookStore
	platform  PlatformReader
	metrics   MetricsRecorder
	signEvent domain.SignEventFunc
	now       func() time.Time
}

type Dependencies struct {
	Accounts  AccountReader
	Financial FinancialCommandStore
	Webhooks  WebhookStore
	Platform  PlatformReader
	Metrics   MetricsRecorder
	SignEvent domain.SignEventFunc
	Now       func() time.Time
}

type PixTransferResult struct {
	Transfer         domain.PixTransfer
	QueuedDeliveries int
}

type PayoutResult struct {
	Payout           domain.Payout
	QueuedDeliveries int
}

type RefundResult struct {
	Refund           domain.Refund
	QueuedDeliveries int
}

func New(deps Dependencies) Service {
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return Service{
		accounts:  deps.Accounts,
		financial: deps.Financial,
		webhooks:  deps.Webhooks,
		platform:  deps.Platform,
		metrics:   deps.Metrics,
		signEvent: deps.SignEvent,
		now:       now,
	}
}

func (s Service) GetBalance(ctx context.Context, partner domain.Partner, accountID string) (domain.Account, error) {
	return s.accounts.GetAccount(ctx, partner.ID, accountID)
}

func (s Service) ListStatements(ctx context.Context, partner domain.Partner, accountID string) ([]domain.StatementEntry, error) {
	return s.accounts.ListStatements(ctx, partner.ID, accountID)
}

func (s Service) CreatePixTransfer(ctx context.Context, partner domain.Partner, request domain.PixTransferRequest, correlationID, requestID string) (PixTransferResult, error) {
	transfer, queued, err := s.financial.CreatePixTransfer(ctx, partner, request, correlationID, s.signEvent)
	if err != nil {
		s.recordFinancialRejection(partner, requestID, correlationID, "pix_transfer", "pix.transfer.create", request.SourceAccountID, err)
		return PixTransferResult{}, err
	}
	s.recordFinancialAcceptance(partner, requestID, correlationID, "pix_transfer", "pix.transfer.create", transfer.ID)
	s.recordWebhookQueue("pix.transfer.created.v1", queued)
	return PixTransferResult{Transfer: transfer, QueuedDeliveries: queued}, nil
}

func (s Service) CreatePayout(ctx context.Context, partner domain.Partner, request domain.PayoutRequest, correlationID, requestID string) (PayoutResult, error) {
	payout, queued, err := s.financial.CreatePayout(ctx, partner, request, correlationID, s.signEvent)
	if err != nil {
		s.recordFinancialRejection(partner, requestID, correlationID, "payout", "payout.create", request.AccountID, err)
		return PayoutResult{}, err
	}
	s.recordFinancialAcceptance(partner, requestID, correlationID, "payout", "payout.create", payout.ID)
	s.recordWebhookQueue("payout.created.v1", queued)
	return PayoutResult{Payout: payout, QueuedDeliveries: queued}, nil
}

func (s Service) CreateRefund(ctx context.Context, partner domain.Partner, request domain.RefundRequest, correlationID, requestID string) (RefundResult, error) {
	refund, queued, err := s.financial.CreateRefund(ctx, partner, request, correlationID, s.signEvent)
	if err != nil {
		s.recordFinancialRejection(partner, requestID, correlationID, "refund", "refund.create", request.AccountID, err)
		return RefundResult{}, err
	}
	s.recordFinancialAcceptance(partner, requestID, correlationID, "refund", "refund.create", refund.ID)
	s.recordWebhookQueue("refund.created.v1", queued)
	return RefundResult{Refund: refund, QueuedDeliveries: queued}, nil
}

func (s Service) RegisterWebhookEndpoint(ctx context.Context, partner domain.Partner, request domain.WebhookEndpointRequest, correlationID, requestID string) (domain.WebhookEndpoint, error) {
	endpoint, err := s.webhooks.RegisterWebhookEndpoint(ctx, partner, request)
	if err != nil {
		s.audit(s.webhooks, partner, requestID, correlationID, "webhook.endpoint.create", "", "rejected", domain.ErrorCode(err))
		return domain.WebhookEndpoint{}, err
	}
	s.audit(s.webhooks, partner, requestID, correlationID, "webhook.endpoint.create", endpoint.ID, "accepted", "")
	return endpoint, nil
}

func (s Service) ListAuditLogs(ctx context.Context, partner domain.Partner) []domain.AuditEntry {
	return s.platform.ListAuditEntries(ctx, partner.ID)
}

func (s Service) ListSandboxScenarios() []domain.SandboxScenario {
	return s.platform.SandboxScenarios()
}

func (s Service) recordFinancialRejection(partner domain.Partner, requestID, correlationID, command, action, resourceID string, err error) {
	if s.metrics != nil {
		s.metrics.RecordFinancialCommand(command, "rejected")
	}
	s.audit(s.financial, partner, requestID, correlationID, action, resourceID, "rejected", domain.ErrorCode(err))
}

func (s Service) recordFinancialAcceptance(partner domain.Partner, requestID, correlationID, command, action, resourceID string) {
	if s.metrics != nil {
		s.metrics.RecordFinancialCommand(command, "accepted")
	}
	s.audit(s.financial, partner, requestID, correlationID, action, resourceID, "accepted", "")
}

func (s Service) recordWebhookQueue(eventType string, queued int) {
	if queued > 0 && s.metrics != nil {
		s.metrics.RecordWebhookDeliveries(eventType, "queued", queued)
	}
}

func (s Service) audit(store interface{ AddAuditEntry(domain.AuditEntry) }, partner domain.Partner, requestID, correlationID, action, resourceID, status, reason string) {
	store.AddAuditEntry(domain.AuditEntry{
		PartnerID:     partner.ID,
		RequestID:     requestID,
		CorrelationID: correlationID,
		Action:        action,
		ResourceID:    resourceID,
		Status:        status,
		Reason:        reason,
		OccurredAt:    s.now(),
	})
}
