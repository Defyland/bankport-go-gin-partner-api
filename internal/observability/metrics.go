package observability

import (
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	Registry              *prometheus.Registry
	HTTPRequests          *prometheus.CounterVec
	HTTPRequestDuration   *prometheus.HistogramVec
	FinancialCommands     *prometheus.CounterVec
	WebhookDeliveries     *prometheus.CounterVec
	RateLimitExceeded     *prometheus.CounterVec
	IdempotencyReplays    *prometheus.CounterVec
	IdempotencyConflicts  *prometheus.CounterVec
	AuthenticatedRequests *prometheus.CounterVec
}

func NewMetrics(namespace string) *Metrics {
	namespace = prometheusNamespace(namespace)
	registry := prometheus.NewRegistry()
	metrics := &Metrics{
		Registry: registry,
		HTTPRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total HTTP requests partitioned by method, route, and status.",
		}, []string{"method", "route", "status"}),
		HTTPRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration partitioned by method, route, and status.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"method", "route", "status"}),
		FinancialCommands: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "financial_commands_total",
			Help:      "Financial write commands accepted by the partner API.",
		}, []string{"command", "status"}),
		WebhookDeliveries: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "webhook_deliveries_total",
			Help:      "Webhook deliveries queued by event type.",
		}, []string{"event_type", "status"}),
		RateLimitExceeded: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "rate_limit_exceeded_total",
			Help:      "Partner requests rejected by rate limiting.",
		}, []string{"partner_id", "route"}),
		IdempotencyReplays: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "idempotency_replays_total",
			Help:      "Idempotent write requests served from cache.",
		}, []string{"partner_id", "route"}),
		IdempotencyConflicts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "idempotency_conflicts_total",
			Help:      "Idempotency-key reuse attempts with a different request body.",
		}, []string{"partner_id", "route"}),
		AuthenticatedRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "authenticated_requests_total",
			Help:      "Authenticated partner requests partitioned by partner and app.",
		}, []string{"partner_id", "developer_app_id"}),
	}

	registry.MustRegister(
		metrics.HTTPRequests,
		metrics.HTTPRequestDuration,
		metrics.FinancialCommands,
		metrics.WebhookDeliveries,
		metrics.RateLimitExceeded,
		metrics.IdempotencyReplays,
		metrics.IdempotencyConflicts,
		metrics.AuthenticatedRequests,
	)
	return metrics
}

func prometheusNamespace(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "bankport_partner_api"
	}
	var builder strings.Builder
	for _, char := range value {
		if unicode.IsLetter(char) || unicode.IsDigit(char) || char == '_' || char == ':' {
			builder.WriteRune(char)
			continue
		}
		builder.WriteByte('_')
	}
	normalized := builder.String()
	if normalized == "" {
		return "bankport_partner_api"
	}
	first := rune(normalized[0])
	if unicode.IsDigit(first) {
		return "bankport_" + normalized
	}
	return normalized
}

func (m *Metrics) ObserveHTTP(method, route string, status int, started time.Time) {
	statusLabel := strconv.Itoa(status)
	m.HTTPRequests.WithLabelValues(method, route, statusLabel).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, route, statusLabel).Observe(time.Since(started).Seconds())
}
