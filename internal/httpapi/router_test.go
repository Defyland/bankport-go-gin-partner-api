package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/config"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/observability"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/store"
)

func TestHealthReadyAndMetrics(t *testing.T) {
	router, _, cfg := newTestRouter(t, 120)

	ready := perform(router, http.MethodGet, "/health/ready", "", "")
	if ready.Code != http.StatusOK {
		t.Fatalf("expected readiness 200, got %d: %s", ready.Code, ready.Body.String())
	}

	metrics := perform(router, http.MethodGet, "/metrics", "", "")
	if metrics.Code != http.StatusOK {
		t.Fatalf("expected metrics 200, got %d: %s", metrics.Code, metrics.Body.String())
	}
	if !strings.Contains(metrics.Body.String(), strings.ReplaceAll(cfg.ServiceName, "-", "_")+"_http_requests_total") {
		t.Fatalf("expected exported http metrics, got %s", metrics.Body.String())
	}
}

func TestRequiresAuthentication(t *testing.T) {
	router, _, _ := newTestRouter(t, 120)

	response := perform(router, http.MethodGet, "/v1/accounts/acct_sandbox_001/balance", "", "")

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "authentication_required") {
		t.Fatalf("expected standardized auth error, got %s", response.Body.String())
	}
}

func TestRejectsOversizedFinancialBody(t *testing.T) {
	cfg := testConfig(120)
	cfg.MaxRequestBodyBytes = 32
	repo := store.NewSeededRepository(cfg)
	router := NewRouter(Dependencies{
		Config:     cfg,
		Logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
		Repository: repo,
		Metrics:    observability.NewMetrics(cfg.ServiceName),
	})

	response := performWithHeaders(router, http.MethodPost, "/v1/pix/transfers", cfg.FullAccessAPIKey, `{
		"source_account_id": "acct_sandbox_001",
		"amount_cents": 100,
		"currency": "BRL",
		"pix_key": "merchant@example.com"
	}`, map[string]string{"Idempotency-Key": "oversized-body"})

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "request_body_too_large") {
		t.Fatalf("expected body size error, got %s", response.Body.String())
	}
}

func TestRejectsInsufficientScope(t *testing.T) {
	router, _, cfg := newTestRouter(t, 120)

	response := perform(router, http.MethodPost, "/v1/pix/transfers", cfg.ReadOnlyAPIKey, `{
		"source_account_id": "acct_sandbox_001",
		"amount_cents": 100,
		"currency": "BRL",
		"pix_key": "merchant@example.com"
	}`)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "insufficient_scope") {
		t.Fatalf("expected insufficient scope error, got %s", response.Body.String())
	}
}

func TestTenantIsolationHidesForeignAccount(t *testing.T) {
	router, _, cfg := newTestRouter(t, 120)

	response := perform(router, http.MethodGet, "/v1/accounts/acct_other_001/balance", cfg.FullAccessAPIKey, "")

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for foreign account, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "account_not_found") {
		t.Fatalf("expected tenant-isolated not found error, got %s", response.Body.String())
	}
}

func TestMetricsUseRoutePatternForAccountIDs(t *testing.T) {
	router, _, cfg := newTestRouter(t, 120)

	response := perform(router, http.MethodGet, "/v1/accounts/acct_sandbox_001/balance", cfg.FullAccessAPIKey, "")
	if response.Code != http.StatusOK {
		t.Fatalf("expected balance 200, got %d: %s", response.Code, response.Body.String())
	}
	metrics := perform(router, http.MethodGet, "/metrics", "", "")
	metricsBody := metrics.Body.String()

	if !strings.Contains(metricsBody, `route="/v1/accounts/:account_id/balance"`) {
		t.Fatalf("expected metrics to use route pattern, got %s", metricsBody)
	}
	if strings.Contains(metricsBody, `route="/v1/accounts/acct_sandbox_001/balance"`) {
		t.Fatalf("expected metrics to avoid account-id cardinality, got %s", metricsBody)
	}
}

func TestIdempotentFinancialWriteReplaysCachedResponse(t *testing.T) {
	router, _, cfg := newTestRouter(t, 120)
	body := `{
		"source_account_id": "acct_sandbox_001",
		"amount_cents": 1000,
		"currency": "BRL",
		"pix_key": "merchant@example.com",
		"description": "coffee settlement"
	}`

	first := performWithHeaders(router, http.MethodPost, "/v1/pix/transfers", cfg.FullAccessAPIKey, body, map[string]string{"Idempotency-Key": "idem-test-1"})
	second := performWithHeaders(router, http.MethodPost, "/v1/pix/transfers", cfg.FullAccessAPIKey, body, map[string]string{"Idempotency-Key": "idem-test-1"})

	if first.Code != http.StatusAccepted {
		t.Fatalf("expected first request 202, got %d: %s", first.Code, first.Body.String())
	}
	if second.Code != http.StatusAccepted {
		t.Fatalf("expected replay 202, got %d: %s", second.Code, second.Body.String())
	}
	if second.Header().Get("Idempotency-Replayed") != "true" {
		t.Fatalf("expected replay header, got %q", second.Header().Get("Idempotency-Replayed"))
	}
	if first.Body.String() != second.Body.String() {
		t.Fatalf("expected cached response body\nfirst: %s\nsecond: %s", first.Body.String(), second.Body.String())
	}
}

