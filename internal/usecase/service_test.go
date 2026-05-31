package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/domain"
)

func TestCreatePixTransferAuditsAndRecordsMetrics(t *testing.T) {
	repo := &fakeRepository{
		pixTransfer: domain.PixTransfer{
			ID:              "pix_123",
			PartnerID:       "partner_1",
			SourceAccountID: "acct_1",
			AmountCents:     100,
			Currency:        "BRL",
			Status:          "accepted",
		},
		queuedDeliveries: 2,
	}
	metrics := &fakeMetrics{}
	service := New(Dependencies{
		Financial: repo,
		Metrics:   metrics,
		SignEvent: func(domain.Event, domain.WebhookEndpoint) string { return "sig" },
		Now:       fixedNow,
	})

	result, err := service.CreatePixTransfer(context.Background(), partner(), domain.PixTransferRequest{
		SourceAccountID: "acct_1",
		AmountCents:     100,
		Currency:        "BRL",
		PixKey:          "merchant@example.com",
	}, "corr_1", "req_1")

	if err != nil {
		t.Fatalf("create pix transfer: %v", err)
	}
	if result.Transfer.ID != "pix_123" || result.QueuedDeliveries != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if got := metrics.financial["pix_transfer:accepted"]; got != 1 {
		t.Fatalf("expected accepted financial metric, got %d", got)
	}
	if got := metrics.webhooks["pix.transfer.created.v1:queued"]; got != 2 {
		t.Fatalf("expected queued webhook metric count 2, got %d", got)
	}
	if len(repo.auditEntries) != 1 || repo.auditEntries[0].Status != "accepted" || repo.auditEntries[0].ResourceID != "pix_123" {
		t.Fatalf("expected accepted audit entry, got %+v", repo.auditEntries)
	}
}

func TestCreatePixTransferAuditsRejectedDomainError(t *testing.T) {
	repo := &fakeRepository{err: domain.ErrInsufficientFunds}
	metrics := &fakeMetrics{}
	service := New(Dependencies{
		Financial: repo,
		Metrics:   metrics,
		SignEvent: func(domain.Event, domain.WebhookEndpoint) string { return "sig" },
		Now:       fixedNow,
	})

	_, err := service.CreatePixTransfer(context.Background(), partner(), domain.PixTransferRequest{
		SourceAccountID: "acct_1",
		AmountCents:     100,
		Currency:        "BRL",
		PixKey:          "merchant@example.com",
	}, "corr_1", "req_1")

	if !errors.Is(err, domain.ErrInsufficientFunds) {
		t.Fatalf("expected insufficient funds, got %v", err)
	}
	if got := metrics.financial["pix_transfer:rejected"]; got != 1 {
		t.Fatalf("expected rejected financial metric, got %d", got)
	}
	if len(repo.auditEntries) != 1 || repo.auditEntries[0].Reason != "insufficient_funds" || repo.auditEntries[0].ResourceID != "acct_1" {
		t.Fatalf("expected rejected audit entry, got %+v", repo.auditEntries)
	}
}

func TestRegisterWebhookEndpointAuditsResult(t *testing.T) {
	repo := &fakeRepository{
		webhookEndpoint: domain.WebhookEndpoint{ID: "wh_123", PartnerID: "partner_1"},
	}
	service := New(Dependencies{
		Webhooks: repo,
		Now:      fixedNow,
	})

	endpoint, err := service.RegisterWebhookEndpoint(context.Background(), partner(), domain.WebhookEndpointRequest{
		URL:        "https://partner.example.com/webhooks",
		EventTypes: []string{"pix.transfer.created.v1"},
	}, "corr_1", "req_1")

	if err != nil {
		t.Fatalf("register webhook: %v", err)
	}
	if endpoint.ID != "wh_123" {
		t.Fatalf("unexpected endpoint: %+v", endpoint)
	}
	if len(repo.auditEntries) != 1 || repo.auditEntries[0].Action != "webhook.endpoint.create" || repo.auditEntries[0].Status != "accepted" {
		t.Fatalf("expected webhook audit entry, got %+v", repo.auditEntries)
	}
}

type fakeRepository struct {
	pixTransfer      domain.PixTransfer
	payout           domain.Payout
	refund           domain.Refund
	webhookEndpoint  domain.WebhookEndpoint
	queuedDeliveries int
	err              error
	auditEntries     []domain.AuditEntry
}

func (r *fakeRepository) CreatePixTransfer(context.Context, domain.Partner, domain.PixTransferRequest, string, domain.SignEventFunc) (domain.PixTransfer, int, error) {
	return r.pixTransfer, r.queuedDeliveries, r.err
}

func (r *fakeRepository) CreatePayout(context.Context, domain.Partner, domain.PayoutRequest, string, domain.SignEventFunc) (domain.Payout, int, error) {
	return r.payout, r.queuedDeliveries, r.err
}

func (r *fakeRepository) CreateRefund(context.Context, domain.Partner, domain.RefundRequest, string, domain.SignEventFunc) (domain.Refund, int, error) {
	return r.refund, r.queuedDeliveries, r.err
}

func (r *fakeRepository) RegisterWebhookEndpoint(context.Context, domain.Partner, domain.WebhookEndpointRequest) (domain.WebhookEndpoint, error) {
	return r.webhookEndpoint, r.err
}

func (r *fakeRepository) AddAuditEntry(entry domain.AuditEntry) {
	r.auditEntries = append(r.auditEntries, entry)
}

type fakeMetrics struct {
	financial map[string]int
	webhooks  map[string]int
}

func (m *fakeMetrics) RecordFinancialCommand(command, outcome string) {
	if m.financial == nil {
		m.financial = make(map[string]int)
	}
	m.financial[command+":"+outcome]++
}

func (m *fakeMetrics) RecordWebhookDeliveries(eventType, status string, count int) {
	if m.webhooks == nil {
		m.webhooks = make(map[string]int)
	}
	m.webhooks[eventType+":"+status] += count
}

func partner() domain.Partner {
	return domain.Partner{ID: "partner_1", DeveloperAppID: "app_1"}
}

func fixedNow() time.Time {
	return time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)
}
