package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/config"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/domain"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/httpapi/middleware"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/observability"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/store"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/webhook"
)

type Dependencies struct {
	Config     config.Config
	Logger     *slog.Logger
	Repository *store.Repository
	Metrics    *observability.Metrics
}

type Server struct {
	cfg              config.Config
	logger           *slog.Logger
	repository       *store.Repository
	metrics          *observability.Metrics
	signer           webhook.Signer
	idempotencyStore *middleware.IdempotencyStore
	rateLimiter      *middleware.RateLimiter
}

func NewRouter(deps Dependencies) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	logger := deps.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	server := &Server{
		cfg:              deps.Config,
		logger:           logger,
		repository:       deps.Repository,
		metrics:          deps.Metrics,
		signer:           webhook.NewSigner(deps.Config.WebhookSigningKey),
		idempotencyStore: middleware.NewIdempotencyStoreWithTTL(deps.Config.IdempotencyTTL),
		rateLimiter:      middleware.NewRateLimiter(),
	}

	router := gin.New()
	router.Use(
		gin.Recovery(),
		middleware.RequestIdentity(),
		middleware.Timeout(deps.Config.RequestTimeout),
		middleware.Tracing(deps.Config.ServiceName),
		middleware.StructuredLogger(logger),
		middleware.Metrics(deps.Metrics),
	)

	router.GET("/health/live", server.live)
	router.GET("/health/ready", server.ready)
	router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(deps.Metrics.Registry, promhttp.HandlerOpts{})))

	v1 := router.Group("/v1")
	v1.Use(middleware.Authenticate(deps.Repository, deps.Metrics), middleware.RateLimit(server.rateLimiter, deps.Metrics))
	v1.GET("/accounts/:account_id/balance", middleware.RequireScopes(domain.ScopeAccountsRead), server.getBalance)
	v1.GET("/accounts/:account_id/statements", middleware.RequireScopes(domain.ScopeAccountsRead), server.listStatements)
	v1.POST("/pix/transfers", middleware.RequireScopes(domain.ScopePixWrite), middleware.Idempotency(server.idempotencyStore, deps.Metrics), server.createPixTransfer)
	v1.POST("/payouts", middleware.RequireScopes(domain.ScopePayoutsWrite), middleware.Idempotency(server.idempotencyStore, deps.Metrics), server.createPayout)
	v1.POST("/refunds", middleware.RequireScopes(domain.ScopeRefundsWrite), middleware.Idempotency(server.idempotencyStore, deps.Metrics), server.createRefund)
	v1.POST("/webhooks/endpoints", middleware.RequireScopes(domain.ScopeWebhooksWrite), server.registerWebhookEndpoint)
	v1.GET("/audit-logs", middleware.RequireScopes(domain.ScopeAuditRead), server.listAuditLogs)
	v1.GET("/sandbox/scenarios", middleware.RequireScopes(domain.ScopeSandboxRead), server.listSandboxScenarios)

	return router
}

func (s *Server) live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":         "ok",
		"service":        s.cfg.ServiceName,
		"request_id":     middleware.RequestID(c),
		"correlation_id": middleware.CorrelationID(c),
	})
}

func (s *Server) ready(c *gin.Context) {
	if !s.cfg.ReadinessEnabled {
		middleware.Abort(c, http.StatusServiceUnavailable, "not_ready", "Readiness is disabled by configuration.", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"checks": gin.H{
			"repository":  "ready",
			"rate_limit":  "ready",
			"idempotency": "ready",
		},
		"request_id":     middleware.RequestID(c),
		"correlation_id": middleware.CorrelationID(c),
	})
}

func (s *Server) getBalance(c *gin.Context) {
	partner, _ := middleware.Partner(c)
	account, err := s.repository.GetAccount(c.Request.Context(), partner.ID, c.Param("account_id"))
	if err != nil {
		s.respondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, success(account, c))
}

func (s *Server) listStatements(c *gin.Context) {
	partner, _ := middleware.Partner(c)
	entries, err := s.repository.ListStatements(c.Request.Context(), partner.ID, c.Param("account_id"))
	if err != nil {
		s.respondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, success(entries, c))
}

func (s *Server) createPixTransfer(c *gin.Context) {
	partner, _ := middleware.Partner(c)
	var request domain.PixTransferRequest
	if !bindJSON(c, &request) {
		return
	}

	transfer, queuedDeliveries, err := s.repository.CreatePixTransfer(c.Request.Context(), partner, request, middleware.CorrelationID(c), s.signer.SignEvent)
	if err != nil {
		s.metrics.FinancialCommands.WithLabelValues("pix_transfer", "rejected").Inc()
		s.audit(c, partner, "pix.transfer.create", request.SourceAccountID, "rejected", store.ErrorCode(err))
		s.respondDomainError(c, err)
		return
	}
	s.metrics.FinancialCommands.WithLabelValues("pix_transfer", "accepted").Inc()
	if queuedDeliveries > 0 {
		s.metrics.WebhookDeliveries.WithLabelValues("pix.transfer.created.v1", "queued").Add(float64(queuedDeliveries))
	}
	s.audit(c, partner, "pix.transfer.create", transfer.ID, "accepted", "")
	c.JSON(http.StatusAccepted, success(gin.H{
		"transfer":                  transfer,
		"queued_webhook_deliveries": queuedDeliveries,
	}, c))
}

