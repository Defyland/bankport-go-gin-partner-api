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
	timestamp := s.clock().Unix()
	payload, _ := json.Marshal(event)
	mac := hmac.New(sha256.New, s.secret)
	_, _ = fmt.Fprintf(mac, "%d.", timestamp)
	_, _ = mac.Write(payload)
	return fmt.Sprintf("t=%d,v1=%s", timestamp, hex.EncodeToString(mac.Sum(nil)))
}