func TestIdempotencyConflict(t *testing.T) {
	router, _, cfg := newTestRouter(t, 120)

	first := performWithHeaders(router, http.MethodPost, "/v1/pix/transfers", cfg.FullAccessAPIKey, `{
		"source_account_id": "acct_sandbox_001",
		"amount_cents": 1000,
		"currency": "BRL",
		"pix_key": "merchant@example.com"
	}`, map[string]string{"Idempotency-Key": "idem-conflict"})
	second := performWithHeaders(router, http.MethodPost, "/v1/pix/transfers", cfg.FullAccessAPIKey, `{
		"source_account_id": "acct_sandbox_001",
		"amount_cents": 2000,
		"currency": "BRL",
		"pix_key": "merchant@example.com"
	}`, map[string]string{"Idempotency-Key": "idem-conflict"})

	if first.Code != http.StatusAccepted {
		t.Fatalf("expected first request 202, got %d: %s", first.Code, first.Body.String())
	}
	if second.Code != http.StatusConflict {
		t.Fatalf("expected conflict 409, got %d: %s", second.Code, second.Body.String())
	}
	if !strings.Contains(second.Body.String(), "idempotency_conflict") {
		t.Fatalf("expected idempotency conflict error, got %s", second.Body.String())
	}
}

func TestRateLimitExceeded(t *testing.T) {
	router, _, cfg := newTestRouter(t, 1)

	first := perform(router, http.MethodGet, "/v1/accounts/acct_sandbox_001/balance", cfg.FullAccessAPIKey, "")
	second := perform(router, http.MethodGet, "/v1/accounts/acct_sandbox_001/balance", cfg.FullAccessAPIKey, "")

	if first.Code != http.StatusOK {
		t.Fatalf("expected first request 200, got %d: %s", first.Code, first.Body.String())
	}
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request 429, got %d: %s", second.Code, second.Body.String())
	}
	if second.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
}

