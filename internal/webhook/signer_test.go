package webhook

import (
	"strings"
	"testing"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/domain"
)

func TestSignerCreatesVersionedHMACSignature(t *testing.T) {
	signer := NewSigner("test-secret")

	signature := signer.SignEvent(domain.Event{ID: "evt_test", Type: "pix.transfer.created.v1"})

	if !strings.Contains(signature, "t=") || !strings.Contains(signature, "v1=") {
		t.Fatalf("expected timestamped v1 signature, got %q", signature)
	}
}
