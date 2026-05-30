package domain

import (
	"errors"
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

func TestWebhookEndpointRequiresHTTPSOutsideLocalhost(t *testing.T) {
	request := WebhookEndpointRequest{
		URL:        "http://partner.example.com/webhooks",
		EventTypes: []string{"pix.transfer.created.v1"},
	}

	if err := request.Validate(); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}
