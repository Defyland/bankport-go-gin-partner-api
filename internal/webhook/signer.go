package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/domain"
)

type Signer struct {
	secret []byte
	clock  func() time.Time
}

func NewSigner(secret string) Signer {
	return Signer{
		secret: []byte(secret),
		clock:  func() time.Time { return time.Now().UTC() },
	}
}

func (s Signer) SignEvent(event domain.Event) string {
	return s.sign(event, s.secret)
}

func (s Signer) SignEventForEndpoint(event domain.Event, endpoint domain.WebhookEndpoint) string {
	return s.sign(event, s.endpointSecret(endpoint.SecretID))
}

func (s Signer) sign(event domain.Event, secret []byte) string {
	timestamp := s.clock().Unix()
	payload, _ := json.Marshal(event)
	mac := hmac.New(sha256.New, secret)
	_, _ = fmt.Fprintf(mac, "%d.", timestamp)
	_, _ = mac.Write(payload)
	return fmt.Sprintf("t=%d,v1=%s", timestamp, hex.EncodeToString(mac.Sum(nil)))
}

func (s Signer) endpointSecret(secretID string) []byte {
	if secretID == "" {
		return s.secret
	}
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(secretID))
	return mac.Sum(nil)
}