func (s *Server) createPayout(c *gin.Context) {
	partner, _ := middleware.Partner(c)
	var request domain.PayoutRequest
	if !bindJSON(c, &request) {
		return
	}

	payout, queuedDeliveries, err := s.repository.CreatePayout(c.Request.Context(), partner, request, middleware.CorrelationID(c), s.signer.SignEvent)
	if err != nil {
		s.metrics.FinancialCommands.WithLabelValues("payout", "rejected").Inc()
		s.audit(c, partner, "payout.create", request.AccountID, "rejected", store.ErrorCode(err))
		s.respondDomainError(c, err)
		return
	}
	s.metrics.FinancialCommands.WithLabelValues("payout", "accepted").Inc()
	if queuedDeliveries > 0 {
		s.metrics.WebhookDeliveries.WithLabelValues("payout.created.v1", "queued").Add(float64(queuedDeliveries))
	}
	s.audit(c, partner, "payout.create", payout.ID, "accepted", "")
	c.JSON(http.StatusAccepted, success(gin.H{
		"payout":                    payout,
		"queued_webhook_deliveries": queuedDeliveries,
	}, c))
}

func (s *Server) createRefund(c *gin.Context) {
	partner, _ := middleware.Partner(c)
	var request domain.RefundRequest
	if !bindJSON(c, &request) {
		return
	}

	refund, queuedDeliveries, err := s.repository.CreateRefund(c.Request.Context(), partner, request, middleware.CorrelationID(c), s.signer.SignEvent)
	if err != nil {
		s.metrics.FinancialCommands.WithLabelValues("refund", "rejected").Inc()
		s.audit(c, partner, "refund.create", request.AccountID, "rejected", store.ErrorCode(err))
		s.respondDomainError(c, err)
		return
	}
	s.metrics.FinancialCommands.WithLabelValues("refund", "accepted").Inc()
	if queuedDeliveries > 0 {
		s.metrics.WebhookDeliveries.WithLabelValues("refund.created.v1", "queued").Add(float64(queuedDeliveries))
	}
	s.audit(c, partner, "refund.create", refund.ID, "accepted", "")
	c.JSON(http.StatusAccepted, success(gin.H{
		"refund":                    refund,
		"queued_webhook_deliveries": queuedDeliveries,
	}, c))
}

func (s *Server) registerWebhookEndpoint(c *gin.Context) {
	partner, _ := middleware.Partner(c)
	var request domain.WebhookEndpointRequest
	if !bindJSON(c, &request) {
		return
	}

	endpoint, err := s.repository.RegisterWebhookEndpoint(c.Request.Context(), partner, request)
	if err != nil {
		s.audit(c, partner, "webhook.endpoint.create", "", "rejected", store.ErrorCode(err))
		s.respondDomainError(c, err)
		return
	}
	s.audit(c, partner, "webhook.endpoint.create", endpoint.ID, "accepted", "")
	c.JSON(http.StatusCreated, success(endpoint, c))
}

func (s *Server) listAuditLogs(c *gin.Context) {
	partner, _ := middleware.Partner(c)
	c.JSON(http.StatusOK, success(s.repository.ListAuditEntries(c.Request.Context(), partner.ID), c))
}

func (s *Server) listSandboxScenarios(c *gin.Context) {
	c.JSON(http.StatusOK, success(s.repository.SandboxScenarios(), c))
}

func (s *Server) respondDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		middleware.Abort(c, http.StatusBadRequest, "validation_failed", err.Error(), nil)
	case errors.Is(err, domain.ErrAccountNotFound):
		middleware.Abort(c, http.StatusNotFound, "account_not_found", "The account was not found for this partner.", nil)
	case errors.Is(err, domain.ErrInsufficientFunds):
		middleware.Abort(c, http.StatusUnprocessableEntity, "insufficient_funds", "The account does not have enough available balance.", nil)
	case errors.Is(err, domain.ErrOriginalTxnNotFound):
		middleware.Abort(c, http.StatusNotFound, "original_transaction_not_found", "The original transaction was not found for this partner and account.", nil)
	case errors.Is(err, domain.ErrRefundExceedsOriginal):
		middleware.Abort(c, http.StatusUnprocessableEntity, "refund_exceeds_original", "The refund amount exceeds the original transaction amount.", nil)
	default:
		s.logger.Error("domain_error_unmapped", "error", err, "request_id", middleware.RequestID(c))
		middleware.Abort(c, http.StatusInternalServerError, "internal_error", "An internal error occurred.", nil)
	}
}

func (s *Server) audit(c *gin.Context, partner domain.Partner, action, resourceID, status, reason string) {
	s.repository.AddAuditEntry(domain.AuditEntry{
		PartnerID:     partner.ID,
		RequestID:     middleware.RequestID(c),
		CorrelationID: middleware.CorrelationID(c),
		Action:        action,
		ResourceID:    resourceID,
		Status:        status,
		Reason:        reason,
	})
}

func bindJSON(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		middleware.Abort(c, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON matching the endpoint schema.", map[string]any{
			"error": err.Error(),
		})
		return false
	}
	return true
}

func success(data any, c *gin.Context) gin.H {
	return gin.H{
		"data":           data,
		"request_id":     middleware.RequestID(c),
		"correlation_id": middleware.CorrelationID(c),
	}
}
