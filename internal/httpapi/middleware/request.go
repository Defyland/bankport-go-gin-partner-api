package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/observability"
)

func RequestIdentity() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = newRequestID("req")
		}
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = requestID
		}
		c.Set(RequestIDKey, requestID)
		c.Set(CorrelationIDKey, correlationID)
		c.Header("X-Request-ID", requestID)
		c.Header("X-Correlation-ID", correlationID)
		c.Next()
	}
}

func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func StructuredLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()

		partnerID := ""
		developerAppID := ""
		if partner, ok := Partner(c); ok {
			partnerID = partner.ID
			developerAppID = partner.DeveloperAppID
		}
		logger.Info("http_request",
			"method", c.Request.Method,
			"route", route(c),
			"status", statusFromWriter(c),
			"duration_ms", time.Since(started).Milliseconds(),
			"request_id", RequestID(c),
			"correlation_id", CorrelationID(c),
			"partner_id", partnerID,
			"developer_app_id", developerAppID,
			"remote_addr", c.ClientIP(),
		)
	}
}

func Metrics(metrics *observability.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		metrics.ObserveHTTP(c.Request.Method, route(c), statusFromWriter(c), started)
	}
}

func Tracing(serviceName string) gin.HandlerFunc {
	tracer := otel.Tracer(serviceName)
	return func(c *gin.Context) {
		ctx, span := tracer.Start(c.Request.Context(), c.Request.Method+" "+c.Request.URL.Path)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
		status := statusFromWriter(c)
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.route", route(c)),
			attribute.Int("http.status_code", status),
			attribute.String("bankport.request_id", RequestID(c)),
			attribute.String("bankport.correlation_id", CorrelationID(c)),
		)
		if status >= http.StatusInternalServerError {
			span.SetStatus(codes.Error, http.StatusText(status))
		}
		span.End()
	}
}
