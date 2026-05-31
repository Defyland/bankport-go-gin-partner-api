package webhook

import (
	"strings"
	"testing"
	"time"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/domain"
)

func TestSignerCreatesVersionedHMACSignature(t *testing.T) {
	signer := NewSigner("test-secret")

	signature := signer.SignEvent(domain.Event{ID: "evt_test", Type: "pix.transfer.created.v1"})

	if !strings.Contains(signature, "t=") || !strings.Contains(signature, "v1=") {
		t.Fatalf("expected timestamped v1 signature, got %q", signature)
	}
}

func TestSignerDerivesEndpointSpecificSignatures(t *testing.T) {
	signer := NewSigner("test-secret")
	signer.clock = func() time.Time { return time.Unix(1_780_199_638, 0).UTC() }
	event := domain.Event{ID: "evt_test", Type: "pix.transfer.created.v1"}

	first := signer.SignEventForEndpoint(event, domain.WebhookEndpoint{SecretID: "whsec_first"})
	second := signer.SignEventForEndpoint(event, domain.WebhookEndpoint{SecretID: "whsec_second"})

	if first == second {
		t.Fatal("expected endpoint-specific signatures to differ")
	}
	if !strings.Contains(first, "t=") || !strings.Contains(first, "v1=") {
		t.Fatalf("expected timestamped v1 signature, got %q", first)
	}
}
