package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/domain"
)

const (
	RequestIDKey     = "request_id"
	CorrelationIDKey = "correlation_id"
	PartnerKey       = "partner"
)

type ErrorEnvelope struct {
	Error         APIError `json:"error"`
	RequestID     string   `json:"request_id"`
	CorrelationID string   `json:"correlation_id"`
}

type APIError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func RequestID(c *gin.Context) string {
	if value, ok := c.Get(RequestIDKey); ok {
		if requestID, ok := value.(string); ok {
			return requestID
		}
	}
	return ""
}

func CorrelationID(c *gin.Context) string {
	if value, ok := c.Get(CorrelationIDKey); ok {
		if correlationID, ok := value.(string); ok {
			return correlationID
		}
	}
	return ""
}

func Partner(c *gin.Context) (domain.Partner, bool) {
	value, ok := c.Get(PartnerKey)
	if !ok {
		return domain.Partner{}, false
	}
	partner, ok := value.(domain.Partner)
	return partner, ok
}

func Abort(c *gin.Context, status int, code, message string, details map[string]any) {
	c.AbortWithStatusJSON(status, ErrorEnvelope{
		Error: APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
		RequestID:     RequestID(c),
		CorrelationID: CorrelationID(c),
	})
}

func newRequestID(prefix string) string {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return prefix + "_fallback"
	}
	return prefix + "_" + hex.EncodeToString(bytes)
}

func bearerToken(header string) string {
	value := strings.TrimSpace(header)
	if value == "" {
		return ""
	}
	parts := strings.Fields(value)
	if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
		return parts[1]
	}
	return ""
}

func route(c *gin.Context) string {
	if fullPath := c.FullPath(); fullPath != "" {
		return fullPath
	}
	return c.Request.URL.Path
}

func statusFromWriter(c *gin.Context) int {
	status := c.Writer.Status()
	if status == 0 {
		return http.StatusOK
	}
	return status
}
