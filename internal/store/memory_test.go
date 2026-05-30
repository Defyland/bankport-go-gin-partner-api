package store

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/config"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/domain"
)

func TestCreatePixTransferDebitsOnlyPartnerOwnedAccount(t *testing.T) {
	cfg := config.Load()
	repo := NewSeededRepository(cfg)
	partner, ok := repo.AuthenticateAPIKey(cfg.FullAccessAPIKey)
	if !ok {
		t.Fatal("expected seeded full access partner")
	}

	_, _, err := repo.CreatePixTransfer(context.Background(), partner, domain.PixTransferRequest{
		SourceAccountID: "acct_other_001",
		AmountCents:     100,
		Currency:        "BRL",
		PixKey:          "merchant@example.com",
	}, "corr_test", func(domain.Event) string { return "signature" })
	if err != domain.ErrAccountNotFound {
		t.Fatalf("expected tenant-isolated account not found, got %v", err)
	}
}

func TestHashAPIKeyUsesPepper(t *testing.T) {
	apiKey := "bp_test_key"
	first := HashAPIKey(apiKey, "pepper-one")
	second := HashAPIKey(apiKey, "pepper-two")
	if first == second {
		t.Fatal("expected different peppers to produce different API key hashes")
	}
	if first == HashAPIKey("different-key", "pepper-one") {
		t.Fatal("expected different API keys to produce different hashes")
	}
}

func TestCreatePixTransferQueuesWebhookDelivery(t *testing.T) {
	cfg := config.Load()
	repo := NewSeededRepository(cfg)
	partner, ok := repo.AuthenticateAPIKey(cfg.FullAccessAPIKey)
	if !ok {
		t.Fatal("expected seeded full access partner")
	}
	_, err := repo.RegisterWebhookEndpoint(context.Background(), partner, domain.WebhookEndpointRequest{
		URL:        "https://partner.example.com/webhooks",
		EventTypes: []string{"pix.transfer.created.v1"},
	})
	if err != nil {
		t.Fatalf("register webhook endpoint: %v", err)
	}

	transfer, queued, err := repo.CreatePixTransfer(context.Background(), partner, domain.PixTransferRequest{
		SourceAccountID: "acct_sandbox_001",
		AmountCents:     1000,
		Currency:        "BRL",
		PixKey:          "merchant@example.com",
	}, "corr_test", func(domain.Event) string { return "signature" })
	if err != nil {
		t.Fatalf("create pix transfer: %v", err)
	}
	if transfer.ID == "" {
		t.Fatal("expected transfer id")
	}
	if queued != 1 {
		t.Fatalf("expected one queued webhook delivery, got %d", queued)
	}
	if len(repo.WebhookDeliveries()) != 1 {
		t.Fatalf("expected persisted webhook delivery")
	}
}

func TestCreateRefundRejectsCumulativeRefundAboveOriginalAmount(t *testing.T) {
	cfg := config.Load()
	repo := NewSeededRepository(cfg)
	partner, ok := repo.AuthenticateAPIKey(cfg.FullAccessAPIKey)
	if !ok {
		t.Fatal("expected seeded full access partner")
	}

	transfer, _, err := repo.CreatePixTransfer(context.Background(), partner, domain.PixTransferRequest{
		SourceAccountID: "acct_sandbox_001",
		AmountCents:     2000,
		Currency:        "BRL",
		PixKey:          "merchant@example.com",
	}, "corr_test", func(domain.Event) string { return "signature" })
	if err != nil {
		t.Fatalf("create pix transfer: %v", err)
	}

	_, _, err = repo.CreateRefund(context.Background(), partner, domain.RefundRequest{
		OriginalTransactionID: transfer.ID,
		AccountID:             transfer.SourceAccountID,
		AmountCents:           1500,
		Currency:              "BRL",
		Reason:                "partial reversal",
	}, "corr_test", func(domain.Event) string { return "signature" })
	if err != nil {
		t.Fatalf("create first refund: %v", err)
	}

	_, _, err = repo.CreateRefund(context.Background(), partner, domain.RefundRequest{
		OriginalTransactionID: transfer.ID,
		AccountID:             transfer.SourceAccountID,
		AmountCents:           600,
		Currency:              "BRL",
		Reason:                "duplicate reversal attempt",
	}, "corr_test", func(domain.Event) string { return "signature" })
	if err != domain.ErrRefundExceedsOriginal {
		t.Fatalf("expected cumulative refund guard, got %v", err)
	}
}

func TestConcurrentPixTransfersDoNotOverspendAccount(t *testing.T) {
	cfg := config.Load()
	repo := NewSeededRepository(cfg)
	partner, ok := repo.AuthenticateAPIKey(cfg.FullAccessAPIKey)
	if !ok {
		t.Fatal("expected seeded full access partner")
	}

	errs := make(chan error, 4)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := repo.CreatePixTransfer(context.Background(), partner, domain.PixTransferRequest{
				SourceAccountID: "acct_sandbox_001",
				AmountCents:     1_000_000,
				Currency:        "BRL",
				PixKey:          "merchant@example.com",
			}, "corr_test", func(domain.Event) string { return "signature" })
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	successes := 0
	insufficientFunds := 0
	for err := range errs {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, domain.ErrInsufficientFunds):
			insufficientFunds++
		default:
			t.Fatalf("unexpected transfer error: %v", err)
		}
	}
	if successes != 2 || insufficientFunds != 2 {
		t.Fatalf("expected 2 accepted and 2 rejected transfers, got %d accepted and %d rejected", successes, insufficientFunds)
	}

	account, err := repo.GetAccount(context.Background(), partner.ID, "acct_sandbox_001")
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if account.AvailableBalanceCts != 500_000 {
		t.Fatalf("expected remaining balance 500000, got %d", account.AvailableBalanceCts)
	}
}

func TestConcurrentRefundsDoNotExceedOriginalAmount(t *testing.T) {
	cfg := config.Load()
	repo := NewSeededRepository(cfg)
	partner, ok := repo.AuthenticateAPIKey(cfg.FullAccessAPIKey)
	if !ok {
		t.Fatal("expected seeded full access partner")
	}

	transfer, _, err := repo.CreatePixTransfer(context.Background(), partner, domain.PixTransferRequest{
		SourceAccountID: "acct_sandbox_001",
		AmountCents:     2000,
		Currency:        "BRL",
		PixKey:          "merchant@example.com",
	}, "corr_test", func(domain.Event) string { return "signature" })
	if err != nil {
		t.Fatalf("create pix transfer: %v", err)
	}

	errs := make(chan error, 3)
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := repo.CreateRefund(context.Background(), partner, domain.RefundRequest{
				OriginalTransactionID: transfer.ID,
				AccountID:             transfer.SourceAccountID,
				AmountCents:           1000,
				Currency:              "BRL",
				Reason:                "concurrent reversal",
			}, "corr_test", func(domain.Event) string { return "signature" })
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	successes := 0
	rejected := 0
	for err := range errs {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, domain.ErrRefundExceedsOriginal):
			rejected++
		default:
			t.Fatalf("unexpected refund error: %v", err)
		}
	}
	if successes != 2 || rejected != 1 {
		t.Fatalf("expected 2 accepted and 1 rejected refunds, got %d accepted and %d rejected", successes, rejected)
	}
}
