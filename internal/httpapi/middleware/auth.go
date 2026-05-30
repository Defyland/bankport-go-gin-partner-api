package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/domain"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/observability"
)

type Authenticator interface {
	AuthenticateAPIKey(apiKey string) (domain.Partner, bool)
}

func Authenticate(authenticator Authenticator, metrics *observability.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := bearerToken(c.GetHeader("Authorization"))
		if apiKey == "" {
			apiKey = c.GetHeader("X-API-Key")
		}
		if apiKey == "" {
			Abort(c, http.StatusUnauthorized, "authentication_required", "Provide a valid Bearer token or X-API-Key.", nil)
			return
		}

		partner, ok := authenticator.AuthenticateAPIKey(apiKey)
		if !ok {
			Abort(c, http.StatusUnauthorized, "invalid_api_key", "The API key is unknown or inactive.", nil)
			return
		}
		c.Set(PartnerKey, partner)
		metrics.AuthenticatedRequests.WithLabelValues(partner.ID, partner.DeveloperAppID).Inc()
		c.Next()
	}
}

func RequireScopes(scopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		partner, ok := Partner(c)
		if !ok {
			Abort(c, http.StatusUnauthorized, "authentication_required", "Authentication context is missing.", nil)
			return
		}
		for _, scope := range scopes {
			if !partner.HasScope(scope) {
				Abort(c, http.StatusForbidden, "insufficient_scope", "The API key does not include the required scope.", map[string]any{
					"required_scope": scope,
				})
				return
			}
		}
		c.Next()
	}
}
