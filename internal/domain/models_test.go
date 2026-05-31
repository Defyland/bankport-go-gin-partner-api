package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestPixTransferRequestValidate(t *testing.T) {
	request := PixTransferRequest{
		SourceAccountID: "acct_sandbox_001",
		AmountCents:     2500,
		Currency:        "brl",
		PixKey:          "merchant@example.com",
	}

	if err := request.Validate(); err != nil {
		t.Fatalf("expected valid pix transfer request: %v", err)
	}
}

func TestPixTransferRequestRejectsInvalidMoney(t *testing.T) {
	request := PixTransferRequest{
		SourceAccountID: "acct_sandbox_001",
		AmountCents:     0,
		Currency:        "BRL",
		PixKey:          "merchant@example.com",
	}

	if err := request.Validate(); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestPixTransferRequestRejectsOversizedPixKey(t *testing.T) {
	request := PixTransferRequest{
		SourceAccountID: "acct_sandbox_001",
		AmountCents:     100,
		Currency:        "BRL",
		PixKey:          strings.Repeat("x", maxPixKeyLength+1),
	}

	if err := request.Validate(); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestPayoutRequestValidatesBankAccountShape(t *testing.T) {
	valid := PayoutRequest{
		AccountID:     "acct_sandbox_001",
		AmountCents:   100,
		Currency:      "BRL",
		BankCode:      "001",
		Branch:        "0001",
		AccountNumber: "12345-6",
		Document:      "12345678909",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid payout request: %v", err)
	}

	invalid := valid
	invalid.BankCode = "1A1"
	if err := invalid.Validate(); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected invalid bank code validation error, got %v", err)
	}

	invalid = valid
	invalid.Document = "123"
	if err := invalid.Validate(); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected invalid document validation error, got %v", err)
	}
}

func TestRefundRequestRejectsOversizedReason(t *testing.T) {
	request := RefundRequest{
		OriginalTransactionID: "pix_test",
		AccountID:             "acct_sandbox_001",
		AmountCents:           100,
		Currency:              "BRL",
		Reason:                strings.Repeat("x", maxReasonLength+1),
	}

	if err := request.Validate(); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestWebhookEndpointRequiresHTTPSOutsideLocalhost(t *testing.T) {
	request := WebhookEndpointRequest{
		URL:        "http://partner.example.com/webhooks",
		EventTypes: []string{"pix.transfer.created.v1"},
	}

	if err := request.Validate(); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestWebhookEndpointAllowsLocalhostWithPort(t *testing.T) {
	request := WebhookEndpointRequest{
		URL:        "http://localhost:3000/webhooks",
		EventTypes: []string{"pix.transfer.created.v1"},
	}

	if err := request.Validate(); err != nil {
		t.Fatalf("expected localhost webhook receiver to be valid in sandbox: %v", err)
	}
}

func TestWebhookEndpointRejectsUnsupportedEventType(t *testing.T) {
	request := WebhookEndpointRequest{
		URL:        "https://partner.example.com/webhooks",
		EventTypes: []string{"unknown.event.v1"},
	}

	if err := request.Validate(); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected unsupported event type validation error, got %v", err)
	}
}

func TestWebhookEndpointRejectsURLUserInfo(t *testing.T) {
	request := WebhookEndpointRequest{
		URL:        "https://user:pass@partner.example.com/webhooks",
		EventTypes: []string{"pix.transfer.created.v1"},
	}

	if err := request.Validate(); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected userinfo validation error, got %v", err)
	}
}