func TestWebhookRegistrationAndDeliveryQueue(t *testing.T) {
	router, repo, cfg := newTestRouter(t, 120)

	created := perform(router, http.MethodPost, "/v1/webhooks/endpoints", cfg.FullAccessAPIKey, `{
		"url": "https://partner.example.com/webhooks",
		"event_types": ["pix.transfer.created.v1"],
		"description": "Primary partner receiver"
	}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("expected webhook endpoint created, got %d: %s", created.Code, created.Body.String())
	}

	transfer := performWithHeaders(router, http.MethodPost, "/v1/pix/transfers", cfg.FullAccessAPIKey, `{
		"source_account_id": "acct_sandbox_001",
		"amount_cents": 1000,
		"currency": "BRL",
		"pix_key": "merchant@example.com"
	}`, map[string]string{"Idempotency-Key": "idem-webhook"})
	if transfer.Code != http.StatusAccepted {
		t.Fatalf("expected pix transfer accepted, got %d: %s", transfer.Code, transfer.Body.String())
	}

	deliveries := repo.WebhookDeliveries()
	if len(deliveries) != 1 {
		t.Fatalf("expected one webhook delivery, got %d", len(deliveries))
	}
	if !strings.Contains(deliveries[0].Signature, "v1=") {
		t.Fatalf("expected HMAC signature, got %q", deliveries[0].Signature)
	}
}

func TestPayoutRefundStatementsAuditAndSandboxEndpoints(t *testing.T) {
	router, _, cfg := newTestRouter(t, 120)

	pix := performWithHeaders(router, http.MethodPost, "/v1/pix/transfers", cfg.FullAccessAPIKey, `{
		"source_account_id": "acct_sandbox_001",
		"amount_cents": 2000,
		"currency": "BRL",
		"pix_key": "merchant@example.com"
	}`, map[string]string{"Idempotency-Key": "idem-refund-source"})
	if pix.Code != http.StatusAccepted {
		t.Fatalf("expected pix transfer accepted, got %d: %s", pix.Code, pix.Body.String())
	}
	var pixPayload struct {
		Data struct {
			Transfer struct {
				ID string `json:"id"`
			} `json:"transfer"`
		} `json:"data"`
	}
	if err := json.Unmarshal(pix.Body.Bytes(), &pixPayload); err != nil {
		t.Fatalf("parse pix response: %v", err)
	}
	if pixPayload.Data.Transfer.ID == "" {
		t.Fatal("expected transfer id in pix response")
	}

	payout := performWithHeaders(router, http.MethodPost, "/v1/payouts", cfg.FullAccessAPIKey, `{
		"account_id": "acct_sandbox_001",
		"amount_cents": 500,
		"currency": "BRL",
		"bank_code": "001",
		"branch": "0001",
		"account_number": "12345-6",
		"document": "12345678909",
		"description": "settlement payout"
	}`, map[string]string{"Idempotency-Key": "idem-payout"})
	if payout.Code != http.StatusAccepted {
		t.Fatalf("expected payout accepted, got %d: %s", payout.Code, payout.Body.String())
	}

	refund := performWithHeaders(router, http.MethodPost, "/v1/refunds", cfg.FullAccessAPIKey, `{
		"original_transaction_id": "`+pixPayload.Data.Transfer.ID+`",
		"account_id": "acct_sandbox_001",
		"amount_cents": 500,
		"currency": "BRL",
		"reason": "partner requested reversal"
	}`, map[string]string{"Idempotency-Key": "idem-refund"})
	if refund.Code != http.StatusAccepted {
		t.Fatalf("expected refund accepted, got %d: %s", refund.Code, refund.Body.String())
	}

	statements := perform(router, http.MethodGet, "/v1/accounts/acct_sandbox_001/statements", cfg.FullAccessAPIKey, "")
	if statements.Code != http.StatusOK || !strings.Contains(statements.Body.String(), "Pix transfer") {
		t.Fatalf("expected statements with command evidence, got %d: %s", statements.Code, statements.Body.String())
	}

	audit := perform(router, http.MethodGet, "/v1/audit-logs", cfg.FullAccessAPIKey, "")
	if audit.Code != http.StatusOK || !strings.Contains(audit.Body.String(), "payout.create") || !strings.Contains(audit.Body.String(), "refund.create") {
		t.Fatalf("expected audit log entries, got %d: %s", audit.Code, audit.Body.String())
	}

	sandbox := perform(router, http.MethodGet, "/v1/sandbox/scenarios", cfg.FullAccessAPIKey, "")
	if sandbox.Code != http.StatusOK || !strings.Contains(sandbox.Body.String(), "scenario_pix_success") {
		t.Fatalf("expected sandbox scenarios, got %d: %s", sandbox.Code, sandbox.Body.String())
	}
}

func BenchmarkGetBalanceRequest(b *testing.B) {
	router, _, cfg := newBenchmarkRouter(b)
	body := ""
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		response := perform(router, http.MethodGet, "/v1/accounts/acct_sandbox_001/balance", cfg.FullAccessAPIKey, body)
		if response.Code != http.StatusOK {
			b.Fatalf("expected 200, got %d", response.Code)
		}
	}
}

func newTestRouter(t *testing.T, rateLimit int) (*gin.Engine, *store.Repository, config.Config) {
	t.Helper()
	cfg := testConfig(rateLimit)
	repo := store.NewSeededRepository(cfg)
	router := NewRouter(Dependencies{
		Config:     cfg,
		Logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
		Repository: repo,
		Metrics:    observability.NewMetrics(cfg.ServiceName),
	})
	return router, repo, cfg
}

func newBenchmarkRouter(b *testing.B) (*gin.Engine, *store.Repository, config.Config) {
	b.Helper()
	cfg := testConfig(1_000_000_000)
	repo := store.NewSeededRepository(cfg)
	router := NewRouter(Dependencies{
		Config:     cfg,
		Logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
		Repository: repo,
		Metrics:    observability.NewMetrics(cfg.ServiceName),
	})
	return router, repo, cfg
}

func testConfig(rateLimit int) config.Config {
	cfg := config.Load()
	cfg.ServiceName = "bankport_partner_api_test"
	cfg.RequestTimeout = 2 * time.Second
	cfg.DefaultRateLimitRPM = rateLimit
	cfg.WebhookSigningKey = "test-webhook-secret"
	return cfg
}

func perform(router *gin.Engine, method, path, apiKey, body string) *httptest.ResponseRecorder {
	return performWithHeaders(router, method, path, apiKey, body, nil)
}

func performWithHeaders(router *gin.Engine, method, path, apiKey, body string, headers map[string]string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	requestBody := bytes.NewBufferString(body)
	request := httptest.NewRequest(method, path, requestBody)
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+apiKey)
	}
	request.Header.Set("X-Correlation-ID", "corr_test")
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	router.ServeHTTP(recorder, request)
	return recorder
}
